package system

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
)

// 存储AccessToken到Redis
func (d *Dao) StoreWXAccessTokenInRedis(ctx context.Context, accessToken string, from string) (err error) {
	if accessToken == "" {
		return
	}
	var (
		key = WXAccessTokenKey(from)
	)
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, 3600, accessToken); err != nil {
		log.Errorc(ctx, "conn.Do(SETEX, %s, %v, %+v) error(%v)", key, 3600, accessToken, err)
		return
	}
	return
}

// Redis中获取AccessToken
func (d *Dao) GetWXAccessTokenFromRedis(ctx context.Context, from string) (accessToken string, err error) {
	var (
		key = WXAccessTokenKey(from)
	)
	if accessToken, err = redis.String(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(ctx, "conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	return
}

// 存储JSAPITicket到Redis
func (d *Dao) StoreWXJSAPITicketInRedis(ctx context.Context, JSAPITicket string, from string) (err error) {
	if JSAPITicket == "" {
		return
	}
	var (
		key = WXJSAPITicketKey(from)
	)
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, 3600, JSAPITicket); err != nil {
		log.Errorc(ctx, "conn.Do(SETEX, %s, %v, %+v) error(%v)", key, 3600, JSAPITicket, err)
		return
	}
	return
}

// Redis中获取JSAPITicket
func (d *Dao) GetWXJSAPITicketFromRedis(ctx context.Context, from string) (JSAPITicket string, err error) {
	var (
		key = WXJSAPITicketKey(from)
	)
	if JSAPITicket, err = redis.String(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(ctx, "conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	return
}

// Redis中获取AccessToken
func (d *Dao) GetOAAccessTokenFromRedis(ctx context.Context) (accessToken string, err error) {
	var (
		key = OAAccessTokenKey()
	)
	if accessToken, err = redis.String(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(ctx, "conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	return
}

// 存储AccessToken到Redis
func (d *Dao) StoreOAAccessTokenInRedis(ctx context.Context, accessToken string) (err error) {
	if accessToken == "" {
		return
	}
	var (
		key = OAAccessTokenKey()
	)
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, 3600, accessToken); err != nil {
		log.Errorc(ctx, "conn.Do(SETEX, %s, %v, %+v) error(%v)", key, 3600, accessToken, err)
		return
	}
	return
}
