package preheat

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/preheat"
)

func (d *Dao) preheatDownloadKey(ID int64) string {
	return fmt.Sprintf("act_preheat_download_%d", ID)
}

func (d *Dao) CacheGetByID(c context.Context, ID int64) (res *preheat.DownInfo, err error) {
	var (
		key  = d.preheatDownloadKey(ID)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheGetByID(%s) return nil", key)
		} else {
			log.Error("CacheGetByID conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (d *Dao) AddCacheGetByID(c context.Context, ID int64, val *preheat.DownInfo) (err error) {
	var (
		key  = d.preheatDownloadKey(ID)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(val); err != nil {
		log.Error("json.Marshal(%v) error (%v)", val, err)
		return
	}
	if err = conn.Send("SETEX", key, 300, bs); err != nil {
		log.Error("AddCacheGetByID conn.Send(SETEX, %s, %v, %s) error(%v)", key, 300, string(bs), err)
	}
	return
}
