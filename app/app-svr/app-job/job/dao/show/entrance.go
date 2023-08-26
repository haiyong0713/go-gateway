package show

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/model/show"
	show2 "go-gateway/app/app-svr/app-show/interface/model/show"

	"github.com/pkg/errors"
)

const (
	_hotChannelName        = "hot-channel"
	_hotTopicName          = "hot-topic"
	_allentrance           = "SELECT id FROM popular_top_entrance WHERE state=0 AND module_id IN (?,?)"
	_allentrances          = "SELECT id,rank FROM popular_top_entrance ORDER BY rank, mtime ASC"
	_prefixAiChannelResKey = "aicr_%d"
	_entranceURL           = "/data/rank/hotword/channel-%d.json"
	_updateEntranceRank    = "UPDATE popular_top_entrance SET rank=? WHERE id=?"
	_prefixRankKey         = "erank_%d"
	_allEntranceIcon       = `SELECT id,share_icon FROM popular_top_entrance`
	// database cron_job
	_validEntranceSQL = `SELECT id,title,icon,redirect_uri,module_id,grey,white_list,black_list,red_dot,red_dot_text,build_limit,version,update_time,top_photo,share_desc,share_title,share_sub_title,share_icon,white_list_bgroup_business,white_list_bgroup_name 
		FROM popular_top_entrance WHERE state=0 ORDER BY rank, mtime ASC`  // state=0 获取所有启用的入口
	_middleTopPhoto    = "SELECT top_photo FROM popular_top_photo WHERE location_id=? LIMIT 1"
	_defaultLocationId = 1
)

// Entrances .
func (d *Dao) Entrances(ctx context.Context) ([]*show2.EntranceMem, error) {
	rows, err := d.db.Query(ctx, _validEntranceSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*show2.EntranceMem
	for rows.Next() {
		var a = new(show2.EntranceDB)
		if err = rows.Scan(&a.ID, &a.Title, &a.Icon, &a.RedirectURI, &a.ModuleID, &a.Grey, &a.WhiteList, &a.BlackList, &a.RedDot, &a.RedDotText,
			&a.BuildLimit, &a.Version, &a.UpdateTime, &a.TopPhoto, &a.ShareDesc, &a.ShareTitle, &a.ShareSubTitle, &a.ShareIcon, &a.BGroup.Business, &a.BGroup.Name); err != nil {
			return nil, errors.Wrapf(err, "SQL %s", _validEntranceSQL)
		}
		entrance := new(show2.EntranceMem)
		if err := entrance.FromEntranceDB(a); err != nil { // json 脏数据忽略
			continue
		}
		res = append(res, entrance)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "SQL %s", _validEntranceSQL)
	}
	return res, err
}

func (d *Dao) AddCacheEntrances(ctx context.Context, list []*show2.EntranceMem) error {
	if len(list) == 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(list)
	if err != nil {
		return errors.WithStack(err)
	}
	key := showActionKey("loadPopEntrances", "entranceMem")
	if _, err := conn.Do("SETEX", key, _showExpire, bs); err != nil {
		return err
	}
	return nil
}

func (d *Dao) MidTopPhoto(ctx context.Context) (string, error) {
	var res string
	err := d.db.QueryRow(ctx, _middleTopPhoto, _defaultLocationId).Scan(&res)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Error("row.Scan error(%v)", err)
			return "", err
		}
	}
	return res, nil
}

func (d *Dao) AddCacheMidTopPhoto(ctx context.Context, res string) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := showActionKey("loadMiddTopPhoto", "string")
	if _, err := conn.Do("SETEX", key, _showExpire, res); err != nil {
		return err
	}
	return nil
}

func (d *Dao) GetAllEntranceIds(ctx context.Context) (res []int64, err error) {
	rows, err := d.db.Query(ctx, _allentrance, _hotChannelName, _hotTopicName)
	if err != nil {
		log.Error("[GetAllEntranceIds] d.db.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			log.Error("[GetAllEntranceIds] rows.Scan() error(%v)", err)
			return
		}
		res = append(res, id)
	}
	if err = rows.Err(); err != nil {
		log.Error("[GetAllEntranceIds] rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) GetAIData(ctx context.Context, ids []int64) (data map[int64][]show.PopAIChannelResource, err error) {
	data = make(map[int64][]show.PopAIChannelResource)
	for _, item := range ids {
		var (
			res struct {
				Code int                         `json:"code"`
				List []show.PopAIChannelResource `json:"list"`
			}
		)
		if err = d.client.Get(ctx, fmt.Sprintf(d.entranceURL, item), "", nil, &res); err != nil {
			log.Error("[GetAIDataByUrl] d.client.Get() url(%s) id(%d) error(%v)", d.entranceURL, item, err)
			err = nil
			continue
		}
		if res.Code != ecode.OK.Code() {
			log.Error("[GetAIDataByUrl] d.client.Get() url(%s) id(%d) code(%d)", d.entranceURL, item, res.Code)
			continue
		}
		data[item] = res.List
	}
	return
}

func (d *Dao) AddCacheAIData(ctx context.Context, data map[int64][]show.PopAIChannelResource) (err error) {
	var (
		item        []byte
		key         string
		keys        []string
		argsRecords = redis.Args{}
		conn        = d.redis.Get(ctx)
	)
	defer conn.Close()
	for id, dataItem := range data {
		if dataItem == nil { // ignore nil record
			continue
		}
		cardDatas := []show.PopularCard{}
		for _, item := range dataItem {
			cardDatas = append(cardDatas, show.PopularCard{
				Type:       item.Goto,
				Value:      item.RID,
				FromType:   item.FromType,
				TagId:      item.TagId,
				Reason:     item.Desc,
				CornerMark: item.CornerMark,
			})
		}
		if item, err = json.Marshal(cardDatas); err != nil {
			log.Error("edge.Marshal error(%v)", err)
			return
		}
		key = fmt.Sprintf(_prefixAiChannelResKey, id)
		keys = append(keys, key)
		argsRecords = argsRecords.Add(key).Add(item)
	}
	if _, err = conn.Do("MSET", argsRecords...); err != nil {
		err = errors.Wrapf(err, "conn.Do(MSET) keys:%+v error", keys)
	}
	return
}

func (d *Dao) AddEntranceCache(ctx context.Context) (err error) {
	var (
		ids  []int64
		rcds map[int64][]show.PopAIChannelResource
	)
	if ids, err = d.GetAllEntranceIds(ctx); err != nil || len(ids) == 0 {
		log.Error("AddEntranceCache GetAllEntranceIds err(%+v) ids(%d)", err, len(ids))
		return
	}
	if rcds, err = d.GetAIData(ctx, ids); err != nil || len(rcds) == 0 {
		log.Error("AddEntranceCache GetAIData err(%+v) rcds(%d)", err, len(rcds))
		return
	}
	if err = d.AddCacheAIData(ctx, rcds); err != nil {
		log.Error("AddEntranceCache AddCacheAIData err(%+v) rcds(%d)", err, len(rcds))
	}
	return
}

func (d *Dao) GetAllEntrances(ctx context.Context) (res []show.PopTopEntrance, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _allentrances); err != nil {
		log.Error("[GetAllEntrances] d.db.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item show.PopTopEntrance
		if err = rows.Scan(&item.ID, &item.Rank); err != nil {
			log.Error("[GetAllEntrances] rows.Scan() error(%v)", err)
			return
		}
		res = append(res, item)
	}
	if err = rows.Err(); err != nil {
		log.Error("[GetAllEntrances] rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) UpdateEntrancesRank(ctx context.Context, res []show.PopTopEntrance) (err error) {
	if len(res) == 0 {
		return
	}
	var tx *sql.Tx
	if tx, err = d.db.Begin(context.Background()); err != nil || tx == nil {
		log.Error("UpdateEntrancesRank db: BeginTran d.db.Begin error(%v)", err)
		return
	}
	for _, item := range res {
		if _, err = tx.Exec(_updateEntranceRank, item.Rank, item.ID); err != nil {
			_ = tx.Rollback()
			err = errors.Wrapf(err, "UpdateEntrancesRank d.db.Exec(%s) error(%v) entrance %v", _updateEntranceRank, err, item)
			return
		}
	}
	err = tx.Commit()
	return
}

func (d *Dao) AddRankCache(ctx context.Context, id, rank int) (err error) {
	var (
		key  = fmt.Sprintf(_prefixRankKey, id)
		conn = d.redis.Get(ctx)
	)
	defer conn.Close()
	if _, err = conn.Do("SET", key, rank); err != nil {
		log.Error("conn.Do(SET, %s, %d) error(%v)", key, rank, err)
	}
	return
}

func (d *Dao) CacheRankCache(c context.Context, id int) (res int, err error) {
	var (
		key  = fmt.Sprintf(_prefixRankKey, id)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
	}
	return
}

func (d *Dao) AllEntranceIcon(c context.Context) (res map[int64]string, err error) {
	rows, err := d.db.Query(c, _allEntranceIcon)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]string)
	for rows.Next() {
		var (
			id   int64
			icon string
		)
		if err = rows.Scan(&id, &icon); err != nil {
			log.Error("%v", err)
			return
		}
		if icon != "" {
			res[id] = icon
		}
	}
	err = rows.Err()
	return
}
