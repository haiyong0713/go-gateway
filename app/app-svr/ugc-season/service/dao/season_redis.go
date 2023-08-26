package dao

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/time"

	"go-gateway/app/app-svr/ugc-season/service/api"
)

const (
	_prefixUpper            = "up_%d"
	_prefixUpperCnt         = "uc_%d"
	_upperCacheTime         = 86400 * 7
	_upperNoSeasonCacheTime = 60 * 30 //30分钟
)

func upperSeason(mid int64) string {
	return fmt.Sprintf(_prefixUpper, mid)
}

func upperNoSeason(mid int64) string {
	return fmt.Sprintf(_prefixUpperCnt, mid)
}

// UpperSeasonCache is
func (d *Dao) UpperSeasonCache(c context.Context, mid, start, end int64) (sids []int64, total int64, err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("EXPIRE", upperSeason(mid), _upperCacheTime); err != nil {
		log.Error("conn.Send(EXPIRE, %d, %d) error(%+v)", mid, _upperCacheTime, err)
		return
	}
	if err = conn.Send("EXISTS", upperNoSeason(mid)); err != nil {
		log.Error("conn.Send(EXISTS, %d) error(%+v)", mid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%+v)", err)
		return
	}
	var (
		exist    bool //up主当前没有剧集
		noSeason bool //up主从未投过剧集
	)
	if exist, err = redis.Bool(conn.Receive()); err != nil {
		log.Error("conn.Receive error(%+v)", err)
		return
	}
	if noSeason, err = redis.Bool(conn.Receive()); err != nil {
		log.Error("conn.Receive error(%+v)", err)
		return
	}
	if noSeason {
		return
	}
	if !exist {
		total = -1
		return
	}
	if err = conn.Send("ZREVRANGE", upperSeason(mid), start, end); err != nil {
		log.Error("conn.Send(ZREVRANGE, %d, %d, %d) error(%+v)", mid, start, end, err)
		return
	}
	if err = conn.Send("ZCARD", upperSeason(mid)); err != nil {
		log.Error("conn.Send(ZCARD, %d) error(%+v)", mid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%+v)", err)
		return
	}
	if sids, err = redis.Int64s(conn.Receive()); err != nil {
		log.Error("redis.Int64s(conn.Receive) error(%+v)", err)
		return
	}
	if total, err = redis.Int64(conn.Receive()); err != nil {
		log.Error("redis.Int64 error(%+v)", err)
		return
	}
	return
}

// AddUpperSeasonCache is
func (d *Dao) AddUpperSeasonCache(c context.Context, mid int64, sids []int64, ptimes []time.Time) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	for k, sid := range sids {
		ptime := ptimes[k]
		if err = conn.Send("ZADD", upperSeason(mid), ptime, sid); err != nil {
			log.Error("conn.Send(ZADD, %s, %d, %d) error(%+v)", upperSeason(mid), ptime, sid, err)
			return
		}
	}
	if err = conn.Send("DEL", upperNoSeason(mid)); err != nil {
		log.Error("conn.Send(DEL, %d) error(%+v)", mid, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush err(%v)", err)
		return
	}
	for _, sid := range sids {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive ZADD sid(%d) mid(%d) err(%v)", sid, mid, err)
		}
	}
	if _, err = conn.Receive(); err != nil {
		log.Error("conn.Receive DEL mid(%d) err(%v)", mid, err)
	}
	return
}

// SetUpperNoSeasonCache is
func (d *Dao) SetUpperNoSeasonCache(c context.Context, mid int64) (err error) {
	conn := d.redis.Get(c)
	if _, err := conn.Do("SETEX", upperNoSeason(mid), _upperNoSeasonCacheTime, 1); err != nil {
		log.Error("conn.Do(SETEX, %d) error(%+v)", mid, err)
	}
	conn.Close()
	return
}

func (d *Dao) AddStCache(c context.Context, st *api.Stat) error {
	key := seasonStatKey(st.SeasonID)
	bs, err := st.Marshal()
	if err != nil {
		log.Error("AddStatCache error(%+v)", err)
		return err
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SETNX", key, bs); err != nil {
		log.Error("rdsConn.Do(SETNX, %s) error(%+v)", key, err)
		return err
	}
	return nil
}

func (d *Dao) AddStsCache(c context.Context, stats map[int64]*api.Stat) error {
	conn := d.redis.Get(c)
	defer conn.Close()
	for _, st := range stats {
		bs, err := st.Marshal()
		if err != nil {
			log.Error("st.Marshal error(%+v)", err)
			return err
		}
		key := seasonStatKey(st.SeasonID)
		if err = conn.Send("SETNX", key, bs); err != nil {
			log.Error("conn.Send(%s) error(%+v)", key, err)
			return err
		}
	}
	if err := conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%+v)", err)
		return err
	}
	return nil
}

// StCache get a season stat from cache.
func (d *Dao) StCache(c context.Context, sid int64) (ss *api.Stat, err error) {
	var (
		key  = seasonStatKey(sid)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		log.Error("conn.Do(GET,%s) error(%+v)", key, err)
		return nil, err
	}
	ss = new(api.Stat)
	if err = ss.Unmarshal(bs); err != nil {
		log.Error("Stat.Unmarshal error(%+v)", err)
		return nil, err
	}
	return ss, nil
}

// StsCache is
func (d *Dao) StsCache(c context.Context, sids []int64) (map[int64]*api.Stat, error) {
	if len(sids) == 0 {
		return nil, nil
	}
	keys := redis.Args{}
	keySidMap := make(map[string]int64, len(sids))
	for _, sid := range sids {
		key := seasonStatKey(sid)
		if _, ok := keySidMap[key]; !ok {
			// duplicate mid
			keySidMap[key] = sid
			keys = keys.Add(key)
		}
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	bss, err := redis.ByteSlices(conn.Do("MGET", keys...))
	if err != nil {
		return nil, err
	}
	stats := make(map[int64]*api.Stat, len(bss))
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		st := &api.Stat{}
		if err := st.Unmarshal(bs); err != nil {
			log.Error("stat.Unmarshal error(%+v)", err)
			continue
		}
		stats[st.SeasonID] = st
	}
	return stats, nil
}

// SeasonRdsCache get a season info from cache.
func (d *Dao) SeasonRdsCache(c context.Context, sid int64) (*api.Season, error) {
	key := seasonKey(sid)
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	season := new(api.Season)
	if err = season.Unmarshal(bs); err != nil {
		return nil, err
	}
	return season, nil
}

// SeasonsRdsCache is
func (d *Dao) SeasonsRdsCache(c context.Context, sids []int64) (map[int64]*api.Season, error) {
	keys := redis.Args{}
	keySidMap := make(map[string]int64, len(sids))
	for _, sid := range sids {
		key := seasonKey(sid)
		if _, ok := keySidMap[key]; !ok {
			// duplicate mid
			keySidMap[key] = sid
			keys = keys.Add(key)
		}
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	bss, err := redis.ByteSlices(conn.Do("MGET", keys...))
	if err != nil {
		return nil, err
	}
	seasons := make(map[int64]*api.Season, len(sids))
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		s := &api.Season{}
		if err = s.Unmarshal(bs); err != nil {
			log.Error("season.Unmarshal error(%+v)", err)
			continue
		}
		seasons[s.ID] = s
	}
	return seasons, nil
}

// ViewRdsCache is
func (d *Dao) ViewRdsCache(c context.Context, sid int64) (*api.View, error) {
	key := viewKey(sid)
	conn := d.redis.Get(c)
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	v := new(api.View)
	if err = v.Unmarshal(bs); err != nil {
		return nil, err
	}
	return v, nil
}

// ViewsRdsCache is
func (d *Dao) ViewsRdsCache(c context.Context, sids []int64) (map[int64]*api.View, error) {
	views := make(map[int64]*api.View, len(sids))

	keys := redis.Args{}
	keySidMap := make(map[string]int64, len(sids))
	for _, sid := range sids {
		key := viewKey(sid)
		if _, ok := keySidMap[key]; !ok {
			keySidMap[key] = sid
			keys = keys.Add(key)
		}
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	bss, err := redis.ByteSlices(conn.Do("MGET", keys...))
	if err != nil {
		return views, err
	}
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		v := &api.View{}
		if err = v.Unmarshal(bs); err != nil {
			log.Error("seasonView.Unmarshal error(%+v)", err)
			continue
		}
		if v.Season != nil && v.Season.ID > 0 {
			views[v.Season.ID] = v
		}
	}
	return views, nil
}
