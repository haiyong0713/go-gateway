package lottery

import (
	"context"
	"go-common/library/log"
)

// DeleteMemberGroup ...
func (d *Dao) DeleteMemberGroup(c context.Context, sid string) (err error) {
	var (
		key  = buildKey(memberGroupKey, sid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DeleteMemberGroup conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}
