package ott

import (
	"context"
	"go-gateway/ecode"

	"go-common/library/cache"
	"go-common/library/log"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

func (d *dao) rawRanks(c context.Context, cid int64) ([]*model.PlayerHonor, error) {
	gameIds, err := d.SelectGamesByCid(c, cid)
	if err != nil {
		log.Error("rawRanks cid(%d) err(%v)", cid, err)
		return nil, err
	}
	playerRanks, err := d.SelectPlayersByGames(c, gameIds)
	if err != nil {
		log.Error("rawRanks gameIds(%v) err(%v)", gameIds, err)
		return nil, err
	}

	var (
		res    = make([]*model.PlayerHonor, 0)
		midMap = make(map[int64]struct{})
	)
	// 去重
	for _, player := range playerRanks {
		if _, ok := midMap[player.Mid]; !ok {
			res = append(res, player)
			midMap[player.Mid] = struct{}{}
		}
	}
	return res, nil
}

// LoadRanks get data from cache if miss will call source method, then add to cache.
func (d *dao) LoadRanks(c context.Context, cid int64, pn int, ps int) (res []*model.PlayerHonor, err error) {
	addCache := true
	res, err = d.CacheRank(c, cid, pn, ps)
	if err != nil {
		addCache = false
		err = nil
	}
	if len(res) != 0 {
		cache.MetricHits.Inc("bts:LoadRanks")
		return
	}
	cache.MetricMisses.Inc("bts:LoadRanks")
	res, err = d.rawRanks(c, cid)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		if err := d.AddCacheRanks(c, cid, miss); err != nil {
			log.Error("LoadRanks cid(%d) err(%v)", cid, err)
		}
	})

	// 只返回请求区间内的数据
	if len(res) < pn*ps {
		return nil, ecode.ReqParamErr
	}
	if len(res) > (pn+1)*ps {
		res = res[pn*ps : (pn+1)*ps]
	} else {
		res = res[pn*ps:]
	}
	return res, nil
}
