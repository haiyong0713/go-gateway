package show

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

func (d *Dao) GetEntryPubTaskLock(c context.Context, id int32) bool {
	var (
		key     = fmt.Sprintf("app-entry-pub-lock-%d", id)
		conn    = d.rds.Get(c)
		timeout = 30
		reply   string
		err     error
	)
	defer conn.Close()
	if reply, err = redis.String(conn.Do("SET", key, "1", "EX", timeout, "NX")); err != nil || reply != "OK" {
		log.Error("[app-entry]GetEntryPubTaskLock conn.Do(SETNX, %s) error(%v)", key, err)
		return false
	}
	return true
}
