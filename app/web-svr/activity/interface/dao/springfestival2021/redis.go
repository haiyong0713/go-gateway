package springfestival2021

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	springfestival2021 "go-gateway/app/web-svr/activity/interface/model/springfestival2021"
)

// MidCardDetail 用户集卡情况
func (d *Dao) MidCardDetail(c context.Context, mid int64) (res *springfestival2021.MidNums, err error) {
	var (
		key = buildKey(midCard, mid)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisStore.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "MidCardDetail conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// MidLimit cache qps limit
func (d *Dao) MidLimit(c context.Context, mid int64) (num int64, err error) {
	timestamp := time.Now().Unix()
	var (
		key = buildKey(qpsLimitKey, mid, timestamp)
		res interface{}
	)

	if res, err = d.redisCache.Do(c, "INCR", key); err != nil {
		log.Errorc(c, "CacheQpsLimit conn.Send(INCR, %s) error(%v)", key, err)
		return
	}
	num = res.(int64)
	if num == int64(1) {
		if res, err = d.redisCache.Do(c, "EXPIRE", key, d.qpsLimitExpire); err != nil {
			log.Errorc(c, "CacheQpsLimit conn.Send(INCEXPIRER, %s) error(%v)", key, err)
		}
	}
	return num, nil
}

// AddMidCardDetail add cache lottery
func (d *Dao) AddMidCardDetail(c context.Context, mid int64, val *springfestival2021.MidNums) (err error) {
	var (
		key = buildKey(midCard, mid)
		bs  []byte
	)
	if bs, err = json.Marshal(val); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", val, err)
		return
	}
	if _, err = d.redisStore.Do(c, "SETEX", key, d.MidCardExpire, bs); err != nil {
		log.Errorc(c, "AddMidCardDetail conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.MidCardExpire, string(bs), err)
		return
	}
	return
}

// DeleteMidCardDetail ...
func (d *Dao) DeleteMidCardDetail(c context.Context, mid int64) (err error) {
	var (
		key = buildKey(midCard, mid)
	)
	if _, err = d.redisStore.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteMidCardDetail conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// InviteTokenToMid token转mid
func (d *Dao) InviteTokenToMid(c context.Context, token string) (res int64, err error) {
	var (
		key = buildKey(tokenKey, token)
	)
	if res, err = redis.Int64(d.redisStore.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "InviteTokenToMid conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	return
}

// AddInviteTokenToMid invite token to mid
func (d *Dao) AddInviteTokenToMid(c context.Context, token string, mid int64) (err error) {
	var (
		key = buildKey(tokenKey, token)
	)
	if _, err = d.redisStore.Do(c, "SETEX", key, d.ActivityEndExpire, mid); err != nil {
		log.Errorc(c, "AddMidCardDetail conn.Send(SETEX, %s, %v, %d) error(%v)", key, d.ActivityEndExpire, mid, err)
		return
	}
	return
}

// DeleteInviteTokenToMid ...
func (d *Dao) DeleteInviteTokenToMid(c context.Context, token string) (err error) {
	var (
		key = buildKey(tokenKey, token)
	)
	if _, err = d.redisStore.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteMidCardDetail conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// InviteMidToToken mid to token
func (d *Dao) InviteMidToToken(c context.Context, mid int64) (res string, err error) {
	var (
		key = buildKey(midKey, mid)
	)
	if res, err = redis.String(d.redisStore.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "InviteMidToToken conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	return
}

// AddInviteMidToToken invite token to mid
func (d *Dao) AddInviteMidToToken(c context.Context, mid int64, token string) (err error) {
	var (
		key = buildKey(midKey, mid)
	)
	if _, err = d.redisStore.Do(c, "SETEX", key, d.ActivityEndExpire, token); err != nil {
		log.Errorc(c, "AddMidCardDetail conn.Send(SETEX, %s, %v, %d) error(%v)", key, d.ActivityEndExpire, mid, err)
		return
	}
	return
}

// MidInviter mid to token
func (d *Dao) MidInviter(c context.Context, mid int64) (res int64, err error) {
	var (
		key = buildKey(inviterKey, mid)
	)
	if res, err = redis.Int64(d.redisStore.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "InviteMidToToken conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	return
}

// AddMidInviter invite token to mid
func (d *Dao) AddMidInviter(c context.Context, mid, inviter int64) (err error) {
	var (
		key = buildKey(inviterKey, mid)
	)
	if _, err = d.redisStore.Do(c, "SETEX", key, d.ActivityEndExpire, inviter); err != nil {
		log.Errorc(c, "AddMidCardDetail conn.Send(SETEX, %s, %v, %d) error(%v)", key, d.ActivityEndExpire, inviter, err)
		return
	}
	return
}

// ArchiveNums 投稿数
func (d *Dao) ArchiveNums(c context.Context, mid int64) (res int64, err error) {
	var (
		key = buildKey(midArchive, mid)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisStore.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "MidCardDetail conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// AddArchiveNums add cache lottery
func (d *Dao) AddArchiveNums(c context.Context, mid int64, nums int64) (err error) {
	var (
		key = buildKey(midArchive, mid)
	)
	if _, err = d.redisStore.Do(c, "SETEX", key, d.MidCardExpire, nums); err != nil {
		log.Errorc(c, "AddMidCardDetail conn.Send(SETEX, %s, %v, %d) error(%v)", key, d.MidCardExpire, nums, err)
		return
	}
	return
}

// DeleteArchiveNums ...
func (d *Dao) DeleteArchiveNums(c context.Context, mid int64) (err error) {
	var (
		key = buildKey(midArchive, mid)
	)
	if _, err = d.redisStore.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteMidCardDetail conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// AddShareCardToken add cache lottery
func (d *Dao) AddShareCardToken(c context.Context, token string, val *springfestival2021.CardTokenMid) (err error) {
	var (
		key = buildKey(midCardToken, token)
		bs  []byte
	)
	if bs, err = json.Marshal(val); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", val, err)
		return
	}
	if _, err = d.redisStore.Do(c, "SETEX", key, d.SevenDayExpire, bs); err != nil {
		log.Errorc(c, "AddShareCardToken conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.SevenDayExpire, string(bs), err)
		return
	}
	return
}

// ShareCardToken 用户集卡情况
func (d *Dao) ShareCardToken(c context.Context, token string) (res *springfestival2021.CardTokenMid, err error) {
	var (
		key = buildKey(midCardToken, token)
		bs  []byte
	)
	if bs, err = redis.Bytes(d.redisStore.Do(c, "GET", key)); err != nil {
		log.Errorc(c, "ShareCardToken conn.Do(GET key(%v)) error(%v)", key, err)
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Errorc(c, "json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

// DeleteShareCardToken ...
func (d *Dao) DeleteShareCardToken(c context.Context, token string) (err error) {
	var (
		key = buildKey(midCardToken, token)
	)
	if _, err = d.redisStore.Do(c, "DEL", key); err != nil {
		log.Errorc(c, "DeleteShareCardToken conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// SetShareCardToken add cache lottery
func (d *Dao) SetShareCardToken(c context.Context, token string, val *springfestival2021.CardTokenMid) (err error) {
	var (
		key = buildKey(midCardToken, token)
		bs  []byte
	)
	if bs, err = json.Marshal(val); err != nil {
		log.Errorc(c, "json.Marshal(%v) error (%v)", val, err)
		return
	}
	if _, err = d.redisStore.Do(c, "SET", key, bs); err != nil {
		log.Errorc(c, "AddShareCardToken conn.Send(SETEX, %s, %v) error(%v)", key, string(bs), err)
		return
	}
	return
}
