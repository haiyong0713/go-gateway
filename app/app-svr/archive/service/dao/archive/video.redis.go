package archive

import (
	"context"
	"math/rand"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model"
)

// addPageCache add pages cache by aid.
func (d *Dao) addPageCache(c context.Context, aid int64, ps []*api.Page) (err error) {
	var (
		vs   = &api.AidVideos{Aid: aid, Pages: ps}
		key  = model.PageKey(aid)
		conn = d.arcRds.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = vs.Marshal(); err != nil {
		log.Error("addPageCache Marshal error(%v)", err)
		return
	}
	// psb_{aid}缓存设置过期时间，默认10小时+48小时内随机数
	rand.Seed(time.Now().UnixNano())
	exp := int64(36000)
	psbExp := exp + rand.Int63n(172800)
	if _, err = conn.Do("SET", key, bs, "EX", psbExp); err != nil {
		log.Error("conn.Do(SET, %s) error(%v)", key, err)
		return
	}
	return
}

// pageCache get page cache by aid.
func (d *Dao) pageCache(c context.Context, aid int64) (ps []*api.Page, err error) {
	var (
		key  = model.PageKey(aid)
		conn = d.arcRds.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		return nil, err
	}
	vs := &api.AidVideos{}
	if err = vs.Unmarshal(bs); err != nil {
		return nil, err
	}
	return vs.Pages, nil
}

// pagesCache get pages cache by aids
func (d *Dao) pagesCache(c context.Context, aids []int64) (cached map[int64][]*api.Page, missed []int64, err error) {
	var (
		conn    = d.arcRds.Get(c)
		bss     [][]byte
		args    = redis.Args{}
		keysMap = make(map[int64]struct{})
	)
	cached = make(map[int64][]*api.Page)
	defer conn.Close()
	for _, aid := range aids {
		if _, ok := keysMap[aid]; ok {
			continue
		}
		args = args.Add(model.PageKey(aid))
		keysMap[aid] = struct{}{}
	}
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		return nil, aids, err
	}
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		vs := &api.AidVideos{}
		if err = vs.Unmarshal(bs); err != nil {
			d.errProm.Incr("pagesCache_Unmarshal")
			log.Error("pagesCache Unmarshal error(%+v)", err)
			continue
		}
		cached[vs.Aid] = vs.Pages
		delete(keysMap, vs.Aid)
	}
	for aid := range keysMap {
		missed = append(missed, aid)
	}
	d.hitProm.Add("pagesCache", int64(len(cached)))
	d.missProm.Add("pagesCache", int64(len(missed)))
	return cached, missed, nil
}

// videoCache get video cache by aid & cid.
func (d *Dao) videoCache(c context.Context, aid, cid int64) (p *api.Page, err error) {
	var (
		key  = model.VideoKey(aid, cid)
		bs   []byte
		conn = d.arcRds.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		return nil, err
	}
	p = &api.Page{}
	if err = p.Unmarshal(bs); err != nil {
		log.Error("videoCache Unmarshal error(%v)", err)
		return nil, err
	}
	return p, nil
}

// videoAidCidsCache get video cache by aid & cid.
func (d *Dao) videoAidCidsCache(c context.Context, aidCids map[int64][]int64) (map[int64][]*api.Page, map[int64][]int64, error) {
	conn := d.arcRds.Get(c)
	args := redis.Args{}
	cached := make(map[int64][]*api.Page)
	tmpMap := make(map[int64]int64)

	defer conn.Close()
	for aid, cids := range aidCids {
		if aid == 0 || len(cids) == 0 {
			continue
		}
		for _, cid := range cids {
			if cid == 0 {
				continue
			}
			args = args.Add(model.VideoKey(aid, cid))
			tmpMap[cid] = aid
		}
	}

	if len(args) == 0 {
		missMap := make(map[int64][]int64, len(tmpMap))
		for cid, aid := range tmpMap {
			missMap[aid] = append(missMap[aid], cid)
		}
		d.infoProm.Incr("videoAidCidsCache_no_args")
		return cached, missMap, nil
	}

	bss, err := redis.ByteSlices(conn.Do("MGET", args...))
	if err != nil {
		log.Error("videoAidCidsCache MGET args(%+v) error(%+v)", args, err)
		return nil, nil, err
	}

	cachedCids := sets.NewInt64()
	for index, bs := range bss {
		if bs == nil {
			continue
		}
		tmp := &api.Page{}
		if err = tmp.Unmarshal(bs); err != nil {
			log.Error("videoAidCidsCache Unmarshal aidCids(%+v) index(%+v) error(%+v)", aidCids, index, err)
			continue
		}
		if aid, ok := tmpMap[tmp.Cid]; ok {
			cached[aid] = append(cached[aid], tmp)
			cachedCids.Insert(tmp.Cid)
		}
	}

	missMap := make(map[int64][]int64, len(aidCids))
	for aid, cids := range aidCids {
		if aid == 0 || len(cids) == 0 {
			continue
		}
		for _, cid := range cids {
			if cid == 0 {
				continue
			}
			if !cachedCids.Has(cid) {
				missMap[aid] = append(missMap[aid], cid)
			}
		}
	}
	return cached, missMap, nil
}

// addVideoCache add video cache by aid & cid.
func (d *Dao) addVideoCache(c context.Context, aid, cid int64, p *api.Page) (err error) {
	var (
		key  = model.VideoKey(aid, cid)
		conn = d.arcRds.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = p.Marshal(); err != nil {
		log.Error("addVideoCache Marshal error(%v)", err)
		return
	}

	// psb_#{aid}_#{cid}缓存设置过期时间，默认10小时+48小时内随机数
	rand.Seed(time.Now().UnixNano())
	exp := int64(36000) + rand.Int63n(172800)

	if _, err = conn.Do("SET", key, bs, "EX", exp); err != nil {
		log.Error("conn.Do(SET, %s) error(%v)", key, err)
		return
	}
	return
}

func (d *Dao) addMultiVideoCache(c context.Context, aid int64, ps []*api.Page) error {
	kvMap := make(map[string][]byte, len(ps))

	for _, page := range ps {
		if page == nil {
			continue
		}
		bs, err := page.Marshal()
		if err != nil {
			log.Error("addMultiVideoCache Marshal error(%v)", err)
			continue
		}
		kvMap[model.VideoKey(aid, page.Cid)] = bs
	}

	rand.Seed(time.Now().UnixNano())
	exp := int64(36000) + rand.Int63n(172800)

	if err := d.redisMSetWithExp(c, kvMap, exp); err != nil {
		log.Error("addMultiVideoCache fail kvMap(%+v) error(%+v)", kvMap, err)
		return err
	}

	return nil
}

func (d *Dao) redisMSetWithExp(c context.Context, kvMap map[string][]byte, exp int64) (err error) {
	if len(kvMap) == 0 {
		return
	}

	var (
		conn        = d.arcRds.Get(c)
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
		log.Error("redisMSetWithExp conn.Send() MSET keys(%+v) err(%+v)", keys, err)
		return err
	}

	for _, key := range keys {
		if err := conn.Send("EXPIRE", key, exp); err != nil {
			log.Error("redisMSetWithExp conn.Send() EXPIRE key(%+v) err(%+v)", key, err)
			return err
		}
	}

	if err := conn.Flush(); err != nil {
		log.Error("redisMSetWithExp conn.Flush() keys(%+v) err(%+v)", keys, err)
		return err
	}

	for i := 0; i < len(keys)+1; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Error("redisMSetWithExp conn.Receive() keys(%+v) err(%+v)", keys, err)
			return err
		}
	}

	return err
}
