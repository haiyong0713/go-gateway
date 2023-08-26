package invite

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	mdl "go-gateway/app/web-svr/activity/interface/model/invite"

	"go-common/library/log"
)

const (
	prefix    = "act_invite"
	separator = ":"
	_keyToken = "tk_%s"
	midToken  = "midToken"
)

func tokenKey(token string) string {
	return fmt.Sprintf(_keyToken, token)
}

const _keyUserShareLog = "usl_%d_%s" // mid activity

func userShareLogKey(mid int64, activityUID string) string {
	return fmt.Sprintf(_keyUserShareLog, mid, activityUID)
}

// AddCacheToken.
func (d *dao) AddCacheToken(ctx context.Context, token string, data *mdl.FiToken) (err error) {
	var (
		bs   []byte
		key  = tokenKey(token)
		conn = d.redis.Get(ctx)
	)
	defer conn.Close()

	if bs, err = json.Marshal(data); err != nil {
		log.Error("AddCacheToken json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SETEX", key, d.actExpire, bs); err != nil {
		log.Errorc(ctx, "AddCacheToken conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.actExpire, bs, err)
	}

	return
}

// CacheMidToken ...
func (d *dao) CacheMidToken(c context.Context, mid, tp int64, activityUID, token string, source int64) (err error) {
	var (
		key  = buildKey(midToken, mid, tp, activityUID, source)
		conn = d.redis.Get(c)
	)
	defer conn.Close()

	if err = conn.Send("SETEX", key, d.tokenExpire, token); err != nil {
		log.Errorc(c, "CacheMidToken conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.tokenExpire, token, err)
	}
	return
}

// CacheGetMidToken ...
func (d *dao) CacheGetMidToken(c context.Context, mid, tp int64, activityUID string, source int64) (res string, err error) {
	var (
		key  = buildKey(midToken, mid, tp, activityUID, source)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		log.Error("conn.Do(GET,%s) error(%v)", key, err)
		return
	}
	res = string(bs)
	return
}

// ClearUserShareLogCache 获取获奖总人数
func (d *dao) ClearUserShareLogCache(c context.Context, mid int64, activityUID string) (err error) {
	var (
		key  = userShareLogKey(mid, activityUID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key, activityUID); err != nil {
		log.Errorc(c, "ClearUserShareLogCache conn.Do(SET) key(%s) error(%v)", key, err)
		return
	}
	return
}

// UserShareLogCache 获得用户获奖情况
func (d *dao) UserShareLogCache(c context.Context, mid int64, activityUID string) (res *mdl.UserShareLog, err error) {
	var (
		bs  []byte
		key = userShareLogKey(mid, activityUID)

		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	res = new(mdl.UserShareLog)
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheToken.
func (d *dao) AddUserShareLogCache(ctx context.Context, mid int64, activityUID string, data *mdl.UserShareLog) (err error) {
	var (
		bs   []byte
		key  = userShareLogKey(mid, activityUID)
		conn = d.redis.Get(ctx)
	)
	defer conn.Close()

	if bs, err = json.Marshal(data); err != nil {
		log.Error("AddUserShareLogCache json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SETEX", key, d.actExpire, bs); err != nil {
		log.Errorc(ctx, "AddUserShareLogCache conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.actExpire, bs, err)
	}

	return
}

// CacheToken 获得用户信息
func (d *dao) CacheToken(ctx context.Context, token string) (res *mdl.FiToken, err error) {
	var (
		bs   []byte
		key  = tokenKey(token)
		conn = d.redis.Get(ctx)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	res = new(mdl.FiToken)
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

const userOldExpire = 24 * 3600
const _keyUserOld = "oldu_%s" // tel hash

func userOldKey(telHash string) string {
	return fmt.Sprintf(_keyUserOld, telHash)
}

// AddCacheToken.
func (d *dao) AddOldUserCache(ctx context.Context, telHash string) (err error) {
	var (
		key  = userOldKey(telHash)
		conn = d.redis.Get(ctx)
	)
	defer conn.Close()

	if err = conn.Send("SETEX", key, userOldExpire, "1"); err != nil {
		log.Errorc(ctx, "AddOldUserCache conn.Send(SETEX, %s, %v, %d) error(%v)", key, userOldExpire, 1, err)
	}
	return
}

// GetMidBindInviter 获取用户的邀请人信息
func (d *dao) GetMidBindInviter(c context.Context, telHash string, activityUID string) (res int64, err error) {
	var (
		key  = buildKey(inviterKey, activityUID, telHash)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int64(conn.Do("GET", key)); err != nil {
		log.Error("conn.Do(GET,%s) error(%v)", key, err)
		return
	}
	return
}

// SetMidBindInviter 设置用户的邀请人信息
func (d *dao) SetMidBindInviter(c context.Context, telHash string, inviter int64, activityUID string) (err error) {
	var (
		key  = buildKey(inviterKey, activityUID, telHash)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("SETEX", key, d.bindExpire, inviter); err != nil {
		log.Errorc(c, "SetMidBindInviter conn.Do(SET) key(%s) error(%v)", key, err)
		return
	}
	return
}
