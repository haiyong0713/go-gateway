package service

import (
	"context"
	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
	"time"
)

const (
	_batchContestSize = 100
	_seasonSlice0     = 0
	_seasonSlice1     = 1
	_seasonSlice2     = 2
	_seasonSlice3     = 3
	_seasonSlice4     = 4
	_seasonSlice5     = 5
	_seasonSlice6     = 6
	_seasonSlice7     = 7
	_seasonSlice8     = 8
	_seasonSlice9     = 9
	_seasonSliceBase  = 10
	_teamCacheTtl     = 300
	_contestCacheTtl  = 300
	_seriesCacheTtl   = 600
)

func (s *Service) storeGameCache(ctx context.Context) {
	games, err := s.dao.GetAllGamesCache(ctx)
	if err != nil {
		log.Errorc(ctx, "[Service][storeGameCache][GetAllGamesCache][Error], err:%+v", err)
		return
	}
	s.gamesCacheMap = games
}

// storeSeasonCache
func (s *Service) storeSeasonCache(ctx context.Context) {
	seasonIds, err := s.dao.GetActiveSeasonsCache(ctx)
	if err != nil && err != memcache.ErrNotFound {
		log.Errorc(ctx, "[Service][storeSeasonCache][GetActiveSeasonsCache][Error], err:%+v", err)
		return
	}
	seasonInfosMap := make(map[int64]*model.SeasonModel)
	if err == memcache.ErrNotFound {
		seasonModels, errG := s.getActiveSeasons(ctx)
		if errG != nil {
			return
		}
		for _, v := range seasonModels {
			seasonInfosMap[v.ID] = v
		}
	} else {
		for _, seasonId := range seasonIds {
			seasonInfo, errG := s.getSeasonModel(ctx, seasonId, false, true)
			if errG != nil {
				continue
			}
			seasonInfosMap[seasonId] = seasonInfo
		}
	}
	s.seasonInfoCacheMap = seasonInfosMap
}

// storeSeasonContestsCache
func (s *Service) storeSeasonContestsCache(ctx context.Context) {
	seasonsContestListMap := make(map[int64][]*model.ContestModel)
	for seasonId := range s.seasonInfoCacheMap {
		contestIds, err := s.getSeasonContestIds(ctx, seasonId, false)
		if err != nil {
			log.Errorc(ctx, "[storeSeasonContestsCache][getSeasonContestIds], sid(%d) error(%+v)", seasonId, err)
			continue
		}
		// 获取所有赛程的缓存
		contestModels := s.getSeasonContestModels(ctx, contestIds)
		seasonsContestListMap[seasonId] = contestModels
	}
	s.rebuildComponentGoingSeasonsContestListMap(seasonsContestListMap)
}

func (s *Service) getSeasonContestModels(ctx context.Context, contestIds []int64) (contestModels []*model.ContestModel) {
	contestModels = make([]*model.ContestModel, 0)
	reqContestIds := make([]int64, 0)
	contestMapList := make(map[int64]*model.ContestModel)
	for _, contestId := range contestIds {
		reqContestIds = append(reqContestIds, contestId)
		if len(reqContestIds) >= _batchContestSize {
			contestsSliceMapList, errG := s.getContestsModel(ctx, reqContestIds, false, true, true)
			if errG != nil {
				time.Sleep(time.Second)
				continue
			}
			for k, v := range contestsSliceMapList {
				contestMapList[k] = v
			}
			reqContestIds = make([]int64, 0)
		}
	}
	if len(reqContestIds) > 0 {
		contestsSliceMapList, errG := s.getContestsModel(ctx, reqContestIds, false, true, true)
		if errG != nil {
			log.Errorc(ctx, "[Service][storeSeasonContestsCache], err:%+v", errG)
		} else {
			for k, v := range contestsSliceMapList {
				contestMapList[k] = v
			}
		}
	}
	for _, v := range contestMapList {
		contestModels = append(contestModels, v)
	}
	return
}

// rebuildComponentGoingSeasonsContestListMap
func (s *Service) rebuildComponentGoingSeasonsContestListMap(goingSeasonsContestListMap map[int64][]*model.ContestModel) {
	tmpAllComponent0Map := make(map[int64][]*model.ContestModel)
	tmpAllComponent1Map := make(map[int64][]*model.ContestModel)
	tmpAllComponent2Map := make(map[int64][]*model.ContestModel)
	tmpAllComponent3Map := make(map[int64][]*model.ContestModel)
	tmpAllComponent4Map := make(map[int64][]*model.ContestModel)
	tmpAllComponent5Map := make(map[int64][]*model.ContestModel)
	tmpAllComponent6Map := make(map[int64][]*model.ContestModel)
	tmpAllComponent7Map := make(map[int64][]*model.ContestModel)
	tmpAllComponent8Map := make(map[int64][]*model.ContestModel)
	tmpAllComponent9Map := make(map[int64][]*model.ContestModel)

	for sid, contestList := range goingSeasonsContestListMap {
		switch sidSharding(sid) {
		case _seasonSlice0:
			tmpAllComponent0Map[sid] = contestList
		case _seasonSlice1:
			tmpAllComponent1Map[sid] = contestList
		case _seasonSlice2:
			tmpAllComponent2Map[sid] = contestList
		case _seasonSlice3:
			tmpAllComponent3Map[sid] = contestList
		case _seasonSlice4:
			tmpAllComponent4Map[sid] = contestList
		case _seasonSlice5:
			tmpAllComponent5Map[sid] = contestList
		case _seasonSlice6:
			tmpAllComponent6Map[sid] = contestList
		case _seasonSlice7:
			tmpAllComponent7Map[sid] = contestList
		case _seasonSlice8:
			tmpAllComponent8Map[sid] = contestList
		case _seasonSlice9:
			tmpAllComponent9Map[sid] = contestList
		}
		// 单赛程的缓存
		s.storeContestCache(contestList)
	}
	// component all api
	seasonContestAllComponent0Map = tmpAllComponent0Map
	seasonContestAllComponent1Map = tmpAllComponent1Map
	seasonContestAllComponent2Map = tmpAllComponent2Map
	seasonContestAllComponent3Map = tmpAllComponent3Map
	seasonContestAllComponent4Map = tmpAllComponent4Map
	seasonContestAllComponent5Map = tmpAllComponent5Map
	seasonContestAllComponent6Map = tmpAllComponent6Map
	seasonContestAllComponent7Map = tmpAllComponent7Map
	seasonContestAllComponent8Map = tmpAllComponent8Map
	seasonContestAllComponent9Map = tmpAllComponent9Map
}

func (s *Service) storeContestCache(contestList []*model.ContestModel) {
	for _, v := range contestList {
		s.activeContestsCacheMap.Add(v.ID, v, _contestCacheTtl)
	}
}

func (s *Service) getSeasonContestsFromCache(seasonId int64) (res []*model.ContestModel) {
	var cache map[int64][]*model.ContestModel
	switch sidSharding(seasonId) {
	case _seasonSlice0:
		cache = seasonContestAllComponent0Map
	case _seasonSlice1:
		cache = seasonContestAllComponent1Map
	case _seasonSlice2:
		cache = seasonContestAllComponent2Map
	case _seasonSlice3:
		cache = seasonContestAllComponent3Map
	case _seasonSlice4:
		cache = seasonContestAllComponent4Map
	case _seasonSlice5:
		cache = seasonContestAllComponent5Map
	case _seasonSlice6:
		cache = seasonContestAllComponent6Map
	case _seasonSlice7:
		cache = seasonContestAllComponent7Map
	case _seasonSlice8:
		cache = seasonContestAllComponent8Map
	case _seasonSlice9:
		cache = seasonContestAllComponent9Map
	}
	if list, ok := cache[seasonId]; ok {
		res = list
	}
	return
}

func sidSharding(sid int64) int64 {
	return sid % _seasonSliceBase
}

// storeActiveTeams
func (s *Service) storeActiveTeams(ctx context.Context) {
	seasonIds, err := s.dao.GetActiveSeasonsCache(ctx)
	if err != nil {
		return
	}
	for _, seasonId := range seasonIds {
		seasonTeams, errG := s.GetSeasonTeams(ctx, seasonId, false, true)
		if errG != nil {
			continue
		}
		// 构造赛季队伍优先级缓存
		s.activeSeasonTeams[seasonId] = seasonTeams
		// 构造队伍级别缓存
		for _, seasonTeam := range seasonTeams {
			team, errGG := s.getTeamsModel(ctx, []int64{seasonTeam.Tid}, false, true)
			if errGG != nil || team[seasonTeam.Tid] == nil {
				continue
			}
			s.activeTeamsCacheMap.Add(seasonTeam.Tid, team, _teamCacheTtl)
		}
	}
}

// storeActiveSeries
func (s *Service) storeActiveSeries(ctx context.Context) {
	seasonIds, err := s.dao.GetActiveSeasonsCache(ctx)
	if err != nil {
		return
	}
	for _, seasonId := range seasonIds {
		seriesModels, errG := s.dao.GetSeasonSeriesListCache(ctx, seasonId)
		if errG != nil {
			continue
		}
		for _, seriesModel := range seriesModels {
			s.seriesCacheMap.Add(seriesModel.ID, seriesModel, _seriesCacheTtl)
		}
	}
}
