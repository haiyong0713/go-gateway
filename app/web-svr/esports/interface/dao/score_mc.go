package dao

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/esports/interface/model"
)

const (
	_lolDataHero2CacheKey  = "es:lol:data:hero2:tournamentID:%v"
	_lolDataPlayerCacheKey = "es:lol:data:player:tournamentID:%v"
	_lolDataTeamCacheKey   = "es:lol:data:team:tournamentID:%v"
)

func lolDataHero2CacheKey(tournamentID int64) string {
	return fmt.Sprintf(_lolDataHero2CacheKey, tournamentID)
}

func lolDataPlayerCacheKey(tournamentID int64) string {
	return fmt.Sprintf(_lolDataPlayerCacheKey, tournamentID)
}

func lolDataTeamCacheKey(tournamentID int64) string {
	return fmt.Sprintf(_lolDataTeamCacheKey, tournamentID)
}

func (d *Dao) FetchLolDataTeamFromCache(ctx context.Context, tournamentID int64) (res []*model.LolTeam, err error) {
	cacheKey := lolDataTeamCacheKey(tournamentID)
	err = d.memcache.Get(ctx, cacheKey).Scan(&res)
	return
}

func (d *Dao) FetchLolDataTeamToCache(ctx context.Context, tournamentID int64, data []*model.LolTeam, expire int32) (err error) {
	item := &memcache.Item{
		Key:        lolDataTeamCacheKey(tournamentID),
		Object:     data,
		Expiration: expire,
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "lol_data_team_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return d.memcache.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "contest component FetchLolTeamToCache d.memcache.Set() tournamentID(%d) error(%+v)", tournamentID, err)
	}
	return
}

func (d *Dao) FetchLolDataPlayerFromCache(ctx context.Context, tournamentID int64) (res []*model.LolPlayer, err error) {
	cacheKey := lolDataPlayerCacheKey(tournamentID)
	err = d.memcache.Get(ctx, cacheKey).Scan(&res)
	return
}

func (d *Dao) FetchLolDataPlayerToCache(ctx context.Context, tournamentID int64, data []*model.LolPlayer, expire int32) (err error) {
	item := &memcache.Item{
		Key:        lolDataPlayerCacheKey(tournamentID),
		Object:     data,
		Expiration: expire,
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "lol_data_player_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return d.memcache.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "contest component FetchLolPlayerToCache d.memcache.Set() tournamentID(%d) error(%+v)", tournamentID, err)
	}
	return
}

func (d *Dao) FetchLolDataHero2FromCache(ctx context.Context, tournamentID int64) (res []*model.LolDataHero2, err error) {
	cacheKey := lolDataHero2CacheKey(tournamentID)
	err = d.memcache.Get(ctx, cacheKey).Scan(&res)
	return
}

func (d *Dao) FetchLolDataHero2ToCache(ctx context.Context, tournamentID int64, data []*model.LolDataHero2, expire int32) (err error) {
	item := &memcache.Item{
		Key:        lolDataHero2CacheKey(tournamentID),
		Object:     data,
		Expiration: expire,
		Flags:      memcache.FlagJSON,
	}
	if err = retry.WithAttempts(ctx, "lol_data_hero2_set_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return d.memcache.Set(ctx, item)
	}); err != nil {
		log.Errorc(ctx, "contest component FetchLolDataHero2ToCache d.memcache.Set() tournamentID(%d) error(%+v)", tournamentID, err)
	}
	return
}

func (d *Dao) DeleteLolDataHero2Cache(ctx context.Context, tournamentID int64) (err error) {
	cacheKey := lolDataHero2CacheKey(tournamentID)
	if err = d.memcache.Delete(ctx, cacheKey); err == memcache.ErrNotFound {
		return nil
	}
	return
}

func (d *Dao) DeleteLolDataPlayerCache(ctx context.Context, tournamentID int64) (err error) {
	cacheKey := lolDataPlayerCacheKey(tournamentID)
	if err = d.memcache.Delete(ctx, cacheKey); err == memcache.ErrNotFound {
		return nil
	}
	return
}
func (d *Dao) DeleteLolDataTeamCache(ctx context.Context, tournamentID int64) (err error) {
	cacheKey := lolDataTeamCacheKey(tournamentID)
	if err = d.memcache.Delete(ctx, cacheKey); err == memcache.ErrNotFound {
		return nil
	}
	return
}
