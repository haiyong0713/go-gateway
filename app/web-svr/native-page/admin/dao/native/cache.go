package native

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
)

func ntTitleUniqueKey(title string) string {
	return fmt.Sprintf("nat_ts_tinu_%s", title)
}

// NtTsTitleUnique 与interface中up主发起活动共用一个key .
func (d *Dao) NtTsTitleUnique(c context.Context, title string) (bool, error) {
	var (
		key = ntTitleUniqueKey(title)
	)
	// 2s 过期
	return d.setNXLockCache(c, key, 2)
}

func (d *Dao) setNXLockCache(c context.Context, key string, expire int32) (bool, error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	reply, err := redis.String(conn.Do("SET", key, 1, "EX", expire, "NX"))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		return false, err
	}
	if reply != "OK" {
		return false, nil
	}
	return true, nil
}
