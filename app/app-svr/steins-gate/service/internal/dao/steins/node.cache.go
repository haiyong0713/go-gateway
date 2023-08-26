package steins

import (
	"context"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/pkg/errors"
)

const _prefixGraphNode = "gn_" // graph node info

func nodeKey(nodeID int64) string {
	return _prefixGraphNode + strconv.FormatInt(nodeID, 10)
}

// CacheNode get a node info from cache.
func (d *Dao) CacheNode(c context.Context, nodeID int64) (res *api.GraphNode, err error) {
	var (
		key  = nodeKey(nodeID)
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
	res = new(api.GraphNode)
	if err = res.Unmarshal(item); err != nil {
		log.Error("node Unmarshal error(%v)", err)
		res = nil
		return
	}
	return
}

// CacheNodes .
func (d *Dao) CacheNodes(c context.Context, nodeIDs []int64) (res map[int64]*api.GraphNode, err error) {
	var (
		conn  = d.rds.Get(c)
		args  = redis.Args{}
		items [][]byte
	)
	defer conn.Close()
	for _, nid := range nodeIDs {
		args = args.Add(nodeKey(nid))
	}
	if items, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		log.Error("MGet conn.Do(mget) error(%v) args(%+v)", err, args)
		return
	}
	res = make(map[int64]*api.GraphNode, len(nodeIDs))
	for _, item := range items {
		if item == nil { // omg
			continue
		}
		a := new(api.GraphNode)
		if e := a.Unmarshal(item); e != nil {
			log.Error("node Unmarshal error(%v)", e)
			continue
		}
		res[a.Id] = a
	}
	return
}

// AddCacheNode .
func (d *Dao) AddCacheNode(c context.Context, nid int64, node *api.GraphNode) (err error) {
	if node == nil {
		return
	}
	var item []byte
	if item, err = node.Marshal(); err != nil {
		log.Error("node.Marshal error(%v)", err)
		return
	}
	key := nodeKey(nid)
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, item); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, node, err)
		return
	}
	return
}

// AddCacheNodes .
func (d *Dao) AddCacheNodes(c context.Context, nodes map[int64]*api.GraphNode) (err error) {
	var (
		item        []byte
		key         string
		keys        []string
		argsRecords = redis.Args{}
		conn        = d.rds.Get(c)
	)
	defer conn.Close()
	for _, node := range nodes {
		if node == nil { // ignore nil record
			continue
		}
		if item, err = node.Marshal(); err != nil {
			log.Error("node.Marshal error(%v)", err)
			return
		}
		key = nodeKey(node.Id)
		keys = append(keys, key)
		argsRecords = argsRecords.Add(key).Add(item)
	}
	if _, err = conn.Do("MSET", argsRecords...); err != nil {
		err = errors.Wrapf(err, "conn.Do(MSET) keys:%+v error", keys)
	}
	return

}
