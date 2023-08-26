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
	_keyArchiveFmt = "va_%d"
	_prefixNx      = "nx_"
)

func keyArchive(aid int64) string {
	return fmt.Sprintf(_keyArchiveFmt, aid)
}

func nxKey(key string) string {
	return _prefixNx + key
}

// SetViewBakCache  set view archive page data to cache.
func (d *Dao) SetViewBakCache(c context.Context, aid int64, a *model.View) (err error) {
	key := keyArchive(aid)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	var bs []byte
	if bs, err = json.Marshal(a); err != nil {
		log.Error("SetViewBakCache json.Marshal(%v) error(%v)", a, err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("SetViewBakCache conn.Send(SET,%s,%s) error(%v)", key, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.redisArchiveBakExpire); err != nil {
		log.Error("SetViewBakCache conn.Send(EXPIRE,%s,%d) error(%v)", key, d.redisArchiveBakExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("SetViewBakCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("SetViewBakCache conn.Recevie(%d) error(%v0", i, err)
		}
	}
	return
}

// ViewBakCache get view archive  page data from cache.
func (d *Dao) ViewBakCache(c context.Context, aid int64) (rs *model.View, err error) {
	key := keyArchive(aid)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	var values []byte
	if values, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("ViewBakCache redis (%s) return nil ", key)
		} else {
			log.Error("ViewBakCache conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(values, &rs); err != nil {
		log.Error("ViewBakCache json.Unmarshal(%v) error(%v)", values, err)
	}
	return
}

// RsSetNX Dao
func (d *Dao) RsSetNX(c context.Context, key string, expire int32) (res bool, err error) {
	var (
		rkey = nxKey(key)
		conn = d.redisBak.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Bool(conn.Do("SETNX", rkey, "1")); err != nil {
		if err == redis.ErrNil {
			log.Error("conn.Do(GET key(%s)) error(%v)", rkey, err)
			err = nil
		} else {
			log.Error("conn.Do(GET key(%s)) error(%v)", rkey, err)
			return
		}
	}
	if expire > 0 {
		if _, err = conn.Do("EXPIRE", rkey, expire); err != nil {
			log.Error("conn.Do(EXPIRE,%s) err(%+v)", rkey, err)
			return
		}
	}
	return
}
