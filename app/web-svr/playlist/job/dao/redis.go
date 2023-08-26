package dao

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	// view
	_viewPrefix = "v_%d_%d_%s"
)

func viewKey(pid, aid int64, ip string) (key string) {
	if ip == "" {
		// let it pass if ip is empty.
		return
	}
	return fmt.Sprintf(_viewPrefix, pid, aid, ip)
}

// Intercept intercepts illegal views.
func (d *Dao) Intercept(c context.Context, pid, aid int64, ip string) (ban bool) {
	var (
		err   error
		exist bool
		key   = viewKey(pid, aid, ip)
		conn  = d.redis.Get(c)
	)
	defer conn.Close()
	if key == "" {
		return
	}
	if exist, err = redis.Bool(conn.Do("EXISTS", key)); err != nil {
		log.Error("conn.Do(EXISTS, %s) error(%v)", key, err)
		return
	}
	if exist {
		ban = true
		return
	}
	if err = conn.Send("SET", key, ""); err != nil {
		log.Error("conn.Send(EXPIRE, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.viewCacheTTL); err != nil {
		log.Error("conn.Send(EXPIRE, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}
