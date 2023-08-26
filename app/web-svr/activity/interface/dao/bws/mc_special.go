package bws

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/log"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

func achieveReloadKey(mid, ctime int64) string {
	return fmt.Sprintf("b_ac_lod_%d_%d", mid, ctime)
}

func requestLimitKey(bid int64, key, ty string) string {
	return fmt.Sprintf("b_re_lt_%d_%s_%s", bid, key, ty)
}

// CacheUsersMids get data from mc
func (d *Dao) CacheUsersMids(c context.Context, bid int64, ids []int64) (res map[int64]*bwsmdl.Users, err error) {
	l := len(ids)
	if l == 0 {
		return
	}
	keysMap := make(map[string]int64, l)
	keys := make([]string, 0, l)
	for _, id := range ids {
		key := midKey(bid, id)
		keysMap[key] = id
		keys = append(keys, key)
	}
	replies, err := d.mc.GetMulti(c, keys)
	if err != nil {
		log.Errorv(c, log.KV("CacheUsersMids", fmt.Sprintf("%+v", err)), log.KV("keys", keys))
		return
	}
	for _, key := range replies.Keys() {
		v := &bwsmdl.Users{}
		err = replies.Scan(key, v)
		if err != nil {
			log.Errorv(c, log.KV("CacheUsersMids", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
		if res == nil {
			res = make(map[int64]*bwsmdl.Users, len(keys))
		}
		res[keysMap[key]] = v
	}
	return
}

// AddCacheUsersMids Set data to mc
func (d *Dao) AddCacheUsersMids(c context.Context, bid int64, values map[int64]*bwsmdl.Users) (err error) {
	if len(values) == 0 {
		return
	}
	for id, val := range values {
		key := midKey(bid, id)
		item := &memcache.Item{Key: key, Object: val, Expiration: d.mcExpire, Flags: memcache.FlagProtobuf}
		if err = d.mc.Set(c, item); err != nil {
			log.Errorv(c, log.KV("AddCacheUsersMids", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
	}
	return
}

// AchieveReloadSet .
func (d *Dao) RequestLimit(c context.Context, bid int64, key, ty string, expire int32) (bool, error) {
	var (
		cacaheKey = requestLimitKey(bid, key, ty)
	)
	return d.setNXCache(c, cacaheKey, expire)
}

// AchieveReloadSet .
func (d *Dao) AchieveReloadSet(c context.Context, mid, ctime int64, expire int32) (bool, error) {
	var (
		cacaheKey = achieveReloadKey(mid, ctime)
	)
	return d.setNXCache(c, cacaheKey, expire)
}

// SetNXCache .
func (d *Dao) setNXCache(c context.Context, key string, times int32) (res bool, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if res, err = redis.Bool(conn.Do("SETNX", key, "1")); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(SETNX(%s)) error(%v)", key, err)
			return
		}
	}
	if res {
		if _, err = redis.Bool(conn.Do("EXPIRE", key, times)); err != nil {
			log.Error("conn.Do(EXPIRE, %s, %d) error(%v)", key, times, err)
			return
		}
	}
	return
}

// UsersMids get data from cache if miss will call source method, then add to cache.
func (d *Dao) UsersMids(c context.Context, bid int64, mids []int64) (res map[int64]*bwsmdl.Users, err error) {
	if len(mids) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheUsersMids(c, bid, mids); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range mids {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*bwsmdl.Users
	missData, err = d.RawUsersMids(c, bid, miss)
	if res == nil {
		res = make(map[int64]*bwsmdl.Users, len(mids))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	d.AddCacheUsersMids(c, bid, missData)
	return
}

// CacheUsersKeys get data from mc
func (d *Dao) CacheUsersKeys(c context.Context, bid int64, userKeys []string) (res map[string]*bwsmdl.Users, err error) {
	l := len(userKeys)
	if l == 0 {
		return
	}
	keysMap := make(map[string]string, l)
	keys := make([]string, 0, l)
	for _, v := range userKeys {
		key := keyKey(bid, v)
		keysMap[key] = v
		keys = append(keys, key)
	}
	replies, err := d.mc.GetMulti(c, keys)
	if err != nil {
		log.Errorv(c, log.KV("CacheUsersMids", fmt.Sprintf("%+v", err)), log.KV("keys", keys))
		return
	}
	for _, key := range replies.Keys() {
		v := &bwsmdl.Users{}
		err = replies.Scan(key, v)
		if err != nil {
			log.Errorv(c, log.KV("CacheUsersMids", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
		if res == nil {
			res = make(map[string]*bwsmdl.Users, len(keys))
		}
		res[keysMap[key]] = v
	}
	return
}

// AddCacheUsersKeys Set data to mc
func (d *Dao) AddCacheUsersKeys(c context.Context, bid int64, values map[string]*bwsmdl.Users) (err error) {
	if len(values) == 0 {
		return
	}
	for id, val := range values {
		key := keyKey(bid, id)
		item := &memcache.Item{Key: key, Object: val, Expiration: d.mcExpire, Flags: memcache.FlagProtobuf}
		if err = d.mc.Set(c, item); err != nil {
			log.Errorv(c, log.KV("AddCacheUsersMids", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
	}
	return
}

// UsersKeys get data from cache if miss will call source method, then add to cache.
func (d *Dao) UsersKeys(c context.Context, bid int64, keys []string) (res map[string]*bwsmdl.Users, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheUsersKeys(c, bid, keys); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []string
	for _, key := range keys {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[string]*bwsmdl.Users
	missData, err = d.RawUsersKeys(c, bid, miss)
	if res == nil {
		res = make(map[string]*bwsmdl.Users, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	d.AddCacheUsersKeys(c, bid, missData)
	return
}
