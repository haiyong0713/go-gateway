package actplat

import (
	"context"
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/log"
	task "go-gateway/app/web-svr/activity/interface/model/task"
)

// GetTaskCache 任务缓存
func (d *Dao) GetTaskCache(c context.Context, activity string) (res []*task.Detail, err error) {
	var (
		key = buildKey(taskPrefix, activity)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisCache.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "MidCardDetail conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// SetTaskCache 设置任务缓存
func (d *Dao) SetTaskCache(c context.Context, activity string, val []*task.Detail) (err error) {
	var (
		key = buildKey(taskPrefix, activity)
		bs  []byte
	)
	if bs, err = json.Marshal(val); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", val, err)
		return
	}
	if _, err = d.redisCache.Do(c, "SETEX", key, d.FiveMinutesExpire, bs); err != nil {
		log.Errorc(c, "SetTaskDetail conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.FiveMinutesExpire, string(bs), err)
		return
	}
	return
}
