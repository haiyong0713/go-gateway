package family

import (
	"context"
	"fmt"

	"go-common/library/log"
)

func (d *Dao) Lock(ctx context.Context, key string) error {
	lk := lockKey(key)
	if _, err := d.redis.Do(ctx, "SET", lk, true, "EX", d.lockExpire, "NX"); err != nil {
		log.Error("Fail to lock, key=%+v error=%+v", lk, err)
		return err
	}
	return nil
}

func (d *Dao) Unlock(ctx context.Context, key string) error {
	lk := lockKey(key)
	if _, err := d.redis.Do(ctx, "DEL", lk); err != nil {
		log.Error("Fail to unlock, key=%+v err=%+v", lk, err)
		return err
	}
	return nil
}

func lockKey(key string) string {
	return fmt.Sprintf("family_lock_%s", key)
}
