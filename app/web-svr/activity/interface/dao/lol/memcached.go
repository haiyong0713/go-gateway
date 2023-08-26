package lol

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-common/library/cache"
	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-common/library/stat/prom"
	esmdl "go-gateway/app/web-svr/activity/interface/model/esports_model"
	"go-gateway/app/web-svr/activity/interface/model/lol"
)

const (
	plKey         = "pl_%d"
	_userGuessOid = "user_guess_oid_%d"
)

func pointListKey(mid int64) string {
	return fmt.Sprintf(plKey, mid)
}

func keyUserGuessOid(mid int64) string {
	return fmt.Sprintf(_userGuessOid, mid)
}

// SetPointList 设置积分列表.
func (d *Dao) SetPointList(c context.Context, mid int64, data []byte) (err error) {
	key := pointListKey(mid)
	if err = d.mc.Set(c, &memcache.Item{
		Key:        key,
		Value:      data,
		Flags:      memcache.FlagRAW,
		Expiration: d.userExpire,
	}); err != nil {
		prom.BusinessErrCount.Incr("mc:SetPointList")
		log.Error("Memcache SetPointList key:%v error:%v", key, err)
		return
	}
	return
}

// PointList 获取积分列表.
func (d *Dao) PointList(c context.Context, mid int64) (list []*lol.PointMsg, err error) {
	list = make([]*lol.PointMsg, 0)
	key := pointListKey(mid)
	if err = d.mc.Get(c, key).Scan(&list); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			return
		}

		prom.BusinessErrCount.Incr("mc:PointList")
		log.Error("Memcache PointList key:%v error:%v", key, err)
	}

	return
}

// ContestList 获取赛程列表.
func (d *Dao) ContestList(c context.Context) (res []int64, err error) {
	if err = component.S10GlobalMC.Get(c, d.eSportsKey).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			return nil, nil
		}
		prom.BusinessErrCount.Incr("mc:ContestList")
		log.Error("Memcache ContestList key:%v error:%v", d.eSportsKey, err)
		return nil, err
	}
	return
}

// ContestListDetail 获取赛程列表详情.
func (d *Dao) ContestListDetail(c context.Context) (res map[int64][]*esmdl.ContestCard, err error) {
	res = make(map[int64][]*esmdl.ContestCard, 0)
	if err = component.S10GlobalMC.Get(c, d.contestDetailKey).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			return res, nil
		}
		prom.BusinessErrCount.Incr("mc:ContestListDetail")
		log.Error("Memcache ContestListDetail key:%v error:%v", d.contestDetailKey, err)
		return nil, err
	}
	return
}

// UserGuessOid get data from cache if miss will call source method, then add to cache.
func (d *Dao) UserGuessOid(c context.Context, mid int64, mainIDs []int64) (res []*lol.UserGuessOid, err error) {
	addCache := true
	res, err = d.CacheUserGuessOid(c, mid, mainIDs)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if len(res) != 0 && res[0].ID == -1 {
			res = nil
		}
	}()
	if len(res) != 0 {
		cache.MetricHits.Inc("bts:UserGuessOid")
		return
	}
	cache.MetricMisses.Inc("bts:UserGuessOid")
	res, err = RawUserGuessOid(c, mid, mainIDs)
	if err != nil {
		return
	}
	miss := res
	if len(miss) == 0 {
		miss = []*lol.UserGuessOid{{ID: -1}}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheUserGuessOid(c, mid, miss, mainIDs)
	})
	return
}

func (d *Dao) CacheUserGuessOid(c context.Context, mid int64, mainIDs []int64) (res []*lol.UserGuessOid, err error) {
	key := keyUserGuessOid(mid)
	if err := component.S10GlobalMC.Get(c, key).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			res = nil
		} else {
			log.Error("CacheUserGuessOid conn.Get error(%v)", err)
		}
	}
	return
}

func (d *Dao) AddCacheUserGuessOid(c context.Context, mid int64, value []*lol.UserGuessOid, mainIDs []int64) (err error) {
	key := keyUserGuessOid(mid)
	if err = component.S10GlobalMC.Set(c, &memcache.Item{
		Key:        key,
		Object:     value,
		Expiration: 86400,
		Flags:      memcache.FlagJSON,
	}); err != nil {
		log.Error("AddCacheUserGuessOid(%d) value(%v) error(%v)", key, value, err)
	}
	return
}

// GuessDetailOptions 获取竞猜信息.
func (d *Dao) GuessDetailOptions(c context.Context) (res map[int64][]*lol.DetailOption, err error) {
	if err = component.S10GlobalMC.Get(c, d.GuessMainDetailsKey).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			return nil, nil
		}
		prom.BusinessErrCount.Incr("mc:GuessDetailOptions")
		log.Error("Memcache GuessDetailOptions key:%v error:%v", d.GuessMainDetailsKey, err)
		return nil, err
	}
	return
}
