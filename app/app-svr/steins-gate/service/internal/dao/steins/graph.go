package steins

import (
	"context"
	"sync"

	xecode "go-gateway/app/app-svr/steins-gate/ecode"
	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/stat/prom"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

// GraphShow .
func (d *Dao) GraphShow(c context.Context, aid, graphID int64) (data *model.Graph, err error) {
	var graphDB *model.GraphDB
	if graphID == 0 {
		if graphDB, err = d.graph(c, aid, false); err != nil {
			log.Error("GraphShow d.graph(%d) error(%v)", aid, err)
			return
		}
		if graphDB == nil {
			err = ecode.NothingFound
			return
		}
		data = &model.Graph{
			ID:     graphDB.Id,
			Aid:    graphDB.Aid,
			Script: graphDB.Script,
			Ctime:  graphDB.Ctime,
		}
		return
	}
	if data, err = d.graphByID(c, graphID); err != nil {
		log.Error("GraphShow d.graphByID(%d) error(%v)", graphID, err)
		return
	}
	if data == nil {
		err = ecode.NothingFound
		return
	}
	return
}

// GraphInfo is
func (d *Dao) GraphInfo(c context.Context, aid int64) (a *api.GraphInfo, err error) {
	if a, err = d.graphCache(c, aid); err != nil {
		log.Error("d.archivePBCache(%d) error(%v)", aid, err)
		err = nil // NOTE ignore error use db
	}
	if a != nil { // found graph info in MC
		prom.CacheHit.Incr("Graph")
		return
	}
	prom.CacheMiss.Incr("Graph")
	if a, err = d.graphWithStarting(c, aid, false); err != nil {
		return
	}
	if fanoutErr := d.cache.Do(c, func(ctx context.Context) {
		//nolint:errcheck
		d.setGraphCache(ctx, a)
	}); fanoutErr != nil {
		log.Error("fanout do aid %d err(%+v) ", aid, fanoutErr)
	}
	return
}

// GraphInfoPreview is
func (d *Dao) GraphInfoPreview(c context.Context, aid int64) (a *api.GraphInfo, err error) {
	return d.graphWithStarting(c, aid, true)
}

func (d *Dao) setGraphAllCache(c context.Context, aid, graphID int64, dimensions map[int64]*model.DimensionInfo) (err error) {
	var (
		graphDB   *model.GraphDB
		nodes     []*api.GraphNode
		edges     []*api.GraphEdge
		nodeEdges = make(map[int64]*model.EdgeFromCache)
	)
	if graphDB, err = d.graph(c, aid, false); err != nil {
		log.Error("setGraphAllCache d.graph(%d) error(%v)", aid, err)
		err = nil
	}
	// set graph cache
	if graphDB == nil {
		return
	}
	if graphDB.IsPass() { // if graph is valid, return it and set it in MC
		a := &graphDB.GraphInfo
		if a.FirstNid, a.FirstCid, err = d.startingPoint(c, graphDB.Id); err != nil {
			return
		}
		//nolint:errcheck
		d.setGraphCache(c, a)
	}
	// set node cache
	if nodes, err = d.GraphNodeList(c, graphID); err != nil {
		log.Error("setGraphAllCache d.GraphNodeList graphID(%d) error(%v)", graphID, err)
		err = nil
	} else {
		for _, node := range nodes {
			dim, ok := dimensions[node.Cid] // 添加缓存时候读取dimensions数据
			if !ok {
				continue
			}
			node.Height = dim.Height
			node.Width = dim.Width
			if e := d.AddCacheNode(c, node.Id, node); e != nil {
				log.Error("setGraphAllCache d.setNodeCache(%v) error(%v)", node, e)
			}
			nodeEdges[node.Id] = &model.EdgeFromCache{
				IsEnd: true,
			}
		}
	}
	// set edge cache
	if edges, err = d.GraphEdgeList(c, graphID); err != nil {
		log.Error("setGraphAllCache d.GraphEdgeList graphID(%d) error(%v)", graphID, err)
		err = nil
	} else {
		var edas = new(model.EdgeAttrsCache)
		for _, edge := range edges {
			if edc, ok := nodeEdges[edge.FromNode]; ok {
				edc.IsEnd = false
				edc.ToEIDs = append(edc.ToEIDs, edge.Id)
			}
			if e := d.AddCacheEdge(c, edge.Id, edge); e != nil {
				log.Error("setGraphAllCache d.setEdgeCache(%v) error(%v)", edge, e)
			}
			if len(edge.Attribute) != 0 { // for the edges with the attribute, we save them in cache
				edas.EdgeAttrs = append(edas.EdgeAttrs, &model.EdgeAttr{
					FromNID:   edge.FromNode,
					ToNID:     edge.ToNode,
					Attribute: edge.Attribute,
				})
			}
		}
		if len(edas.EdgeAttrs) > 0 {
			edas.HasAttrs = true
		}
		if e := d.AddCacheEdgeAttrs(c, graphID, edas); e != nil {
			log.Error("setGraphAllCache d.AddCacheEdgeAttrs Gid %d Err %v", graphID, err)
		}
		if len(nodeEdges) > 0 {
			for fromNodeID, edgeIDs := range nodeEdges {
				if e := d.setEdgeFromNodeCache(c, fromNodeID, edgeIDs); e != nil {
					log.Error("setGraphAllCache d.setEdgeFromNodeCache fromNodeID(%d) edgeIDs(%v) error(%v)", fromNodeID, edgeIDs, e)
				}
			}
		}
	}
	return
}

// GraphInfos get data from cache if miss will call source method, then add to cache.
func (d *Dao) GraphInfos(c context.Context, keys []int64) (res map[int64]*api.GraphInfo) {
	if len(keys) == 0 {
		return
	}
	var err error
	addCache := true
	if res, err = d.CacheGraphs(c, keys); err != nil {
		addCache = false
		res = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	prom.CacheHit.Add("Graph", int64(len(keys)-len(miss)))
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	missData := make(map[int64]*api.GraphInfo, missLen)
	prom.CacheMiss.Add("Graph", int64(missLen))
	var mutex sync.Mutex
	group := errgroup.WithContext(c)
	//nolint:gomnd
	if missLen > 10 {
		group.GOMAXPROCS(10)
	}
	var run = func(ms []int64) {
		group.Go(func(ctx context.Context) (err error) { // for循环回查graph信息，任何一个err直接return
			var data *api.GraphInfo
			for _, aid := range ms {
				if data, err = d.graphWithStarting(ctx, aid, false); err != nil {
					continue // graph信息缺失，continue
				}
				mutex.Lock()
				missData[aid] = data
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
	if res == nil {
		res = make(map[int64]*api.GraphInfo, len(keys))
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
		d.setGraphsCache(c, missData)
	})
	return
}

// GraphAuditMigrate 过审时将审核表的数据搬移到结果表
func (d *Dao) GraphAuditMigrate(c context.Context, graphDB *model.GraphAuditDB, mid int64) (resultGID int64, err error) {
	var (
		edges      []*api.GraphEdge
		nodes      []*api.GraphNode
		dimensions map[int64]*model.DimensionInfo
		auditGID   = graphDB.Id
	)
	resultGID = graphDB.ResultGID
	eg := errgroup.WithCancel(c)
	eg.Go(func(c context.Context) (err error) {
		edges, err = d.GraphEdgeList(c, auditGID, true)
		return
	})
	eg.Go(func(c context.Context) (err error) {
		nodes, err = d.GraphNodeList(c, auditGID, true)
		return
	})
	if err = eg.Wait(); err != nil {
		err = errors.Wrapf(err, "resultGID %d auditGID %d", resultGID, auditGID)
		return
	}
	var cids []int64
	for _, v := range nodes {
		cids = append(cids, v.Cid)
	}
	if dimensions, err = d.BvcDimensions(c, cids); err != nil {
		log.Error("BvcDimensions Cids %+v, Err %v", cids, err)
		err = xecode.GraphGetDimensionErr
		return
	}
	saveParam := new(model.SaveGraphParam)
	saveParam.FromAudit(graphDB, edges, nodes)
	if resultGID, err = d.SaveGraph(c, 0, false, saveParam, dimensions, mid); err != nil {
		err = errors.Wrapf(err, "resultGID %d auditGID %d", resultGID, auditGID)
	}
	return

}
