package dao

import (
	"context"
	"fmt"

	"go-common/library/log"
)

const (
	_timeLock = "act_vogue_%d"
)

func TimeLockKey(mid int64) string {
	return fmt.Sprintf(_timeLock, mid)
}

func (d *Dao) AddTimeLock(c context.Context, mid int64) (ok bool, err error) {
	var reply interface{}
	conn := d.redis.Get(c)
	defer conn.Close()
	key := TimeLockKey(mid)
	if reply, err = conn.Do("SET", key, "lock", "EX", 1, "NX"); err != nil {
		log.Error("rs.lock(%v) error(%v)", key, err)
		return
	} else if reply == nil {
		return
	}
	ok = true
	return
}

func (d *Dao) DelTimeLock(c context.Context, mid int64) (ok bool, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	key := TimeLockKey(mid)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("rs.release(%v) error(%v)", key, err)
		return
	}
	return
}
