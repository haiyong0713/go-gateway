package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/common/helper"
	pb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/model"
)

func (s *Service) GetTeamsInfoBySeasonContests(ctx context.Context, seasonId int64, contests []*pb.ContestBattleCardComponent) (response map[int64][]*model.ContestTeamInfo, err error) {
	contestIds := make([]int64, 0)
	contestInfoMap := make(map[int64]*pb.ContestBattleCardComponent)
	for _, v := range contests {
		contestInfoMap[v.ID] = v
		contestIds = append(contestIds, v.ID)
	}
	res, err := s.fetchTeamsScoreInfoFromCache(ctx, contestIds)
	if err != nil {
		log.Errorc(ctx, "[Service][GetTeamsInfoBySeasonContests][fetchTeamsScoreInfoFromCache][Error], err:(%+v)", err)
		return
	}
	cacheMissContestIds := make([]int64, 0)
	for _, contestId := range contestIds {
		if _, isOk := res[contestId]; !isOk {
			cacheMissContestIds = append(cacheMissContestIds, contestId)
		}
	}
	response = s.formatTeamsInfoByCache(ctx, res)
	rebuildContestTeamsInfo, err := s.RawContestTeamsCacheByContestIds(ctx, seasonId, cacheMissContestIds, contestInfoMap)
	if err != nil {
		return
	}
	for contestId, contestTeamList := range rebuildContestTeamsInfo {
		response[contestId] = contestTeamList
	}
	return
}

func (s *Service) fetchTeamsScoreInfoFromCache(ctx context.Context, contestIds []int64) (res map[int64][]*model.ContestTeamScoreInfo, err error) {
	res = s.fetchTeamsLocalCacheByContestIds(ctx, contestIds)

	fetchMcContestIds := make([]int64, 0)
	for _, contestId := range contestIds {
		if _, isOK := res[contestId]; !isOK {
			fetchMcContestIds = append(fetchMcContestIds, contestId)
		}
	}
	if len(fetchMcContestIds) == 0 {
		return
	}

	mcRes, err := s.dao.FetchTeamsMcCacheByContestIds(ctx, contestIds)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTeams][GetTeamsInfoByContestIds][FetchCache][Error], err:(%+v)", err)
		return
	}
	for contestId, contestTeamsList := range mcRes {
		res[contestId] = contestTeamsList
	}
	return
}

func (s *Service) fetchTeamsLocalCacheByContestIds(ctx context.Context, contestIds []int64) (res map[int64][]*model.ContestTeamScoreInfo) {
	res = make(map[int64][]*model.ContestTeamScoreInfo)
	if contestIds == nil {
		return
	}
	if len(contestIds) == 0 {
		return
	}
	for _, contestId := range contestIds {
		v, isOk := goingSeasonsContestsTeams.Load(contestId)
		if !isOk {
			continue
		}
		value, ok := v.([]*model.ContestTeamScoreInfo)
		if !ok {
			log.Errorc(ctx, "[Service][ContestTeams][LoadSyncMap][Error], value valid, ok:%+v, v:%+v", ok, value)
			continue
		}
		contestsTeams := make([]*model.ContestTeamScoreInfo, 0)
		for _, contestTeam := range value {
			contestsTeams = append(contestsTeams, &model.ContestTeamScoreInfo{
				TeamId:         contestTeam.TeamId,
				Score:          contestTeam.Score,
				KillNumber:     contestTeam.KillNumber,
				SurvivalRank:   contestTeam.SurvivalRank,
				SeasonTeamRank: contestTeam.SeasonTeamRank,
				Rank:           contestTeam.Rank,
			})
		}
		res[contestId] = contestsTeams
	}
	return
}

func (s *Service) GetTeamsInfoBySeasonContestsSkipLocalCache(ctx context.Context, seasonId int64, contests []*pb.ContestBattleCardComponent) (teamsInfoMap map[int64][]*model.ContestTeamInfo, err error) {
	contestIds := make([]int64, 0)
	contestInfoMap := make(map[int64]*pb.ContestBattleCardComponent)
	for _, v := range contests {
		contestInfoMap[v.ID] = v
		contestIds = append(contestIds, v.ID)
	}
	res, err := s.dao.FetchTeamsMcCacheByContestIds(ctx, contestIds)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTemas][GetTeamsInfoByContestIds][FetchCache][Error], err:(%+v)", err)
		return
	}
	cacheMissContestIds := make([]int64, 0)
	teamsInfoMap = make(map[int64][]*model.ContestTeamInfo)
	for _, contestId := range contestIds {
		if _, isOk := res[contestId]; !isOk {
			cacheMissContestIds = append(cacheMissContestIds, contestId)
		}
	}
	response := s.formatTeamsInfoByCache(ctx, res)
	rebuildContestTeamsInfo, err := s.RawContestTeamsCacheByContestIds(ctx, seasonId, cacheMissContestIds, contestInfoMap)
	if err != nil {
		return
	}
	for contestId, contestTeamList := range rebuildContestTeamsInfo {
		response[contestId] = contestTeamList
	}
	return
}

func (s *Service) formatTeamsInfoByCache(ctx context.Context, res map[int64][]*model.ContestTeamScoreInfo) (response map[int64][]*model.ContestTeamInfo) {
	teamsInfoMap := s.GetAllTeamsOfComponent(ctx)
	response = make(map[int64][]*model.ContestTeamInfo)
	for contestId, contestTeamsInfo := range res {
		contestTeamList := make([]*model.ContestTeamInfo, 0)
		for _, teamInfo := range contestTeamsInfo {
			contestTeamList = append(contestTeamList, &model.ContestTeamInfo{
				TeamId:   teamInfo.TeamId,
				TeamInfo: teamsInfoMap[teamInfo.TeamId],
				ScoreInfo: &model.ContestTeamScoreInfo{
					TeamId:         teamInfo.TeamId,
					Score:          teamInfo.Score,
					KillNumber:     teamInfo.KillNumber,
					SurvivalRank:   teamInfo.SurvivalRank,
					SeasonTeamRank: teamInfo.SeasonTeamRank,
					Rank:           teamInfo.Rank,
				},
			})
		}
		response[contestId] = contestTeamList
	}
	return
}

func (s *Service) RawContestTeamsCacheByContestIds(
	ctx context.Context,
	seasonId int64,
	contestIds []int64,
	contestInfoMap map[int64]*pb.ContestBattleCardComponent,
) (finalContestTeamsMap map[int64][]*model.ContestTeamInfo, err error) {
	finalContestTeamsMap = make(map[int64][]*model.ContestTeamInfo)
	if len(contestIds) == 0 {
		return
	}
	teamsDbInfo, err := s.dao.GetTeamsByContestIds(ctx, contestIds)
	if err != nil {
		log.Errorc(ctx, "[Dao][ContestTeams][RawContestTeamsCacheByContestIds][Error], err(%+v)", err)
		return
	}
	rebuildInfoMap, nameSortContestIds := s.rebuildContestInfoFilter(teamsDbInfo, contestInfoMap)
	seasonInfoMap, err := s.getSeasonTeamsBySeasonId(ctx, seasonId)
	if err != nil {
		log.Errorc(ctx, "[Dao][ContestTeams][getSeasonTeamsBySeasonId][Error], err(%+v)", err)
		return
	}
	for _, contestId := range contestIds {
		_, isRebuild := rebuildInfoMap[contestId]
		if !isRebuild {
			continue
		}
		if _, isNameSortContest := nameSortContestIds[contestId]; isNameSortContest {
			finalContestTeamsMap[contestId] = s.doTeamsSecondSortByScoreAndName(ctx, rebuildInfoMap[contestId])
			continue
		}
		_, validContest := contestInfoMap[contestId]
		seasonInfo, validSeason := seasonInfoMap[contestInfoMap[contestId].SeasonID]
		if !(validSeason && validContest) {
			finalContestTeamsMap[contestId] = make([]*model.ContestTeamInfo, 0)
		} else {
			finalContestTeamsMap[contestId] = s.doTeamsSecondSortBySeasonRank(ctx, rebuildInfoMap[contestId], seasonInfo)
		}
	}
	err = s.dao.RebuildTeamsMcCacheByContestTeamInfos(ctx, finalContestTeamsMap)
	return
}

func (s *Service) rebuildContestInfoFilter(teamsDbInfo []*model.ContestTeamDbInfo, contestInfoMap map[int64]*pb.ContestBattleCardComponent) (rebuildInfoMap map[int64][]*model.ContestTeamDbInfo, nameSortContestIds map[int64]bool) {
	rebuildInfoMap = make(map[int64][]*model.ContestTeamDbInfo)
	nameSortContestIds = make(map[int64]bool)
	for _, v := range teamsDbInfo {
		if _, isOk := rebuildInfoMap[v.ContestId]; !isOk {
			list := make([]*model.ContestTeamDbInfo, 0)
			rebuildInfoMap[v.ContestId] = append(list, v)
		} else {
			rebuildInfoMap[v.ContestId] = append(rebuildInfoMap[v.ContestId], v)
		}

		contestInfo, isOk := contestInfoMap[v.ContestId]
		if !isOk {
			continue
		}
		if _, isOk := nameSortContestIds[v.ContestId]; !isOk && checkIsScore(contestInfo, v.SurvivalRank) {
			// 已编辑结果信息
			nameSortContestIds[v.ContestId] = true
		}
	}
	return
}

func (s *Service) doTeamsSecondSortByScoreAndName(ctx context.Context, teams []*model.ContestTeamDbInfo) (contestTeamInfos []*model.ContestTeamInfo) {
	teamsInfoMap := s.GetAllTeamsOfComponent(ctx)
	contestTeamInfos = make([]*model.ContestTeamInfo, 0)
	for _, v := range teams {
		if _, isOk := teamsInfoMap[v.TeamId]; !isOk {
			continue
		}
		contestTeamInfos = append(contestTeamInfos, &model.ContestTeamInfo{
			TeamId:   v.TeamId,
			TeamInfo: teamsInfoMap[v.TeamId],
			ScoreInfo: &model.ContestTeamScoreInfo{
				TeamId:         v.TeamId,
				Score:          v.Score,
				KillNumber:     v.KillNumber,
				SurvivalRank:   v.SurvivalRank,
				SeasonTeamRank: 0,
				Rank:           0,
			},
		})
	}
	helper.SortByScoreAndTitle(contestTeamInfos)
	for index, v := range contestTeamInfos {
		v.ScoreInfo.Rank = int64(index + 1)
	}
	return
}

func (s *Service) doTeamsSecondSortBySeasonRank(
	ctx context.Context, teams []*model.ContestTeamDbInfo, seasonTeams map[int64]*model.TeamInSeason) (contestTeamInfos []*model.ContestTeamInfo) {
	teamsInfoMap := s.GetAllTeamsOfComponent(ctx)
	contestTeamInfos = make([]*model.ContestTeamInfo, 0)
	for _, v := range teams {
		if _, isOk := teamsInfoMap[v.TeamId]; !isOk {
			continue
		}
		seasonTeamRank := int64(0)
		if seasonTeam, isOk := seasonTeams[v.TeamId]; isOk {
			seasonTeamRank = seasonTeam.Rank
		}
		contestTeamInfos = append(contestTeamInfos, &model.ContestTeamInfo{
			TeamId:   v.TeamId,
			TeamInfo: teamsInfoMap[v.TeamId],
			ScoreInfo: &model.ContestTeamScoreInfo{
				TeamId:         v.TeamId,
				Score:          v.Score,
				KillNumber:     v.KillNumber,
				SurvivalRank:   v.SurvivalRank,
				SeasonTeamRank: seasonTeamRank,
				Rank:           0,
			},
		})
	}
	helper.SortByRankAndTitle(contestTeamInfos)
	for index, v := range contestTeamInfos {
		v.ScoreInfo.Rank = int64(index + 1)
	}
	return
}

func (s *Service) getSeasonTeamsBySeasonId(ctx context.Context, seasonId int64) (res map[int64]map[int64]*model.TeamInSeason, err error) {
	res = make(map[int64]map[int64]*model.TeamInSeason)
	if seasonId == 0 {
		return
	}
	teamsInfo, err := s.GetTeamsInSeason(ctx, seasonId)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTeams][GetTeamsInSeason][Error], err:(%+v)", err)
		return
	}
	res[seasonId] = make(map[int64]*model.TeamInSeason)
	for _, teamInfo := range teamsInfo {
		res[seasonId][teamInfo.TeamId] = teamInfo
	}
	return
}

func checkIsScore(contestInfo *pb.ContestBattleCardComponent, survivalRank int64) bool {
	return contestInfo.Status == model.ContestStatusOfEnd && survivalRank != 0
}
