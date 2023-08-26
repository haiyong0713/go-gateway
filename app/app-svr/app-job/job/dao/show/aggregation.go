package show

import (
	"context"
	xsql "database/sql"
	"fmt"
	"strings"

	"go-common/library/cache/memcache"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-job/job/model/show"

	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

// nolint:gosec
const (
	_AIInsertHotNameSQL = "INSERT INTO hotword_aggregation(plat,hot_title,title,subtitle,image) VALUES('bili_popular',?,?,?,?)"
	_AIInsertTagIDSQL   = "INSERT INTO hotword_aggregation_tag(hotword_id,tag_id) VALUES (?,?)"
	_AIAlertSQL         = "SELECT COUNT(1) FROM hotword_aggregation_tag AS hat,hotword_aggregation AS ha WHERE " +
		"hat.hotword_id=ha.id AND ha.plat='bili_popular' AND hat.tag_id=? AND hat.state=0"
	_hotNameSQL         = "SELECT hot_title,state,id FROM hotword_aggregation WHERE hot_title=? AND state!=4 AND plat='bili_popular'"
	_aggregationPassSQL = "SELECT id,plat,hot_title,title,subtitle,image,state,active_state FROM hotword_aggregation WHERE state=1"
	_aggMaterial        = "SELECT oid FROM hotword_aggregation_video WHERE hotword_id=?"
	_offlineAggregation = "UPDATE hotword_aggregation_video SET active_state=2 WHERE id IN (%s)"
)

const (
	_subtitle       = "当前热门视频"
	_title          = "更多关于【%s】的热门视频"
	_key            = "%d_hot_word"
	_aggregation    = "aggregation_%s"
	_aggregationAI  = "aggregationAI_%d"
	_aggregationArc = "aggregationArc_%d"

	_aggURL = "/data/rank/hotword/list-%d.json"
)

func _fmtMcKey(hotID int64) string {
	return fmt.Sprintf(_key, hotID)
}

func _fmtTitle(title string) string {
	return fmt.Sprintf(_title, title)
}

func _fmtAggregation(spiderType string) string {
	return fmt.Sprintf(_aggregation, spiderType)
}

func formAggAI(hotID int64) string {
	return fmt.Sprintf(_aggregationAI, hotID)
}

func formAggArc(hotID int64) string {
	return fmt.Sprintf(_aggregationArc, hotID)
}

// TagIDCount .
func (d *Dao) TagIDCount(ctx context.Context, tagID int64) (count int, err error) {
	err = d.db.QueryRow(ctx, _AIAlertSQL, tagID).Scan(&count)
	return
}

// HotWordName .
func (d *Dao) HotWordName(ctx context.Context, tagName string) (res []*show.Aggregation, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _hotNameSQL, tagName); err != nil {
		log.Error("[HotWordName] d.db.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &show.Aggregation{}
		if err = rows.Scan(&a.HotTitle, &a.State, &a.ID); err != nil {
			log.Error("[HotWordName] rows.Scan() error(%v)", err)
			return
		}
		res = append(res, a)
	}
	if err = rows.Err(); err != nil {
		log.Error("[HotWordName] rows.Err() error(%v)", err)
	}
	return
}

// AddAggregation .
func (d *Dao) AddAggregation(ctx context.Context, hotTitle string, tagID int64) (err error) {
	var (
		sqlRes xsql.Result
		tx, _  = d.db.Begin(ctx)
		id     int64
	)
	if sqlRes, err = tx.Exec(_AIInsertHotNameSQL, hotTitle, _fmtTitle(hotTitle), _subtitle, d.conf.Aggregation.Image); err != nil {
		_ = tx.Rollback()
		log.Error("[AddAggregation] tx.Exec(%s) error(%v)", _AIInsertHotNameSQL, err)
		return
	}
	if id, err = sqlRes.LastInsertId(); err != nil {
		_ = tx.Rollback()
		log.Error("[AddAggregation] sqlRes.LastInsertId() error(%v)", err)
		return
	}
	if _, err = tx.Exec(_AIInsertTagIDSQL, id, tagID); err != nil {
		_ = tx.Rollback()
		log.Error("[AddAggregation] tx.Exec(%s) error(%v)", _AIInsertTagIDSQL, err)
		return
	}
	return tx.Commit()
}

// AddTagID .
func (d *Dao) AddTagID(ctx context.Context, hotID, tagID int64) (err error) {
	if _, err = d.db.Exec(ctx, _AIInsertTagIDSQL, hotID, tagID); err != nil {
		log.Error("[AddTagID] tx.Exec(%s) error(%v)", _AIInsertTagIDSQL, err)
	}
	return
}

// Tag .
func (d *Dao) Tag(ctx context.Context, tagID int64) (res *taggrpc.TagReply, err error) {
	res, err = d.tagGRPC.Tag(ctx, &taggrpc.TagReq{Tid: tagID})
	return
}

// AddHotMC .
func (d *Dao) AddHotMC(ctx context.Context, aggregation *show.Aggregation) (err error) {
	var (
		key  = _fmtMcKey(aggregation.ID)
		conn = d.mc.Get(ctx)
	)
	defer conn.Close()
	item := &memcache.Item{Key: key, Object: aggregation, Flags: memcache.FlagJSON, Expiration: d.expireMC}
	if err = conn.Set(item); err != nil {
		log.Error("[AddHotMC] conn.Set() error()")
	}
	return
}

func (d *Dao) SetAggregations(c context.Context, spiderType string, aggregations map[string][]*show.AggregationItem) (err error) {
	var (
		key  = _fmtAggregation(spiderType)
		conn = d.aggregationmc.Get(c)
	)
	defer conn.Close()
	item := &memcache.Item{Key: key, Object: aggregations, Flags: memcache.FlagJSON, Expiration: d.expireMC}
	if err = conn.Set(item); err != nil {
		log.Error("[SetAggregations] conn.Set() error()")
	}
	return
}

func (d *Dao) Aggregations(c context.Context, spiderType string) (res map[string][]*show.AggregationItem, err error) {
	var (
		key  = _fmtAggregation(spiderType)
		as   *memcache.Item
		conn = d.aggregationmc.Get(c)
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
	res = make(map[string][]*show.AggregationItem)
	if err = conn.Scan(as, &res); err != nil {
		log.Error("conn.Scan(%s) error(%v)", as.Value, err)
	}
	return
}

// AggFromAI list .
func (d *Dao) AggFromAI(ctx context.Context, hotID int64) (views []*show.CardList, err error) {
	var (
		res struct {
			Code int              `json:"code"`
			List []*show.CardList `json:"list"`
		}
	)
	if err = d.client.Get(ctx, fmt.Sprintf(d.aggURL, hotID), "", nil, &res); err != nil {
		log.Error("[AggView] d.client.Get() url(%s) hotID(%d) error(%v)", d.aggURL, hotID, err)
		err = ecode.NothingFound
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("[AggView] d.client.Get() url(%s) hotID(%d) error(%v)", d.aggURL, hotID, err)
		err = ecode.Int(res.Code)
		return
	}
	views = res.List
	return
}

func (d *Dao) Materials(c context.Context, hotID int64) (ids []int64, err error) {
	rows, err := d.db.Query(c, _aggMaterial, hotID)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			log.Error("%v", err)
			return
		}
		if id != 0 {
			ids = append(ids, id)
		}
	}
	err = rows.Err()
	return
}

func (d *Dao) AggAI(c context.Context, hotID int64) (res *show.AggAI, err error) {
	var (
		key  = formAggAI(hotID)
		as   *memcache.Item
		conn = d.aggregationmc.Get(c)
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
	res = &show.AggAI{}
	if err = conn.Scan(as, &res); err != nil {
		log.Error("conn.Scan(%s) error(%v)", as.Value, err)
	}
	return
}

func (d *Dao) SetAggAI(c context.Context, hotID int64, ai *show.AggAI) (err error) {
	var (
		key  = formAggAI(hotID)
		conn = d.aggregationmc.Get(c)
	)
	defer conn.Close()
	item := &memcache.Item{Key: key, Object: ai, Flags: memcache.FlagJSON, Expiration: d.expireMC}
	if err = conn.Set(item); err != nil {
		log.Error("[SetAggAI] conn.Set() error()")
	}
	return
}

func (d *Dao) AggArc(c context.Context, hotID int64) (res map[int64]*show.ArcInfo, err error) {
	var (
		key  = formAggArc(hotID)
		as   *memcache.Item
		conn = d.aggregationmc.Get(c)
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
	res = make(map[int64]*show.ArcInfo)
	if err = conn.Scan(as, &res); err != nil {
		log.Error("conn.Scan(%s) error(%v)", as.Value, err)
	}
	return
}

func (d *Dao) SetAggArc(c context.Context, hotID int64, arcm map[int64]*show.ArcInfo) (err error) {
	var (
		key  = formAggArc(hotID)
		conn = d.aggregationmc.Get(c)
	)
	defer conn.Close()
	item := &memcache.Item{Key: key, Object: arcm, Flags: memcache.FlagJSON, Expiration: d.expireMC}
	if err = conn.Set(item); err != nil {
		log.Error("[SetAggArc] conn.Set() error()")
	}
	return
}

func (d *Dao) Tags(c context.Context, tagIDs []int64) (res map[int64]*taggrpc.Tag, err error) {
	var (
		args   = &taggrpc.TagsReq{Tids: tagIDs}
		resTmp *taggrpc.TagsReply
	)
	if resTmp, err = d.tagGRPC.Tags(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetTags()
	return
}

func (d *Dao) AggregationPass(c context.Context) (res []*show.Aggregation, err error) {
	rows, err := d.db.Query(c, _aggregationPassSQL)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &show.Aggregation{}
		if err = rows.Scan(&re.ID, &re.Plat, &re.HotTitle, &re.Title, &re.Subtitle, &re.Image, &re.State, &re.ActiveState); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) OffLine(c context.Context, ids []int64) (raw int64, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range ids {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	res, err := d.db.Exec(c, fmt.Sprintf(_offlineAggregation, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	return res.RowsAffected()
}
