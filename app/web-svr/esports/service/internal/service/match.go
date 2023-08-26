package service

import (
	"context"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (s *Service) GetMatchInfo(ctx context.Context, matchId int64, skipCache bool) (matchInfo *pb.MatchModel, err error) {
	matchModel, err := s.getMatchModel(ctx, matchId, false)
	if err != nil {
		return
	}
	matchInfo = s.matchModel2External(matchModel)
	return
}

func (s *Service) getMatchModel(ctx context.Context, matchId int64, skipCache bool) (matchInfo *model.MatchModel, err error) {
	if !skipCache {
		cacheInfo, hit := s.matchesCacheMap.Get(matchId)
		if hit {
			matchCache, ok := cacheInfo.(*model.MatchModel)
			if ok {
				matchInfo = matchCache
				return
			}
		}
		matchModel, errG := s.dao.GetMatchCache(ctx, matchId)
		if errG != nil {
			err = errG
			return
		}
		if matchModel != nil {
			matchInfo = matchModel
			return
		}
	}
	matchInfo, err = s.dao.GetMatchModel(ctx, matchId)
	return
}

func (s *Service) matchModel2External(fromModel *model.MatchModel) (toModel *pb.MatchModel) {
	return &pb.MatchModel{
		ID:       fromModel.ID,
		Title:    fromModel.Title,
		SubTitle: fromModel.SubTitle,
		CYear:    fromModel.CYear,
		Sponsor:  fromModel.Sponsor,
		Logo:     fromModel.Logo,
		Dic:      fromModel.Dic,
		Status:   fromModel.Status,
		Rank:     fromModel.Rank,
	}
}
