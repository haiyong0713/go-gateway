package service

import (
	"context"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (s *Service) RefreshSeasonContestIdsCache(ctx context.Context, in *pb.RefreshSeasonContestIdsReq) (response *pb.RefreshSeasonContestIdsResponse, err error) {
	response = new(pb.RefreshSeasonContestIdsResponse)
	response.ContestIds = make([]int64, 0)
	if in == nil || in.SeasonId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	// 刷新赛季信息
	_, _ = s.getSeasonModel(ctx, in.SeasonId, true, true)
	// 刷新赛程
	contestIds, err := s.getSeasonContestIds(ctx, in.SeasonId, true)
	if err != nil {
		log.Errorc(ctx, "[Service][CacheRefresh][RefreshSeasonContestIdsCache][Error], err:%+v", err)
		return
	}
	response.ContestIds = contestIds
	return
}
func (s *Service) RefreshContestCache(ctx context.Context, in *pb.RefreshContestCacheReq) (response *pb.NoArgsResponse, err error) {
	response = new(pb.NoArgsResponse)
	if in == nil || in.ContestId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	_, err = s.getContestsModel(ctx, []int64{in.ContestId}, true, true, true)
	if err != nil {
		log.Errorc(ctx, "[Service][CacheRefresh][RefreshContestCache][Error], err:%+v", err)
		return
	}
	return
}

func (s *Service) RefreshSeriesCache(ctx context.Context, req *pb.RefreshSeriesCacheReq) (response *pb.NoArgsResponse, err error) {
	response = new(pb.NoArgsResponse)
	if req == nil || req.SeriesId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	_, err = s.getSeriesByIds(ctx, []int64{req.SeriesId}, true, true)
	if err != nil {
		log.Errorc(ctx, "[Service][CacheRefresh][RefreshSeriesCache][Error], err:%+v", err)
		return
	}
	return
}
func (s *Service) RefreshGameCache(ctx context.Context, req *pb.NoArgsRequest) (response *pb.NoArgsResponse, err error) {
	response = new(pb.NoArgsResponse)
	gameModels, err := s.dao.GetAllGames(ctx)
	if err != nil {
		log.Errorc(ctx, "[Service][CacheRefresh][RefreshGameCache][Error], err:%+v", err)
		return
	}
	gameMaps := make(map[int64]*model.GameModel)
	for _, game := range gameModels {
		gameMaps[game.ID] = game
	}
	err = s.dao.SetGamesCache(ctx, gameMaps)
	if err != nil {
		log.Errorc(ctx, "[Service][CacheRefresh][RefreshGameCache][Error], err:%+v", err)
		return
	}
	return
}

func (s *Service) RefreshTeamCache(ctx context.Context, req *pb.RefreshTeamCacheReq) (response *pb.NoArgsResponse, err error) {
	response = new(pb.NoArgsResponse)
	if req == nil || req.TeamId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	_, err = s.getTeamsModel(ctx, []int64{req.TeamId}, true, true)
	if err != nil {
		log.Errorc(ctx, "[Service][CacheRefresh][RefreshTeamCache][Error], err:%+v", err)
		return
	}
	return
}
