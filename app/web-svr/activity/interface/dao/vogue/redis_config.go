package dao

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	_vogueConfigList = "vogue_config_%s"
)

func keyVogueConfigList(name string) string {
	return fmt.Sprintf(_vogueConfigList, name)
}

func (d *Dao) CacheConfig(c context.Context, name string) (res string, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keyVogueConfigList(name)
	if res, err = redis.String(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET) key:%v error:%v", key, err)
		return
	}
	return
}

func (d *Dao) AddCacheConfig(c context.Context, name string, value string) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := keyVogueConfigList(name)
	if _, err = conn.Do("SET", key, value, "EX", d.confExpire); err != nil {
		log.Error("d.AddCacheConfig(%+v) error(%v)", key, err)
	}
	return
}
