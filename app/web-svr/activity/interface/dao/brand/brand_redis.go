package brand

import (
	"context"
	"time"

	"go-common/library/log"
)

const (
	midKey = "mid"
	vipKey = "vip"
)

// CacheAddCouponTimes cache add mid coupon times
func (d *dao) CacheAddCouponTimes(c context.Context, mid int64) (couponNum int64, err error) {
	var (
		key  = buildKey(midKey, mid)
		conn = d.redis.Get(c)
		res  interface{}
	)
	defer conn.Close()

	if res, err = conn.Do("INCR", key); err != nil {
		log.Error("CacheSetPlaylist conn.Send(INCR, %s) error(%v)", key, err)
		return
	}
	return res.(int64), nil
}

// CacheSetMinusCouponTimes cache minus coupon times
func (d *dao) CacheSetMinusCouponTimes(c context.Context, mid int64) (err error) {
	var (
		key  = buildKey(midKey, mid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()

	if _, err = conn.Do("DECR", key); err != nil {
		log.Error("CacheSetPlaylist conn.Send(DECR, %s) error(%v)", key, err)
	}
	return err
}

// CacheQPSLimit cache qps limit
func (d *dao) CacheQPSLimit(c context.Context, typeName string) (num int64, err error) {
	timestamp := time.Now().Unix()
	var (
		key  = buildKey(vipKey, typeName, timestamp)
		conn = d.redis.Get(c)
		res  interface{}
	)
	defer conn.Close()

	if res, err = conn.Do("INCR", key); err != nil {
		log.Error("CacheQpsLimit conn.Send(INCR, %s) error(%v)", key, err)
		return
	}
	num = res.(int64)
	if num == int64(1) {
		if res, err = conn.Do("EXPIRE", key, d.qpsLimitExpire); err != nil {
			log.Error("CacheQpsLimit conn.Send(INCEXPIRER, %s) error(%v)", key, err)
		}
	}
	return num, nil
}
