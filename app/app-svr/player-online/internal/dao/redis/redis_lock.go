package redis

import (
	"context"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	SetLockSuccess     = "OK" // 操作成功
	DelLockSuccess     = 1    // lock 删除成功
	DelLockNonExistent = 0    // 删除lock key时,并不存在
)

// TryLock 尝试获取分布式锁
func (d *Dao) TryLock(c context.Context, key string, value string, timeout int32) (bool, error) {
	value, err := redis.String(d.redis.Do(c, "SET", key, value, "EX", timeout, "NX"))
	log.Error("TryLock key(%+v) value(%+v) err(%+v)", key, value, err)
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		return false, err
	}
	if value == SetLockSuccess {
		return true, nil
	}
	return false, nil
}

// UnLock 解锁
func (d *Dao) UnLock(c context.Context, key string, value string) bool {
	log.Error("UnLock key(%+v) value(%+v)", key, value)
	if d.GetLock(c, key) == value {
		msg, _ := redis.Int64(d.redis.Do(c, "DEL", key))
		if msg == DelLockSuccess || msg == DelLockNonExistent {
			return true
		}
	}
	return false
}

func (d *Dao) GetLock(c context.Context, key string) string {
	msg, err := redis.String(d.redis.Do(c, "GET", key))
	log.Error("GetLock key(%+v) value(%+v) err(%+v)", key, msg, err)
	return msg
}
