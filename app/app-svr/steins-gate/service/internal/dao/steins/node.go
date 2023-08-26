package steins

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/stat/prom"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/pkg/errors"
)

// Nodes get data from cache if miss will call source method, then add to cache.
func (d *Dao) Nodes(c context.Context, keys []int64) (res map[int64]*api.GraphNode, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheNodes(c, keys); err != nil {
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
	prom.CacheHit.Add("Nodes", int64(len(keys)-len(miss)))
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	missData := make(map[int64]*api.GraphNode, missLen)
	prom.CacheMiss.Add("Nodes", int64(missLen))
	var mutex sync.Mutex
	group := errgroup.WithContext(c)
	//nolint:gomnd
	if missLen > 10 {
		group.GOMAXPROCS(10)
	}
	var run = func(ms []int64) {
		group.Go(func(ctx context.Context) (err error) {
			data, err := d.RawNodes(ctx, ms)
			if err != nil {
				return
			}
			var (
				cids       []int64
				dimensions map[int64]*model.DimensionInfo
			)
			for _, node := range data {
				cids = append(cids, node.Cid)
			}
			if dimensions, err = d.BvcDimensions(ctx, cids); err != nil {
				log.Error("BvcDimensions Nids %v, Cids %v, err %+v", ms, cids, err)
				err = ecode.GraphGetDimensionErr
				return
			}
			for k, node := range data { // bts 逻辑增加读取视频云的dimension
				if dimension, ok := dimensions[node.Cid]; ok {
					node.Width = dimension.Width
					node.Height = dimension.Height
					node.Sar = dimension.Sar
				}
				mutex.Lock()
				missData[k] = node
				mutex.Unlock()
			}
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
	if err != nil {
		return
	}
	if res == nil {
		res = make(map[int64]*api.GraphNode, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		//nolint:errcheck
		d.AddCacheNodes(c, missData)
	})
	return
}

// Node bts: -batch=50 -max_group=10 -batch_err=continue
func (d *Dao) Node(c context.Context, key int64) (res *api.GraphNode, err error) {
	addCache := true
	res, err = d.CacheNode(c, key)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != nil {
		prom.CacheHit.Incr("Node")
		return
	}
	prom.CacheMiss.Incr("Node")
	var bvcRes *model.DimensionInfo
	if res, err = d.RawNode(c, key); err != nil {
		err = errors.Wrapf(err, "NodeID %d", key)
		return
	}
	if res == nil {
		log.Warn("NodeID %d Not found", key)
		return
	}
	if bvcRes, err = d.BvcDimension(c, res.Cid); err != nil {
		err = errors.Wrapf(err, "NodeID %d", key)
		return
	}
	if bvcRes != nil {
		res.Height = bvcRes.Height
		res.Width = bvcRes.Width
		res.Sar = bvcRes.Sar
	}
	miss := res
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		//nolint:errcheck
		d.AddCacheNode(c, key, miss)
	})
	return

}
