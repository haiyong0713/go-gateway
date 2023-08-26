package dao

import (
	"context"
	"fmt"
	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (d *dao) GetSeasonTeamsCache(ctx context.Context, seasonId int64) (seasonTeams []*model.SeasonTeamModel, err error) {
	mcKey := fmt.Sprintf(seasonTeamsList, seasonId)
	seasonTeams = make([]*model.SeasonTeamModel, 0)
	if err = d.mc.Get(ctx, mcKey).Scan(&seasonTeams); err != nil {
		log.Errorc(ctx, "[Dao][GetSeasonTeamsCache][Get][Scan][Error], err:%+v", err)
		return
	}
	return
}
func (d *dao) StoreSeasonTeamsCache(ctx context.Context, seasonTeams []*model.SeasonTeamModel, seasonId int64) (err error) {
	mcKey := fmt.Sprintf(seasonTeamsList, seasonId)
	item := &memcache.Item{
		Key:        mcKey,
		Object:     seasonTeams,
		Expiration: seasonTeamsListTtl,
		Flags:      memcache.FlagJSON}
	if err = d.mc.Set(ctx, item); err != nil {
		log.Errorc(ctx, "[Dao][StoreSeasonTeamsCache][Set][Error], err:%+v", err)
		return
	}
	return
}
