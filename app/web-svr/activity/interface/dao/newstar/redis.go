package newstar

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/newstar"
)

func (d *Dao) creationKey(mid int64, activityUID string) string {
	return fmt.Sprintf("creation_%d_%s", mid, activityUID)
}

func (d *Dao) invitesKey(inviterMid int64, activityUID string) string {
	return fmt.Sprintf("invites_%d_%s", inviterMid, activityUID)
}

func (d *Dao) inviteCountKey(inviterMid int64, activityUID string) string {
	return fmt.Sprintf("invite_c_%d_%s", inviterMid, activityUID)
}

func (d *Dao) CacheCreationByMid(ctx context.Context, mid int64, activityUID string) (res *newstar.Newstar, err error) {
	var (
		key  = d.creationKey(mid, activityUID)
		conn = d.redis.Get(ctx)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheCreationByID(%s) return nil", key)
		} else {
			log.Error("CacheCreationByID conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("CacheCreationByID json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (d *Dao) AddCacheCreationByMid(c context.Context, mid int64, activityUID string, data *newstar.Newstar) (err error) {
	var (
		key  = d.creationKey(mid, activityUID)
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal(%v) error (%+v)", data, err)
		return
	}
	if err = conn.Send("SETEX", key, d.newstarExpire, bs); err != nil {
		log.Error("AddCacheCreationByID conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.newstarExpire, string(bs), err)
	}
	return
}

func (d *Dao) CacheInvites(c context.Context, inviterMid int64, activityUID string) (res []*newstar.Newstar, err error) {
	var (
		key  = d.invitesKey(inviterMid, activityUID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZREVRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		cale := &newstar.Newstar{}
		if err = json.Unmarshal(bs, cale); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, cale)
	}
	return
}

func (d *Dao) AddCacheInvites(c context.Context, inviterMid int64, activityUID string, data []*newstar.Newstar) (err error) {
	var (
		key  = d.invitesKey(inviterMid, activityUID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	args := redis.Args{}.Add(key)
	for _, newstar := range data {
		bs, _ := json.Marshal(newstar)
		args = args.Add(newstar.ID).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return
	}
	count++
	if err = conn.Send("EXPIRE", key, d.newstarExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.newstarExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// CacheInviteCount .
func (d *Dao) CacheInviteCount(ctx context.Context, inviterMid int64, activityUID string) (r int64, err error) {
	var (
		key  = d.inviteCountKey(inviterMid, activityUID)
		conn = d.redis.Get(ctx)
	)
	defer conn.Close()
	if r, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheInviteCount(%s) return nil", key)
		} else {
			log.Error("CacheInviteCount conn.Do(GET key(%v)) error(%v)", key, err)
		}
	}
	return
}

// AddCacheInviteCount .
func (d *Dao) AddCacheInviteCount(ctx context.Context, inviterMid int64, activityUID string, val int64) (err error) {
	var (
		key  = d.inviteCountKey(inviterMid, activityUID)
		conn = d.redis.Get(ctx)
	)
	defer conn.Close()
	if err = conn.Send("SETEX", key, d.newstarExpire, val); err != nil {
		log.Error("AddCacheInviteCount conn.Send(SETEX, %s, %v, %s) error(%v)", key, d.newstarExpire, val, err)
	}
	return
}

func (d *Dao) DelCacheInviteCount(ctx context.Context, inviterMid int64, activityUID string) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := d.inviteCountKey(inviterMid, activityUID)
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("DelCacheInviteCount key(%s) error(%v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) DelCacheCacheInvite(ctx context.Context, inviterMid int64, activityUID string) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	key := d.invitesKey(inviterMid, activityUID)
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("DelCacheCacheInvite key(%s) error(%v)", key, err)
		return err
	}
	return nil
}
