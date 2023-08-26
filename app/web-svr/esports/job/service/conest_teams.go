package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/common/helper"
	pb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/model"
)

func (s *Service) RawContestTeamsCacheByContestIds(
	ctx context.Context,
	seasonId int64,
	contestIds []int64,
	contestInfoMap map[int64]*pb.ContestBattleCardComponent,
) (err error) {
	if len(contestIds) == 0 {
		return
	}
	for _, contestId := range contestIds {
		contestInfo, isOk := contestInfoMap[contestId]
		if !isOk {
			continue
		}
		errSingle := s.rawSingleContestCache(ctx, seasonId, contestInfo)
		if errSingle != nil {
			err = errSingle
		}
	}
	return
}

func (s *Service) rawSingleContestCache(ctx context.Context, seasonId int64, contestInfo *pb.ContestBattleCardComponent) (err error) {
	contestId := contestInfo.ID
	teamsInSeasonMap, err := s.getSeasonTeamsBySeasonId(ctx, seasonId)
	if err != nil {
		log.Errorc(ctx, "[Job][ContestTeams][RawContestTeamsCacheByContestIds][getSeasonTeamsBySeasonId][Error], err(%+v)", err)
		return
	}
	teamsDbInfo, err := s.dao.GetTeamsByContestIds(ctx, []int64{contestId})
	if err != nil {
		log.Errorc(ctx, "[Job][ContestTeams][RawContestTeamsCacheByContestIds][GetTeamsByContestIds][Error], err(%+v)", err)
		return
	}
	rebuildInfo := make([]*model.ContestTeamDbInfo, 0)
	nameSort := false
	for _, v := range teamsDbInfo {
		rebuildInfo = append(rebuildInfo, v)
		if checkIsScore(contestInfo, v.SurvivalRank) {
			// 已编辑结果信息
			nameSort = true
		}
	}
	var finalContestTeams []*model.ContestTeamInfo
	if nameSort {
		finalContestTeams = s.doTeamsSecondSortByScoreAndName(rebuildInfo, teamsInSeasonMap)
	} else {
		finalContestTeams = s.doTeamsSecondSortBySeasonRank(rebuildInfo, teamsInSeasonMap)
	}
	err = s.dao.RebuildTeamsMcCacheByContestTeamInfos(ctx, contestId, finalContestTeams)
	if err != nil {
		log.Errorc(ctx, "[Job][ContestTeams][RawContestTeamsCacheByContestIds][getSeasonTeamsBySeasonId][Error], err(%+v)", err)
	}
	return
}

func (s *Service) doTeamsSecondSortByScoreAndName(teams []*model.ContestTeamDbInfo, teamsInfoMap map[int64]*model.TeamInSeason) (contestTeamInfos []*model.ContestTeamInfo) {
	contestTeamInfos = make([]*model.ContestTeamInfo, 0)
	for _, v := range teams {
		if _, isOk := teamsInfoMap[v.TeamId]; !isOk {
			continue
		}
		contestTeamInfos = append(contestTeamInfos, &model.ContestTeamInfo{
			TeamId: v.TeamId,
			TeamInfo: &model.Team2TabComponent{
				ID:          v.TeamId,
				Title:       teamsInfoMap[v.TeamId].TeamTitle,
				SubTitle:    teamsInfoMap[v.TeamId].TeamTitle,
				Logo:        "",
				RegionID:    0,
				Region:      "",
				ScoreTeamID: 0,
			},
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
	teams []*model.ContestTeamDbInfo, teamsInfoMap map[int64]*model.TeamInSeason) (contestTeamInfos []*model.ContestTeamInfo) {
	contestTeamInfos = make([]*model.ContestTeamInfo, 0)
	for _, v := range teams {
		if _, isOk := teamsInfoMap[v.TeamId]; !isOk {
			continue
		}
		seasonTeamRank := int64(0)
		if seasonTeam, isOk := teamsInfoMap[v.TeamId]; isOk {
			seasonTeamRank = seasonTeam.Rank
		}
		contestTeamInfos = append(contestTeamInfos, &model.ContestTeamInfo{
			TeamId: v.TeamId,
			TeamInfo: &model.Team2TabComponent{
				ID:          v.TeamId,
				Title:       teamsInfoMap[v.TeamId].TeamTitle,
				SubTitle:    teamsInfoMap[v.TeamId].TeamTitle,
				Logo:        "",
				RegionID:    0,
				Region:      "",
				ScoreTeamID: 0,
			},
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

func (s *Service) getSeasonTeamsBySeasonId(ctx context.Context, seasonId int64) (res map[int64]*model.TeamInSeason, err error) {
	res = make(map[int64]*model.TeamInSeason)
	teamsInfo, err := s.dao.GetTeamsInSeasonFromDB(ctx, seasonId)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTeams][GetTeamsInSeason][Error], err:(%+v)", err)
		return
	}
	for _, teamInfo := range teamsInfo {
		res[teamInfo.TeamId] = teamInfo
	}
	return
}

func checkIsScore(contestInfo *pb.ContestBattleCardComponent, survivalRank int64) bool {
	return contestInfo.Status == model.ContestStatusOfEnd && survivalRank != 0
}
