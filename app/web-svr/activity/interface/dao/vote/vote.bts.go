package vote

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/api"

	"go-common/library/cache"
)

// Activity get data from cache if miss will call source method, then add to cache.
func (d *Dao) Activity(c context.Context, key int64) (res *api.VoteActivity, err error) {
	addCache := true
	res, err = d.CacheActivity(c, key)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res.Id == -1 {
			res = nil
		}
	}()
	if res != nil {
		cache.MetricHits.Inc("bts:Activity")
		return
	}
	cache.MetricMisses.Inc("bts:Activity")
	res, err = d.RawActivity(c, key)
	if err != nil {
		return
	}
	miss := res
	if miss == nil {
		miss = &api.VoteActivity{Id: -1}
	}
	if !addCache {
		return
	}
	d.AddCacheActivity(c, key, miss)
	return
}

// DataSourceGroup get data from cache if miss will call source method, then add to cache.
func (d *Dao) DataSourceGroup(c context.Context, key int64) (res *api.VoteDataSourceGroupItem, err error) {
	addCache := true
	res, err = d.CacheDataSourceGroup(c, key)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res.GroupId == -1 {
			res = nil
		}
	}()
	if res != nil {
		cache.MetricHits.Inc("bts:Activity")
		return
	}
	cache.MetricMisses.Inc("bts:Activity")
	res, err = d.rawActivityDataSourceGroup(c, key)
	if err != nil {
		return
	}
	miss := res
	if miss == nil {
		miss = &api.VoteDataSourceGroupItem{GroupId: -1}
	}
	if !addCache {
		return
	}
	d.AddCacheDataSourceGroup(c, key, miss)
	return
}
