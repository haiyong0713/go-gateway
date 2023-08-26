package steins

import (
	"context"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/pkg/errors"
)

const (
	_prefixGraphEdgeGroup = "geg_" // graph edge info
)

func edgeGroupKey(edgeGroupID int64) string {
	return _prefixGraphEdgeGroup + strconv.FormatInt(edgeGroupID, 10)
}

// CacheEdge get a edge info from cache.
func (d *Dao) CacheEdgeGroup(c context.Context, edgeID int64) (ge *api.EdgeGroup, err error) {
	var (
		key  = edgeGroupKey(edgeID)
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
	ge = new(api.EdgeGroup)
	if err = ge.Unmarshal(item); err != nil {
		log.Error("edgeGroup Unmarshal error(%v)", err)
		ge = nil
		return
	}
	return
}

// CacheEdges .
func (d *Dao) CacheEdgeGroups(c context.Context, edgeGroupIDs []int64) (res map[int64]*api.EdgeGroup, err error) {
	var (
		conn  = d.rds.Get(c)
		args  = redis.Args{}
		items [][]byte
	)
	defer conn.Close()
	for _, egid := range edgeGroupIDs {
		args = args.Add(edgeGroupKey(egid))
	}
	if items, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		log.Error("MGet conn.Do(mget) error(%v) args(%+v)", err, args)
		return
	}
	res = make(map[int64]*api.EdgeGroup, len(edgeGroupIDs))
	for _, item := range items {
		if item == nil { // omg
			continue
		}
		a := new(api.EdgeGroup)
		if e := a.Unmarshal(item); e != nil {
			log.Error("edgeGroup Unmarshal error(%v)", e)
			continue
		}
		res[a.Id] = a
	}
	return
}

// AddCacheEdgeGroup .
func (d *Dao) AddCacheEdgeGroup(c context.Context, edgeGroup *api.EdgeGroup) (err error) {
	if edgeGroup == nil {
		return
	}
	var item []byte
	if item, err = edgeGroup.Marshal(); err != nil {
		log.Error("edge.Marshal error(%v)", err)
		return
	}
	key := edgeGroupKey(edgeGroup.Id)
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, item); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, edgeGroup, err)
		return
	}
	return
}

// AddCacheEdges .
func (d *Dao) AddCacheEdgeGroups(c context.Context, edgegroups map[int64]*api.EdgeGroup) (err error) {
	var (
		item        []byte
		key         string
		keys        []string
		argsRecords = redis.Args{}
		conn        = d.rds.Get(c)
	)
	defer conn.Close()
	for _, edgeGroup := range edgegroups {
		if edgeGroup == nil { // ignore nil record
			continue
		}
		if item, err = edgeGroup.Marshal(); err != nil {
			log.Error("edge.Marshal error(%v)", err)
			return
		}
		key = edgeGroupKey(edgeGroup.Id)
		keys = append(keys, key)
		argsRecords = argsRecords.Add(key).Add(item)
	}
	if _, err = conn.Do("MSET", argsRecords...); err != nil {
		err = errors.Wrapf(err, "conn.Do(MSET) keys:%+v error", keys)
	}
	return

}
