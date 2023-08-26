package archive

import (
	"context"
	"math/rand"
	"time"

	"go-common/library/cache"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	mdl "go-gateway/app/app-svr/archive/service/model"
)

// ArcsInner get data from cache if miss will call source method, then add to cache.
func (d *Dao) ArcsInner(c context.Context, aids []int64) (res map[int64]*arcmdl.ArcInternal, err error) {
	if len(aids) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheArcsInner(c, aids); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range aids {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	cache.MetricHits.Add(float64(len(aids)-len(miss)), "bts:ArcsInner")
	for k, v := range res {
		if v != nil && v.ID == -1 {
			delete(res, k)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*arcmdl.ArcInternal
	cache.MetricMisses.Add(float64(len(miss)), "bts:ArcsInner")
	missData, err = d.RawArcsInner(c, miss) //只返回存在的数据
	if res == nil {
		res = make(map[int64]*arcmdl.ArcInternal, len(aids))
	}
	for k, v := range missData {
		var ca = &arcmdl.ArcInternal{}
		*ca = *v //防止并发操作
		res[k] = ca
	}
	if err != nil {
		return
	}
	for _, key := range miss {
		if res[key] == nil {
			missData[key] = &arcmdl.ArcInternal{ID: -1, Aid: key}
		}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		_ = d.AddCacheArcsInner(c, missData)
	})
	return
}

// CacheArcsInner get data from redis
func (d *Dao) CacheArcsInner(c context.Context, ids []int64) (res map[int64]*arcmdl.ArcInternal, err error) {
	l := len(ids)
	if l == 0 {
		return
	}
	keysMap := make(map[string]int64, l)
	idxMap := make(map[int]string, l)
	args := redis.Args{}
	for idx, id := range ids {
		key := mdl.InternalArcKey(id)
		idxMap[idx] = key
		keysMap[key] = id
		args = args.Add(key)
	}
	conn := d.sArcRds.Get(c)
	defer conn.Close()
	values, err := redis.ByteSlices(conn.Do("MGET", args...))
	if err != nil {
		log.Errorc(c, "d.CacheArcsInner() error(%v)", err)
		return
	}
	for idx, temp := range values {
		if temp == nil {
			continue
		}
		v := &arcmdl.ArcInternal{}
		err = v.Unmarshal(temp)
		if err != nil {
			log.Errorc(c, "d.CacheArcsInner() err: %+v", err)
			return
		}
		if res == nil {
			res = make(map[int64]*arcmdl.ArcInternal, len(values))
		}
		key := idxMap[idx]
		res[keysMap[key]] = v
	}
	return
}

// AddCacheArcsInner Set data to redis
func (d *Dao) AddCacheArcsInner(c context.Context, values map[int64]*arcmdl.ArcInternal) error {
	if len(values) == 0 {
		return nil
	}
	rand.Seed(time.Now().UnixNano())
	exp := int64(36000)
	arcExp := exp + rand.Int63n(172800)
	kvMap := make(map[string][]byte)
	for id, val := range values {
		key := mdl.InternalArcKey(id)
		bs, err := val.Marshal()
		if err != nil {
			log.Errorc(c, "d.AddCacheArcsInner() err: %+v", err)
			return err
		}
		kvMap[key] = bs
	}
	return d.simpleMSetWithExp(c, kvMap, arcExp)
}

func (d *Dao) simpleMSetWithExp(c context.Context, kvMap map[string][]byte, exp int64) error {
	if len(kvMap) == 0 {
		return nil
	}
	var (
		conn        = d.sArcRds.Get(c)
		argsRecords = redis.Args{}
		keys        = make([]string, 0, len(kvMap))
	)
	defer conn.Close()
	for key, value := range kvMap {
		keys = append(keys, key)
		argsRecords = argsRecords.Add(key).Add(value)
	}
	// 使用redis pipeline批量提交
	if err := conn.Send("MSET", argsRecords...); err != nil {
		log.Errorc(c, "simpleMSetWithExp conn.Send() MSET keys(%+v) err(%+v)", keys, err)
		return err
	}
	for _, key := range keys {
		if err := conn.Send("EXPIRE", key, exp); err != nil {
			log.Errorc(c, "simpleMSetWithExp conn.Send() EXPIRE key(%+v) err(%+v)", key, err)
			return err
		}
	}
	if err := conn.Flush(); err != nil {
		log.Errorc(c, "simpleMSetWithExp conn.Flush() keys(%+v) err(%+v)", keys, err)
		return err
	}
	for i := 0; i < len(keys)+1; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Errorc(c, "simpleMSetWithExp conn.Receive() keys(%+v) err(%+v)", keys, err)
			return err
		}
	}
	return nil
}

func (d *Dao) RawArcsInner(c context.Context, aids []int64) (map[int64]*arcmdl.ArcInternal, error) {
	// form  taishan
	rly, miss, err := d.ArcsInnerFromTaishan(c, aids)
	if err != nil { //taishan错误，直接返回error，不回源
		return nil, err
	}
	//剔除reply中空缓存
	for k, v := range rly {
		if v != nil && v.ID == -1 {
			delete(rly, k)
		}
	}
	if len(miss) == 0 {
		return rly, nil
	}
	if rly == nil {
		rly = make(map[int64]*arcmdl.ArcInternal)
	}
	//暂时关闭回源db，依赖泰山数据
	//dbRly, err := d.RawInternals(c, miss)
	//if err != nil { //db错误，返回taishan查询接口和err,不回源
	//	return rly, err
	//}
	missData := make(map[int64]*arcmdl.ArcInternal, len(miss))
	for _, v := range miss {
		//if _, ok := dbRly[v]; ok {
		//	var ca = &arcmdl.ArcInternal{}
		//	*ca = *dbRly[v] //防止并发操作
		//	missData[v] = ca
		//	rly[v] = dbRly[v]
		//} else { //写入id=0的空缓存
		missData[v] = &arcmdl.ArcInternal{Aid: v, ID: -1}
		//}
	}
	//db 回源taishan
	d.cache.Do(c, func(c context.Context) {
		d.AddArcsInnerFromTaishan(c, missData)
	})
	return rly, nil
}

func (d *Dao) AddArcsInnerFromTaishan(c context.Context, vs map[int64]*arcmdl.ArcInternal) {
	if len(vs) == 0 {
		return
	}
	kvMap := make(map[string][]byte, len(vs))
	for aid, val := range vs {
		res, err := val.Marshal()
		if err != nil {
			log.Error("d.AddArcsInnerFromTaishan Marshal(%+v) error(%+v)", val, err)
			continue
		}
		kvMap[mdl.InternalArcKey(aid)] = res
	}
	if err := d.batchPutTaishan(c, kvMap); err != nil {
		log.Error("d.AddArcsInnerFromTaishan(%+v) error(%+v)", kvMap, err)
	}
}

func (d *Dao) ArcsInnerFromTaishan(c context.Context, aids []int64) (map[int64]*arcmdl.ArcInternal, []int64, error) {
	var (
		keys   []string
		keyMap = make(map[int64]struct{}, len(aids))
		missed []int64
	)
	for _, aid := range aids {
		if _, ok := keyMap[aid]; ok {
			continue
		}
		keyMap[aid] = struct{}{}
		keys = append(keys, mdl.InternalArcKey(aid))
	}
	bss, err := d.batchGetFromTaishan(c, keys)
	if err != nil {
		if ecode.EqualError(ecode.NothingFound, err) { //taishan miss all
			return make(map[int64]*arcmdl.ArcInternal), aids, nil
		}
		return nil, aids, err
	}
	am := make(map[int64]*arcmdl.ArcInternal, len(bss))
	for _, bs := range bss {
		a := &arcmdl.ArcInternal{}
		if err := a.Unmarshal(bs); err != nil {
			continue
		}
		am[a.Aid] = a
		delete(keyMap, a.Aid)
	}
	for aid := range keyMap {
		missed = append(missed, aid)
	}
	return am, missed, nil
}
