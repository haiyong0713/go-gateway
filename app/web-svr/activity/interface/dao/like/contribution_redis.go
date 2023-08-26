package like

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/component"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_preAddTimes = "c_add_times_%d_%d_%s"
	_preLock     = "c_Lock_%d"
	_partIn      = "part_in_%d"
)

func lightCount(mid int64) string {
	return fmt.Sprintf("light_count_%d", mid)
}

// userContriKey .
func userContriKey(mid int64) string {
	return fmt.Sprintf("contri_mid_%d", mid)
}

func contriAddTimes(mid int64, actionType int, day string) string {
	return fmt.Sprintf(_preAddTimes, mid, actionType, day)
}

func contriLokc(mid int64) string {
	return fmt.Sprintf(_preLock, mid)
}

func partIn(mid int64) string {
	return fmt.Sprintf(_partIn, mid)
}

// GetAddTimesRecord 增加锁，同时避免多次领取
func (d *Dao) GetAddTimesRecord(c context.Context, mid int64, actionType int, day string) (res string, err error) {
	var (
		bs  []byte
		key = contriAddTimes(mid, actionType, day)
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	return string(bs), err
}

// AddTimeLock 增加锁，同时避免多次领取
func (d *Dao) AddTimesRecord(c context.Context, mid int64, actionType int, day string) (err error) {
	var reply interface{}
	conn := component.GlobalRedis.Conn(c)
	var key = contriAddTimes(mid, actionType, day)
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

// AddTimeLock 增加锁
func (d *Dao) AddTimeLock(c context.Context, mid int64) (err error) {
	var reply interface{}
	conn := component.GlobalRedis.Conn(c)
	var key = contriLokc(mid)
	defer conn.Close()

	if reply, err = conn.Do("SET", key, "LOCK", "EX", 100000, "NX"); err != nil {
		log.Error("SETEX(%v) error(%v)", key, err)
		return
	}
	if reply == nil {
		err = ecode.ActivityWriteHandAddtimesTooFastErr
	}
	return
}

// UserContribution
func (d *Dao) UserContribution(c context.Context, mid int64) (data *likemdl.ContributionUser, err error) {
	var (
		key = userContriKey(mid)
		bs  []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("UserContribution redis.String(conn.Do(GET,%s)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &data); err != nil {
		log.Error("UserContribution json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheLightVideoJoin
func (d *Dao) AddCacheLightVideoJoin(ctx context.Context, mid int64) (err error) {
	key := partIn(mid)
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, 5184000, 1); err != nil {
		log.Error("LightVideoJoin conn.Send(SETEX, %s, %v, %s) error(%v)", key, 5184000, 1, err)
	}
	return
}

// CacheLightVideoJoin .
func (d *Dao) CacheLightVideoJoin(ctx context.Context, mid int64) (r int64, err error) {
	key := partIn(mid)
	if r, err = redis.Int64(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLightVideoJoin(%s) return nil", key)
		} else {
			log.Errorc(ctx, "CacheLightVideoJoin conn.Do(GET key(%v)) error(%v)", key, err)
		}
	}
	return
}

// AddCacheLightCount
func (d *Dao) AddCacheLightCount(ctx context.Context, mid int64, val int64) (err error) {
	key := lightCount(mid)
	if val == 0 {
		val = -1
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, d.lightCountExpire, val); err != nil {
		log.Errorc(ctx, "AddCacheKnowRule conn.Send(SETEX, %s, %d) error(%v)", key, d.lightCountExpire, 1, err)
	}
	return
}

// CacheLightCount .
func (d *Dao) CacheLightCount(ctx context.Context, mid int64) (r int64, err error) {
	key := lightCount(mid)
	if r, err = redis.Int64(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheLightCount(%s) return nil", key)
		} else {
			log.Error("CacheLightCount conn.Do(GET key(%v)) error(%v)", key, err)
		}
	}
	return
}
