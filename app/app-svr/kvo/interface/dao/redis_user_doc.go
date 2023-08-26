package dao

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	_rdsuserDocKeyPrefix = "dm_u_doc"
)

func keyUserDoc(mid int64, moduleKeyId int, buvid string) string {
	if mid == 0 {
		return fmt.Sprintf("%s_%d_%d_%s", _rdsuserDocKeyPrefix, mid, moduleKeyId, buvid)
	}
	return fmt.Sprintf("%s_%d_%d", _rdsuserDocKeyPrefix, mid, moduleKeyId)
}

// DelUserDoc .
func (d *Dao) DelUserDoc(c context.Context, mid int64, buvid string, moduleKeyId int) (ok bool, err error) {
	key := keyUserDoc(mid, moduleKeyId, buvid)
	ok, err = redis.Bool(d.rds.Do(c, "DEL", key))
	if err != nil {
		log.Error("d.DelUserDoc(mid:%d,moduleKeyId:%d) err(%v)", mid, moduleKeyId, err)
	}
	return
}

// HMsetUserDoc .
func (d *Dao) HMsetUserDoc(c context.Context, mid int64, buvid string, moduleKeyId int, m map[string]string) (err error) {
	key := keyUserDoc(mid, moduleKeyId, buvid)
	args := redis.Args{}.Add(key)
	for k, v := range m {
		args = args.Add(k).Add(v)
	}
	p := d.rds.Pipeline()
	p.Send("HMSET", args...)
	p.Send("EXPIRE", key, d.rdsIncrExpire)
	replies, nerr := p.Exec(c)
	if nerr != nil {
		err = nerr
		return
	}
	for replies.Next() {
		if _, nerr := replies.Scan(); nerr != nil {
			err = nerr
			return
		}
	}
	return
}

// HgetAllUserDoc .
func (d *Dao) HgetAllUserDoc(c context.Context, mid int64, buvid string, moduleKeyId int) (res map[string]string, err error) {
	key := keyUserDoc(mid, moduleKeyId, buvid)
	if res, err = redis.StringMap(d.rds.Do(c, "HGETALL", key)); err != nil {
		return
	}
	return
}
