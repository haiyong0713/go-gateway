// Code generated by kratos tool redisgen. DO NOT EDIT.

/*
  Package native is a generated redis cache package.
  It is generated from:
  type _redis interface {
		// redis: -key=natPageExtendKey -struct_name=Dao
		CacheNativeExtend(c context.Context, pid int64) (*v1.NativePageExtend, error)
		// redis: -key=natPageExtendKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
		AddCacheNativeExtend(c context.Context, pid int64, val *v1.NativePageExtend) error
		// redis: -key=natPageExtendKey -struct_name=Dao
		DelCacheNativeExtend(c context.Context, id int64) error
		// redis: -key=natTagIDExistKey -struct_name=Dao
		CacheNatTagIDExist(c context.Context, id int64) (int64, error)
		// redis: -key=natTagIDExistKey -expire=d.mcRegularExpire -encode=raw -struct_name=Dao
		AddCacheNatTagIDExist(c context.Context, id int64, val int64) error
		// redis: -key=natTagIDExistKey -struct_name=Dao
		DelCacheNatTagIDExist(c context.Context, id int64) error
		// redis: -key=nativeTabBindKey -struct_name=Dao
		DelCacheNativeTabBind(c context.Context, id int64, category int32) error
		// redis: -key=nativeParticipationKey -struct_name=Dao
		DelCacheNativePart(c context.Context, id int64) error
		// redis: -key=nativeTabSortKey -struct_name=Dao
		CacheNativeTabSort(c context.Context, id int64) ([]int64, error)
		// redis: -key=nativeTabSortKey -struct_name=Dao -expire=d.mcRegularExpire -encode=json
		AddCacheNativeTabSort(c context.Context, id int64, val []int64) error
		// redis: -key=nativeTabSortKey -struct_name=Dao
		DelNativeTabSort(c context.Context, id int64) error
		// redis: -key=nativeClickIDsKey -struct_name=Dao
		CacheNativeClickIDs(c context.Context, id int64) ([]int64, error)
		// redis: -key=nativeClickIDsKey -struct_name=Dao -expire=d.mcRegularExpire -encode=json
		AddCacheNativeClickIDs(c context.Context, id int64, val []int64) error
		// redis: -key=nativeClickIDsKey -struct_name=Dao
		DelNativeClickIDs(c context.Context, id int64) error
		// redis: -key=nativeDynamicIDsKey -struct_name=Dao
		CacheNativeDynamicIDs(c context.Context, id int64) ([]int64, error)
		// redis: -key=nativeDynamicIDsKey -struct_name=Dao -expire=d.mcRegularExpire -encode=json
		AddCacheNativeDynamicIDs(c context.Context, id int64, val []int64) error
		// redis: -key=nativeDynamicIDsKey -struct_name=Dao
		DelNativeDynamicIDs(c context.Context, id int64) error
		// redis: -key=nativeActIDsKey -struct_name=Dao
		CacheNativeActIDs(c context.Context, id int64) ([]int64, error)
		// redis: -key=nativeActIDsKey -struct_name=Dao -expire=d.mcRegularExpire -encode=json
		AddCacheNativeActIDs(c context.Context, id int64, val []int64) error
		// redis: -key=nativeActIDsKey -struct_name=Dao
		DelNativeActIDs(c context.Context, id int64) error
		// redis: -key=nativeVideoIDsKey -struct_name=Dao
		CacheNativeVideoIDs(c context.Context, id int64) ([]int64, error)
		// redis: -key=nativeVideoIDsKey -struct_name=Dao -expire=d.mcRegularExpire -encode=json
		AddCacheNativeVideoIDs(c context.Context, id int64, val []int64) error
		// redis: -key=nativeVideoIDsKey -struct_name=Dao
		DelNativeVideoIDs(c context.Context, id int64) error
		// redis: -key=nativeTabKey -struct_name=Dao
		DelCacheNativeTab(c context.Context, id int64) error
		// redis: -key=nativeTabModuleKey -struct_name=Dao
		DelCacheNativeTabModule(c context.Context, ids int64) error
		// redis: -key=nativePageKey -struct_name=Dao
		DelCacheNativePages(c context.Context, ids []int64) error
		// redis: -key=ntTsPageKey -struct_name=Dao
		DelCacheNtTsPages(c context.Context, ids []int64) error
		// redis: -key=ntTsModuleExtKey -struct_name=Dao
		DelCacheNtTsModulesExt(c context.Context, ids []int64) error
		// redis: -key=nativeForeignKey -struct_name=Dao
		CacheNativeForeign(c context.Context, id int64, pageType int64) (int64, error)
		// redis: -key=nativeForeignKey -struct_name=Dao
		DelCacheNativeForeign(c context.Context, id int64, pageType int64) error
		// redis: -key=nativeModuleKey -struct_name=Dao
		DelCacheNativeModules(c context.Context, ids []int64) error
		// redis: -key=nativeClickKey -struct_name=Dao
		DelCacheNativeClicks(c context.Context, ids []int64) error
		// redis: -key=nativeDynamicKey -struct_name=Dao
		DelCacheNativeDynamics(c context.Context, ids []int64) error
		// redis: -key=nativeVideoKey -struct_name=Dao
		DelCacheNativeVideos(c context.Context, ids []int64) error
		// redis: -key=nativeUkeyKey -struct_name=Dao
		CacheNativeUkey(c context.Context, pid int64, ukey string) (int64, error)
		// redis: -key=nativeUkeyKey -struct_name=Dao
		DelCacheNativeUkey(c context.Context, pid int64, ukey string) error
		// redis: -key=nativeMixtureKey -struct_name=Dao
		DelCacheNativeMixtures(c context.Context, ids []int64) error
		// redis: -key=natIDsByActTypeKey -struct_name=Dao
		CacheNatIDsByActType(c context.Context, actType int64) ([]int64, error)
		// redis: -key=natIDsByActTypeKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
		AddCacheNatIDsByActType(c context.Context, actType int64, val []int64) error
		// redis: -key=natIDsByActTypeKey -struct_name=Dao
		DelCacheNatIDsByActType(c context.Context, actType int64) error
		// redis: -key=userSpaceByMidKey -struct_name=Dao
		CacheUserSpaceByMid(c context.Context, mid int64) (*v1.NativeUserSpace, error)
		// redis: -key=userSpaceByMidKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
		AddCacheUserSpaceByMid(c context.Context, mid int64, val *v1.NativeUserSpace) error
		// redis: -key=userSpaceByMidKey -struct_name=Dao
		DelCacheUserSpaceByMid(c context.Context, mid int64) error
		// redis: -key=sponsoredUpKey -struct_name=Dao
		CacheSponsoredUp(c context.Context, mid int64) (bool, error)
	}
*/

package native

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
	v1 "go-gateway/app/web-svr/native-page/interface/api"
)

var _ _redis

// CacheNativeExtend get data from redis
func (d *Dao) CacheNativeExtend(c context.Context, id int64) (res *v1.NativePageExtend, err error) {
	key := natPageExtendKey(id)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheNativeExtend(get key: %v) err: %+v", key, err)
		return
	}
	res = &v1.NativePageExtend{}
	err = json.Unmarshal(reply, res)
	if err != nil {
		log.Errorc(c, "d.CacheNativeExtend(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheNativeExtend Set data to redis
func (d *Dao) AddCacheNativeExtend(c context.Context, id int64, val *v1.NativePageExtend) (err error) {
	if val == nil {
		return
	}
	key := natPageExtendKey(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheNativeExtend(get key: %v) err: %+v", key, err)
		return
	}
	expire := d.mcRegularExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheNativeExtend(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheNativeExtend delete data from redis
func (d *Dao) DelCacheNativeExtend(c context.Context, id int64) (err error) {
	key := natPageExtendKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheNativeExtend(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheNatTagIDExist get data from redis
func (d *Dao) CacheNatTagIDExist(c context.Context, id int64) (res int64, err error) {
	key := natTagIDExistKey(id)
	var temp []byte
	temp, err = redis.Bytes(d.redis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.CacheNatTagIDExist(get key: %v) err: %+v", key, err)
		return
	}
	var v string
	v = string(temp)
	r, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Errorc(c, "d.CacheNatTagIDExist(get key: %v) err: %+v", key, err)
		return
	}
	res = int64(r)
	return
}

// AddCacheNatTagIDExist Set data to redis
func (d *Dao) AddCacheNatTagIDExist(c context.Context, id int64, val int64) (err error) {
	key := natTagIDExistKey(id)
	var bs []byte
	bs = []byte(strconv.FormatInt(int64(val), 10))
	expire := d.mcRegularExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheNatTagIDExist(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheNatTagIDExist delete data from redis
func (d *Dao) DelCacheNatTagIDExist(c context.Context, id int64) (err error) {
	key := natTagIDExistKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheNatTagIDExist(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheNativeTabBind delete data from redis
func (d *Dao) DelCacheNativeTabBind(c context.Context, id int64, category int32) (err error) {
	key := nativeTabBindKey(id, category)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheNativeTabBind(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheNativePart delete data from redis
func (d *Dao) DelCacheNativePart(c context.Context, id int64) (err error) {
	key := nativeParticipationKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheNativePart(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheNativeTabSort get data from redis
func (d *Dao) CacheNativeTabSort(c context.Context, id int64) (res []int64, err error) {
	key := nativeTabSortKey(id)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheNativeTabSort(get key: %v) err: %+v", key, err)
		return
	}
	res = []int64{}
	err = json.Unmarshal(reply, &res)
	if err != nil {
		log.Errorc(c, "d.CacheNativeTabSort(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheNativeTabSort Set data to redis
func (d *Dao) AddCacheNativeTabSort(c context.Context, id int64, val []int64) (err error) {
	if len(val) == 0 {
		return
	}
	key := nativeTabSortKey(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheNativeTabSort(get key: %v) err: %+v", key, err)
		return
	}
	expire := d.mcRegularExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheNativeTabSort(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelNativeTabSort delete data from redis
func (d *Dao) DelNativeTabSort(c context.Context, id int64) (err error) {
	key := nativeTabSortKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelNativeTabSort(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheNativeClickIDs get data from redis
func (d *Dao) CacheNativeClickIDs(c context.Context, id int64) (res []int64, err error) {
	key := nativeClickIDsKey(id)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheNativeClickIDs(get key: %v) err: %+v", key, err)
		return
	}
	res = []int64{}
	err = json.Unmarshal(reply, &res)
	if err != nil {
		log.Errorc(c, "d.CacheNativeClickIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheNativeClickIDs Set data to redis
func (d *Dao) AddCacheNativeClickIDs(c context.Context, id int64, val []int64) (err error) {
	if len(val) == 0 {
		return
	}
	key := nativeClickIDsKey(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheNativeClickIDs(get key: %v) err: %+v", key, err)
		return
	}
	expire := d.mcRegularExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheNativeClickIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelNativeClickIDs delete data from redis
func (d *Dao) DelNativeClickIDs(c context.Context, id int64) (err error) {
	key := nativeClickIDsKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelNativeClickIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheNativeDynamicIDs get data from redis
func (d *Dao) CacheNativeDynamicIDs(c context.Context, id int64) (res []int64, err error) {
	key := nativeDynamicIDsKey(id)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheNativeDynamicIDs(get key: %v) err: %+v", key, err)
		return
	}
	res = []int64{}
	err = json.Unmarshal(reply, &res)
	if err != nil {
		log.Errorc(c, "d.CacheNativeDynamicIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheNativeDynamicIDs Set data to redis
func (d *Dao) AddCacheNativeDynamicIDs(c context.Context, id int64, val []int64) (err error) {
	if len(val) == 0 {
		return
	}
	key := nativeDynamicIDsKey(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheNativeDynamicIDs(get key: %v) err: %+v", key, err)
		return
	}
	expire := d.mcRegularExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheNativeDynamicIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelNativeDynamicIDs delete data from redis
func (d *Dao) DelNativeDynamicIDs(c context.Context, id int64) (err error) {
	key := nativeDynamicIDsKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelNativeDynamicIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheNativeActIDs get data from redis
func (d *Dao) CacheNativeActIDs(c context.Context, id int64) (res []int64, err error) {
	key := nativeActIDsKey(id)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheNativeActIDs(get key: %v) err: %+v", key, err)
		return
	}
	res = []int64{}
	err = json.Unmarshal(reply, &res)
	if err != nil {
		log.Errorc(c, "d.CacheNativeActIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheNativeActIDs Set data to redis
func (d *Dao) AddCacheNativeActIDs(c context.Context, id int64, val []int64) (err error) {
	if len(val) == 0 {
		return
	}
	key := nativeActIDsKey(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheNativeActIDs(get key: %v) err: %+v", key, err)
		return
	}
	expire := d.mcRegularExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheNativeActIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelNativeActIDs delete data from redis
func (d *Dao) DelNativeActIDs(c context.Context, id int64) (err error) {
	key := nativeActIDsKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelNativeActIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheNativeVideoIDs get data from redis
func (d *Dao) CacheNativeVideoIDs(c context.Context, id int64) (res []int64, err error) {
	key := nativeVideoIDsKey(id)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheNativeVideoIDs(get key: %v) err: %+v", key, err)
		return
	}
	res = []int64{}
	err = json.Unmarshal(reply, &res)
	if err != nil {
		log.Errorc(c, "d.CacheNativeVideoIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheNativeVideoIDs Set data to redis
func (d *Dao) AddCacheNativeVideoIDs(c context.Context, id int64, val []int64) (err error) {
	if len(val) == 0 {
		return
	}
	key := nativeVideoIDsKey(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheNativeVideoIDs(get key: %v) err: %+v", key, err)
		return
	}
	expire := d.mcRegularExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheNativeVideoIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelNativeVideoIDs delete data from redis
func (d *Dao) DelNativeVideoIDs(c context.Context, id int64) (err error) {
	key := nativeVideoIDsKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelNativeVideoIDs(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheNativeTab delete data from redis
func (d *Dao) DelCacheNativeTab(c context.Context, id int64) (err error) {
	key := nativeTabKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheNativeTab(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheNativeTabModule delete data from redis
func (d *Dao) DelCacheNativeTabModule(c context.Context, id int64) (err error) {
	key := nativeTabModuleKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheNativeTabModule(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheNativePages delete data from redis
func (d *Dao) DelCacheNativePages(c context.Context, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	args := redis.Args{}
	for _, id := range ids {
		key := nativePageKey(id)
		args = args.Add(key)
	}
	if _, err = d.redis.Do(c, "del", args...); err != nil {
		log.Errorc(c, "d.DelCacheNativePages() err: %+v", err)
		return
	}
	return
}

// DelCacheNtTsPages delete data from redis
func (d *Dao) DelCacheNtTsPages(c context.Context, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	args := redis.Args{}
	for _, id := range ids {
		key := ntTsPageKey(id)
		args = args.Add(key)
	}
	if _, err = d.redis.Do(c, "del", args...); err != nil {
		log.Errorc(c, "d.DelCacheNtTsPages() err: %+v", err)
		return
	}
	return
}

// DelCacheNtTsModulesExt delete data from redis
func (d *Dao) DelCacheNtTsModulesExt(c context.Context, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	args := redis.Args{}
	for _, id := range ids {
		key := ntTsModuleExtKey(id)
		args = args.Add(key)
	}
	if _, err = d.redis.Do(c, "del", args...); err != nil {
		log.Errorc(c, "d.DelCacheNtTsModulesExt() err: %+v", err)
		return
	}
	return
}

// CacheNativeForeign get data from redis
func (d *Dao) CacheNativeForeign(c context.Context, id int64, pageType int64) (res int64, err error) {
	key := nativeForeignKey(id, pageType)
	var temp []byte
	temp, err = redis.Bytes(d.redis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.CacheNativeForeign(get key: %v) err: %+v", key, err)
		return
	}
	var v string
	v = string(temp)
	r, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Errorc(c, "d.CacheNativeForeign(get key: %v) err: %+v", key, err)
		return
	}
	res = int64(r)
	return
}

// DelCacheNativeForeign delete data from redis
func (d *Dao) DelCacheNativeForeign(c context.Context, id int64, pageType int64) (err error) {
	key := nativeForeignKey(id, pageType)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheNativeForeign(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheNativeModules delete data from redis
func (d *Dao) DelCacheNativeModules(c context.Context, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	args := redis.Args{}
	for _, id := range ids {
		key := nativeModuleKey(id)
		args = args.Add(key)
	}
	if _, err = d.redis.Do(c, "del", args...); err != nil {
		log.Errorc(c, "d.DelCacheNativeModules() err: %+v", err)
		return
	}
	return
}

// DelCacheNativeClicks delete data from redis
func (d *Dao) DelCacheNativeClicks(c context.Context, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	args := redis.Args{}
	for _, id := range ids {
		key := nativeClickKey(id)
		args = args.Add(key)
	}
	if _, err = d.redis.Do(c, "del", args...); err != nil {
		log.Errorc(c, "d.DelCacheNativeClicks() err: %+v", err)
		return
	}
	return
}

// DelCacheNativeDynamics delete data from redis
func (d *Dao) DelCacheNativeDynamics(c context.Context, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	args := redis.Args{}
	for _, id := range ids {
		key := nativeDynamicKey(id)
		args = args.Add(key)
	}
	if _, err = d.redis.Do(c, "del", args...); err != nil {
		log.Errorc(c, "d.DelCacheNativeDynamics() err: %+v", err)
		return
	}
	return
}

// DelCacheNativeVideos delete data from redis
func (d *Dao) DelCacheNativeVideos(c context.Context, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	args := redis.Args{}
	for _, id := range ids {
		key := nativeVideoKey(id)
		args = args.Add(key)
	}
	if _, err = d.redis.Do(c, "del", args...); err != nil {
		log.Errorc(c, "d.DelCacheNativeVideos() err: %+v", err)
		return
	}
	return
}

// CacheNativeUkey get data from redis
func (d *Dao) CacheNativeUkey(c context.Context, id int64, ukey string) (res int64, err error) {
	key := nativeUkeyKey(id, ukey)
	var temp []byte
	temp, err = redis.Bytes(d.redis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.CacheNativeUkey(get key: %v) err: %+v", key, err)
		return
	}
	var v string
	v = string(temp)
	r, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		log.Errorc(c, "d.CacheNativeUkey(get key: %v) err: %+v", key, err)
		return
	}
	res = int64(r)
	return
}

// DelCacheNativeUkey delete data from redis
func (d *Dao) DelCacheNativeUkey(c context.Context, id int64, ukey string) (err error) {
	key := nativeUkeyKey(id, ukey)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheNativeUkey(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheNativeMixtures delete data from redis
func (d *Dao) DelCacheNativeMixtures(c context.Context, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	args := redis.Args{}
	for _, id := range ids {
		key := nativeMixtureKey(id)
		args = args.Add(key)
	}
	if _, err = d.redis.Do(c, "del", args...); err != nil {
		log.Errorc(c, "d.DelCacheNativeMixtures() err: %+v", err)
		return
	}
	return
}

// CacheNatIDsByActType get data from redis
func (d *Dao) CacheNatIDsByActType(c context.Context, id int64) (res []int64, err error) {
	key := natIDsByActTypeKey(id)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheNatIDsByActType(get key: %v) err: %+v", key, err)
		return
	}
	res = []int64{}
	err = json.Unmarshal(reply, &res)
	if err != nil {
		log.Errorc(c, "d.CacheNatIDsByActType(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheNatIDsByActType Set data to redis
func (d *Dao) AddCacheNatIDsByActType(c context.Context, id int64, val []int64) (err error) {
	if len(val) == 0 {
		return
	}
	key := natIDsByActTypeKey(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheNatIDsByActType(get key: %v) err: %+v", key, err)
		return
	}
	expire := d.mcRegularExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheNatIDsByActType(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheNatIDsByActType delete data from redis
func (d *Dao) DelCacheNatIDsByActType(c context.Context, id int64) (err error) {
	key := natIDsByActTypeKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheNatIDsByActType(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheUserSpaceByMid get data from redis
func (d *Dao) CacheUserSpaceByMid(c context.Context, id int64) (res *v1.NativeUserSpace, err error) {
	key := userSpaceByMidKey(id)
	reply, err1 := redis.Bytes(d.redis.Do(c, "GET", key))
	if err1 != nil {
		if err1 == redis.ErrNil {
			return
		}
		err = err1
		log.Errorc(c, "d.CacheUserSpaceByMid(get key: %v) err: %+v", key, err)
		return
	}
	res = &v1.NativeUserSpace{}
	err = json.Unmarshal(reply, res)
	if err != nil {
		log.Errorc(c, "d.CacheUserSpaceByMid(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// AddCacheUserSpaceByMid Set data to redis
func (d *Dao) AddCacheUserSpaceByMid(c context.Context, id int64, val *v1.NativeUserSpace) (err error) {
	if val == nil {
		return
	}
	key := userSpaceByMidKey(id)
	var bs []byte
	bs, err = json.Marshal(val)
	if err != nil {
		log.Errorc(c, "d.AddCacheUserSpaceByMid(get key: %v) err: %+v", key, err)
		return
	}
	expire := d.mcRegularExpire
	if _, err = d.redis.Do(c, "set", key, bs, "EX", expire); err != nil {
		log.Errorc(c, "d.AddCacheUserSpaceByMid(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// DelCacheUserSpaceByMid delete data from redis
func (d *Dao) DelCacheUserSpaceByMid(c context.Context, id int64) (err error) {
	key := userSpaceByMidKey(id)
	if _, err = d.redis.Do(c, "del", key); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.DelCacheUserSpaceByMid(get key: %v) err: %+v", key, err)
		return
	}
	return
}

// CacheSponsoredUp get data from redis
func (d *Dao) CacheSponsoredUp(c context.Context, id int64) (res bool, err error) {
	key := sponsoredUpKey(id)
	var temp []byte
	temp, err = redis.Bytes(d.redis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "d.CacheSponsoredUp(get key: %v) err: %+v", key, err)
		return
	}
	var v string
	v = string(temp)
	r, err := strconv.ParseBool(v)
	if err != nil {
		log.Errorc(c, "d.CacheSponsoredUp(get key: %v) err: %+v", key, err)
		return
	}
	res = bool(r)
	return
}
