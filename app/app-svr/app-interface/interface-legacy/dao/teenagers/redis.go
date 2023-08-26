package teenagers

import (
	"context"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	_prefixAttention = "space_att_%s_%d"
)

func keyAttention(day string, mid int64) string {
	return fmt.Sprintf(_prefixAttention, day, mid)
}

// CacheAttention .
func (d *Dao) CacheAttention(ctx context.Context, mid int64) (num int, err error) {
	day := time.Now().Format("2006-01-02")
	key := keyAttention(day, mid)
	conn := d.attRedis.Get(ctx)
	defer conn.Close()
	if num, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		}
	}
	return
}

// IncrCacheAttention .
func (d *Dao) IncrCacheAttention(ctx context.Context, mid int64) (err error) {
	day := time.Now().Format("2006-01-02")
	key := keyAttention(day, mid)
	conn := d.attRedis.Get(ctx)
	defer conn.Close()
	if err = conn.Send("INCR", key); err != nil {
		log.Error("IncrCacheAttention(INCR %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, 86400); err != nil {
		log.Error("IncrCacheAttention EXPIRE %s error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("IncrCacheAttention conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("IncrCacheAttention conn.Receive() error(%v)", err)
			return
		}
	}
	return
}
