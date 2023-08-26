package dao

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"
)

const (
	_taskList  = "act_task_%d"
	_prizeList = "act_prize_%d"
)

func keyCacheTask(uid int64) string {
	return fmt.Sprintf(_taskList, uid)
}
func keyCachePrize() string {
	return _prizeList
}

func (d *Dao) CacheTask(c context.Context, uid int64) (res *model.Task, err error) {
	key := keyCacheTask(uid)
	if err := d.mc.Get(c, key).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			res = nil
		} else {
			log.Error("conn.Get error(%v)", err)
		}
	}
	return
}

func (d *Dao) AddCacheTask(c context.Context, uid int64, value *model.Task) (err error) {
	key := keyCacheTask(uid)
	if err = d.mc.Set(c, &memcache.Item{
		Key:        key,
		Object:     value,
		Expiration: d.taskExpire,
		Flags:      memcache.FlagJSON,
	}); err != nil {
		log.Error("AddCacheGoodsList(%d) value(%v) error(%v)", key, value, err)
		return
	}
	return
}

// DelCacheTask delete data from mc
func (d *Dao) DelCacheTask(c context.Context, id int64) (err error) {
	key := keyCacheTask(id)
	if err = d.mc.Delete(c, key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}
		log.Error("DelCacheTask(%d) error(%v)", key, err)
		return
	}
	return
}

func (d *Dao) CachePrizeList(c context.Context) (res []*model.Task, err error) {
	key := keyCachePrize()
	if err := d.mc.Get(c, key).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			res = nil
		} else {
			log.Error("conn.Get error(%v)", err)
		}
	}
	return
}

func (d *Dao) AddCachePrizeList(c context.Context, value []*model.Task) (err error) {
	key := keyCachePrize()
	if err = d.mc.Set(c, &memcache.Item{
		Key:        key,
		Object:     value,
		Expiration: d.confExpire,
		Flags:      memcache.FlagJSON,
	}); err != nil {
		log.Error("AddCacheGoodsList(%d) value(%v) error(%v)", key, value, err)
		return
	}
	return
}
