// Code generated by kratos tool redisgen. DO NOT EDIT.

/*
  Package family is a generated redis cache package.
  It is generated from:
  type _redis interface {
		// redis: -struct_name=Dao -key=qrcodeKey
		CacheQrcode(ctx context.Context, ticket string) (int64, error)
		// redis: -struct_name=Dao -key=qrcodeKey -expire=d.qrcodeExpire
		AddCacheQrcode(ctx context.Context, ticket string, mid int64) error
		// redis: -struct_name=Dao -key=qrcodeKey
		DelCacheQrcode(ctx context.Context, ticket string) error
		// redis: -struct_name=Dao -key=qrcodeBindKey
		CacheQrcodeBind(ctx context.Context, ticket string) (int64, error)
		// redis: -struct_name=Dao -key=qrcodeBindKey -expire=d.qrcodeStatusExpire
		AddCacheQrcodeBind(ctx context.Context, ticket string, mid int64) error
		// redis: -struct_name=Dao -key=timelockPwdKey
		CacheTimelockPwd(ctx context.Context, mid int64) (string, error)
		// redis: -struct_name=Dao -key=timelockPwdKey -expire=d.timelockPwdExpire
		AddCacheTimelockPwd(ctx context.Context, mid int64, pwd string) error
		// redis: -struct_name=Dao -key=timelockPwdKey
		DelCacheTimelockPwd(ctx context.Context, mid int64) error
	}
*/

package family

import (
	"context"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

var _ _redis

// CacheQrcode get data from redis
func (d *Dao) CacheQrcode(c context.Context, id string) (res int64, err error) {
	key := qrcodeKey(id)
	var temp []byte
	temp, err = redis.Bytes(d.redis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.CacheQrcode(get key: %v) err: %+v", key, err)
		return
	}
	var v string
	v = string(temp)
	r, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Errorc(c, "d.CacheQrcode(get key: %v) err: %+v", key, err)
		return
	}
	res = int64(r)
	return
}

// AddCacheQrcode Set data to redis
func (d *Dao) AddCacheQrcode(c context.Context, id string, val int64) (err error) {
	key := qrcodeKey(id)
	var bs []byte
	bs = []byte(strconv.FormatInt(int64(val), 10))
	expire := d.qrcodeExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheQrcode(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheQrcode delete data from redis
func (d *Dao) DelCacheQrcode(c context.Context, id string) (err error) {
	key := qrcodeKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheQrcode(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheQrcodeBind get data from redis
func (d *Dao) CacheQrcodeBind(c context.Context, id string) (res int64, err error) {
	key := qrcodeBindKey(id)
	var temp []byte
	temp, err = redis.Bytes(d.redis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.CacheQrcodeBind(get key: %v) err: %+v", key, err)
		return
	}
	var v string
	v = string(temp)
	r, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Errorc(c, "d.CacheQrcodeBind(get key: %v) err: %+v", key, err)
		return
	}
	res = int64(r)
	return
}

// AddCacheQrcodeBind Set data to redis
func (d *Dao) AddCacheQrcodeBind(c context.Context, id string, val int64) (err error) {
	key := qrcodeBindKey(id)
	var bs []byte
	bs = []byte(strconv.FormatInt(int64(val), 10))
	expire := d.qrcodeStatusExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheQrcodeBind(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheTimelockPwd get data from redis
func (d *Dao) CacheTimelockPwd(c context.Context, id int64) (res string, err error) {
	key := timelockPwdKey(id)
	var temp []byte
	temp, err = redis.Bytes(d.redis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.CacheTimelockPwd(get key: %v) err: %+v", key, err)
		return
	}
	res = string(temp)
	return
}

// AddCacheTimelockPwd Set data to redis
func (d *Dao) AddCacheTimelockPwd(c context.Context, id int64, val string) (err error) {
	if len(val) == 0 {
		return
	}
	key := timelockPwdKey(id)
	var bs []byte
	bs = []byte(val)
	expire := d.timelockPwdExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheTimelockPwd(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheTimelockPwd delete data from redis
func (d *Dao) DelCacheTimelockPwd(c context.Context, id int64) (err error) {
	key := timelockPwdKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheTimelockPwd(get key: %v) err: %+v", key, err)
		return
	}
	return
}
