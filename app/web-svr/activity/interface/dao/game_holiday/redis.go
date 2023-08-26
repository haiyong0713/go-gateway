package gameholiday

import (
	"context"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
)

// AddTimeLock 增加锁
func (d *dao) AddTimeLock(c context.Context, mid int64) (err error) {
	var reply interface{}
	conn := d.redis.Get(c)
	var key = buildKey(addTimesLock, mid)
	defer conn.Close()

	if reply, err = conn.Do("SET", key, "LOCK", "EX", 1, "NX"); err != nil {
		log.Errorc(c, "SETEX(%v) error(%v)", key, err)
		return
	}
	if reply == nil {
		err = ecode.ActivityWriteHandAddtimesTooFastErr
	}
	return
}

// AddTimeLock 增加锁，同时避免多次领取
func (d *dao) AddTimesRecord(c context.Context, mid int64, day string) (err error) {
	var reply interface{}
	conn := d.redis.Get(c)
	var key = buildKey(alreadyAddTimes, mid, day)
	defer conn.Close()

	if reply, err = conn.Do("SET", key, "True", "EX", 86400, "NX"); err != nil {
		log.Errorc(c, "SETEX(%v) error(%v)", key, err)
		return
	}
	if reply == nil {
		err = ecode.ActivityWriteHandAddtimesTooFastErr
	}
	return
}

// GetAddTimesRecord 增加锁，同时避免多次领取
func (d *dao) GetAddTimesRecord(c context.Context, mid int64, day string) (res string, err error) {
	var (
		bs   []byte
		key  = buildKey(alreadyAddTimes, mid, day)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}

	return string(bs), err
}
