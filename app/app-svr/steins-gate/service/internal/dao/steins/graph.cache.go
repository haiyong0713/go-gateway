package steins

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"github.com/pkg/errors"
)

const (
	_preFixGraph = "gh_"
)

func graphKey(aid int64) string {
	return _preFixGraph + strconv.FormatInt(aid, 10)
}

func (d *Dao) setGraphCache(c context.Context, graphInfo *api.GraphInfo) (err error) {
	if graphInfo == nil {
		return
	}
	var (
		item []byte
		key  = graphKey(graphInfo.Aid)
	)
	if item, err = graphInfo.Marshal(); err != nil {
		log.Error("graph.Marshal error(%v)", err)
		return
	}
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, item); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, graphInfo, err)
	}
	return
}

func (d *Dao) graphCache(c context.Context, aid int64) (a *api.GraphInfo, err error) {
	var (
		key  = graphKey(aid)
		conn = d.rds.Get(c)
		item []byte
	)
	defer conn.Close()
	if item, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		err = errors.Wrapf(err, "aid %d", aid)
		return
	}
	a = new(api.GraphInfo)
	if err = a.Unmarshal(item); err != nil {
		err = errors.Wrapf(err, "aid %d", aid)
		return
	}
	return
}

// CacheGraphs .
func (d *Dao) CacheGraphs(c context.Context, aids []int64) (res map[int64]*api.GraphInfo, err error) {
	var (
		conn  = d.rds.Get(c)
		args  = redis.Args{}
		items [][]byte
	)
	defer conn.Close()
	for _, aid := range aids {
		args = args.Add(graphKey(aid))
	}
	if items, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		log.Error("MGet conn.Do(mget) error(%v) args(%+v)", err, args)
		return
	}
	res = make(map[int64]*api.GraphInfo, len(aids))
	for _, item := range items {
		if item == nil { // omg
			continue
		}
		a := new(api.GraphInfo)
		if e := a.Unmarshal(item); e != nil {
			log.Error("graphInfo Unmarshal error(%v)", e)
			continue
		}
		res[a.Aid] = a
	}
	return
}

func (d *Dao) setGraphsCache(c context.Context, graphInfos map[int64]*api.GraphInfo) (err error) {
	if graphInfos == nil {
		return
	}
	var (
		item        []byte
		key         string
		keys        []string
		argsRecords = redis.Args{}
		conn        = d.rds.Get(c)
	)
	defer conn.Close()
	for _, graph := range graphInfos {
		if graph == nil { // ignore nil record
			continue
		}
		if item, err = graph.Marshal(); err != nil {
			log.Error("record.Marshal error(%v)", err)
			return
		}
		key = graphKey(graph.Aid)
		keys = append(keys, key)
		argsRecords = argsRecords.Add(key).Add(item)
	}
	if _, err = conn.Do("MSET", argsRecords...); err != nil {
		err = errors.Wrapf(err, "conn.Do(MSET) keys:%+v error", keys)
	}
	return
}

// DelGraphCache def.
func (d *Dao) DelGraphCache(c context.Context, aid int64) (err error) {
	var (
		key  = graphKey(aid)
		conn = d.rds.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelGraphCache Aid %d, Err %v", aid, err)
	}
	return

}
