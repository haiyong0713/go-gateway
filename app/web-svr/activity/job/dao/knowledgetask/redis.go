package knowledgetask

import (
	"context"

	xecode "go-common/library/ecode"
	"go-common/library/log"
)

// RsSetNX .
func (d *Dao) RsSetNX(ctx context.Context, key string, expire int32) (err error) {
	var reply interface{}
	if reply, err = d.redis.Do(ctx, "SET", key, "LOCK", "EX", expire, "NX"); err != nil {
		log.Errorc(ctx, "RsSetNX(%v) error(%v)", key, err)
		return
	}
	if reply == nil {
		err = xecode.AccessDenied
	}
	return
}
