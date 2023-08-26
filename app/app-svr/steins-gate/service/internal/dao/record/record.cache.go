package record

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

const (
	_prefixRecord = "r_%d_%d_%s"
)

func recordKey(mid, graphID int64, buvid string) string {
	return fmt.Sprintf(_prefixRecord, mid, graphID, buvid)
}

// AddCacheRecord .
func (d *Dao) AddCacheRecord(c context.Context, record *api.GameRecords, params *model.NodeInfoParam) (err error) {
	if record == nil {
		return
	}
	if record.Buvid == "" {
		log.Error("AddCacheRecord Aid %d, GraphID %d, Mid %d, Empty Buvid, Params %+v", record.Aid, record.GraphId, record.Mid, params)
		return
	}
	var item []byte
	if item, err = record.Marshal(); err != nil {
		log.Error("record.Marshal error(%v)", err)
		return
	}
	key := recordKey(record.Mid, record.GraphId, record.Buvid)
	conn := d.rds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SETEX", key, d.recordExpire, item); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %+v) error(%v)", key, d.recordExpire, record, err)
		return
	}
	return
}

// CacheRecord .
func (d *Dao) CacheRecord(c context.Context, mid, graphID int64, buvid string) (a *api.GameRecords, err error) {
	var (
		key  = recordKey(mid, graphID, buvid)
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
	a = new(api.GameRecords)
	if err = a.Unmarshal(item); err != nil {
		log.Error("gameRecord Unmarshal error(%v)", err)
	}
	return
}

// CacheRecords 注意这里使用的是graphID作为map的key！！
func (d *Dao) CacheRecords(c context.Context, mid int64, graphIDs []int64, buvid string) (res map[int64]*api.GameRecords, err error) {
	var (
		conn  = d.rds.Get(c)
		args  = redis.Args{}
		items [][]byte
	)
	defer conn.Close()
	for _, gid := range graphIDs {
		args = args.Add(recordKey(mid, gid, buvid))
	}
	if items, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		log.Error("MGet conn.Do(mget) error(%v) args(%+v)", err, args)
		return
	}
	res = make(map[int64]*api.GameRecords, len(graphIDs))
	for _, item := range items {
		if item == nil { // omg
			continue
		}
		a := new(api.GameRecords)
		if e := a.Unmarshal(item); e != nil {
			log.Error("gameRecord Unmarshal error(%v)", e)
			continue
		}
		res[a.GraphId] = a
	}
	return
}

// AddCacheRecord .
func (d *Dao) AddCacheRecords(c context.Context, records map[int64]*api.GameRecords) (err error) {
	var (
		item []byte
		conn = d.rds.Get(c)
	)
	defer conn.Close()
	for _, record := range records {
		if item, err = record.Marshal(); err != nil {
			log.Error("record.Marshal error(%v)", err)
			return
		}
		key := recordKey(record.Mid, record.GraphId, record.Buvid)
		if err = conn.Send("SETEX", key, d.recordExpire, item); err != nil {
			log.Error("setex SEND key %s error(%v)", key, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		return
	}
	for range records {
		if _, err = conn.Receive(); err != nil {
			return
		}
	}
	return

}
