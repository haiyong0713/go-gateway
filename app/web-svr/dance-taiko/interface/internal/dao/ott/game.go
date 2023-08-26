package ott

import (
	"context"
	"go-common/library/cache"
	"go-common/library/log"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

// LoadGame get data from cache if miss will call source method, then add to cache.
func (d *dao) LoadGame(c context.Context, gameId int64) (res *model.OttGame, err error) {
	addCache := true
	res, err = d.cacheGame(c, gameId)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res != nil && res.GameId == -1 {
			res = nil
		}
	}()
	if res != nil {
		cache.MetricHits.Inc("bts:LoadGame")
		return
	}
	cache.MetricMisses.Inc("bts:LoadGame")
	res, err = d.rawGame(c, gameId)
	if err != nil {
		return
	}
	miss := res
	if miss == nil {
		miss = &model.OttGame{GameId: -1}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		if err := d.addCacheGame(c, gameId, miss); err != nil {
			log.Error("addCacheGame gameId(%d) err(%v)", gameId, err)
		}
	})
	return
}
