package service

import (
	"context"
	"go-common/library/log"

	xecode "go-common/library/ecode"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (s *Service) GetGameInfo(ctx context.Context, gameId int64) (gameInfo *pb.GameModel, err error) {
	gamesInfo, err := s.getGamesModel(ctx, []int64{gameId}, false)
	if err != nil {
		return
	}
	if gamesInfo == nil || gamesInfo[gameId] == nil {
		err = xecode.Errorf(xecode.RequestErr, "游戏不存在或已被冻结")
		return
	}
	gameInfo = s.gameModel2External(gamesInfo[0])
	return
}

func (s *Service) getGamesModel(ctx context.Context, gameIds []int64, skipCache bool) (
	gamesInfoMap map[int64]*model.GameModel,
	err error,
) {
	missGameIds := gameIds
	gamesInfoMap = make(map[int64]*model.GameModel)
	if !skipCache {
		cacheMap, missIds, errG := s.getGameInfoFromCache(ctx, gameIds)
		if errG != nil {
			err = errG
			return
		}
		gamesInfoMap = cacheMap
		missGameIds = missIds
	}
	if len(missGameIds) == 0 {
		return
	}
	gameModels, err := s.dao.GetGamesByIds(ctx, missGameIds)
	if err != nil {
		return
	}
	rebuildCacheMap := make(map[int64]*model.GameModel)
	for _, gameInfo := range gameModels {
		rebuildCacheMap[gameInfo.ID] = gameInfo
		gamesInfoMap[gameInfo.ID] = gameInfo
	}
	_ = s.dao.SetGamesCache(ctx, rebuildCacheMap)
	return
}

func (s *Service) getGameInfoFromCache(ctx context.Context, gameIds []int64) (gamesInfoMap map[int64]*model.GameModel, missGameIds []int64, err error) {
	missGameIds = make([]int64, 0)
	//获取内存
	gamesInfoMap = s.getGamesInfoFromLocal(gameIds)
	for _, gameId := range gameIds {
		if _, ok := gamesInfoMap[gameId]; !ok {
			missGameIds = append(missGameIds, gameId)
		}
	}
	//获取缓存
	if len(missGameIds) > 0 {
		redisCacheGameGames, missIds, errC := s.dao.GetGamesCache(ctx, missGameIds)
		if errC != nil {
			err = errC
			log.Errorc(ctx, "[Service][getGameInfoFromCache][GetGamesCache], err:%+v", err)
			return
		}
		for gameId, gameInfo := range redisCacheGameGames {
			gamesInfoMap[gameId] = gameInfo
		}
		missGameIds = missIds
	}
	return
}

func (s *Service) getGamesInfoFromLocal(gameIds []int64) (gamesMap map[int64]*model.GameModel) {
	gamesMap = make(map[int64]*model.GameModel)
	for _, v := range gameIds {
		if gameInfoCache, ok := s.gamesCacheMap[v]; ok {
			gamesMap[gameInfoCache.ID] = gameInfoCache
		}
	}
	return
}

// GetGamesModel
func (s *Service) GetGamesModel(ctx context.Context, gameIds []int64) (gameInfos []*pb.GameModel, err error) {
	gameModels, err := s.dao.GetGamesByIds(ctx, gameIds)
	if err != nil {
		return
	}
	gameInfos = s.gameModels2External(gameModels)
	return
}

// GetGamesInfoMap
func (s *Service) GetGamesInfoMap(ctx context.Context, gameIds []int64, skipCache bool) (
	GamesInfoMap map[int64]*pb.GameModel,
	err error,
) {
	return
}

func (s *Service) gameModels2External(fromModel []*model.GameModel) (toModel []*pb.GameModel) {
	toModel = make([]*pb.GameModel, 0)
	for _, v := range fromModel {
		single := s.gameModel2External(v)
		toModel = append(toModel, single)
	}
	return
}

func (s *Service) gameModel2External(fromModel *model.GameModel) (toModel *pb.GameModel) {
	return &pb.GameModel{
		ID:         fromModel.ID,
		Title:      fromModel.Title,
		SubTitle:   fromModel.SubTitle,
		ETitle:     fromModel.ETitle,
		Plat:       fromModel.Plat,
		Type:       fromModel.Type,
		Logo:       fromModel.Logo,
		Publisher:  fromModel.Publisher,
		Operations: fromModel.Operations,
		PbTime:     fromModel.PbTime,
		Dic:        fromModel.Dic,
		Status:     fromModel.Status,
		Rank:       fromModel.Rank,
	}
}

func (s *Service) gameModel2ExternalDetail(fromModel *model.GameModel) (toModel *pb.GameDetail) {
	return &pb.GameDetail{
		ID:         fromModel.ID,
		Title:      fromModel.Title,
		SubTitle:   fromModel.SubTitle,
		ETitle:     fromModel.ETitle,
		Plat:       fromModel.Plat,
		GameType:   fromModel.Type,
		Logo:       fromModel.Logo,
		Publisher:  fromModel.Publisher,
		Operations: fromModel.Operations,
		PbTime:     fromModel.PbTime,
		Dic:        fromModel.Dic,
		LogoFull:   s.formatFullLogoPath(fromModel.Logo),
		Rank:       fromModel.Rank,
	}
}
