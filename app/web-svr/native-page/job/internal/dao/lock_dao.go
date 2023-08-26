package dao

import (
	"context"
	"fmt"
	"go-common/library/cache/credis"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	DelLockScript = `
		if redis.call('exists', KEYS[1]) == 0 then
			return 0
		elseif redis.call('get', KEYS[1]) == ARGV[1] then
			return redis.call('del', KEYS[1])
		else 
			return -1
		end
	`
)

type lockDao struct {
	redis credis.Redis
}

func newLockDao(r credis.Redis) *lockDao {
	return &lockDao{redis: r}
}

// return lockID, locked, error
func (d *lockDao) Lock(c context.Context, key string, expire int64) (string, bool, error) {
	lockKey := lockKey(key)
	lockID := lockIdentify()
	_, err := d.redis.Do(c, "SET", lockKey, lockID, "EX", expire, "NX")
	if err == redis.ErrNil {
		return "", false, nil
	}
	if err != nil {
		log.Errorc(c, "Fail to get lock, key=%+v id=%+v error=%+v", lockKey, lockID, err)
		return "", false, err
	}
	return lockID, true, nil
}

func (d *lockDao) Unlock(c context.Context, key, id string) error {
	lockKey := lockKey(key)
	conn := d.redis.Conn(c)
	defer conn.Close()
	script := redis.NewScript(1, DelLockScript)
	// 0: key不存在；1 删除成功；-1 key已被占用；
	if err := script.Send(conn, lockKey, id); err != nil {
		log.Error("Fail to send script, key=%+v id=%+v error=%+v", lockKey, id, err)
		return err
	}
	if err := conn.Flush(); err != nil {
		log.Error("Fail to flush script, key=%+v id=%+v error=%+v", lockKey, id, err)
		return err
	}
	rly, err := redis.Int(conn.Receive())
	if err != nil {
		log.Errorc(c, "Fail to unlock, key=%+v id=%+v error=%+v", lockKey, id, err)
		return err
	}
	if rly == -1 {
		log.Errorc(c, "Fail to unlock, lock has been taken, key=%+v id=%+v", lockKey, id)
		return errors.New("lock has been taken")
	}
	return nil
}

func lockIdentify() string {
	return uuid.New().String()
}

func lockKey(key string) string {
	return fmt.Sprintf("natpage_lock_%s", key)
}
