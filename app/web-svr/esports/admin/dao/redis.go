package dao

import (
	"context"
	"fmt"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/admin/component"
	"time"
)

var (
	_contestTeamsUpdateLockKey    = "esports-admin:contest:%d:teams:update:lock"
	_contestTeamsUpdateLockKeyTtl = int64(60)
)

func (d *Dao) RedisLock(ctx context.Context, key string, value string, ttlSecond int64, retryTimes int, sleepMs int64) (err error) {
	conn := component.GlobalAutoSubCache.Get(ctx)
	defer conn.Close()
	var reply interface{}
	for {
		reply, err = conn.Do("SET", key, value, "EX", ttlSecond, "NX")
		if "OK" == reply {
			return
		}
		if err != nil {
			log.Errorc(ctx, "[Dao][Redis][Lock][Error], err:(%+v)", err)
		}
		retryTimes--
		if retryTimes == 0 {
			break
		}
		time.Sleep(time.Duration(sleepMs) * time.Microsecond)
	}
	if err != nil {
		return
	}
	if reply == nil {
		err = xecode.Errorf(xecode.RequestErr, "setnx do failed")
	}
	return
}

func (d *Dao) RedisUnLock(ctx context.Context, key string, value string) (err error) {
	conn := component.GlobalAutoSubCache.Get(ctx)
	defer conn.Close()
	reply, err := conn.Do("DEL", key)
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][Unlock][Error], err:(%+v), reply:(%+v)", err, reply)
	}
	return
}

func (d *Dao) GetContestTeamsUpdateLockInfo(contestId int64) (key string, value string, ttl int64) {
	key = fmt.Sprintf(_contestTeamsUpdateLockKey, contestId)
	value = "1"
	ttl = _contestTeamsUpdateLockKeyTtl
	return
}
