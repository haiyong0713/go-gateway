package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/web-svr/web/interface/model"
)

const (
	_keyIndex        = "index_set:%d"
	_keyIndexVersion = "index_set:%d_%d"
)

func keyIndexSort(uid, version int64) string {
	if version > 0 {
		return fmt.Sprintf(_keyIndexVersion, uid, version)
	}
	return fmt.Sprintf(_keyIndex, uid)
}

// IndexCache get index setting cache.
func (d *Dao) IndexSortCache(c context.Context, uid, version int64) (set string, err error) {
	var (
		conn = d.RedisIndex.Get(c)
		key  = keyIndexSort(uid, version)
	)
	defer conn.Close()
	if set, err = redis.String(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("IndexSortCache conn.Do(GET, %s) error(%v)", key, err)
		}
	}
	return
}

func (d *Dao) SetIndexSortCache(c context.Context, uid int64, set model.IndexSet, version int64) (err error) {
	var (
		key  = keyIndexSort(uid, version)
		conn = d.RedisIndex.Get(c)
	)
	defer conn.Close()
	var bs []byte
	if bs, err = json.Marshal(set); err != nil {
		log.Error("json.Marshal(%v) error(%v)", set, err)
		return
	}
	if _, err = conn.Do("SET", key, bs); err != nil {
		log.Error("SetIndexSortCache add conn.Do SET error(%v)", err)
		return
	}
	return
}
