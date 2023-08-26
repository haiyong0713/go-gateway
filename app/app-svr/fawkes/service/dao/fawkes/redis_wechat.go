package fawkes

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
)

const (
	_wechatKey = "FAWKES:WECHATTOKEN:%v"
)

func (d *Dao) GetCacheWechatToken(c context.Context, appSecret string) (value string, err error) {
	key := fmt.Sprintf(_wechatKey, appSecret)
	bs, err := redis.Bytes(d.redis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		return
	}
	value = string(bs)
	return
}

func (d *Dao) SetCacheWechatToken(c context.Context, appSecret, token string, expireTime int64) (err error) {
	key := fmt.Sprintf(_wechatKey, appSecret)
	_, err = d.redis.Do(c, "SETEX", key, expireTime, token)
	return
}
