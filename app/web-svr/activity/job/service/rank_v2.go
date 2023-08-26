package service

import (
	"context"
	"go-common/library/net/trace"
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
)

var rankCtx context.Context

func rankInit() {
	rankCtx = trace.SimpleServerTrace(context.Background(), "rank")
}

// RankJob ...
func (s *Service) RankJob(id int64, attributeType uint) (err error) {
	rankInit()
	// 查询rank
	rank, err := s.rankSvr.GetRankConfig(rankCtx, id)
	if err != nil {
		return err
	}
	s.rankSvr.Rank(rank, attributeType)
	return nil
}

// RankCronMap ...
func (s *Service) RankCronMap() map[string]*rankmdl.Rank {
	return s.rankSvr.GetRankCronMap()
}

// StartRankJob ...
func (s *Service) StartRankJob() {
	defer s.waiter.Done()
	s.rankSvr.RankJob()
}

// CronStartJob ...
func (s *Service) CronStartJob() {
	s.rankSvr.RankJob()
}
