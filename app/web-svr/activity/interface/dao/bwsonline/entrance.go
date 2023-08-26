package bwsonline

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	EntranceStatusOpened = iota + 1
	EntranceStatusClosed
	EntranceStatusUnknown
)

func userEntranceKey(mid int64) string {
	return fmt.Sprintf("bws_user_entrance_%d", mid)
}

func (d *Dao) CacheUserEntrance(ctx context.Context, mid int64) (int, error) {
	key := userEntranceKey(mid)
	status, err := redis.Int(d.redis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return EntranceStatusUnknown, nil
		}
		log.Errorc(ctx, "CacheUserEntrance key[%s] err[%v]", key, err)
		return EntranceStatusClosed, err
	}
	return status, nil
}

func (d *Dao) AddCacheUserEntrance(ctx context.Context, mid int64, status int) error {
	key := userEntranceKey(mid)
	// 开关打开后不会关闭，所以打开情况缓存有效期长一些，其他情况10分钟失效，及时更新
	var expired int = 40 * 86400
	if status == EntranceStatusClosed {
		expired = 600
	}
	_, err := d.redis.Do(ctx, "SETEX", key, expired, status)
	if err != nil {
		log.Errorc(ctx, "AddCacheUserEntrance key[%s] status[%d] expired[%d] err[%v]", key, status, expired, err)
	}
	return err
}
