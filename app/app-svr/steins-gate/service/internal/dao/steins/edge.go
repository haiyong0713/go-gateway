package steins

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/stat/prom"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

// 这个方法必须传graphInfo，并且所有的edge属于传进来这个graph，并且可以传一个edgeid=0过来。
func (d *Dao) Edges(c context.Context, keys []int64, graphInfo *api.GraphInfo) (res map[int64]*api.GraphEdge, err error) {
	var (
		newKeys       []int64
		needFirstEdge bool
	)
	if graphInfo == nil {
		return
	}
	//nolint:staticcheck
	res = make(map[int64]*api.GraphEdge)
	for _, item := range keys {
		if item == model.RootEdge {
			needFirstEdge = true
		} else {
			newKeys = append(newKeys, item)
		}
	}
	if res, err = d.edges(c, newKeys); err != nil {
		return
	}
	if needFirstEdge {
		if res == nil {
			res = make(map[int64]*api.GraphEdge)
		}
		res[model.RootEdge] = model.GetFirstEdge(graphInfo)
	}
	return
}

// 这个方法支持不同树的edge返回，但是不支持edge_id=0的传参，会忽略掉0的传参
func (d *Dao) EdgesWithoutRoot(c context.Context, keys []int64) (res map[int64]*api.GraphEdge, err error) {
	var newKeys []int64
	for _, item := range keys {
		if item == model.RootEdge {
			continue
		}
		newKeys = append(newKeys, item)
	}
	if res, err = d.edges(c, newKeys); err != nil {
		log.Error("edges keys %v, err %v", newKeys, err)
	}
	return
}

// Edges get data from cache if miss will call source method, then add to cache.
func (d *Dao) edges(c context.Context, keys []int64) (res map[int64]*api.GraphEdge, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheEdges(c, keys); err != nil {
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
	prom.CacheHit.Add("Edges", int64(len(keys)-len(miss)))
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	missData := make(map[int64]*api.GraphEdge, missLen)
	prom.CacheMiss.Add("Edges", int64(missLen))
	var mutex sync.Mutex
	group := errgroup.WithContext(c)
	//nolint:gomnd
	if missLen > 10 {
		group.GOMAXPROCS(10)
	}
	var run = func(ms []int64) {
		group.Go(func(ctx context.Context) (err error) {
			data, err := d.RawEdges(ctx, ms)
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
		res = make(map[int64]*api.GraphEdge, len(keys))
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
		d.AddCacheEdges(c, missData)
	})
	return
}

// Edge 方法新增graphInfo作为参数，edge_id = 0时根据graph拼出默认值
func (d *Dao) Edge(c context.Context, key int64, graphInfo *api.GraphInfo) (res *api.GraphEdge, err error) {
	if key == model.RootEdge {
		res = model.GetFirstEdge(graphInfo)
		return
	}
	return d.edge(c, key)
}

// edge bts: -batch=50 -max_group=10 -batch_err=continue
func (d *Dao) edge(c context.Context, key int64) (res *api.GraphEdge, err error) {
	addCache := true
	res, err = d.CacheEdge(c, key)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != nil {
		prom.CacheHit.Incr("Edge")
		return
	}
	prom.CacheMiss.Incr("Edge")
	res, err = d.RawEdge(c, key)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		//nolint:errcheck
		d.AddCacheEdge(c, key, miss)
	})
	return
}

// EdgesByFromNode get edges by from node_id
func (d *Dao) EdgesByFromNode(c context.Context, fromNodeID int64) (edges []*api.GraphEdge, err error) {
	var (
		eds    *model.EdgeFromCache
		edgem  map[int64]*api.GraphEdge
		cached = true
	)
	if eds, err = d.edgeByNodeCache(c, fromNodeID); err != nil {
		log.Error("d.edgeByNodeCache err(%+v) fromNodeID(%d)", err, fromNodeID)
		err = nil
	}
	if eds != nil {
		if eds.IsEnd { // 结束节点
			return
		}
		if len(eds.ToEIDs) > 0 {
			if edgem, err = d.edges(c, eds.ToEIDs); err != nil {
				log.Error("d.edgesCache err(%+v) edgeIDs(%+v)", err, eds.ToEIDs)
				edgem = make(map[int64]*api.GraphEdge, len(eds.ToEIDs))
				err = nil
			}
			for _, eid := range eds.ToEIDs {
				if ed, ok := edgem[eid]; !ok {
					log.Error("unknown edge_id (%d)", eid)
					cached = false
				} else {
					edges = append(edges, ed)
				}
			}
		}
	}
	if !cached || eds == nil {
		prom.CacheMiss.Incr("EdgeFrom")
		if edges, err = d.edgeByNode(c, fromNodeID); err != nil {
			log.Error("d.edgeByNode err(%+v) fromNodeID(%d)", err, fromNodeID)
			return
		}
		var eds = new(model.EdgeFromCache)
		if len(edges) == 0 {
			eds.IsEnd = true
		} else {
			for _, v := range edges {
				eds.ToEIDs = append(eds.ToEIDs, v.Id)
			}
		}
		d.cache.Do(c, func(c context.Context) {
			//nolint:errcheck
			d.setEdgeFromNodeCache(c, fromNodeID, eds)
		})
	} else {
		prom.CacheHit.Incr("EdgeFrom")
	}
	return
}

// EdgeFrameAnimations is
func (d *Dao) EdgeFrameAnimations(c context.Context, keys []int64) (res map[int64]*api.EdgeFrameAnimations, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheEdgeFrameAnimations(c, keys); err != nil {
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
	prom.CacheHit.Add("EdgeFrameAnimations", int64(len(keys)-len(miss)))
	for k, v := range res {
		if v.Animations == nil {
			delete(res, k)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	missData := make(map[int64]*api.EdgeFrameAnimations, missLen)
	prom.CacheMiss.Add("EdgeFrameAnimations", int64(missLen))
	var mutex sync.Mutex
	group := errgroup.WithContext(c)
	//nolint:gomnd
	if missLen > 10 {
		group.GOMAXPROCS(10)
	}
	var run = func(ms []int64) {
		group.Go(func(ctx context.Context) (err error) {
			data, err := d.RawEdgeFrameAnimations(ctx, ms)
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
		res = make(map[int64]*api.EdgeFrameAnimations, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	for _, key := range miss {
		if res[key] == nil {
			missData[key] = &api.EdgeFrameAnimations{Animations: nil}
		}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		//nolint:errcheck
		d.AddCacheEdgeFrameAnimations(c, missData)
	})
	return

}
