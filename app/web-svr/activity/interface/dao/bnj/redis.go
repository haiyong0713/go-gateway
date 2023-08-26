package bnj

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

func resetKey(mid int64) string {
	return fmt.Sprintf("bnj_%d", mid)
}

func rewardKey(mid, subID int64, step int) string {
	return fmt.Sprintf("bnj_rwd_%d_%d_%d", mid, subID, step)
}

func rewardsKey(mid, sid int64) string {
	return fmt.Sprintf("bnj20_rwd_%d_%d", mid, sid)
}

func reddotKey(mid, sid int64) string {
	return fmt.Sprintf("bnj20_reddot_%d_%d", mid, sid)
}

func hotpotKey(mid int64, day int) string {
	return fmt.Sprintf("bnj20_hotpot_incr_%d_%d", mid, day)
}

// CacheResetCD .
func (d *Dao) CacheResetCD(c context.Context, mid int64, cd int32) (bool, error) {
	resetCD := d.resetExpire
	if cd > 0 {
		resetCD = cd
	}
	return d.setNXLockCache(c, resetKey(mid), resetCD)
}

// TTLResetCD get reset cd ttl
func (d *Dao) TTLResetCD(c context.Context, mid int64) (ttl int64, err error) {
	key := resetKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if ttl, err = redis.Int64(conn.Do("TTL", key)); err != nil {
		log.Error("TTLResetCD conn.Do(TTL, %s), error(%v)", key, err)
	}
	return
}

// CacheHasReward .
func (d *Dao) CacheHasReward(c context.Context, mid, subID int64, step int) (bool, error) {
	return d.setNXLockCache(c, rewardKey(mid, subID, step), d.rewardExpire)
}

// DelCacheHasReward .
func (d *Dao) DelCacheHasReward(c context.Context, mid, subID int64, step int) error {
	return d.delNXLockCache(c, rewardKey(mid, subID, step))
}

// HasReward .
func (d *Dao) HasReward(c context.Context, mid, subID int64, step int) (res bool, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := rewardKey(mid, subID, step)
	if res, err = redis.Bool(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("HasReward conn.Do(GET(%s)) error(%v)", key, err)
	}
	return
}

func (d *Dao) setNXLockCache(c context.Context, key string, times int32) (res bool, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if res, err = redis.Bool(conn.Do("SETNX", key, "1")); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(SETNX(%s)) error(%v)", key, err)
			return
		}
	}
	if res {
		if _, err = redis.Bool(conn.Do("EXPIRE", key, times)); err != nil {
			log.Error("conn.Do(EXPIRE, %s, %d) error(%v)", key, times, err)
			return
		}
	}
	return
}

func (d *Dao) delNXLockCache(c context.Context, key string) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	_, err = conn.Do("DEL", key)
	return
}

// CacheRewards .
func (d *Dao) CacheRewards(c context.Context, mid, sid int64) (res map[int64]int, err error) {
	var (
		values map[string]int
		key    = rewardsKey(mid, sid)
		conn   = d.redis.Get(c)
	)
	defer conn.Close()
	if values, err = redis.IntMap(conn.Do("HGETALL", key)); err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("conn.Do(HGETALL %s) error(%v)", key, err)
		return
	}
	res = make(map[int64]int, len(res))
	for k, v := range values {
		field, e := strconv.ParseInt(k, 10, 64)
		if e != nil {
			log.Warn("CacheRewards field(%s) strconv.ParseInt error(%v)", k, e)
			continue
		}
		res[field] = v
	}
	return
}

// AddCacheRewards .
func (d *Dao) AddCacheRewards(c context.Context, mid, sid, id int64) (res bool, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := rewardsKey(mid, sid)
	if res, err = redis.Bool(conn.Do("HSETNX", key, id, "1")); err != nil {
		log.Error("conn.Do(HSETNX(%s)) error(%v)", key, err)
		return
	}
	if res {
		if _, err = redis.Bool(conn.Do("EXPIRE", key, d.rewardExpire)); err != nil {
			log.Error("conn.Do(EXPIRE, %s, %d) error(%v)", key, d.rewardExpire, err)
			return
		}
	}
	return
}

// DelCacheRewards .
func (d *Dao) DelCacheRewards(c context.Context, mid, sid, id int64) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := rewardsKey(mid, sid)
	if _, err = conn.Do("HDEL", key, id); err != nil {
		log.Error("HDEL key(%s) field(%d) error(%v)", key, id, err)
	}
	return
}

// CacheClearRedDot .
func (d *Dao) CacheClearRedDot(c context.Context, mid, sid int64) (res int, err error) {
	var (
		key  = reddotKey(mid, sid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int(conn.Do("GET", key)); err == redis.ErrNil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("GET key(%s) error(%v)", key, err)
	}
	return
}

// AddCacheClearRedDot .
func (d *Dao) AddCacheClearRedDot(c context.Context, mid, sid int64) (res bool, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := reddotKey(mid, sid)
	return d.setNXLockCache(c, key, d.rewardExpire)
}

// DelCacheClearRedDot .
func (d *Dao) DelCacheClearRedDot(c context.Context, mid, sid int64) (err error) {
	return d.delNXLockCache(c, reddotKey(mid, sid))
}

// AddCacheDecreaseCD .
func (d *Dao) AddCacheDecreaseCD(c context.Context, mid int64, cd int32) (bool, error) {
	return d.setNXLockCache(c, resetKey(mid), cd)
}

// TTLCacheDecreaseCD get reset cd ttl
func (d *Dao) TTLCacheDecreaseCD(c context.Context, mid int64) (ttl int64, err error) {
	key := resetKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if ttl, err = redis.Int64(conn.Do("TTL", key)); err != nil {
		log.Error("TTLResetCD conn.Do(TTL, %s), error(%v)", key, err)
	}
	return
}

// TTLCacheDecreaseCD get reset cd ttl
func (d *Dao) DelCacheDecreaseCD(c context.Context, mid int64) (ttl int64, err error) {
	key := resetKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	_, err = conn.Do("DEL", key)
	return
}

// HotpotIncreaseCount is
func (d *Dao) HotpotIncreaseCount(ctx context.Context, mid int64) int64 {
	key := hotpotKey(mid, time.Now().Day())
	conn := d.redis.Get(ctx)
	defer conn.Close()
	v, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return 0
		}
		log.Error("Failed to get hotpot increase count: %s: %+v", key, err)
		return 0
	}
	return v
}

// IncreaseHotpotCount is
func (d *Dao) IncreaseHotpotCount(ctx context.Context, mid int64) {
	key := hotpotKey(mid, time.Now().Day())
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err := redis.Int64(conn.Do("INCR", key)); err != nil {
		log.Error("Failed to increase hotpot count: %s: %+v", key, err)
		return
	}
	if _, err := conn.Do("EXPIRE", key, 86400); err != nil {
		log.Error("Failed to set expiration increase hotpot count: %s: %+v", key, err)
		return
	}
	return
}
