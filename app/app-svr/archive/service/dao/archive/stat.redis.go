package archive

import (
	"context"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/archive/service/api"
)

const (
	_prefixStatRedisPB = "stpr_"
)

func statPBRedisKey(aid int64) (key string) {
	return _prefixStatRedisPB + strconv.FormatInt(aid, 10)
}

// statCache3 get a archive stat from redis cache.
func (d *Dao) StatRedisCache(c context.Context, aid int64) (st *api.Stat, err error) {
	var (
		key  = statPBRedisKey(aid)
		conn = d.arcRds.Get(c)
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
	st = new(api.Stat)
	if err = st.Unmarshal(item); err != nil {
		log.Error("stat Unmarshal error(%v)", err)
		st = nil
		return
	}
	return
}

// statRedisCaches multi get archives stat, return map[aid]*Stat and missed aids.
func (d *Dao) statRedisCaches(c context.Context, aids []int64) (map[int64]*api.Stat, []int64, error) {
	var (
		keyMap = make(map[int64]struct{}, len(aids))
		args   = redis.Args{}
		cached = make(map[int64]*api.Stat, len(aids))
	)
	for _, aid := range aids {
		if _, ok := keyMap[aid]; ok {
			continue
		}
		keyMap[aid] = struct{}{}
		args = args.Add(statPBRedisKey(aid))
	}

	conn := d.arcRds.Get(c)
	defer conn.Close()
	bss, err := redis.ByteSlices(conn.Do("MGET", args...))
	if err != nil {
		return cached, aids, err
	}

	for _, bs := range bss {
		if bs == nil {
			continue
		}
		st := &api.Stat{}
		if err = st.Unmarshal(bs); err != nil {
			log.Error("statRedisCaches Unmarshal error(%+v)", err)
			continue
		}
		cached[st.Aid] = st
	}
	var missed []int64
	for _, aid := range aids {
		if _, ok := cached[aid]; !ok {
			missed = append(missed, aid)
		}
	}
	return cached, missed, nil
}

func (d *Dao) addStatRedisCache(c context.Context, st *api.Stat) (err error) {
	if st == nil {
		return
	}
	var item []byte
	if item, err = st.Marshal(); err != nil {
		log.Error("st.Marshal error(%v)", err)
		return
	}
	key := statPBRedisKey(st.Aid)
	conn := d.arcRds.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, item); err != nil {
		log.Error("conn.Do(SET, %s, %+v) error(%v)", key, st, err)
	}
	return
}
