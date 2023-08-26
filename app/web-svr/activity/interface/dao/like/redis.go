package like

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/interface/component"
	dynmdl "go-gateway/app/web-svr/activity/interface/model/dynamic"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_prefixAttention = "lg_"
	_esLikesKey      = "go_es_lsn_%d"
	_esTypeLikesKey  = "go_es_to_lsn_%d_%d"
	_sfEsLilesKey    = "go_sf_es_lks_%d_%d"
	_reserveTotalKey = "go_reserve_total_%d"
	_actDomainList   = "act_domain_prefix:list"
)

var (
	reserveCount42021Player int64
)

func redisKey(key string) string {
	return _prefixAttention + key
}

func reserveTotalKey(sid int64) string {
	return fmt.Sprintf(_reserveTotalKey, sid)
}

func esLikesKey(sid, ltype int64) string {
	if ltype > 0 {
		return fmt.Sprintf(_esTypeLikesKey, sid, ltype)
	}
	return fmt.Sprintf(_esLikesKey, sid)
}

// cacheSFActEsLikesIDs .
func (dao *Dao) cacheSFActEsLikesIDs(sid, ltype int64, _, _ int64) string {
	return fmt.Sprintf(_sfEsLilesKey, ltype, sid)
}

// AddCacheReservesTotal .
func (dao *Dao) AddCacheReservesTotal(c context.Context, miss map[int64]int64) (err error) {
	if len(miss) == 0 {
		return
	}
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	var reserveKey []string
	args := redis.Args{}
	for k, v := range miss {
		keyStr := reserveTotalKey(k)
		args = args.Add(keyStr).Add(v)
		reserveKey = append(reserveKey, keyStr)
	}
	var count int
	if err = conn.Send("MSET", args...); err != nil {
		log.Error("AddCacheReservesTotal redis.Ints(conn.Do(MSET,%v) error(%v)", miss, err)
		return
	}
	count++
	for _, v := range reserveKey {
		if err = conn.Send("EXPIRE", v, dao.likeTotalExpire); err != nil {
			log.Error("AddCacheReservesTotal EXPIRE %v error(%v)", miss, err)
			return
		}
		count++
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheReservesTotal Flush %v error(%v)", miss, err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheReservesTotal Receive %v error(%v)", miss, err)
			return
		}
	}
	return
}

func (dao *Dao) ResetPlayerReserveTotal(ctx context.Context, id int64) (count int64, err error) {
	m := make(map[int64]int64, 0)
	m, err = dao.RawReservesTotal(ctx, []int64{id})
	if err == nil {
		count, _ = m[id]
		reserveCount42021Player = count
	}

	return
}

// CacheReservesTotal .
func (dao *Dao) CacheReservesTotal(c context.Context, sids []int64) (rly map[int64]int64, err error) {
	if len(sids) == 0 {
		return
	}
	var (
		args = redis.Args{}
		ss   []int64
	)
	for _, sid := range sids {
		args = args.Add(reserveTotalKey(sid))
	}
	if ss, err = redis.Int64s(component.GlobalRedis.Do(c, "MGET", args...)); err != nil {
		err = errors.Wrapf(err, "redis.Ints(conn.Do(MGET,%v)", args)
		return
	}
	rly = make(map[int64]int64, len(sids))
	for key, val := range ss {
		if val == 0 {
			continue
		}

		rly[sids[key]] = val
	}

	return

}

// IncrCacheReserveTotal .
func (dao *Dao) IncrCacheReserveTotal(c context.Context, sid int64, num int32) (err error) {
	var ok bool
	key := reserveTotalKey(sid)
	if ok, err = redis.Bool(component.GlobalRedis.Do(c, "EXPIRE", key, dao.likeTotalExpire)); err != nil {
		log.Errorc(c, "IncrCacheReserveTotal conn.Do(EXPIRE) key(%s) error(%v)", key, err)
		return
	}
	if !ok {
		return
	}
	if _, err = component.GlobalRedis.Do(c, "INCRBY", key, num); err != nil {
		log.Errorc(c, "IncrCacheReserveTotal conn.Do(INCR key(%s)) error(%v)", key, err)
	}
	return
}

// RsSet Dao
func (dao *Dao) RsSet(c context.Context, key string, value string) (err error) {
	var (
		rkey = redisKey(key)
	)
	if _, err = component.GlobalRedisStore.Do(c, "SET", rkey, value); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", rkey, value, err)
		return
	}
	return
}

// RsGet Dao
func (dao *Dao) RsGet(c context.Context, key string) (res string, err error) {
	var (
		rkey = redisKey(key)
	)
	if res, err = redis.String(component.GlobalRedisStore.Do(c, "GET", rkey)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET %s) error(%v)", rkey, err)
		}
	}
	return
}

// RiGet get int value.
func (dao *Dao) RiGet(c context.Context, key string) (res int, err error) {
	var (
		rkey = redisKey(key)
	)
	if res, err = redis.Int(component.GlobalRedisStore.Do(c, "GET", rkey)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET key(%s)) error(%v)", rkey, err)
		}
		return
	}
	return
}

// RsSetNX Dao
func (dao *Dao) RsSetNX(c context.Context, key string, expire int32) (res bool, err error) {
	var (
		rkey = redisKey(key)
		rly  interface{}
	)
	if rly, err = component.GlobalRedisStore.Do(c, "SET", rkey, "1", "EX", expire, "NX"); err != nil {
		log.Error("conn.Do(GET key(%s)) error(%v)", rkey, err)
		return
	}
	if rly != nil {
		res = true
	}
	return
}

// RsDelNx .
func (dao *Dao) RsDelNX(c context.Context, key string) (err error) {
	_, err = component.GlobalRedisStore.Do(c, "DEL", redisKey(key))
	return
}

// Rb Dao
func (dao *Dao) Rb(c context.Context, key string) (res []byte, err error) {
	var (
		rkey = redisKey(key)
	)
	if res, err = redis.Bytes(component.GlobalRedisStore.Do(c, "GET", rkey)); err != nil {
		if err == redis.ErrNil {
			res = nil
			err = nil
		} else {
			log.Error("conn.Do(GET key(%v)) error(%v)", rkey, err)
		}
	}
	return
}

// Incr Dao
func (dao *Dao) Incr(c context.Context, key string) (res bool, err error) {
	var (
		rkey = redisKey(key)
	)
	if res, err = redis.Bool(component.GlobalRedisStore.Do(c, "INCR", rkey)); err != nil {
		log.Error("conn.Do(INCR key(%s)) error(%v)", rkey, err)
	}
	return
}

// IncrWithExpire Dao
func (dao *Dao) IncrWithExpire(c context.Context, key string, expire int32) (res bool, err error) {
	var (
		rkey = redisKey(key)
		conn = component.GlobalRedisStore.Conn(c)
	)
	defer conn.Close()
	if err = conn.Send("INCR", rkey); err != nil {
		log.Error("IncrWithExpire conn.Do(INCR key(%s)) error(%v)", rkey, err)
		return
	}
	if err = conn.Send("EXPIRE", rkey, expire); err != nil {
		log.Error("IncrWithExpire conn.Do(expire key(%s)) error(%v)", rkey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("IncrWithExpire conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("IncrWithExpire Receive error(%v)", err)
			return
		}
	}
	return
}

// Incrby Dao
func (dao *Dao) Incrby(c context.Context, key string) (res bool, err error) {
	var (
		rkey = redisKey(key)
	)
	if res, err = redis.Bool(component.GlobalRedisStore.Do(c, "INCRBY", rkey, 222)); err != nil {
		log.Error("conn.Do(INCRBY key(%s)) error(%v)", rkey, err)
	}
	return
}

// CacheActEsLikesIDs .
func (dao *Dao) CacheActEsLikesIDs(c context.Context, sid, ltype, start, end int64) (*lmdl.EsLikesReply, error) {
	var (
		key = esLikesKey(sid, ltype)
	)
	return dao.zrangeCommon(c, start, end, key)
}

// zrevrangeCommon .
func (dao *Dao) zrangeCommon(c context.Context, start, end int64, key string) (res *lmdl.EsLikesReply, err error) {
	var (
		conn = component.GlobalRedisStore.Conn(c)
	)
	defer conn.Close()
	if err = conn.Send("ZRANGE", key, start, end); err != nil {
		log.Error("zrangeCommon conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("ZCARD", key); err != nil {
		log.Error("zrangeCommon conn.Do(ZCARD, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("ZRANGE", key, 0, 1); err != nil {
		log.Error("zrangeCommon conn.Do(ZRANGE, %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("zrangeCommon conn.Flush() error(%v)", err)
		return
	}
	var (
		lids, checkLids []int64
		count           int64
	)
	if lids, err = redis.Int64s(conn.Receive()); err != nil {
		log.Error("ModuleCache conn.Receive() error(%v)", err)
		return
	}
	if count, err = redis.Int64(conn.Receive()); err != nil {
		log.Error("ModuleCache conn.Receive() error(%v)", err)
		return
	}
	if checkLids, err = redis.Int64s(conn.Receive()); err != nil {
		log.Error("ModuleCache conn.Receive() error(%v)", err)
		return
	}
	if count > 0 {
		res = &lmdl.EsLikesReply{Lids: lids, Count: count}
		// 空缓存过滤
		if count == 1 && len(checkLids) == 1 && checkLids[0] == -1 {
			res.Count = 0
		}
	}
	return
}

// zrevrangeCommon .
func (dao *Dao) zrevrangeCommon(c context.Context, start, end int64, key string) (res []int64, err error) {
	if res, err = redis.Int64s(component.GlobalRedisStore.Do(c, "ZREVRANGE", key, start, end)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrap(err, "conn.Do(ZREVRANGE)")
		}
	}
	return
}

// RawActEsLikesIDs .
func (dao *Dao) RawActEsLikesIDs(c context.Context, sid, ltype, start, end int64) (*lmdl.EsLikesReply, *lmdl.EsLikesReply, error) {
	var lidsMap []*dynmdl.EsLikesReply
	mu := sync.Mutex{}
	eg := errgroup.WithContext(c)
	for i := 1; i < 5; i++ {
		temp := i
		eg.Go(func(ctx context.Context) error {
			if lids := dao.RawEsLikesIDs(ctx, sid, ltype, temp); len(lids) > 0 {
				mu.Lock()
				lidsMap = append(lidsMap, lids...)
				mu.Unlock()
			}
			return nil
		})
	}
	eg.Wait()
	sort.Slice(lidsMap, func(i, j int) bool {
		if lidsMap[i].Score == lidsMap[j].Score {
			return lidsMap[i].Lid > lidsMap[j].Lid
		}
		return lidsMap[i].Score > lidsMap[j].Score
	})
	missIDs := make([]int64, 0)
	for _, v := range lidsMap {
		missIDs = append(missIDs, v.Lid)
	}
	lidLent := int64(len(missIDs))
	if end == -1 {
		end = lidLent
	} else {
		end += 1
		if lidLent < end {
			end = lidLent
		}
	}
	var lids []int64
	if start > end {
		lids = []int64{}
	} else {
		lids = missIDs[start:end]
	}
	count := int64(len(missIDs))
	var missRly *lmdl.EsLikesReply
	if len(missIDs) > 0 {
		missRly = &lmdl.EsLikesReply{Lids: missIDs, Count: count}
	}
	return &lmdl.EsLikesReply{Lids: lids, Count: count}, missRly, nil
}

// RawEsLikesIDs .
func (dao *Dao) RawEsLikesIDs(c context.Context, sid, ltype int64, pn int) (lidReply []*dynmdl.EsLikesReply) {
	esRes, err := dao.ListFromES(c, sid, EsOrderLikes, 500, pn, 0, ltype)
	if err != nil {
		log.Error("RawEsLikesIDs:d.ListFromES(%d) error(%+v)", sid, err)
		return
	}
	// 没有数据直接return
	if esRes == nil || len(esRes.List) == 0 {
		return
	}
	for _, val := range esRes.List {
		lidReply = append(lidReply, &dynmdl.EsLikesReply{Lid: val.Item.ID, Score: val.Likes})
	}
	return
}

// AddCacheEsLikesIDs .
func (dao *Dao) AddCacheActEsLikesIDs(c context.Context, sid int64, val *lmdl.EsLikesReply, ltype int64) (err error) {
	var (
		conn = component.GlobalRedisStore.Conn(c)
		key  = esLikesKey(sid, ltype)
		lids = val.Lids
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for k, v := range lids {
		args = args.Add(k).Add(v)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddCacheEsLikesIDs conn.Send(%v)", err)
	}
	expire := dao.EsLikesExpire
	if len(lids) == 1 && lids[0] == -1 {
		expire = 30
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		log.Error("conn.Send(EXPIRE %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush(%s) error(%v)", key, err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// RsNXGet nx get.
func (dao *Dao) RsNXGet(c context.Context, k string) (res string, err error) {
	key := redisKey(k)
	if res, err = redis.String(component.GlobalRedisStore.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET %s) error(%v)", key, err)
		}
	}
	return
}

func (dao *Dao) DeleteActSubjectCache(c context.Context, id int64) (err error) {
	var (
		key = actSubjectKey(id)
	)
	for i := 0; i < 3; i++ {
		err = component.GlobalMC.Delete(c, key)
		if err != nil {
			log.Error("DeleteActSubjectCache conn.Do(DEL, %s) error(%v)", key, err)
		} else {
			break
		}
	}
	return
}

func (dao *Dao) DeleteActSubjectWithStateCache(c context.Context, id int64) (err error) {
	var (
		key = actSubjectWithStateKey(id)
	)
	for i := 0; i < 3; i++ {
		err = component.GlobalMC.Delete(c, key)
		if err != nil {
			log.Error("DeleteActSubjectWithStateCache conn.Do(DEL, %s) error(%v)", key, err)
		} else {
			break
		}
	}
	return
}

func (dao *Dao) HGetAllDomain(ctx context.Context) (list []*lmdl.Record, err error) {
	var (
		strMap map[string]string
		conn   = component.GlobalRedisStore.Conn(ctx)
	)
	defer conn.Close()
	if strMap, err = redis.StringMap(conn.Do("HGETALL", _actDomainList)); err != nil {
		if err != redis.ErrNil {
			log.Errorc(ctx, "HGetAllDomain conn.Do(HGETALL %s) error(%v)", _actDomainList, err)
			return
		}
	}

	for _, v := range strMap {
		value := &lmdl.Record{}
		if jsonErr := json.Unmarshal([]byte(v), value); jsonErr != nil {
			log.Errorc(ctx, "HGetAllDomain json Unmarshal err:%v , value:%v", jsonErr, value)
			continue
		}
		if value == nil || value.Etime.Time().Before(time.Now()) || value.Stime.Time().After(time.Now()) {
			log.Infoc(ctx, "HGetAllDomain activity not  start or is over :%v , nowtime:%v", v, time.Now())
			continue
		}
		list = append(list, value)
	}
	return
}
