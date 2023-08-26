package model

import (
	"fmt"
	"go-gateway/app/web-svr/esports/interface/model"
)

var (
	_contestTeamsCacheKey          = "esports:features:contestId:%d:teamsCache"
	_contestTeamsCacheKeyTtlSecond = int32(1800)
)

type ContestTeamsMcCache struct {
	ContestId   int64                         `json:"contest_id"`
	Teams       []*model.ContestTeamScoreInfo `json:"teams"`
	BuildMethod string                        `json:"build_method"`
}

func GetContestTeamsCacheKey(ContestId int64) string {
	return fmt.Sprintf(_contestTeamsCacheKey, ContestId)
}

func GetContestTeamsCacheKeyTtlSecond() int32 {
	return _contestTeamsCacheKeyTtlSecond
}
