package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/cache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"

	"go-gateway/app/web-svr/web/interface/model"

	"git.bilibili.co/bapis/bapis-go/activity/service"
	emotegrpc "git.bilibili.co/bapis/bapis-go/community/service/emote"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
)

func (d *Dao) CacheWeeklySeries(ctx context.Context, typ string) ([]*model.SeriesConfig, error) {
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	key := fmt.Sprintf("popular_series_%s", typ)
	value, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, errors.WithStack(ecode.NothingFound)
		}
		return nil, errors.Wrap(err, key)
	}
	var res []*model.SeriesConfig
	if err = json.Unmarshal(value, &res); err != nil {
		return nil, errors.Wrapf(err, "%s", value)
	}
	return res, nil
}

func (d *Dao) CacheSeriesDetail(ctx context.Context, seriesIDs ...int64) (map[int64][]*model.SeriesList, error) {
	if len(seriesIDs) == 0 {
		return nil, nil
	}
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	var args redis.Args
	for _, id := range seriesIDs {
		key := fmt.Sprintf("popular_series_detail_%d", id)
		args = args.Add(key)
	}
	bss, err := redis.ByteSlices(conn.Do("MGET", args...))
	if err != nil {
		return nil, err
	}
	res := map[int64][]*model.SeriesList{}
	for k, bs := range bss {
		if bs == nil {
			continue
		}
		var v []*model.SeriesList
		if err := json.Unmarshal(bs, &v); err != nil {
			return nil, errors.Wrapf(err, "%s", bs)
		}
		res[seriesIDs[k]] = v
	}
	return res, nil
}

func (d *Dao) UserEmoteUnlock(ctx context.Context, req *emotegrpc.UserEmoteUnlockReq) (*empty.Empty, error) {
	return d.EmoteClient.UserEmoteUnlock(ctx, req)
}

func (d *Dao) RewardsSendAwardV2(ctx context.Context, req *api.RewardsSendAwardV2Req) (*api.RewardsSendAwardReply, error) {
	return d.ActivityClient.RewardsSendAwardV2(ctx, req)
}

const _popularBadgeAwardSQL = "SELECT id FROM popular_badge_award WHERE mid = ?"

func (d *Dao) RawPopularBadgeAward(ctx context.Context, mid int64) (int64, error) {
	var id int64
	rows := d.showDB.QueryRow(ctx, _popularBadgeAwardSQL, mid)
	if err := rows.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, errors.Wrapf(err, "RawPopularBadgeAward() rows.Scan id=%d", id)
	}
	return id, nil
}

const (
	_popularAwardSQL   = "SELECT id, award_name FROM %s WHERE mid = ?"
	_popularAwardTable = "popular_award_%d"
)

func (d *Dao) RawPopularAward(ctx context.Context, mid int64) (map[string]int64, error) {
	rows, err := d.showDB.Query(ctx, fmt.Sprintf(_popularAwardSQL, awardTable(mid)), mid)
	if err != nil {
		return nil, errors.Wrapf(err, "d.showDB.Query error")
	}
	defer rows.Close()
	var (
		id        int64
		awardName string
	)
	res := map[string]int64{model.AwardStep1: 0, model.AwardStep2: 0, model.AwardStep3: 0, model.AwardStep4: 0}
	for rows.Next() {
		if err = rows.Scan(&id, &awardName); err != nil {
			return nil, errors.Wrapf(err, "rows.Scan error")
		}
		res[awardName] = id
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "rows.Err() error")
	}
	return res, nil
}

func awardTable(mid int64) string {
	return fmt.Sprintf(_popularAwardTable, mid%10)
}

const (
	_updatePopularBadgeAwardSQL = "UPDATE popular_badge_award set mid = ? WHERE mid = 0 ORDER BY id LIMIT 1"
	_countPopularBadgeAwardSQL  = "SELECT COUNT(*) FROM popular_badge_award WHERE mid > 0"
)

func (d *Dao) CountPopularBadgeAward(ctx context.Context) (int64, error) {
	var cnt int64
	row := d.showDB.QueryRow(ctx, _countPopularBadgeAwardSQL)
	if err := row.Scan(&cnt); err != nil {
		return 0, err
	}
	return cnt, nil
}

func (d *Dao) UpdatePopularBadgeAward(ctx context.Context, mid int64) (int64, error) {
	res, err := d.showDB.Exec(ctx, _updatePopularBadgeAwardSQL, mid)
	if err != nil {
		return 0, errors.Wrapf(err, "UpdatePopularBadgeAward() d.showDB.Exec mid:%d", mid)
	}
	return res.RowsAffected()
}

const _addPopularAwardSQL = "INSERT INTO %s (`mid`,`award_name`,`token`) VALUES (?,?,?)"

func (d *Dao) AddPopularAward(ctx context.Context, mid int64, awardName, token string) (int64, error) {
	res, err := d.showDB.Exec(ctx, fmt.Sprintf(_addPopularAwardSQL, awardTable(mid)), mid, awardName, token)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

const _popularRankSQL = "SELECT id FROM popular_rank WHERE mid = ?"

func (d *Dao) RawPopularRank(ctx context.Context, mid int64) (int64, error) {
	id := int64(0)
	rows := d.showDB.QueryRow(ctx, _popularRankSQL, mid)
	if err := rows.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}

const (
	_addPopularRankSQL = "INSERT INTO popular_rank (`mid`,`rank_from`) VALUES (?,?)"
	_rankFrom          = "makeup"
)

func (d *Dao) AddPopularRank(ctx context.Context, mid int64) (int64, error) {
	res, err := d.showDB.Exec(ctx, _addPopularRankSQL, mid, _rankFrom)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

const (
	_popularWatchSQL = "SELECT wtime FROM %s WHERE mid = ?"
	_popularWatch    = "popular_watch_%d"
	_popularWatch0   = "popular_watch_0_%d"
)

func (d *Dao) RawPopularWatchTime(ctx context.Context, mid, stage int64) (xtime.Time, error) {
	t := xtime.Time(0)
	rows := d.showDB.QueryRow(ctx, fmt.Sprintf(_popularWatchSQL, watchTable(mid, stage)), mid)
	if err := rows.Scan(&t); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return t, nil
}

func watchTable(mid, stage int64) string {
	if stage > 0 {
		return fmt.Sprintf(_popularWatch, stage)
	}
	return fmt.Sprintf(_popularWatch0, mid%10)
}

func popularWatchKey(mid, stage int64) string {
	return fmt.Sprintf("precious:watch:%d:%d", mid, stage)
}

// CachePopularWatchTime get data from redis
func (d *Dao) CachePopularWatchTime(ctx context.Context, id int64, stage int64) (res xtime.Time, err error) {
	key := popularWatchKey(id, stage)
	var temp []byte
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	temp, err = redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		log.Errorc(ctx, "d.CachePopularWatchTime(get key: %v) err: %+v", key, err)
		return 0, err
	}
	v := string(temp)
	r, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Errorc(ctx, "d.CachePopularWatchTime(get key: %v) err: %+v", key, err)
		return 0, err
	}
	res = xtime.Time(r)
	return res, nil
}

// AddCachePopularWatchTime Set data to redis
func (d *Dao) AddCachePopularWatchTime(ctx context.Context, id int64, val xtime.Time, stage int64) error {
	key := popularWatchKey(id, stage)
	bs := []byte(strconv.FormatInt(int64(val), 10))
	expire := 86400
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("set", key, bs, "EX", expire); err != nil {
		log.Errorc(ctx, "d.AddCachePopularWatchTime(get key: %v) err: %+v", key, err)
		return err
	}
	return nil
}

func (d *Dao) PopularWatchTime(ctx context.Context, id int64, stage int64) (res xtime.Time, err error) {
	addCache := true
	res, err = d.CachePopularWatchTime(ctx, id, stage)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res == -1 {
			res = 0
		}
	}()
	if res != 0 {
		cache.MetricHits.Inc("bts:PopularWatchTime")
		return
	}
	cache.MetricMisses.Inc("bts:PopularWatchTime")
	res, err = d.RawPopularWatchTime(ctx, id, stage)
	if err != nil {
		return
	}
	miss := res
	if miss == 0 {
		miss = -1
	}
	if !addCache {
		return
	}
	if err := d.AddCachePopularWatchTime(ctx, id, miss, stage); err != nil {
		log.Error("Failed to AddCachePopularWatchTime: %d, %d, %d, %+v", id, miss, stage, err)
	}
	return
}

func popularRankKey(mid int64) string {
	return fmt.Sprintf("precious:rank:%d", mid)
}

// CachePopularRank get data from redis
func (d *Dao) CachePopularRank(ctx context.Context, mid int64) (res int64, err error) {
	key := popularRankKey(mid)
	var temp []byte
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	temp, err = redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		log.Errorc(ctx, "d.CachePopularRank(get key: %v) err: %+v", key, err)
		return 0, err
	}
	v := string(temp)
	r, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Errorc(ctx, "d.CachePopularRank(get key: %v) err: %+v", key, err)
		return 0, err
	}
	res = r
	return res, nil
}

// AddCachePopularRank Set data to redis
func (d *Dao) AddCachePopularRank(ctx context.Context, mid int64, val int64) error {
	key := popularRankKey(mid)
	bs := []byte(strconv.FormatInt(val, 10))
	expire := 86400
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("set", key, bs, "EX", expire); err != nil {
		log.Errorc(ctx, "d.AddCachePopularRank(get key: %v) err: %+v", key, err)
		return err
	}
	return nil
}

func (d *Dao) PopularRank(ctx context.Context, mid int64) (res int64, err error) {
	addCache := true
	res, err = d.CachePopularRank(ctx, mid)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res == -1 {
			res = 0
		}
	}()
	if res != 0 {
		cache.MetricHits.Inc("bts:PopularRank")
		return
	}
	cache.MetricMisses.Inc("bts:PopularRank")
	res, err = d.RawPopularRank(ctx, mid)
	if err != nil {
		return
	}
	miss := res
	if miss == 0 {
		miss = -1
	}
	if !addCache {
		return
	}
	if err := d.AddCachePopularRank(ctx, mid, miss); err != nil {
		log.Error("Failed to AddCachePopularRank: %d, %d, %+v", mid, miss, err)
	}
	return
}

func (d *Dao) DelCachePopularRank(ctx context.Context, mid int64) error {
	cacheKey := popularRankKey(mid)
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("DEL", cacheKey); err != nil {
		return err
	}
	return nil
}

func popularAwardKey(mid int64) string {
	return fmt.Sprintf("precious:award:%d", mid)
}

func popularBadgeAwardKey(mid int64) string {
	return fmt.Sprintf("precious:badge_award:%d", mid)
}

func (d *Dao) CachePopularBadgeAward(ctx context.Context, mid int64) (int64, error) {
	const (
		_redisErrNil = -1
	)
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	key := popularBadgeAwardKey(mid)
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return _redisErrNil, nil
		}
		return 0, errors.Wrapf(err, "CachePopularBadgeAward() redis.Bytes key:%s", key)
	}
	res, err := strconv.ParseInt(string(bs), 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "CachePopularBadgeAward() strconv.ParseInt bs:%s", string(bs))
	}
	return res, nil
}

func (d *Dao) AddCachePopularBadgeAward(ctx context.Context, mid int64, val int64) error {
	key := popularBadgeAwardKey(mid)
	bs := []byte(strconv.FormatInt(val, 10))
	expire := 86400
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("set", key, bs, "EX", expire); err != nil {
		return errors.Wrapf(err, "AddCachePopularBadgeAward() error key: %s", key)
	}
	return nil
}

func (d *Dao) DelCachePopularBadgeAward(ctx context.Context, mid int64) error {
	cacheKey := popularBadgeAwardKey(mid)
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("DEL", cacheKey); err != nil {
		return err
	}
	return nil
}

func (d *Dao) CachePopularAward(ctx context.Context, mid int64) (map[string]int64, bool, error) {
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	key := popularAwardKey(mid)
	bt, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, true, nil
		}
		return nil, false, errors.Wrapf(err, "CachePopularAward() redis.Bytes key:%s", key)
	}
	var res map[string]int64
	if err = json.Unmarshal(bt, &res); err != nil {
		return nil, false, errors.WithStack(err)
	}
	if res == nil {
		return nil, true, nil
	}
	return res, false, nil
}

func (d *Dao) AddCachePopularAward(ctx context.Context, mid int64, val map[string]int64) error {
	if val == nil {
		return nil
	}
	key := popularAwardKey(mid)
	bs, err := json.Marshal(val)
	if err != nil {
		return errors.WithStack(err)
	}
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	if _, err = conn.Do("set", key, bs, "EX", 86400); err != nil {
		return errors.Wrapf(err, "AddCachePopularAward() conn.Do key:%s", key)
	}
	return nil
}

func (d *Dao) PopularBadgeAward(ctx context.Context, mid int64) (int64, error) {
	// res>0：已领取， =0未领取；-1不存在记录
	res, err := d.CachePopularBadgeAward(ctx, mid)
	if err != nil {
		return 0, err
	}
	// 如果key存在，该用户领过/未领取徽章，直接返回
	// 如果key不存在则回源
	if res >= 0 {
		return res, nil
	}
	// 读mysql, mysql查询出错返回err，不存在会返回0
	res, err = d.RawPopularBadgeAward(ctx, mid)
	if err != nil {
		return 0, err
	}
	// 写redis 如果遇到错误直接返回res
	if err = d.AddCachePopularBadgeAward(ctx, mid, res); err != nil {
		log.Error("Failed to AddCachePopularAward: %d, %d, %+v", mid, res, err)
		return res, nil
	}
	return res, nil
}

func (d *Dao) PopularAward(ctx context.Context, mid int64) (map[string]int64, error) {
	res, isNil, err := d.CachePopularAward(ctx, mid)
	if err != nil {
		return nil, errors.Wrapf(err, "PopularAward() d.CachePopularAward mid:%d", mid)
	}
	if !isNil {
		return res, nil
	}
	res, err = d.RawPopularAward(ctx, mid)
	if err != nil {
		return nil, errors.Wrapf(err, "PopularAward() d.RawPopularAward mid:%d", mid)
	}
	if err = d.AddCachePopularAward(ctx, mid, res); err != nil {
		log.Error("Failed to AddCachePopularAward: %d, %+v, %+v", mid, res, err)
		return res, nil
	}
	return res, nil
}

func (d *Dao) DelCachePopularAward(ctx context.Context, mid int64) error {
	cacheKey := popularAwardKey(mid)
	conn := d.redisBak.Get(ctx)
	defer conn.Close()
	if _, err := conn.Do("DEL", cacheKey); err != nil {
		return err
	}
	return nil
}
