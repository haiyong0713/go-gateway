package wechat

import (
	"context"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	_weChatKey = "wechat_auth"
)

func (d *Dao) AddWeChatCache(c context.Context, wechatTicket string) error {
	conn := d.redis.Conn(c)
	defer conn.Close()
	if _, err := conn.Do("SETEX", _weChatKey, d.keyExpired, wechatTicket); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

func (d *Dao) WeChatCache(c context.Context) (string, error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	wechatTicket, err := redis.String(conn.Do("GET", _weChatKey))
	if err != nil {
		log.Error("%+v", err)
		return "", err
	}
	return wechatTicket, nil
}
