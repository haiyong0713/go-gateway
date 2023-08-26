package steins

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/pkg/errors"
)

const (
	_prefixGraphEdge          = "ge_"      // graph edge info
	_prefixGraphEdgeByNode    = "new_ebn_" // graph edges by from_node
	_postfxGraphEdgeAttrs     = "_atre"    // graph edges' attrs
	_prefixEdgeFrameAnimation = "efa_"     // edge frame animation
)

func edgeKey(edgeID int64) string {
	return _prefixGraphEdge + strconv.FormatInt(edgeID, 10)
}

func edgeByNodeKey(nodeID int64) string {
	return _prefixGraphEdgeByNode + strconv.FormatInt(nodeID, 10)
}

func edgeAttrsKey(graphID int64) string {
	return strconv.FormatInt(graphID, 10) + _postfxGraphEdgeAttrs
}

func edgeFrameAnimationKey(edgeID int64) string {
	return _prefixEdgeFrameAnimation + strconv.FormatInt(edgeID, 10)
}

// CacheEdgeFrameAnimations is
func (d *Dao) CacheEdgeFrameAnimations(c context.Context, edgeIDs []int64) (map[int64]*api.EdgeFrameAnimations, error) {
	if len(edgeIDs) <= 0 {
		return map[int64]*api.EdgeFrameAnimations{}, nil
	}
	args := redis.Args{}
	for _, eid := range edgeIDs {
		key := edgeFrameAnimationKey(eid)
		args = args.Add(key)
	}

	conn := d.rds.Get(c)
	defer conn.Close()
	items, err := redis.ByteSlices(conn.Do("MGET", args...))
	if err != nil {
		return nil, err
	}
	if len(items) != len(edgeIDs) {
		log.Error("Inconsistent item in reply: %+v: %+v", items, edgeIDs)
		return nil, errors.New("Inconsistent item count in reply")
	}

	res := make(map[int64]*api.EdgeFrameAnimations, len(edgeIDs))
	for i, edgeID := range edgeIDs {
		efa := new(api.EdgeFrameAnimations)
		if items[i] == nil {
			continue
		}
		if err := efa.Unmarshal(items[i]); err != nil {
			log.Error("Failed to unmarshal edge frame animation: %+v", err)
			continue
		}
		res[edgeID] = efa
	}
	return res, nil
}

// CacheEdge get a edge info from cache.
func (d *Dao) CacheEdge(c context.Context, edgeID int64) (ge *api.GraphEdge, err error) {
	var (
		key  = edgeKey(edgeID)
		conn = d.rds.Get(c)
		item []byte
	)
	defer conn.Close()
	if item, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	ge = new(api.GraphEdge)
	if err = ge.Unmarshal(item); err != nil {
		log.Error("edge Unmarshal error(%v)", err)
		ge = nil
		return
	}
	return
}

// CacheEdges .
func (d *Dao) CacheEdges(c context.Context, edgeIDs []int64) (res map[int64]*api.GraphEdge, err error) {
	var (
		conn  = d.rds.Get(c)
		args  = redis.Args{}
		items [][]byte
	)
	defer conn.Close()
	for _, eid := range edgeIDs {
		args = args.Add(edgeKey(eid))
	}
	if items, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		log.Error("MGet conn.Do(mget) error(%v) args(%+v)", err, args)
		return
	}
	res = make(map[int64]*api.GraphEdge, len(edgeIDs))
	for _, item := range items {
		if item == nil { // omg
			continue
		}
		a := new(api.GraphEdge)
		if e := a.Unmarshal(item); e != nil {
			log.Error("edge Unmarshal error(%v)", e)
			continue
		}
		res[a.Id] = a
	}
	return
}

// AddCacheEdge .
func (d *Dao) AddCacheEdge(c context.Context, eid int64, edge *api.GraphEdge) (err error) {
	if edge == nil {
		return
	}
	var item []byte
	if item, err = edge.Marshal(); err != nil {
		log.Error("edge.Marshal error(%v)", err)
		return
	}
	key := edgeKey(eid)
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, item); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, edge, err)
		return
	}
	return
}

// AddCacheEdges .
func (d *Dao) AddCacheEdges(c context.Context, edges map[int64]*api.GraphEdge) (err error) {
	var (
		item        []byte
		key         string
		keys        []string
		argsRecords = redis.Args{}
		conn        = d.rds.Get(c)
	)
	defer conn.Close()
	for _, edge := range edges {
		if edge == nil { // ignore nil record
			continue
		}
		if item, err = edge.Marshal(); err != nil {
			log.Error("edge.Marshal error(%v)", err)
			return
		}
		key = edgeKey(edge.Id)
		keys = append(keys, key)
		argsRecords = argsRecords.Add(key).Add(item)
	}
	if _, err = conn.Do("MSET", argsRecords...); err != nil {
		err = errors.Wrapf(err, "conn.Do(MSET) keys:%+v error", keys)
	}
	return
}

// edgeByNodeCache get a edge info from cache.
func (d *Dao) edgeByNodeCache(c context.Context, nodeID int64) (eds *model.EdgeFromCache, err error) {
	var (
		key  = edgeByNodeKey(nodeID)
		conn = d.rds.Get(c)
		item []byte
	)
	defer conn.Close()
	if item, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
		return
	}
	eds = new(model.EdgeFromCache)
	if err = json.Unmarshal(item, &eds); err != nil { // json反序列化
		log.Error("edgeByNode Unmarshal error(%v)", err)
		eds = nil
		return
	}
	return
}

func (d *Dao) setEdgeFromNodeCache(c context.Context, fromNodeID int64, edgeFromCache *model.EdgeFromCache) (err error) {
	if edgeFromCache == nil || fromNodeID == 0 {
		return
	}
	var item []byte
	if item, err = json.Marshal(edgeFromCache); err != nil { // json序列化
		log.Error("edgeFrom.Marshal error(%v)", err)
		return
	}
	key := edgeByNodeKey(fromNodeID)
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, item); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, edgeFromCache, err)
	}
	return
}

// AddCacheEdgeAttrs . TODO 这里被saveGraph调用了，所以在两个dao里都写了该方法，后续优化为只写一处
func (d *Dao) AddCacheEdgeAttrs(c context.Context, graphID int64, edgeAttrs *model.EdgeAttrsCache) (err error) {
	if edgeAttrs == nil {
		return
	}
	var item []byte
	if item, err = json.Marshal(edgeAttrs); err != nil { // json序列化
		log.Error("edgeAttrs.Marshal error(%v)", err)
		return
	}
	key := edgeAttrsKey(graphID)
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, item); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, edgeAttrs, err)
	}
	return
}

// AddCacheEdges .
func (d *Dao) AddCacheEdgeFrameAnimations(c context.Context, in map[int64]*api.EdgeFrameAnimations) error {
	if len(in) <= 0 {
		return nil
	}

	args := redis.Args{}
	for edgeID, efa := range in {
		item, err := efa.Marshal()
		if err != nil {
			log.Error("Failed to marshal edge frame animation: %+v", err)
			continue
		}
		key := edgeFrameAnimationKey(edgeID)
		args = args.Add(key).Add(item)
	}

	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err := conn.Do("MSET", args...); err != nil {
		return err
	}
	return nil

}
