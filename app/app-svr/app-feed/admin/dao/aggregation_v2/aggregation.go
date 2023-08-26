package aggregation_v2

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/cache/memcache"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"

	aggmdl "go-gateway/app/app-svr/app-feed/admin/model/aggregation_v2"
	showmdl "go-gateway/app/app-svr/app-feed/admin/model/show"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

const (
	// mc key
	_aggregation    = "aggregation_rank_data"
	_aggregationAI  = "aggregationAI_%d"
	_aggregationArc = "aggregationArc_%d"
	// ai url
	_aiAggregationURL = "/data/rank/hotword/list-%d.json"
	// sql
	_aggregationSQL              = "SELECT id,plat,hot_title,title,subtitle,image,state,active_state FROM hotword_aggregation WHERE state!=4 %s ORDER BY mtime DESC"
	_aggregationByIDSQL          = "SELECT id,plat,hot_title,title,subtitle,image,state,active_state FROM hotword_aggregation WHERE id=? ORDER BY mtime DESC"
	_aggregationByHotWordState   = "SELECT id,plat,hot_title,title,subtitle,image,state,active_state FROM hotword_aggregation WHERE plat=? AND hot_title=?"
	_addAggregationSQL           = "INSERT INTO hotword_aggregation(plat,hot_title,title,subtitle,image,state,active_state) VALUES (?,?,?,?,?,?,?)"
	_upAggregationStateSQL       = "UPDATE hotword_aggregation SET state=? WHERE plat=? AND hot_title=?"
	_upAggregationActiveStateSQL = "UPDATE hotword_aggregation SET active_state=? WHERE plat=? AND hot_title=?"
	_upAggregationOlineConfig    = "UPDATE hotword_aggregation SET title=?,subtitle=?,image=? WHERE id=?"

	_aggregationTags = "SELECT id,hotword_id,tag_id,state FROM hotword_aggregation_tag WHERE hotword_id=? AND state=0"

	_viewsSQL    = "SELECT id,source,hotword_id,oid,state FROM hotword_aggregation_video WHERE hotword_id=?"
	_viewSQL2    = "SELECT id FROM hotword_aggregation_video WHERE hotword_id=? AND oid=? AND source=?"
	_addViewSQL  = "INSERT INTO hotword_aggregation_video(source,hotword_id,oid,state) VALUES (?,?,?,?)"
	_addViewsSQL = "INSERT INTO hotword_aggregation_video(source,hotword_id,oid,state) VALUES %s"
	_upViewSQL   = "UPDATE hotword_aggregation_video SET state=? WHERE id=?"

	_tagSQL    = "SELECT id FROM hotword_aggregation_tag WHERE hotword_id=? AND tag_id=?"
	_addTagSQL = "INSERT INTO hotword_aggregation_tag (hotword_id,tag_id) VALUES (?,?)"
	_upTagSQL  = "UPDATE hotword_aggregation_tag SET state=? WHERE id=?"

	_delHotWord      = "DELETE FROM hotword_aggregation WHERE id=?"
	_delHotWordTag   = "DELETE FROM hotword_aggregation_tag WHERE hotword_id=?"
	_delHotWordVideo = "DELETE FROM hotword_aggregation_video WHERE hotword_id=?"
)

func formAggAI(hotID int64) string {
	return fmt.Sprintf(_aggregationAI, hotID)
}

func formAggArc(hotID int64) string {
	return fmt.Sprintf(_aggregationArc, hotID)
}

func (d *Dao) ListCache(c context.Context, filters []string) (res map[string][]*showmdl.AggregationItem, err error) {
	var (
		key  = _aggregation
		as   *memcache.Item
		conn = d.mc.Get(c)
	)
	defer conn.Close()
	if as, err = conn.Get(key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
		} else {
			log.Error("memcache.Get(%s) error(%v)", key, err)
		}
		return
	}
	var resTmp = make(map[string][]*showmdl.AggregationItem)
	if err = conn.Scan(as, &resTmp); err != nil {
		log.Error("conn.Scan(%s) error(%v)", as.Value, err)
		return
	}
	res = make(map[string][]*showmdl.AggregationItem)
	if len(filters) > 0 {
		for _, plat := range filters {
			if aggCache, ok := resTmp[plat]; ok {
				res[plat] = aggCache
			}
		}
	} else {
		res = resTmp
	}
	return
}

func (d *Dao) Aggregations(c context.Context, hotWord, filter string) (res []*aggmdl.Aggregation, err error) {
	var (
		sqls []string
		args []interface{}
	)
	if hotWord != "" {
		sqls = append(sqls, "AND hot_title LIKE ?")
		args = append(args, "%"+hotWord+"%")
	}
	if fs := strings.Split(filter, ","); len(fs) > 0 {
		var (
			sqlsTmp []string
			argsTmp []interface{}
		)
		for _, f := range fs {
			if f == "" {
				continue
			}
			sqlsTmp = append(sqlsTmp, "?")
			argsTmp = append(argsTmp, f)
		}
		if len(sqlsTmp) != 0 && len(argsTmp) != 0 && (len(sqlsTmp) == len(argsTmp)) {
			sqls = append(sqls, fmt.Sprintf("AND plat IN (%v)", strings.Join(sqlsTmp, ",")))
			args = append(args, argsTmp...)
		}
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_aggregationSQL, strings.Join(sqls, " ")), args...)
	//nolint:staticcheck
	defer rows.Close()
	if err != nil {
		log.Error("%v", err)
		return
	}
	for rows.Next() {
		re := &aggmdl.Hotword{}
		if err = rows.Scan(&re.ID, &re.Plat, &re.HotTitle, &re.Title, &re.Subtitle, &re.Image, &re.State, &re.ActiveState); err != nil {
			log.Error("%v", err)
			return
		}
		agg := &aggmdl.Aggregation{}
		agg.FormHot(re)
		res = append(res, agg)
	}
	err = rows.Err()
	return
}

func (d *Dao) AggregationByID(c context.Context, hotID int64) (re *aggmdl.Hotword, err error) {
	row := d.db.QueryRow(c, _aggregationByIDSQL, hotID)
	re = &aggmdl.Hotword{}
	if err = row.Scan(&re.ID, &re.Plat, &re.HotTitle, &re.Title, &re.Subtitle, &re.Image, &re.State, &re.ActiveState); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("AggregationByID %v", err)
		}
	}
	return
}

func (d *Dao) AggregationByPlatHotWord(c context.Context, plat, hotword string) (re *aggmdl.Hotword, err error) {
	row := d.db.QueryRow(c, _aggregationByHotWordState, plat, hotword)
	re = &aggmdl.Hotword{}
	if err = row.Scan(&re.ID, &re.Plat, &re.HotTitle, &re.Title, &re.Subtitle, &re.Image, &re.State, &re.ActiveState); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("AggregationByHotWordState %v", err)
		}
	}
	return
}

func (d *Dao) AddAggregation(tx *sql.Tx, plat, hotTitle, title, subtitle, image string, state, activeState int) (raw int64, err error) {
	res, err := tx.Exec(_addAggregationSQL, plat, hotTitle, title, subtitle, image, state, activeState)
	if err != nil {
		log.Error("AddAggregation %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpAggregationState(tx *sql.Tx, plat, hotTitle string, state int) (raw int64, err error) {
	res, err := tx.Exec(_upAggregationStateSQL, state, plat, hotTitle)
	if err != nil {
		log.Error("UpAggregationState %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpAggregationActiveState(tx *sql.Tx, plat, hotTitle string, activeState int) (raw int64, err error) {
	res, err := tx.Exec(_upAggregationActiveStateSQL, activeState, plat, hotTitle)
	if err != nil {
		log.Error("UpAggregationActiveState %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpAggregationOlineConfig(tx *sql.Tx, id int64, title, subtitle, cover string) (raw int64, err error) {
	res, err := tx.Exec(_upAggregationOlineConfig, title, subtitle, cover, id)
	if err != nil {
		log.Error("UpAggregationOlineConfig %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) Tags(c context.Context, hotID int64) (res []*aggmdl.AggregationTag, err error) {
	rows, err := d.db.Query(c, _aggregationTags, hotID)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &aggmdl.AggregationTag{}
		if err = rows.Scan(&re.ID, &re.HotID, &re.TagID, &re.State); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	if err = rows.Err(); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) TagByTagID(ctx context.Context, tagIDs []int64) (res map[int64]*taggrpc.Tag, err error) {
	var (
		args      = &taggrpc.TagsReq{Tids: tagIDs}
		tagsReply *taggrpc.TagsReply
	)
	if tagsReply, err = d.tagClient.Tags(ctx, args); err != nil {
		log.Error("[NameByTagID] d.tagClient.Tag() tag_id(%s) error(%v)", xstr.JoinInts(tagIDs), err)
		return
	}
	res = tagsReply.GetTags()
	return
}

func (d *Dao) Materiels(c context.Context, hotID int64) (res []*aggmdl.Materiel, err error) {
	rows, err := d.db.Query(c, _viewsSQL, hotID)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &aggmdl.Materiel{}
		if err = rows.Scan(&re.ID, &re.Source, &re.HotID, &re.OID, &re.State); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	if err = rows.Err(); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) AggAI(c context.Context, hotID int64) (res *showmdl.AggAI, err error) {
	var (
		key  = formAggAI(hotID)
		as   *memcache.Item
		conn = d.mc.Get(c)
	)
	defer conn.Close()
	if as, err = conn.Get(key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
		} else {
			log.Error("memcache.Get(%s) error(%v)", key, err)
		}
		return
	}
	res = &showmdl.AggAI{}
	if err = conn.Scan(as, &res); err != nil {
		log.Error("conn.Scan(%s) error(%v)", as.Value, err)
	}
	return
}

func (d *Dao) AggArc(c context.Context, hotID int64) (res map[int64]*showmdl.ArcInfo, err error) {
	var (
		key  = formAggArc(hotID)
		as   *memcache.Item
		conn = d.mc.Get(c)
	)
	defer conn.Close()
	if as, err = conn.Get(key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
		} else {
			log.Error("memcache.Get(%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64]*showmdl.ArcInfo)
	if err = conn.Scan(as, &res); err != nil {
		log.Error("conn.Scan(%s) error(%v)", as.Value, err)
	}
	return
}

func (d *Dao) View(c context.Context, hotID, oid int64, source string) (id int64, err error) {
	row := d.db.QueryRow(c, _viewSQL2, hotID, oid, source)
	if err = row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			id = 0
		} else {
			log.Error("ViewID %v", err)
		}
	}
	return
}

func (d *Dao) AddView(tx *sql.Tx, source string, hotID, oid int64, state int) (raw int64, err error) {
	res, err := tx.Exec(_addViewSQL, source, hotID, oid, state)
	if err != nil {
		log.Error("AddView %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) AddViews(tx *sql.Tx, source string, hotID int64, oids []int64, state int) (raw int64, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, oid := range oids {
		sqls = append(sqls, "(?,?,?,?)")
		args = append(args, source, hotID, oid, state)
	}
	res, err := tx.Exec(fmt.Sprintf(_addViewsSQL, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("AddViews %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpView(tx *sql.Tx, id int64, state int) (raw int64, err error) {
	res, err := tx.Exec(_upViewSQL, state, id)
	if err != nil {
		log.Error("UpView %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TagID(c context.Context, hotID, tagID int64) (id int64, err error) {
	row := d.db.QueryRow(c, _tagSQL, hotID, tagID)
	if err = row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			id = 0
		} else {
			log.Error("TagID %v", err)
		}
	}
	return
}

func (d *Dao) AddTag(tx *sql.Tx, hotID, tagID int64) (raw int64, err error) {
	res, err := tx.Exec(_addTagSQL, hotID, tagID)
	if err != nil {
		log.Error("AddTag %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpTag(tx *sql.Tx, id int64, state int) (raw int64, err error) {
	res, err := tx.Exec(_upTagSQL, state, id)
	if err != nil {
		log.Error("AddTag %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) DelHotWord(tx *sql.Tx, id int64) (raw int64, err error) {
	res, err := tx.Exec(_delHotWord, id)
	if err != nil {
		log.Error("DelHotWord %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) DelHotWordTag(tx *sql.Tx, hotID int64) (raw int64, err error) {
	res, err := tx.Exec(_delHotWordTag, hotID)
	if err != nil {
		log.Error("DelHotWordTag %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) DelHotWordVideo(tx *sql.Tx, hotID int64) (raw int64, err error) {
	res, err := tx.Exec(_delHotWordVideo, hotID)
	if err != nil {
		log.Error("DelHotWordVideo %v", err)
		return
	}
	return res.RowsAffected()
}
