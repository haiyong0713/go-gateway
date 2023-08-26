package thirdsdk

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"

	"github.com/pkg/errors"
)

func userWhitelistKey(mid int64) string {
	return fmt.Sprintf("USER_WHITELIST_%d", mid)
}

func (d *Dao) Invited(ctx context.Context, mid int64) (bool, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := userWhitelistKey(mid)
	reply, err := redis.Strings(conn.Do("SMEMBERS", key))
	if err != nil {
		return false, errors.WithStack(err)
	}
	if len(reply) != 0 {
		return true, nil
	}
	return false, nil
}
