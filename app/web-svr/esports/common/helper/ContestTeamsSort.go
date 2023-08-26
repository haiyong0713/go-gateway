package helper

import (
	"go-gateway/app/web-svr/esports/interface/model"
	"sort"
)

func SortByScoreAndTitle(contestTeamInfos []*model.ContestTeamInfo) {
	sort.Slice(contestTeamInfos, func(i, j int) bool {
		if contestTeamInfos[i].ScoreInfo.Score > contestTeamInfos[j].ScoreInfo.Score {
			return true
		}
		if contestTeamInfos[i].ScoreInfo.Score == contestTeamInfos[j].ScoreInfo.Score {
			if contestTeamInfos[i].ScoreInfo.KillNumber > contestTeamInfos[j].ScoreInfo.KillNumber {
				return true
			}
			if contestTeamInfos[i].ScoreInfo.KillNumber == contestTeamInfos[j].ScoreInfo.KillNumber {
				return contestTeamInfos[i].TeamInfo.Title < contestTeamInfos[j].TeamInfo.Title
			}
		}
		return false
	})
}

func SortByRankAndTitle(contestTeamInfos []*model.ContestTeamInfo) {
	sort.Slice(contestTeamInfos, func(i, j int) bool {
		if contestTeamInfos[i].ScoreInfo.SeasonTeamRank > contestTeamInfos[j].ScoreInfo.SeasonTeamRank {
			return true
		}
		if contestTeamInfos[i].ScoreInfo.SeasonTeamRank == contestTeamInfos[j].ScoreInfo.SeasonTeamRank {
			return contestTeamInfos[i].TeamInfo.Title < contestTeamInfos[j].TeamInfo.Title
		}
		return false
	})
}
