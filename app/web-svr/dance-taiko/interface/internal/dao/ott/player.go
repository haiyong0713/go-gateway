package ott

import (
	"context"

	"go-common/library/cache"
	"go-common/library/log"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	accClient "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/pkg/errors"
)

func (d *dao) UserCards(c context.Context, mids []int64) (map[int64]*accClient.Card, error) {
	req := &accClient.MidsReq{Mids: mids}
	res, err := d.accClient.Cards3(c, req)
	if err != nil {
		return nil, errors.Wrapf(err, "UserCards mids(%v)", mids)
	}
	return res.Cards, nil
}

// LoadPlayers get data from cache if miss will call source method, then add to cache.
func (d *dao) LoadPlayers(c context.Context, gameId int64) (res []*model.PlayerHonor, err error) {
	addCache := true
	res, err = d.CachePlayer(c, gameId)
	if err != nil {
		addCache = false
		err = nil
	}
	if len(res) != 0 {
		cache.MetricHits.Inc("bts:LoadPlayers")
		return
	}
	cache.MetricMisses.Inc("bts:LoadPlayers")
	res, err = d.RawPlayers(c, gameId)
	if err != nil {
		return
	}
	miss := res
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		if err := d.AddCachePLayer(c, gameId, miss); err != nil {
			log.Error("LoadPlayers players(%v) err(%v)", miss, err)
		}
	})
	return
}
