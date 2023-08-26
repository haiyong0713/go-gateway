package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/currency"
)

func lockNumKey(sid int64) string {
	return fmt.Sprintf("cur_lock_%d", sid)
}

func unlockStateKey(date string) string {
	return fmt.Sprintf("curr_unlock_%s", date)
}

func mikuStateKey(sid, mid int64) string {
	return fmt.Sprintf("miku_%d_%d", sid, mid)
}

func singleStateKey(sid, mid int64) string {
	return fmt.Sprintf("single2_%d_%d", sid, mid)
}

// CacheLockNum .
func (d *Dao) CacheLockNum(c context.Context, sid int64) (res int, err error) {
	key := lockNumKey(sid)
	if res, err = redis.Int(component.GlobalRedisStore.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("CacheLockNum conn.Do(GET(%s)) error(%v)", key, err)
	}
	return
}

// AddCacheLockNum .
func (d *Dao) AddCacheLockNum(c context.Context, sid int64) (count int, err error) {
	key := lockNumKey(sid)
	if count, err = redis.Int(component.GlobalRedisStore.Do(c, "INCR", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("AddCacheLockNum conn.Do(GET(%s)) error(%v)", key, err)
	}
	return

}

// CacheUnlockState .
func (d *Dao) CacheUnlockState(c context.Context, date string) (res int, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := unlockStateKey(date)
	if res, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("CacheUnlockState conn.Do(GET(%s)) error(%v)", key, err)
	}
	return
}

// AddCacheUnlockState Set data to .
func (d *Dao) AddCacheUnlockState(c context.Context, date string, val int) (res bool, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := unlockStateKey(date)
	if res, err = redis.Bool(conn.Do("SETNX", key, val)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("AddCacheUnlockState conn.Do(GET(%s)) error(%v)", key, err)
		return
	}
	if res {
		if _, err = redis.Bool(conn.Do("EXPIRE", key, 86400)); err != nil {
			log.Error("conn.Do(EXPIRE, %s) error(%v)", key, err)
			err = nil
		}
	}
	return
}

// CacheMikuAward .
func (d *Dao) CacheMikuAward(c context.Context, sid, mid int64) (res []*currency.MikuAward, err error) {
	var (
		key  = mikuStateKey(sid, mid)
		conn = d.redis.Get(c)
		val  [][]byte
	)
	defer conn.Close()
	if val, err = redis.ByteSlices(conn.Do("HGETALL", key)); err != nil {
		log.Error("CacheMikuAward conn.Do(HGETALL(%s)) error(%v)", key, err)
		return
	}
	for i := 1; i <= len(val); i = i + 2 {
		data := new(currency.MikuAward)
		if err = json.Unmarshal(val[i], &data); err != nil {
			log.Error("json.Unmarshal(%s,%v) err(%v)", val[i], data, err)
			continue
		}
		res = append(res, data)
	}
	return
}

// CacheSingleAward .
func (d *Dao) CacheSingleAward(c context.Context, sid, mid int64) (res []*currency.SingleAward, err error) {
	var (
		key  = singleStateKey(sid, mid)
		conn = d.redis.Get(c)
		val  [][]byte
	)
	defer conn.Close()
	if val, err = redis.ByteSlices(conn.Do("HGETALL", key)); err != nil {
		log.Error("CacheSingleAward conn.Do(HGETALL(%s)) error(%v)", key, err)
		return
	}
	for i := 1; i <= len(val); i = i + 2 {
		data := new(currency.SingleAward)
		if err = json.Unmarshal(val[i], &data); err != nil {
			log.Error("json.Unmarshal(%s,%v) err(%v)", val[i], data, err)
			continue
		}
		res = append(res, data)
	}
	return
}

// SetCacheSingleAward .
func (d *Dao) SetCacheSingleAward(c context.Context, sid, mid int64, num int, data *currency.SingleAward) (err error) {
	var (
		bs   []byte
		key  = singleStateKey(sid, mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("HSET", key, num, bs); err != nil {
		log.Error("conn.Send(HSET,%s,%d) error(%v)", key, num, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.singleCurrencyExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("set conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("set conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

func (d *Dao) CacheLikeTotal(c context.Context, key string, mid int64) (res int, err error) {
	var (
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLikeTotal(%s) return nil", key)
		} else {
			log.Error("CacheLikeTotal conn.Do(GET key(%v)) error(%v)", key, err)
		}
	}
	return
}
