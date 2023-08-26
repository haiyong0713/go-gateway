package dao

import (
	"context"
	"fmt"
	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (d *dao) GetSeasonContestIdsCache(ctx context.Context, seasonId int64) (contestIds []int64, err error) {
	contestIds = make([]int64, 0)
	mcKey := fmt.Sprintf(seasonContestIds, seasonId)
	if err = d.mc.Get(ctx, mcKey).Scan(&contestIds); err != nil {
		log.Errorc(ctx, "[Dao][GetActiveSeasonsCache][Get][Scan][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) StoreSeasonContestIdsCache(ctx context.Context, seasonId int64, contestIds []int64) (err error) {
	mcKey := fmt.Sprintf(seasonContestIds, seasonId)
	item := &memcache.Item{
		Key:        mcKey,
		Object:     contestIds,
		Expiration: seasonContestIdsTtl,
		Flags:      memcache.FlagJSON}
	if err = d.mc.Set(ctx, item); err != nil {
		log.Errorc(ctx, "[Dao][StoreActiveSeasonsCache][Set][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) GetActiveSeasonsCache(ctx context.Context) (seasonIds []int64, err error) {
	seasonsMap := make(map[int64]*model.SeasonModel)
	if err = d.mc.Get(ctx, activeSeasonList).Scan(&seasonsMap); err != nil {
		log.Errorc(ctx, "[Dao][GetActiveSeasonsCache][Get][Scan][Error], err:%+v", err)
		return
	}
	seasonIds = make([]int64, 0)
	for _, v := range seasonsMap {
		seasonIds = append(seasonIds, v.ID)
	}
	return
}
func (d *dao) StoreActiveSeasonsCache(ctx context.Context, seasonsMap map[int64]*model.SeasonModel) (err error) {
	item := &memcache.Item{
		Key:        activeSeasonList,
		Object:     seasonsMap,
		Expiration: activeSeasonListTtl,
		Flags:      memcache.FlagJSON}
	if err = d.mc.Set(ctx, item); err != nil {
		log.Errorc(ctx, "[Dao][StoreActiveSeasonsCache][Set][Error], err:%+v", err)
		return
	}
	return
}
