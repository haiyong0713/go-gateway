package dao

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	smodel "go-gateway/app/app-svr/playurl/service/model"
)

func (d *Dao) AddOnlineInfo(ctx context.Context, aid int64, onlineInfo *smodel.OnlineInfo) error {
	conn := d.mixRedis.Get(ctx)
	defer conn.Close()
	key := smodel.OnlineKey(aid)
	value, err := json.Marshal(onlineInfo)
	if err != nil {
		return err
	}
	if _, err := conn.Do("SET", key, value, "EX", d.c.Custom.ExTime, "NX"); err != nil {
		if err == redis.ErrNil {
			return nil
		}
		return err
	}
	return nil
}
