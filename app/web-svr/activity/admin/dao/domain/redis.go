package domain

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	mdomain "go-gateway/app/web-svr/activity/admin/model/domain"
	"time"
)

const domainList = "list"

// cacheDomain ...
func (d *Dao) CacheDomain(ctx context.Context, value *mdomain.Record) (err error) {
	var (
		effect    interface{}
		key       = buildKey(domainList)
		conn      = d.redis.Get(ctx)
		byteValue []byte
	)
	defer conn.Close()

	if value.Etime.Time().Before(time.Now()) {
		effect, err = conn.Do("HDEL", key, value.Id)
		log.Infoc(ctx, "cacheDomain conn.Do.HDEL, key:%s , field:%v , effect:%v , error:%v", key, value.Id, effect, err)
		return
	}
	if byteValue, err = json.Marshal(*value); err == nil {
		effect, err = conn.Do("HSET", key, value.Id, byteValue)
	}
	log.Infoc(ctx, "cacheDomain conn.Do.HSET, key:%s , field:%v , effect:%v , error:%v", key, value.Id, effect, err)
	return
}
