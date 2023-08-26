package service

import (
	"context"
	"github.com/jinzhu/gorm"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/model"
	"go-gateway/app/web-svr/esports/common/helper"
	model2 "go-gateway/app/web-svr/esports/interface/model"
	"sort"
)

var (
	_contestUpdateLockRetryTimes = 5
	_contestUpdateLockSleepMs    = int64(300)
)

func (s *Service) ContestTeamsAdd(ctx context.Context, contestId int64, seasonId int64, teamIdsStr string) (err error) {
	teamIds, err := s.teamsStringSplit(ctx, teamIdsStr)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "填写的战队信息解析失败，请先检查战队填写")
		return
	}
	if len(teamIds) == 0 {
		return
	}
	teamsCheckResponse, err := s.ContestTeamsCheck(ctx, seasonId, teamIds)
	if err != nil || teamsCheckResponse == nil || teamsCheckResponse.Teams == nil || len(teamsCheckResponse.Teams) == 0 {
		if err != nil {
			return
		}
		err = xecode.Errorf(xecode.RequestErr, "获取战队信息失败，请重试")
		return
	}
	err = s.dao.BatchAddTeams(ctx, nil, contestId, teamIds)
	return
}

func (s *Service) ContestTeamsUpdate(ctx context.Context, contest *model.Contest, tx *gorm.DB) (err error) {
	teamIds, _, err := s.checkTeamsCheckResponse(ctx, contest)
	if err != nil {
		return
	}
	contestId := contest.ID
	lockKey, lockValue, ttl := s.dao.GetContestTeamsUpdateLockInfo(contestId)
	err = s.dao.RedisLock(ctx, lockKey, lockValue, ttl, _contestUpdateLockRetryTimes, _contestUpdateLockSleepMs)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTeamsUpdate][RedisLock][Error], error:(%+v)", err)
		return
	}
	defer func() {
		err = s.dao.RedisUnLock(ctx, lockKey, lockValue)
		if err != nil {
			log.Errorc(ctx, "[Service][ContestTeamsUpdate][RedisUnLock][Error], error:(%+v)", err)
		}
	}()

	contestTeams, err := s.dao.GetTeamsOrderBySurvivalRank(ctx, contestId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "获取战队配置失败，请重试")
		return
	}
	log.Infoc(ctx, "[Service][DB][Transaction][Begin][Success]")
	// 如果有队伍变更, 则将已配置的生存排名顺序重建，重建规则：被保留的战队排名依次前移即可
	newTeamIdsMap := make(map[int64]bool)
	for _, teamId := range teamIds {
		newTeamIdsMap[teamId] = true
	}
	rebuildRank := false
	oldTeamIds := make([]int64, 0)
	waitForReBuildRankTeamsMap := make(map[int64]*model.ContestTeam)
	index := 0
	for _, contestTeam := range contestTeams {
		oldTeamIds = append(oldTeamIds, contestTeam.TeamId)
		if contestTeam.RankEditStatus != model.ContestTeamRankEditStatusOn {
			continue
		}
		rebuildRank = true
		contestTeam.SurvivalRank = int64(index + 1)
		index++
		waitForReBuildRankTeamsMap[contestTeam.TeamId] = contestTeam
	}
	// 队伍未更新的话不进行后续操作
	if int64SliceCompare(teamIds, oldTeamIds) {
		return
	}
	err = s.dao.BatchDeleteTeamsByContestId(ctx, tx, contestId)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTeamsUpdate][Do][Error], error:(%+v)", err)
		return
	}
	if rebuildRank {
		err = s.batchAddTeamsHandlerWhenUpdate(ctx, contest, teamIds, waitForReBuildRankTeamsMap, tx)
	} else {
		err = s.dao.BatchAddTeams(ctx, tx, contestId, teamIds)
	}
	return
}

func (s *Service) batchAddTeamsHandlerWhenUpdate(ctx context.Context, contest *model.Contest, teamIds []int64, waitForReBuildRankTeamsMap map[int64]*model.ContestTeam, tx *gorm.DB) (err error) {
	contestId := contest.ID
	insertTeamInfos := make([]*model.ContestTeam, 0)
	for _, teamId := range teamIds {
		if v, isOk := waitForReBuildRankTeamsMap[teamId]; isOk {
			insertTeamInfos = append(insertTeamInfos, v)
		} else {
			insertTeamInfos = append(insertTeamInfos, s.contestTeamDefaultFormat(contestId, teamId))
		}
	}
	err = s.teamsScoreCalculate(ctx, contest.SeriesID, insertTeamInfos)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTeamsUpdate][teamsScoreCalculate][Error], error:(%+v)", err)
		return
	}
	err = s.dao.BatchAddTeamsWhenUpdate(ctx, tx, insertTeamInfos)
	return
}

func (s *Service) checkTeamsCheckResponse(ctx context.Context, contest *model.Contest) (teamIds []int64, teamsCheckResponse *model.ContestTeamsCheckResponse, err error) {
	if contest == nil || contest.ID == 0 {
		err = xecode.Errorf(xecode.RequestErr, "赛程信息缺失，赛程战队保存失败")
		return
	}
	seasonId := contest.Sid
	teamIdsStr := contest.TeamIds
	teamIds, err = s.teamsStringSplit(ctx, teamIdsStr)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTeamsUpdate][Do][Error], error:(%+v)", err)
		err = xecode.Errorf(xecode.RequestErr, "战队输入格式不正确")
		return
	}
	teamsCheckResponse, err = s.ContestTeamsCheck(ctx, seasonId, teamIds)
	if err != nil {
		return
	}
	if len(teamIds) != 0 &&
		(teamsCheckResponse == nil || teamsCheckResponse.Teams == nil || len(teamsCheckResponse.Teams) == 0) {
		err = xecode.Errorf(xecode.RequestErr, "获取战队信息失败，请重试")
		return
	}
	return
}

func (s *Service) teamsScoreCalculate(ctx context.Context, contestSeriesId int64, contestTeams []*model.ContestTeam) (err error) {
	scoreRules, err := s.GetScoreRules(ctx, contestSeriesId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "获取赛程阶段信息失败, 请重试")
		return
	}
	for _, contestTeam := range contestTeams {
		survivalScore := int64(0)
		if contestTeam.SurvivalRank != 0 && contestTeam.SurvivalRank <= int64(len(scoreRules.RankScores)) {
			survivalScore = scoreRules.RankScores[contestTeam.SurvivalRank-1]
		}
		contestTeam.Score = survivalScore + contestTeam.KillNumber*scoreRules.KillScore
	}
	return
}

func (s *Service) contestTeamDefaultFormat(contestId int64, teamId int64) (contestTeam *model.ContestTeam) {
	return &model.ContestTeam{
		ContestId:      contestId,
		TeamId:         teamId,
		SurvivalRank:   0,
		KillNumber:     0,
		Score:          0,
		RankEditStatus: model.ContestTeamRankEditStatusOff,
	}
}

func (s *Service) ContestTeamsCheck(ctx context.Context, seasonId int64, teamIds []int64) (checkResponse *model.ContestTeamsCheckResponse, err error) {
	checkResponse = &model.ContestTeamsCheckResponse{}
	if len(teamIds) == 0 {
		return
	}
	// teams重复性校验
	idsMap := make(map[int64]bool)
	for _, v := range teamIds {
		if _, isOk := idsMap[v]; isOk {
			err = xecode.Errorf(xecode.RequestErr, "包含重复战队id: %d，请先检查战队id列表", v)
			return
		} else {
			idsMap[v] = true
		}
	}

	teamsInSeasonResponse, err := s.ListTeamInSeason(ctx, seasonId)
	if err != nil {
		log.Errorc(ctx, "teamsValidCheck error By method[ListTeamSeason], err:(%+v)", err)
		return
	}
	teamsInfoMap := make(map[int64]*model.Team)
	for _, team := range teamsInSeasonResponse {
		if team == nil {
			continue
		}
		teamsInfoMap[team.Team.ID] = team.Team
	}
	teamsMiss := make([]int64, 0)
	teamsInfo := make([]*model.Team, 0)
	for _, teamId := range teamIds {
		if teamInfo, isOk := teamsInfoMap[teamId]; isOk {
			teamsInfo = append(teamsInfo, teamInfo)
		} else {
			teamsMiss = append(teamsMiss, teamId)
		}
	}
	if len(teamsMiss) > 0 {
		log.Errorc(ctx, "输入的战队id：%s不在该赛季下的战队名单中", xstr.JoinInts(teamsMiss))
		err = xecode.Errorf(xecode.RequestErr, "输入的战队id：%s不在该赛季下的战队名单中", xstr.JoinInts(teamsMiss))
		return
	}
	checkResponse = &model.ContestTeamsCheckResponse{
		Teams: teamsInfo,
	}
	return
}

func (s *Service) GetContestTeams(ctx context.Context, contestId int64) (contestTeams []*model.ContestTeam, err error) {
	contestTeams, err = s.dao.GetTeamList(ctx, contestId)
	return
}

func (s *Service) teamsStringSplit(ctx context.Context, teamIdsStr string) (teamIds []int64, err error) {
	if teamIds, err = xstr.SplitInts(teamIdsStr); err != nil {
		log.Errorc(ctx, "teamsStringSplit Error, err:(%+v)", err)
		return
	}
	return
}

func (s *Service) GetContestTeamsOrderBySurvivalRank(ctx context.Context, contestId int64) (response *model.ContestTeamScoresResponse, err error) {
	contestTeamsFromDb, err := s.dao.GetTeamsOrderBySurvivalRank(ctx, contestId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "获取赛程战队比赛结果失败，请重试")
		return
	}
	contestTeams := make([]*model.ContestTeam, 0)
	for _, contestTeam := range contestTeamsFromDb {
		if contestTeam.RankEditStatus != model.ContestTeamRankEditStatusOn {
			break
		}
		contestTeams = append(contestTeams, contestTeam)
	}
	response = &model.ContestTeamScoresResponse{
		TeamScores: contestTeams,
	}
	return
}

func (s *Service) SaveContestTeamsBySurvivalRank(ctx context.Context, contestId int64, teamScores []*model.TeamScores) (err error) {
	contestInfo, err := s.dao.GetContest(ctx, contestId)
	if err != nil {
		log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][ContestInfo][Get][Error], error:(%+v)", err)
		return
	}
	lockKey, lockValue, ttl := s.dao.GetContestTeamsUpdateLockInfo(contestId)
	err = s.dao.RedisLock(ctx, lockKey, lockValue, ttl, _contestUpdateLockRetryTimes, _contestUpdateLockSleepMs)
	if err != nil {
		log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][RedisLock][Error], error:(%+v)", err)
		return
	}
	defer func() {
		err = s.dao.RedisUnLock(ctx, lockKey, lockValue)
		if err != nil {
			log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][RedisUnLock][Error], error:(%+v)", err)
		}
	}()
	contestTeams, err := s.dao.GetTeamList(ctx, contestId)
	if err != nil {
		log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][GetTeamList][Error], error:(%+v)", err)
		err = xecode.Errorf(xecode.RequestErr, "获取战队配置失败，请重试")
		return
	}
	waitForReBuildRankTeamsMap := make(map[int64]*model.ContestTeam)
	for index, teamInfo := range teamScores {
		survivalRank := int64(index + 1)
		waitForReBuildRankTeamsMap[teamInfo.ID] = &model.ContestTeam{
			ID:             0,
			ContestId:      contestId,
			TeamId:         teamInfo.ID,
			SurvivalRank:   survivalRank,
			KillNumber:     teamInfo.KillNumber,
			Score:          0,
			RankEditStatus: 1,
			IsDeleted:      0,
		}
	}
	err = s.teamsUpdateByTransaction(ctx, contestInfo, waitForReBuildRankTeamsMap, contestTeams)
	if err != nil {
		return
	}
	err = s.rebuildContestTeamsCache(ctx, contestInfo)
	if err != nil {
		log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][RebuildContestTeamsCache][Error], error:(%+v)", err)
		err = xecode.Errorf(xecode.RequestErr, "更新战队结果失败，请重试")
	}
	return
}

func (s *Service) teamsUpdateByTransaction(
	ctx context.Context,
	contestInfo *model.Contest,
	waitForReBuildRankTeamsMap map[int64]*model.ContestTeam,
	contestTeams []*model.ContestTeam,
) (err error) {
	contestId := contestInfo.ID
	tx := s.dao.DB.Begin()
	err = s.dao.BatchDeleteTeamsByContestId(ctx, nil, contestId)
	if err != nil {
		log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][BatchDeleteTeamsByContestId][Error], error:(%+v)", err)
		err = tx.Rollback().Error
		if err != nil {
			log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][BatchDeleteTeamsByContestId][Error], error:(%+v)", err)
		}
		err = xecode.Errorf(xecode.RequestErr, "保存战队生存排名与击杀数失败，请重试")
		return
	}

	insertTeamInfos := make([]*model.ContestTeam, 0)
	for _, team := range contestTeams {
		if v, isOk := waitForReBuildRankTeamsMap[team.TeamId]; isOk {
			insertTeamInfos = append(insertTeamInfos, v)
			continue
		}
		insertTeamInfos = append(insertTeamInfos, s.contestTeamDefaultFormat(contestId, team.TeamId))
	}
	err = s.teamsScoreCalculate(ctx, contestInfo.SeriesID, insertTeamInfos)
	if err != nil {
		log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][teamsScoreCalculate][Error], error:(%+v)", err)
		err = tx.Rollback().Error
		if err != nil {
			log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][TeamScores][Rollback][Error], error:(%+v)", err)
		}
		return
	}
	err = s.dao.BatchAddTeamsWhenUpdate(ctx, nil, insertTeamInfos)
	if err != nil {
		log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][BatchAddTeamsWhenUpdate][Error], error:(%+v)", err)
		err = tx.Rollback().Error
		if err != nil {
			log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][TeamScores][Rollback][Error], error:(%+v)", err)
		}
		err = xecode.Errorf(xecode.RequestErr, "保存战队生存排名与击杀数失败，请重试")
		return
	}
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		log.Errorc(ctx, "[Service][SaveContestTeamsBySurvivalRank][TeamScores][Commit][Error], error:(%+v)", err)
	}
	return
}

func int64SliceCompare(slice1 []int64, slice2 []int64) bool {
	int64SliceSort(slice1)
	int64SliceSort(slice2)
	if len(slice1) != len(slice2) {
		return false
	}
	for index, v := range slice1 {
		if v != slice2[index] {
			return false
		}
	}
	return true
}

func int64SliceSort(slice []int64) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
}

func (s *Service) AsyncRebuildContestTeamsCache(ctx context.Context, contest *model.Contest) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				return
			}
		}()
		errG := s.rebuildContestTeamsCache(ctx, contest)
		if errG != nil {
			log.Errorc(ctx, "[Service][Contest][Add][rebuildContestTeamsCache][Error], err:(%+v)", errG)
		}
	}()
}

func (s *Service) rebuildContestTeamsCache(ctx context.Context, contest *model.Contest) (err error) {
	contestId := contest.ID
	contestTeams, err := s.dao.GetTeamList(ctx, contestId)
	if err != nil {
		log.Errorc(ctx, "[Service][rebuildContestTeamsCache][contestId:%d][GetTeamList], err:(%+v)", contestId, err)
		return
	}
	seasonTeamsInfoMap, err := s.getSeasonTeamsBySeasonId(ctx, contest.Sid)
	if err != nil {
		log.Errorc(ctx, "[Service][rebuildContestTeamsCache][getSeasonTeamsBySeasonId][contestId:%d], err:(%+v)", contestId, err)
		return
	}
	editSurvivalRank := false
	for _, contestTeam := range contestTeams {
		if contestTeam.RankEditStatus == model.ContestTeamRankEditStatusOn {
			editSurvivalRank = true
			break
		}
	}
	var contestTeamsList []*model2.ContestTeamInfo
	if contest.CalculateStatus() == model2.ContestStatusOfEnd && editSurvivalRank {
		contestTeamsList = s.doTeamsSecondSortByScoreAndName(contestTeams, seasonTeamsInfoMap)
	} else {
		contestTeamsList = s.doTeamsSecondSortBySeasonRank(contestTeams, seasonTeamsInfoMap)
	}
	err = s.dao.RebuildTeamsMcCacheByContestTeamInfos(ctx, contestId, contestTeamsList)
	if err != nil {
		log.Errorc(ctx, "[Service][rebuildContestTeamsCache][RebuildTeamsMcCacheByContestTeamInfos][Error][contestId:%d], err:(%+v)", contestId, err)
	}
	return
}

func (s *Service) doTeamsSecondSortByScoreAndName(teams []*model.ContestTeam, teamsInfoMap map[int64]*model.TeamInSeasonResponse) (contestTeamInfos []*model2.ContestTeamInfo) {
	contestTeamInfos = make([]*model2.ContestTeamInfo, 0)
	for _, v := range teams {
		if _, isOk := teamsInfoMap[v.TeamId]; !isOk {
			continue
		}
		contestTeamInfos = append(contestTeamInfos, &model2.ContestTeamInfo{
			TeamId: v.TeamId,
			TeamInfo: &model2.Team2TabComponent{
				ID:          v.TeamId,
				Title:       teamsInfoMap[v.TeamId].Title,
				SubTitle:    teamsInfoMap[v.TeamId].SubTitle,
				Logo:        "",
				RegionID:    0,
				Region:      "",
				ScoreTeamID: 0,
			},
			ScoreInfo: &model2.ContestTeamScoreInfo{
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
	teams []*model.ContestTeam,
	teamsInfoMap map[int64]*model.TeamInSeasonResponse,
) (contestTeamInfos []*model2.ContestTeamInfo) {
	contestTeamInfos = make([]*model2.ContestTeamInfo, 0)
	for _, v := range teams {
		if _, isOk := teamsInfoMap[v.TeamId]; !isOk {
			continue
		}
		seasonTeamRank := int64(0)
		if seasonTeam, isOk := teamsInfoMap[v.TeamId]; isOk {
			seasonTeamRank = seasonTeam.Rank
		}
		contestTeamInfos = append(contestTeamInfos, &model2.ContestTeamInfo{
			TeamId: v.TeamId,
			TeamInfo: &model2.Team2TabComponent{
				ID:          v.TeamId,
				Title:       teamsInfoMap[v.TeamId].Title,
				SubTitle:    teamsInfoMap[v.TeamId].SubTitle,
				Logo:        "",
				RegionID:    0,
				Region:      "",
				ScoreTeamID: 0,
			},
			ScoreInfo: &model2.ContestTeamScoreInfo{
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

func (s *Service) getSeasonTeamsBySeasonId(ctx context.Context, seasonId int64) (res map[int64]*model.TeamInSeasonResponse, err error) {
	res = make(map[int64]*model.TeamInSeasonResponse)
	teamsInfo, err := s.ListTeamInSeason(ctx, seasonId)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTeams][GetTeamsInSeason][Error], err:(%+v)", err)
		return
	}
	for _, teamInfo := range teamsInfo {
		res[teamInfo.ID] = teamInfo
	}
	return
}
