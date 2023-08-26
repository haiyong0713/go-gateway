package service

import (
	"context"
	"go-common/library/cache/memcache"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

// GetTeamModel 获取战队信息 .
func (s *Service) GetTeamModel(ctx context.Context, req *pb.GetTeamModelReq) (res *pb.TeamModel, err error) {
	res = &pb.TeamModel{}
	teamInfo, err := s.getTeamsModel(ctx, []int64{req.GetTeamId()}, true, true)
	if err != nil {
		log.Errorc(ctx, "GetTeamModel  s.getTeamsModel() teamID(%d) error(%+v)", req.TeamId, err)
		return
	}
	if len(teamInfo) == 0 {
		err = xecode.Errorf(xecode.RequestErr, "战队不存在或已被冻结")
		return
	}
	res = s.teamModel2External(teamInfo[0])
	return
}

func (s *Service) GetTeamsModelMap(ctx context.Context, teamIds []int64) (
	teamInfoMap map[int64]*pb.TeamModel,
	err error,
) {
	teamInfoInternalMap, err := s.getTeamsModel(ctx, teamIds, false, false)
	if err != nil {
		return
	}
	teamInfoMap = make(map[int64]*pb.TeamModel)
	for teamId, teamModel := range teamInfoInternalMap {
		teamInfoMap[teamId] = s.teamModel2External(teamModel)
	}
	return
}

func (s *Service) ClearTeamCache(ctx context.Context, req *pb.ClearTeamCacheReq) (res *pb.NoArgsResponse, err error) {
	res = &pb.NoArgsResponse{}
	if req == nil || req.TeamId == 0 {
		return
	}
	err = s.dao.DeleteTeamCache(ctx, req.TeamId)
	return
}

func (s *Service) GetSeasonTeams(ctx context.Context, seasonId int64, skipCache bool, skipMemory bool) (seasonTeamModels []*model.SeasonTeamModel, err error) {
	if !skipMemory {
		if teamModelsCache, ok := s.activeSeasonTeams[seasonId]; ok {
			seasonTeamModels = teamModelsCache
			return
		}
	}
	if !skipCache {
		seasonTeamModels, err = s.dao.GetSeasonTeamsCache(ctx, seasonId)
		if err != nil && err != memcache.ErrNotFound {
			log.Errorc(ctx, "[Service][Team][GetSeasonTeams][GetSeasonTeamsCache][Error], err:%+v", err)
			return
		}
		if err == nil {
			return
		}
	}
	seasonTeamModels, err = s.dao.GetSeasonTeamsModel(ctx, seasonId)
	if err != nil {
		log.Errorc(ctx, "[Service][Team][GetSeasonTeams][GetSeasonTeamsModel][Error], err:%+v", err)
		return
	}
	if errD := s.dao.StoreSeasonTeamsCache(ctx, seasonTeamModels, seasonId); errD != nil {
		log.Errorc(ctx, "[Service][Team][GetSeasonTeams][StoreSeasonTeamsCache][Error], err:%+v", err)
	}
	return
}

func (s *Service) GetTeamsModel(ctx context.Context, teamIds []int64) (teamInfos []*pb.TeamModel, err error) {
	teamModels, err := s.dao.GetTeamsByIds(ctx, teamIds)
	if err != nil {
		return
	}
	teamInfos = make([]*pb.TeamModel, 0)
	for _, v := range teamModels {
		single := s.teamModel2External(v)
		teamInfos = append(teamInfos, single)
	}
	return
}

func (s *Service) getTeamsModel(ctx context.Context, teamIds []int64, skipCache bool, skipMemory bool) (
	teamsInfoMap map[int64]*model.TeamModel,
	err error,
) {
	teamsInfoMap = make(map[int64]*model.TeamModel)
	if len(teamIds) == 0 {
		return
	}
	missTeamIds := teamIds
	teamsInfoMap = make(map[int64]*model.TeamModel)
	if !skipMemory {
		teamsInfoMap, missTeamIds = s.getTeamsModelFromMemory(teamIds)
	}
	if !skipCache {
		teamsInfoMapFromCache, missTeamIdsFromCache, errG := s.getTeamsFromCache(ctx, missTeamIds)
		if errG != nil {
			err = errG
			log.Errorc(ctx, "[getTeamsModel][getTeamsInfoFormCache][Error], err:%+v", err)
			return
		}
		for k, v := range teamsInfoMapFromCache {
			teamsInfoMap[k] = v
		}
		missTeamIds = missTeamIdsFromCache
	}
	if len(missTeamIds) == 0 {
		return
	}
	teamsInfo, err := s.dao.GetTeamsByIds(ctx, missTeamIds)
	if err != nil {
		return
	}
	rebuildCacheMap := make(map[int64]*model.TeamModel)
	for _, teamInfo := range teamsInfo {
		rebuildCacheMap[teamInfo.ID] = teamInfo
		teamsInfoMap[teamInfo.ID] = teamInfo
	}
	_ = s.dao.SetTeamsCache(ctx, rebuildCacheMap)
	return
}

func (s *Service) getTeamsModelFromMemory(teamIds []int64) (
	teamsInfoMap map[int64]*model.TeamModel,
	missTeamIds []int64,
) {
	teamsInfoMap = make(map[int64]*model.TeamModel)
	missTeamIds = make([]int64, 0)
	for _, v := range teamIds {
		if cache, ok := s.activeTeamsCacheMap.Get(v); ok {
			if team, valid := cache.(*model.TeamModel); valid {
				teamsInfoMap[v] = team
				continue
			}
		}
		missTeamIds = append(missTeamIds, v)
	}
	return
}

func (s *Service) getTeamsFromCache(ctx context.Context, teamIds []int64) (
	teamsInfoMap map[int64]*model.TeamModel,
	missTeamIds []int64,
	err error,
) {
	teamsInfoMap = make(map[int64]*model.TeamModel)
	missTeamIds = make([]int64, 0)
	//获取缓存
	if len(teamIds) == 0 {
		return
	}
	redisCacheTeams, missIds, errC := s.dao.GetTeamsCache(ctx, teamIds)
	if errC != nil {
		err = errC
		return
	}
	for gameId, teamInfo := range redisCacheTeams {
		teamsInfoMap[gameId] = teamInfo
	}
	missTeamIds = missIds
	return
}

func (s *Service) teamModel2Internal(fromModel *pb.TeamModel) (toModel *model.TeamModel) {
	if fromModel == nil {
		return nil
	}
	return &model.TeamModel{
		ID:         fromModel.ID,
		Title:      fromModel.Title,
		SubTitle:   fromModel.SubTitle,
		ETitle:     fromModel.ETitle,
		Area:       fromModel.Area,
		Logo:       fromModel.Logo,
		UID:        fromModel.Uid,
		Members:    fromModel.Members,
		Dic:        fromModel.Dic,
		VideoUrl:   fromModel.VideoUrl,
		Profile:    fromModel.Profile,
		LeidaTId:   fromModel.LeidaTId,
		ReplyId:    fromModel.ReplyId,
		TeamType:   fromModel.TeamType,
		RegionId:   fromModel.RegionId,
		PictureUrl: fromModel.PictureUrl,
	}
}

func (s *Service) teamModel2External(fromModel *model.TeamModel) (toModel *pb.TeamModel) {
	if fromModel == nil {
		return nil
	}
	return &pb.TeamModel{
		ID:         fromModel.ID,
		Title:      fromModel.Title,
		SubTitle:   fromModel.SubTitle,
		ETitle:     fromModel.ETitle,
		Area:       fromModel.Area,
		Logo:       fromModel.Logo,
		Uid:        fromModel.UID,
		Members:    fromModel.Members,
		Dic:        fromModel.Dic,
		VideoUrl:   fromModel.VideoUrl,
		Profile:    fromModel.Profile,
		LeidaTId:   fromModel.LeidaTId,
		ReplyId:    fromModel.ReplyId,
		TeamType:   fromModel.TeamType,
		RegionId:   fromModel.RegionId,
		PictureUrl: fromModel.PictureUrl,
	}
}

func (s *Service) teamModel2ExternalInfo(fromModel *model.TeamModel) (toModel *pb.TeamDetail) {
	if fromModel == nil {
		return nil
	}
	return &pb.TeamDetail{
		ID:       fromModel.ID,
		Title:    fromModel.Title,
		SubTitle: fromModel.SubTitle,
		ETitle:   fromModel.ETitle,
		Area:     fromModel.Area,
		Logo:     fromModel.Logo,
		Uid:      fromModel.UID,
		Members:  fromModel.Members,
		Dic:      fromModel.Dic,
		TeamType: fromModel.TeamType,
		LogoFull: s.formatFullLogoPath(fromModel.Logo),
		RegionId: fromModel.RegionId,
	}
}
