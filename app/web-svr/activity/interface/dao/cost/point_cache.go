package cost

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

func userCostKey(mid int64, sid string) string {
	return fmt.Sprintf(userCostRedisKey, sid, mid)
}

// CacheGetUserCostPoint
func (d *dao) CacheGetUserCostPoint(ctx context.Context, mid int64, activityId string) (points int64, err error) {
	var (
		key = buildKey(userCostKey(mid, activityId))
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redis.Do(ctx, "GET", key)); err != nil {
		log.Errorc(ctx, "CacheGetUserCostPoint conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &points); err != nil {
		log.Errorc(ctx, "CacheGetUserCostPoint json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return

}

// CacheSetUserCostPoint
func (d *dao) CacheSetUserCostPoint(ctx context.Context, mid int64, activityId string, data int64) (err error) {
	var (
		key = buildKey(userCostKey(mid, activityId))
		bs  []byte
	)
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(ctx, "CacheSetUserCostPoint json.Marshal(%v) error (%v)", data, err)
		return
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.userCostExpire, bs); err != nil {
		log.Errorc(ctx, "CacheSetUserCostPoint conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.userCostExpire, string(bs), err)
	}
	return
}

// CacheDelUserCostPoint 删除缓存
func (d *dao) CacheDelUserCostPoint(ctx context.Context, mid int64, activityId string) (err error) {
	key := buildKey(userCostKey(mid, activityId))
	if _, err = d.redis.Do(ctx, "DEL", key); err != nil {
		log.Errorc(ctx, "CacheDelUserCostPoint conn.Do(DEL, %s) error(%v)", key, err)
		return
	}
	return
}

func userExchangeKey(orderId string) string {
	return fmt.Sprintf(userExchangeFlagKey, orderId)
}

// CacheGetUserExchangeFlag
func (d *dao) CacheGetUserExchangeFlag(ctx context.Context, orderId string) (flag int, err error) {
	var (
		key = buildKey(userExchangeKey(orderId))
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redis.Do(ctx, "GET", key)); err != nil {
		log.Errorc(ctx, "CacheGetUserExchangeFlag conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &flag); err != nil {
		log.Errorc(ctx, "CacheGetUserExchangeFlag json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return

}

// CacheSetUserExchangeFlag
func (d *dao) CacheSetUserExchangeFlag(ctx context.Context, orderId string, flag int) (err error) {
	var (
		key = buildKey(userExchangeKey(orderId))
		bs  []byte
	)
	if bs, err = json.Marshal(flag); err != nil {
		log.Errorc(ctx, "CacheSetUserExchangeFlag json.Marshal(%v) error (%v)", flag, err)
		return
	}
	if _, err = d.redis.Do(ctx, "SETEX", key, d.userExangeExpire, bs); err != nil {
		log.Errorc(ctx, "CacheSetUserExchangeFlag conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.userExangeExpire, string(bs), err)
	}
	return
}
