package bws

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/log"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

// CacheUserDetail .
func (d *Dao) CacheUserDetail(c context.Context, bid, mid int64, date string) (res *bwsmdl.UserDetail, err error) {

	var (
		bs       []byte
		cacheKey = keyUserDetail(bid, mid, date)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", cacheKey)); err != nil {

		log.Errorc(c, "CacheLotteryTimesConfig conn.Do(GET key(%v)) error(%v)", cacheKey, err)
		if err == redis.ErrNil {
			err = nil
		}
		return
	}
	res = &bwsmdl.UserDetail{}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// IncrCouponMid Dao
func (d *Dao) IncrCouponMid(c context.Context, bid, mid int64) (res int64, err error) {

	key := keyMidCoupon(bid, mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if res, err = redis.Int64(conn.Do("INCR", key)); err != nil {
		log.Error("Failed to increase hotpot count: %s: %+v", key, err)
		return
	}
	if _, err = conn.Do("EXPIRE", key, d.bwsRankUserExpire); err != nil {
		log.Error("Failed to set expiration increase hotpot count: %s: %+v", key, err)
		return
	}
	return
}

// CacheUserDetails .
func (d *Dao) CacheUserDetails(c context.Context, mids []int64, bid int64, date string) (res map[int64]*bwsmdl.UserDetail, err error) {

	var bss [][]byte
	if len(mids) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	for _, v := range mids {
		args = args.Add(keyUserDetail(bid, v, date))
	}
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheUserDetails redis.Ints(conn.Do(MGET,%v) error(%v)", args, err)
		}
		return
	}
	res = make(map[int64]*bwsmdl.UserDetail)
	for _, val := range bss {
		data := &bwsmdl.UserDetail{}
		if err = json.Unmarshal(val, &data); err != nil {
			log.Errorc(c, "json.Unmarshal(%v) error(%v)", string(val), err)
			continue
		}
		res[data.Mid] = data
	}
	return
}

// AddCacheUserDetails .
func (d *Dao) AddCacheUserDetails(c context.Context, userDetail map[int64]*bwsmdl.UserDetail, bid int64, date string) (err error) {

	if len(userDetail) == 0 {
		return nil
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	argsMDs := redis.Args{}
	var keys []string
	for mid, v := range userDetail {
		var bs = []byte{}
		if bs, err = json.Marshal(v); err != nil {
			log.Error("json.Marshal(%v) error (%v)", userDetail, err)
			return
		}
		key := keyUserDetail(bid, mid, date)
		argsMDs = argsMDs.Add(key).Add(string(bs))
		keys = append(keys, key)
	}
	if err := conn.Send("MSET", argsMDs...); err != nil {
		log.Errorc(c, "AddCacheUserDetails MSET error(%v)", err)
		return err
	}
	count := 1
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, d.bwsOfflineUserExpire); err != nil {
			log.Errorc(c, "AddCacheAwardByIDs conn.Send(Expire, %s, %d) error(%v)", v, d.bwsOfflineUserExpire, err)
			return err
		}
		count++
	}
	if err := conn.Flush(); err != nil {
		log.Errorc(c, "AddCachePrintByIDs Flush error(%v)", err)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Errorc(c, "AddCacheAwardByIDs conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

// AddCacheUserDetail .
func (d *Dao) AddCacheUserDetail(c context.Context, bid int64, userDetail *bwsmdl.UserDetail, mid int64, date string) (res *bwsmdl.UserDetail, err error) {
	var (
		bs       []byte
		cacheKey = keyUserDetail(bid, mid, date)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(userDetail); err != nil {
		log.Error("json.Marshal(%v) error (%v)", userDetail, err)
		return
	}
	if _, err = conn.Do("SETEX", cacheKey, d.bwsOfflineUserExpire, bs); err != nil {
		log.Errorc(c, "conn.Send(SETEX, %s, %v, %s) error(%v)", cacheKey, d.bwsOfflineUserExpire, string(bs), err)
	}
	return
}

// DelCacheUserDetail .
func (d *Dao) DelCacheUserDetail(c context.Context, bid int64, mid int64, date string) (err error) {
	var (
		cacheKey = keyUserDetail(bid, mid, date)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", cacheKey); err != nil {
		log.Errorc(c, "conn.Send(DEL, %s) error(%v)", cacheKey, err)
	}
	return
}

// AddCacheInsertUserScore ...
func (d *Dao) AddCacheInsertUserScore(c context.Context, bid int64, userRank *bwsmdl.UserRank, date string) (err error) {
	if userRank == nil {
		return
	}
	cacheKey := keyUserRank(bid, date)
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(cacheKey)
	args = args.Add(userRank.Score).Add(userRank.Mid)
	if err = conn.Send("ZADD", args...); err != nil {
		log.Errorc(c, "AddCacheUserRank conn.Send(ZADD, %s, %v) error(%v)", cacheKey, args, err)
		return
	}
	if err = conn.Send("EXPIRE", cacheKey, d.bwsRankUserExpire); err != nil {
		log.Errorc(c, "AddCacheUserRank conn.Send(Expire, %s) error(%v)", cacheKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Errorc(c, "AddCacheUserRank conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Errorc(c, "AddCacheUserRank conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// DelRankMid  删除用户排行，干预
func (d *Dao) DelRankMid(c context.Context, bid int64, date string, mid int64) (err error) {
	var (
		cacheKey = keyUserRank(bid, date)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("ZREM", cacheKey, mid); err != nil {
		log.Errorc(c, "DelRankMid conn.Do(ZREM) err(%v)", err)
		return err
	}
	return nil
}

// CacheUserRankMinScore 最低分
func (d *Dao) CacheUserRankMinScore(c context.Context, bid int64, date string, rank int64) (score float64, err error) {
	var (
		cacheKey = keyUserRank(bid, date)
		conn     = d.redis.Get(c)
		values   []interface{}
	)
	defer conn.Close()

	if values, err = redis.Values(conn.Do("ZRANGE", cacheKey, rank, rank, "WITHSCORES")); err != nil {

		log.Errorc(c, "conn.Send(ZRANGE, %s) error(%v)", cacheKey, err)
		return
	}
	if len(values) == 0 {
		return float64(0), nil
	}
	if len(values) > 0 {
		var mid int64
		if values, err = redis.Scan(values, &mid, &score); err != nil {
			log.Errorc(c, "CacheUserRankMinScore redis.Scan(%v) error(%v)", values, err)
			return
		}
		return score, nil
	}
	return
}

// UserRank ...
func (d *Dao) UserRank(c context.Context, bid, mid int64, date string) (rank int64, err error) {
	var (
		cacheKey = keyUserRank(bid, date)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()

	if rank, err = redis.Int64(conn.Do("ZRANK", cacheKey, mid)); err != nil {
		if err == redis.ErrNil {
			err = nil
		}
		log.Errorc(c, "UserRank conn.Send(zrank, %s) error(%v)", cacheKey, err)
		return
	}
	return rank + 1, nil
}

// CacheUserRank 获取排行榜
func (d *Dao) CacheUserRank(c context.Context, bid int64, date string, rank int64) (list []*bwsmdl.UserRank, err error) {
	var (
		cacheKey = keyUserRank(bid, date)
		conn     = d.redis.Get(c)
		values   []interface{}
	)
	list = make([]*bwsmdl.UserRank, 0)
	defer conn.Close()

	if values, err = redis.Values(conn.Do("ZRANGE", cacheKey, 0, rank, "WITHSCORES")); err != nil {
		log.Errorc(c, "CacheUserRank conn.Send(ZRANGE, %s) error(%v)", cacheKey, err)
		return
	}
	if len(values) == 0 {
		return list, nil
	}
	if len(values) > 0 {
		for len(values) > 0 {
			var mid int64
			var score float64
			if values, err = redis.Scan(values, &mid, &score); err != nil {
				log.Errorc(c, "CacheUserRank redis.Scan(%v) error(%v)", values, err)
				return
			}
			object := &bwsmdl.UserRank{
				Mid:   mid,
				Score: score,
			}
			list = append(list, object)
		}

		return list, nil
	}
	return
}

// AddRankCache .
func (d *Dao) AddRankCache(c context.Context, bid int64, rank map[int64]*bwsmdl.MidScore, date string) (err error) {
	var (
		bs       []byte
		cacheKey = keyBwsRank(bid, date)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(rank); err != nil {
		log.Error("json.Marshal(%v) error (%v)", rank, err)
		return
	}
	if _, err = conn.Do("SETEX", cacheKey, d.bwsRankUserExpire, bs); err != nil {
		log.Errorc(c, "conn.Send(SETEX, %s, %v, %s) error(%v)", cacheKey, d.bwsRankUserExpire, string(bs), err)
	}
	return
}

// GetRankCache ...
func (d *Dao) GetRankCache(c context.Context, bid int64, date string) (res map[int64]*bwsmdl.MidScore, err error) {
	var (
		bs       []byte
		cacheKey = keyBwsRank(bid, date)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", cacheKey)); err != nil {

		log.Errorc(c, "GetRankCache conn.Do(GET key(%v)) error(%v)", cacheKey, err)
		if err == redis.ErrNil {
			err = nil
		}
		return
	}
	res = make(map[int64]*bwsmdl.MidScore)
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}
