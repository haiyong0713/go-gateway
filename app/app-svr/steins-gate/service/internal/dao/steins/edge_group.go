package steins

import (
	"context"
	"sync"

	"go-common/library/stat/prom"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/steins-gate/service/api"
)

func (d *Dao) EdgeGroups(c context.Context, keys []int64) (res map[int64]*api.EdgeGroup, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheEdgeGroups(c, keys); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	prom.CacheHit.Add("EdgeGroups", int64(len(keys)-len(miss)))
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	missData := make(map[int64]*api.EdgeGroup, missLen)
	prom.CacheMiss.Add("EdgeGroups", int64(missLen))
	var mutex sync.Mutex
	group := errgroup.WithContext(c)
	//nolint:gomnd
	if missLen > 10 {
		group.GOMAXPROCS(10)
	}
	var run = func(ms []int64) {
		group.Go(func(ctx context.Context) (err error) {
			data, err := d.RawEdgeGroups(ctx, ms)
			mutex.Lock()
			for k, v := range data {
				missData[k] = v
			}
			mutex.Unlock()
			return
		})
	}
	var (
		i int
		n = missLen / 50
	)
	for i = 0; i < n; i++ {
		run(miss[i*50 : (i+1)*50])
	}
	if len(miss[i*50:]) > 0 {
		run(miss[i*50:])
	}
	err = group.Wait()
	if res == nil {
		res = make(map[int64]*api.EdgeGroup, len(keys))
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
	d.cache.Do(c, func(c context.Context) {
		//nolint:errcheck
		d.AddCacheEdgeGroups(c, missData)
	})
	return

}
