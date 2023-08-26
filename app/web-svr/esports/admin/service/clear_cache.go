package service

import (
	"context"
	"encoding/json"

	"go-common/library/log"
	"go-gateway/app/web-svr/esports/admin/tool"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
)

const (
	clearCacheBiz4ESportOfTeam    = "team"
	clearCacheBiz4ESportOfContest = "contest"
	clearCacheBiz4ESportOfSeason  = "season"
)

func (s *Service) ClearESportCacheByType(cacheType v1.ClearCacheType, list []int64) (err error) {
	if len(list) == 0 {
		return
	}

	bizName := ""
	switch cacheType {
	case v1.ClearCacheType_CONTEST:
		bizName = clearCacheBiz4ESportOfContest
	case v1.ClearCacheType_SEASON:
		bizName = clearCacheBiz4ESportOfSeason
	case v1.ClearCacheType_TEAM:
		bizName = clearCacheBiz4ESportOfTeam
	}

	req := new(v1.ClearCacheRequest)
	{
		req.CacheType = cacheType
		req.CacheKeys = list
	}
	reqBs, _ := json.Marshal(req)
	for i := 1; i <= 3; i++ {
		_, err = s.espClient.ClearCache(context.Background(), req)
		if err == nil {
			break
		}

		log.Error("ClearCache occur err: %v, req: %v, try times: %v", err, string(reqBs), i)
	}

	if bizName != "" {
		status := tool.StatusOfSucceed
		if err != nil {
			status = tool.StatusOfFailed
		}

		tool.AddClearCacheMetric(bizName, status)
	}

	return
}

func (s *Service) ClearComponentContestCacheByGRPC(param *v1.ClearComponentContestCacheRequest) (err error) {
	if _, err = s.espClient.ClearComponentContestCache(context.Background(), param); err != nil {
		log.Error("contest component ClearComponentContestCacheGRPC param(%+v) error(%+v)", param, err)
	}
	return
}

func (s *Service) ClearMatchSeasonsCacheByGRPC(param *v1.ClearMatchSeasonsCacheRequest) (err error) {
	if _, err = s.espClient.ClearMatchSeasonsCache(context.Background(), param); err != nil {
		log.Error("MatchSeasonsInfo ClearMatchSeasonsCacheByGRPC param(%+v) error(%+v)", param, err)
	}
	return
}

func (s *Service) ClearVideoListCacheByGRPC(id int64) (err error) {
	if id == 0 {
		return
	}
	ctx := context.Background()
	arg := &v1.ClearTopicVideoListRequest{ID: id}
	if _, err = s.espClient.ClearTopicVideoListCache(ctx, arg); err != nil {
		log.Errorc(ctx, "contest component ClearTopicVideoListCache param(%+v) error(%+v)", arg, err)
	}
	return
}
