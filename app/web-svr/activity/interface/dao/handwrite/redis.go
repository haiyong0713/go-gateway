package handwrite

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	mdl "go-gateway/app/web-svr/activity/interface/model/handwrite"
)

// GetMidAward 获得用户获奖情况
func (d *dao) GetMidAward(c context.Context, mid int64) (res *mdl.MidAward, err error) {
	var (
		bs   []byte
		key  = d.buildKey(awardKey, midKey, mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// GetAwardCount 获取获奖总人数
func (d *dao) GetAwardCount(c context.Context) (res *mdl.AwardCount, err error) {
	var (
		bs   []byte
		key  = d.buildKey(awardCountKey)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddTimeLock 增加锁
func (d *dao) AddTimeLock(c context.Context, mid int64) (err error) {
	var reply interface{}
	conn := d.redis.Get(c)
	var key = d.buildKey(addTimesLock, mid)
	defer conn.Close()

	if reply, err = conn.Do("SET", key, "LOCK", "EX", 1, "NX"); err != nil {
		log.Error("SETEX(%v) error(%v)", key, err)
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
	var key = d.buildKey(alreadyAddTimes, mid, day)
	defer conn.Close()

	if reply, err = conn.Do("SET", key, "True", "EX", 86400, "NX"); err != nil {
		log.Error("SETEX(%v) error(%v)", key, err)
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
		key  = d.buildKey(alreadyAddTimes, mid, day)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}

	return string(bs), err
}

// GetMidTask 获得用户获奖情况
func (d *dao) GetMidTask(c context.Context, mid int64) (res *mdl.MidTaskAll, err error) {
	var (
		bs   []byte
		key  = d.buildKey(taskKey, midKey, mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// GetTaskCount 获取获奖总人数
func (d *dao) GetTaskCount(c context.Context) (res *mdl.AwardCountNew, err error) {
	var (
		bs   []byte
		key  = d.buildKey(taskCountKey)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}
