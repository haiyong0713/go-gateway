package service

import (
	"context"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

func (s *Service) getSeriesByIds(ctx context.Context, seriesIds []int64, skipCache bool, skipMemory bool) (seriesMap map[int64]*model.ContestSeriesModel, err error) {
	missIds := seriesIds
	seriesMap = make(map[int64]*model.ContestSeriesModel)
	if !skipMemory {
		seriesMap, missIds = s.getSeriesInfoFromMemory(missIds)
		if len(missIds) == 0 {
			return
		}
	}
	if !skipCache {
		cacheMap, cacheMissIds, errG := s.dao.GetSeriesCacheByIds(ctx, missIds)
		if errG != nil {
			err = errG
			log.Errorc(ctx, "[Service][GetSeriesByIds][GetSeriesCacheByIds][Error], err:%+v", err)
			return
		}
		for k, v := range cacheMap {
			seriesMap[k] = v
		}
		missIds = cacheMissIds
	}
	if len(missIds) == 0 {
		return
	}
	dbMap, err := s.dao.GetSeriesByIds(ctx, missIds)
	if err != nil {
		log.Errorc(ctx, "[Service][GetSeriesByIds][GetSeriesByIds][Error], err:%+v", err)
		return
	}
	rebuildList := make([]*model.ContestSeriesModel, 0)
	for k, v := range dbMap {
		seriesMap[k] = v
		rebuildList = append(rebuildList, v)
	}
	if len(rebuildList) > 0 {
		errSet := s.dao.SetSeriesCacheByIds(ctx, rebuildList)
		if errSet != nil {
			log.Errorc(ctx, "[Service][GetSeriesByIds][SetSeriesCacheByIds][Error], err:%+v", err)
		}
	}
	return
}

func (s *Service) getSeriesInfoFromMemory(seriesIds []int64) (contestSeriesMap map[int64]*model.ContestSeriesModel, missIds []int64) {
	contestSeriesMap = make(map[int64]*model.ContestSeriesModel)
	missIds = make([]int64, 0)
	for _, v := range seriesIds {
		if cache, ok := s.seriesCacheMap.Get(v); ok {
			if cacheValue, valid := cache.(*model.ContestSeriesModel); valid {
				contestSeriesMap[v] = cacheValue
			}
		}
		missIds = append(missIds, v)
	}
	return
}

func (s *Service) GetSeasonSeriesModel(ctx context.Context, req *pb.GetSeasonSeriesReq) (response *pb.GetSeasonSeriesResponse, err error) {
	response = new(pb.GetSeasonSeriesResponse)
	response.Series = make([]*pb.SeriesModel, 0)
	if req == nil || req.SeasonId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	seriesModels, err := s.dao.GetSeriesBySeasonId(ctx, req.SeasonId)
	if err != nil {
		log.Errorc(ctx, "[Service][GetSeasonSeriesModel][GetSeriesBySeasonId][Error], err:%+v", err)
		return
	}
	for _, v := range seriesModels {
		response.Series = append(response.Series, seriesModel2Internal(v))
	}
	return
}

func seriesModel2Internal(model *model.ContestSeriesModel) *pb.SeriesModel {
	if model == nil {
		return nil
	}
	return &pb.SeriesModel{
		ID:          model.ID,
		ParentTitle: model.ParentTitle,
		ChildTitle:  model.ChildTitle,
		StartTime:   model.StartTime,
		EndTime:     model.EndTime,
		ScoreId:     model.ScoreId,
	}
}

func seriesModel2ExternalInfo(model *model.ContestSeriesModel) *pb.SeriesDetail {
	if model == nil {
		return nil
	}
	return &pb.SeriesDetail{
		ID:          model.ID,
		ParentTitle: model.ParentTitle,
		ChildTitle:  model.ChildTitle,
		StartTime:   model.StartTime,
		EndTime:     model.EndTime,
		ScoreId:     model.ScoreId,
	}
}
