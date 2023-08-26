package archive

import (
	"context"
	"math/rand"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model"
)

func (d *Dao) arcCache(c context.Context, aid int64) (a *api.Arc, err error) {
	var (
		key  = model.ArcKey(aid)
		bs   []byte
		conn = d.arcRds.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		return nil, err
	}
	a = &api.Arc{}
	if err = a.Unmarshal(bs); err != nil {
		return nil, err
	}
	return a, nil
}

// setArcRdsCache set archive into cache.
func (d *Dao) setArcRdsCache(c context.Context, a *api.Arc) (err error) {
	var (
		key  = model.ArcKey(a.Aid)
		conn = d.arcRds.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = a.Marshal(); err != nil {
		log.Error("Arc marshal error(%v)", err)
		return
	}
	// a3p_{aid}缓存设置过期时间，默认10小时+48小时内随机数
	rand.Seed(time.Now().UnixNano())
	exp := int64(36000)
	arcExp := exp + rand.Int63n(172800)
	if _, err = conn.Do("SET", key, bs, "EX", arcExp); err != nil {
		log.Error("conn.Do(SET, %s) error(%v)", key, err)
		return
	}
	return
}

// arcCaches multi get archives, return cached map[aid]*Archive
func (d *Dao) arcCaches(c context.Context, aids []int64) (map[int64]*api.Arc, []int64, error) {
	var (
		missed []int64
		args   = redis.Args{}
		keyMap = make(map[int64]struct{}, len(aids))
	)
	for _, aid := range aids {
		if _, ok := keyMap[aid]; ok {
			continue
		}
		keyMap[aid] = struct{}{}
		args = args.Add(model.ArcKey(aid))
	}
	conn := d.arcRds.Get(c)
	defer conn.Close()
	bss, err := redis.ByteSlices(conn.Do("MGET", args...))
	if err != nil {
		return nil, aids, err
	}
	am := make(map[int64]*api.Arc, len(bss))
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		a := &api.Arc{}
		if err = a.Unmarshal(bs); err != nil {
			log.Error("%+v", err)
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

func (d *Dao) sArcCache(c context.Context, aid int64) (*api.SimpleArc, error) {
	conn := d.sArcRds.Get(c)
	defer conn.Close()
	bs, err := redis.Bytes(conn.Do("GET", model.SimpleArcKey(aid)))
	if err != nil {
		return nil, err
	}
	d.hitProm.Incr("simpleArc")
	a := &api.SimpleArc{}
	if err = a.Unmarshal(bs); err != nil {
		d.errProm.Incr("sArcCache_Unmarshal")
		return nil, err
	}
	return a, nil
}

// setSArcCache set simple archive into cache.
func (d *Dao) setSArcCache(c context.Context, a *api.SimpleArc) error {
	conn := d.sArcRds.Get(c)
	defer conn.Close()
	bs, err := a.Marshal()
	if err != nil {
		return err
	}
	_, err = conn.Do("SET", model.SimpleArcKey(a.Aid), bs)
	return err
}

// sArcCaches multi get simple archives
func (d *Dao) sArcCaches(c context.Context, aids []int64) (map[int64]*api.SimpleArc, error) {
	args := redis.Args{}
	keysMap := make(map[int64]struct{})
	for _, aid := range aids {
		if _, ok := keysMap[aid]; ok {
			continue
		}
		args = args.Add(model.SimpleArcKey(aid))
		keysMap[aid] = struct{}{}
	}
	conn := d.sArcRds.Get(c)
	defer conn.Close()
	bss, err := redis.ByteSlices(conn.Do("MGET", args...))
	if err != nil {
		log.Error("conn.Do(MGET, %+v) error(%v)", args, err)
		return nil, err
	}
	sas := make(map[int64]*api.SimpleArc)
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		a := &api.SimpleArc{}
		if err = a.Unmarshal(bs); err != nil {
			d.errProm.Incr("sArcCaches_Unmarshal")
			log.Error("sArcCaches Unmarshal error(%v)", err)
			continue
		}
		sas[a.Aid] = a
	}
	d.hitProm.Add("batchArc", int64(len(sas)))
	d.missProm.Add("batchArc", int64(len(keysMap)-len(sas)))
	return sas, nil
}
