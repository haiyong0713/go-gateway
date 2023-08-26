package dao

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/archive-honor/service/api"
)

const (
	_prefixHonor = "ah_%d"
)

func honorKey(aid int64) string {
	return fmt.Sprintf(_prefixHonor, aid)
}

// HonorsCacheByAid is
func (d *Dao) HonorsCacheByAid(c context.Context, aid int64) (honors map[int32]*api.Honor, noCache bool, err error) {
	var (
		values map[string]string
		conn   = d.redis.Get(c)
	)
	honors = make(map[int32]*api.Honor)
	defer conn.Close()
	if values, err = redis.StringMap(conn.Do("HGETALL", honorKey(aid))); err != nil {
		log.Error("redis.Values(conn.Do(HGETALL)) error(%v) key(%s)", err, honorKey(aid))
		return
	}
	if len(values) == 0 {
		noCache = true
		return
	}
	for k, v := range values {
		var typ int64
		typ, err = strconv.ParseInt(k, 10, 64)
		h := &api.Honor{}
		if err = h.Unmarshal([]byte(v)); err != nil {
			log.Error("h.Unmarshal(%s) error(%v)", v, err)
			return
		}
		honors[int32(typ)] = h
	}
	return
}

// AddHonorCache is
func (d *Dao) AddHonorCache(c context.Context, aid int64, honor *api.Honor) (err error) {
	if honor == nil {
		return
	}
	var (
		hm   []byte
		conn = d.redis.Get(c)
	)
	if hm, err = honor.Marshal(); err != nil {
		log.Error("honor.Marshal honor(%v) err(%v)", honor, err)
		return
	}
	defer conn.Close()
	if _, err = conn.Do("HSET", honorKey(aid), honor.Type, hm); err != nil {
		log.Error("conn.Do(HSET, %s, %v) error(%v)", honorKey(aid), honor, err)
		return
	}
	return
}

// AddHonorsCache is
func (d *Dao) AddHonorsCache(c context.Context, aid int64, honors map[int32]*api.Honor) (err error) {
	var (
		conn = d.redis.Get(c)
		args = redis.Args{}.Add(honorKey(aid))
	)
	defer conn.Close()
	for k, h := range honors {
		var hm []byte
		if hm, err = h.Marshal(); err != nil {
			log.Error("honor.Marshal honor(%v) err(%v)", h, err)
			continue
		}
		args = args.Add(k).Add(hm)
	}
	if _, err = conn.Do("HMSET", args...); err != nil {
		log.Error("conn.Do(HMSET, %s) error(%v)", honorKey(aid), err)
		return
	}
	return
}

// DelHonorCache is
func (d *Dao) DelHonorCache(c context.Context, aid int64, typ int32) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("HDEL", honorKey(aid), typ); err != nil {
		log.Error("conn.Do(HDEL, %s, %d) error(%v)", honorKey(aid), typ, err)
		return
	}
	return
}

// HonorsCacheByAids is
func (d *Dao) HonorsCacheByAids(c context.Context, aids []int64) (honors map[int64]map[int32]*api.Honor, noCacheAids []int64, err error) {
	honors = make(map[int64]map[int32]*api.Honor)
	conn := d.redis.Get(c)
	defer conn.Close()
	for _, aid := range aids {
		if err = conn.Send("HGETALL", honorKey(aid)); err != nil {
			log.Error("conn.Send HGETALL key(%s) err(%v)", honorKey(aid), err)
			err = nil
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush err(%v)", err)
		return
	}
	for _, aid := range aids {
		var (
			honorStr map[string]string
			hs       = make(map[int32]*api.Honor)
		)
		if honorStr, err = redis.StringMap(conn.Receive()); err != nil {
			log.Error("conn.Receive aid(%d) err(%v)", aid, err)
			err = nil
			continue
		}
		if len(honorStr) == 0 {
			noCacheAids = append(noCacheAids, aid)
			continue
		}
		for k, v := range honorStr {
			var typ int64
			typ, err = strconv.ParseInt(k, 10, 64)
			h := &api.Honor{}
			if err = h.Unmarshal([]byte(v)); err != nil {
				log.Error("h.Unmarshal(%s) error(%v)", v, err)
				continue
			}
			hs[int32(typ)] = h
		}
		honors[aid] = hs
	}
	return
}
