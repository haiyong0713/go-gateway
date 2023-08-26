package archive

import (
	"context"

	"go-common/library/log"
	"go-common/library/stat/prom"

	"go-gateway/app/app-svr/archive/service/api"
)

// Stat3 get archive stat.
func (d *Dao) Stat3(c context.Context, aid int64) (st *api.Stat, err error) {
	var cached = true
	if st, err = d.StatRedisCache(c, aid); err != nil {
		log.Error("d.statRedisCache(%d) error(%v)", aid, err)
		cached = false
	}
	prom.BusinessInfoCount.Incr("StatRedisCache")
	if st != nil {
		return
	}
	if st, err = d.RawStat(c, aid); err != nil {
		log.Error("d.stat(%d) error(%v)", aid, err)
		return
	}
	if st == nil {
		st = &api.Stat{Aid: aid}
		return
	}
	if cached {
		d.addCache(func() {
			_ = d.addStatRedisCache(context.TODO(), st)
		})
	}
	return
}

// Stats3 get archives stat.
func (d *Dao) Stats3(c context.Context, aids []int64) (stm map[int64]*api.Stat, err error) {
	if len(aids) == 0 {
		return
	}
	var (
		missed []int64
		missm  map[int64]*api.Stat
		cached = true
	)
	if stm, missed, err = d.statRedisCaches(c, aids); err != nil {
		log.Error("d.statCaches(%d) error(%v)", aids, err)
		missed = aids
		stm = make(map[int64]*api.Stat, len(aids))
		err = nil // ignore error
		cached = false
	}
	if stm != nil && len(missed) == 0 {
		return
	}
	if missm, err = d.RawStats(c, missed); err != nil {
		log.Error("d.stats(%v) error(%v)", missed, err)
		err = nil // ignore error
	}
	for aid, st := range missm {
		stm[aid] = st
		if cached {
			var cst = &api.Stat{}
			*cst = *st
			d.addCache(func() {
				_ = d.addStatRedisCache(context.TODO(), cst)
			})
		}
	}
	return
}

// InitStatCache3 if db is nil, set nil cache
func (d *Dao) InitStatCache3(c context.Context, aid int64) (err error) {
	var st *api.Stat
	if st, err = d.RawStat(c, aid); err != nil {
		log.Error("d.stat(%d) error(%v)", aid, err)
		return
	}
	if st == nil {
		d.addCache(func() {
			_ = d.addStatRedisCache(context.TODO(), &api.Stat{Aid: aid})
		})
	}
	return
}
