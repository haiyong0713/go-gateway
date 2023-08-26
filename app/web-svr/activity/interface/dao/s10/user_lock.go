package s10

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
)

const userLock = "s10:sl:%d"

func userLockKey(mid int64) string {
	return fmt.Sprintf(userLock, mid)
}

func (d *Dao) UserLock(ctx context.Context, mid int64) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ulkey := userLockKey(mid)
	reply, err := conn.Do("SET", ulkey, 1, "EX", 2, "NX")
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", ulkey, err)
		return err
	}
	if reply == nil {
		return fmt.Errorf("已经存在")
	}
	return nil
}

func (d *Dao) UserUnlock(ctx context.Context, mid int64) error {
	conn := component.S10PointShopRedis.Get(ctx)
	defer conn.Close()
	ulkey := userLockKey(mid)
	_, err := conn.Do("DEL", ulkey)
	if err != nil {
		log.Errorc(ctx, "s10 conn.Do(key:%s) error(%v)", ulkey, err)
		return err
	}
	return nil
}
