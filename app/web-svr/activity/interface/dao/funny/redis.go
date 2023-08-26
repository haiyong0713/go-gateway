package funny

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"time"
)

func (d *dao) GetUserTodayIsAdded(c context.Context, mid int64) (IsAdd int, err error) {
	var (
		key   = buildKey(mid, time.Now().Format("20060102"))
		isAdd = 0
	)
	conn := d.redis.Get(c)

	defer conn.Close()
	if isAdd, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err != redis.ErrNil {
			log.Warnc(c, "api GetUserTodayIsAdded conn.Do(GET %s) error(%v)", key, err)
			return
		}
	}

	return isAdd, nil
}

func (d *dao) SetUserAddedTimes(c context.Context, mid int64) (err error) {
	var (
		key = buildKey(mid, time.Now().Format("20060102"))
	)
	conn := d.redis.Get(c)

	defer conn.Close()
	if _, err = redis.Int(conn.Do("SET", key, 1)); err != nil {
		if err != redis.ErrNil {
			log.Warnc(c, "api SetUserAddedTimes conn.Do(SET %s 1) error(%v)", key, err)
			return
		}
	}

	if _, err = redis.Int(conn.Do("EXPIRE", key, 60*60*24*3)); err != nil {
		if err != redis.ErrNil {
			log.Errorc(c, "api SetUserAddedTimes set expire time err key:%v err:%v", key, err)
		}
	}

	return nil
}

func (d *dao) GetTask1Num(c context.Context) (count int, err error) {
	var (
		key = buildKey("task1")
	)
	conn := d.redis.Get(c)

	defer conn.Close()
	if count, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err != redis.ErrNil {
			log.Warnc(c, "api GetTask1Num conn.Do(GET %s) error(%v)", key, err)
			return
		}
	}

	return count, nil
}

func (d *dao) GetTask2Num(c context.Context) (count int, err error) {
	var (
		key = buildKey("task2")
	)
	conn := d.redis.Get(c)

	defer conn.Close()
	if count, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err != redis.ErrNil {
			log.Warnc(c, "api GetTask2Num conn.Do(GET %s) error(%v)", key, err)
			return
		}
	}

	return count, nil
}
