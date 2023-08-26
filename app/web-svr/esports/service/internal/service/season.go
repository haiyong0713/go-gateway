package service

import (
	"context"
	"time"

	"go-common/library/cache/memcache"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

// GetSeasonModel 获取赛季信息
func (s *Service) GetSeasonModel(ctx context.Context, req *pb.GetSeasonModelReq) (seasonInfo *pb.SeasonModel, err error) {
	seasonModel, err := s.dao.GetSeasonByID(ctx, req.SeasonId)
	seasonInfo = s.seasonModel2External(seasonModel)
	return
}

func (s *Service) GetSeasonInfo(ctx context.Context, seasonId int64) (
	seasonInfo *pb.SeasonModel, err error,
) {
	seasonModel, err := s.getSeasonModel(ctx, seasonId, false, false)
	seasonInfo = s.seasonModel2External(seasonModel)
	return
}

func (s *Service) GetSeasonDetail(ctx context.Context, req *pb.GetSeasonModelReq) (response *pb.SeasonDetail, err error) {
	response = new(pb.SeasonDetail)
	if req == nil || req.SeasonId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	seasonModel, err := s.getSeasonModel(ctx, req.SeasonId, false, false)
	if err != nil {
		log.Errorc(ctx, "[Service][GetSeasonDetail][getSeasonModel][Error], err:%+v", err)
		return
	}
	response = s.formatSeason(seasonModel)
	return
}

func (s *Service) ClearSeasonCache(ctx context.Context, req *pb.ClearSeasonCacheReq) (res *pb.NoArgsResponse, err error) {
	res = &pb.NoArgsResponse{}
	if req == nil || req.SeasonId == 0 {
		return
	}
	err = s.dao.DeleteSeasonCache(ctx, req.SeasonId)
	return
}

func (s *Service) getSeasonModel(ctx context.Context, seasonId int64, skipCache bool, skipMemory bool) (
	seasonInfo *model.SeasonModel, err error,
) {
	if !skipMemory {
		seasonInfo = s.getSeasonFromMemory(seasonId)
		if seasonInfo != nil && seasonInfo.Status != model.SeasonStatusFalse {
			return
		}
	}
	if !skipCache {
		// 再读缓存
		seasonInfo, err = s.dao.GetSeasonInfoCache(ctx, seasonId)
		if err != nil {
			return
		}
		if seasonInfo != nil {
			return
		}
	}
	seasonInfo, err = s.dao.GetSeasonByID(ctx, seasonId)

	if err != nil {
		return
	}
	if err = s.dao.SetSeasonInfoCache(ctx, seasonInfo); err != nil {
		return
	}
	// seasonInfo 接口不返回已冻结信息
	if seasonInfo.Status == model.SeasonStatusFalse {
		log.Warnc(ctx, "[Service][Season][Freeze][Season][Get], season:%+v", seasonInfo)
		err = xecode.Errorf(xecode.NothingFound, "赛季不存在或已冻结")
		seasonInfo = nil
		return
	}
	return
}

func (s *Service) getSeasonFromMemory(seasonId int64) *model.SeasonModel {
	if cache, ok := s.seasonInfoCacheMap[seasonId]; ok {
		return cache
	}
	return nil
}

func (s *Service) getSeasonsModel(ctx context.Context, seasonIds []int64, skipCache bool) (seasonModelsMap map[int64]*model.SeasonModel, err error) {
	seasonModelsMap = make(map[int64]*model.SeasonModel)
	missIds := seasonIds
	if !skipCache {
		seasonModelsMap, missIds, err = s.getSeasonsModelFromCache(ctx, seasonIds)
		if err != nil {
			log.Errorc(ctx, "[Service][getSeasonsModel][getSeasonsModelFromCache][Error], err:%+v", err)
			return
		}
	}
	if len(missIds) == 0 {
		return
	}
	// 回源
	seasons, err := s.dao.GetSeasonsByIDs(ctx, missIds)
	if err != nil {
		log.Errorc(ctx, "[Service][getSeasonsModel][GetSeasonsByIDs][Error], err:%+v", err)
		return
	}
	for _, season := range seasons {
		seasonModelsMap[season.ID] = season
	}
	errS := s.dao.SetSeasonsInfoCache(ctx, seasons)
	if errS != nil {
		log.Errorc(ctx, "[Service][getSeasonsModel][SetSeasonsInfoCache][Error], err:%+v", err)
	}
	return
}
func (s *Service) getSeasonsModelFromCache(ctx context.Context, seasonIds []int64) (seasonModelsMap map[int64]*model.SeasonModel, missIds []int64, err error) {
	seasonModelsMap = make(map[int64]*model.SeasonModel)
	missIds = make([]int64, 0)
	for _, seasonId := range seasonIds {
		if seasonInfoCache, ok := s.seasonInfoCacheMap[seasonId]; ok {
			seasonModelsMap[seasonId] = seasonInfoCache
		} else {
			missIds = append(missIds, seasonId)
		}
	}
	if len(missIds) == 0 {
		return
	}
	// 再读缓存
	seasonsModelMapCache, missIds, err := s.dao.GetSeasonsInfoCache(ctx, missIds)
	if err != nil {
		return
	}
	for key, v := range seasonsModelMapCache {
		seasonModelsMap[key] = v
	}
	return
}

func (s *Service) getSeasonContestIds(ctx context.Context, seasonId int64, skipCache bool) (contestIds []int64, err error) {
	if !skipCache {
		contestIds, err = s.dao.GetSeasonContestIdsCache(ctx, seasonId)
		if err != nil && err != memcache.ErrNotFound {
			log.Errorc(ctx, "[getSeasonContestIds][GetSeasonContestIdsCache][Error], err:%+v", err)
			return
		}
		if err == nil || len(contestIds) != 0 {
			return
		}
	}
	// 回源
	contestIds, err = s.dao.GetSeasonContestIds(ctx, seasonId)
	if err != nil {
		log.Errorc(ctx, "[getSeasonContestIds][GetSeasonContestIds][Error], err:%+v", err)
		return
	}
	err = s.dao.StoreSeasonContestIdsCache(ctx, seasonId, contestIds)
	if err != nil {
		log.Errorc(ctx, "[getSeasonContestIds][StoreSeasonContestIdsCache][Error], err:%+v", err)
		return
	}
	return
}

func (s *Service) RefreshActiveSeasons(ctx context.Context, req *pb.NoArgsRequest) (response *pb.ActiveSeasonsResponse, err error) {
	response = new(pb.ActiveSeasonsResponse)
	response.Seasons = make([]*pb.SeasonModel, 0)
	seasonModels, err := s.getActiveSeasons(ctx)
	if err != nil {
		return
	}
	response.Seasons = s.seasonModels2External(seasonModels)
	return
}

func (s *Service) getActiveSeasons(ctx context.Context) (seasons []*model.SeasonModel, err error) {
	nowTime := time.Now().Unix()
	startTime := nowTime + int64(time.Duration(s.conf.SeasonContestComponent.StartTimeBefore).Seconds())
	endTime := nowTime - int64(time.Duration(s.conf.SeasonContestComponent.EndTimeAfter).Seconds())
	seasons, err = s.dao.GetSeasonsBySETime(ctx, startTime, endTime)
	if err != nil {
		log.Errorc(ctx, "[GetActiveSeasons][Error], err:%+v", err)
	}
	// 重构缓存
	seasonsMap := make(map[int64]*model.SeasonModel)
	for _, v := range seasons {
		seasonsMap[v.ID] = v
	}
	err = s.dao.StoreActiveSeasonsCache(ctx, seasonsMap)
	return
}

func (s *Service) seasonModels2External(fromModels []*model.SeasonModel) (toModels []*pb.SeasonModel) {
	toModels = make([]*pb.SeasonModel, 0)
	for _, v := range fromModels {
		single := s.seasonModel2External(v)
		toModels = append(toModels, single)
	}
	return
}

func (s *Service) seasonModel2External(fromModel *model.SeasonModel) (toModels *pb.SeasonModel) {
	if fromModel == nil {
		return nil
	}
	return &pb.SeasonModel{
		ID:           fromModel.ID,
		Mid:          fromModel.Mid,
		Title:        fromModel.Title,
		SubTitle:     fromModel.SubTitle,
		Stime:        fromModel.Stime,
		Etime:        fromModel.Etime,
		Sponsor:      fromModel.Sponsor,
		Logo:         fromModel.Logo,
		Dic:          fromModel.Dic,
		Status:       fromModel.Status,
		Rank:         fromModel.Rank,
		IsApp:        fromModel.IsApp,
		URL:          fromModel.URL,
		DataFocus:    fromModel.DataFocus,
		FocusURL:     fromModel.FocusURL,
		ForbidIndex:  fromModel.ForbidIndex,
		LeidaSid:     fromModel.LeidaSid,
		SerieType:    fromModel.SerieType,
		SearchImage:  fromModel.SearchImage,
		SyncPlatform: fromModel.SyncPlatform,
		GuessVersion: fromModel.GuessVersion,
		SeasonType:   fromModel.SeasonType,
	}
}
