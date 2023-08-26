package show

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"strconv"
)

// CustomConfigs is
func (d *Dao) GetBWListItemFromRedis(ctx context.Context, key string) (itemValue int64, err error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	var val string

	if val, err = redis.String(conn.Do("GET", key)); err != nil {
		if err != redis.ErrNil {
			log.Error("GetBWListItemFromRedis Get fail: %+v, key: %+v", err, key)
			return -1, err
		}
		return -1, nil
	}
	if itemValue, err = strconv.ParseInt(val, 10, 64); err != nil {
		log.Error("GetBWListItemFromRedis ParseInt fail: %+v, data: %+v", err, val)
	}

	return itemValue, err
}
