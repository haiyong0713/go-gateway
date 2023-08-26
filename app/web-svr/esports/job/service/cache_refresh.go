package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/job/component"
	v1 "go-gateway/app/web-svr/esports/service/api/v1"
)

func (s *Service) gamesCacheRefresh(ctx context.Context) {
	log.Infoc(ctx, "[Cron][gamesCacheRefresh][Begin]")
	_, err := component.EspServiceClient.RefreshGameCache(ctx, &v1.NoArgsRequest{})
	if err != nil {
		log.Errorc(ctx, "[Job][Refresh][gamesCacheRefresh][Error], err:%+v", err)
		return
	}
	log.Infoc(ctx, "[Cron][gamesCacheRefresh][End]")
}

func (s *Service) activeSeasonRefresh(ctx context.Context) {
	log.Infoc(ctx, "[Cron][activeSeasonRefresh][Begin]")
	activeSeasons, err := component.EspServiceClient.RefreshActiveSeasons(ctx, &v1.NoArgsRequest{})
	if err != nil {
		log.Errorc(ctx, "[Job][Refresh][activeSeasonRefresh][Error], err:%+v", err)
		return
	}
	seasons := activeSeasons.Seasons
	seasonIds := make([]int64, 0)
	if seasons == nil || len(seasons) == 0 {
		log.Warnc(ctx, "[Job][Refresh][activeSeasonRefresh][Warn][activeSeasons][Empty]")
		return
	}
	for _, season := range seasons {
		seasonId := season.ID
		seasonIds = append(seasonIds, seasonId)
		s.refreshSeasonContests(seasonId)
		s.refreshSeasonSeries(seasonId)
		// 赛季的战队缓存更新
		//component.EspServiceClient.RefreshTeamCache(ctx, &v1.RefreshTeamCacheReq{
		//	TeamId:               0,
		//})
	}
	log.Infoc(ctx, "[Cron][activeSeasonRefresh][End], season:%+v", seasonIds)
}

func (s *Service) refreshSeasonSeries(seasonId int64) {
	if seasonId == 0 {
		return
	}
	seriesRes, err := component.EspServiceClient.GetSeasonSeriesModel(ctx, &v1.GetSeasonSeriesReq{
		SeasonId: seasonId,
	})
	if err != nil {
		log.Errorc(ctx, "[Job][Refresh][GetSeasonSeriesModel][Error], seasonId:%d, err:%+v", seasonId, err)
		return
	}
	seriesList := seriesRes.Series
	if seriesList == nil || len(seriesList) == 0 {
		log.Warnc(ctx, "[Job][Refresh][GetSeasonSeriesModel][Warn][series][Empty], seasonId:%d", seasonId)
		return
	}
	for _, series := range seriesList {
		component.EspServiceClient.RefreshSeriesCache(ctx, &v1.RefreshSeriesCacheReq{
			SeriesId: series.ID,
		})
	}
}

func (s *Service) refreshSeasonContests(seasonId int64) {
	if seasonId == 0 {
		return
	}
	res, err := component.EspServiceClient.RefreshSeasonContestIdsCache(ctx, &v1.RefreshSeasonContestIdsReq{
		SeasonId: seasonId,
	})
	if err != nil {
		log.Errorc(ctx, "[Job][Refresh][RefreshSeasonContestIdsCache][Error], seasonId:%d, err:%+v", seasonId, err)
		return
	}
	contestIds := res.ContestIds
	if len(contestIds) == 0 {
		return
	}
	for _, contestId := range contestIds {
		_, errG := component.EspServiceClient.RefreshContestCache(ctx, &v1.RefreshContestCacheReq{
			ContestId: contestId,
		})
		if errG != nil {
			log.Errorc(ctx, "[Job][Refresh][RefreshContestCache][Error], seasonId:%d, contestId:%d, err:%+v", seasonId, contestId, errG)
		}
	}
}
