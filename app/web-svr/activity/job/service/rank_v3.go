package service

import (
	"context"
	"go-common/library/log"
	"go-common/library/net/trace"
)

var rankV3Ctx context.Context

func rankV3Init() {
	rankV3Ctx = trace.SimpleServerTrace(context.Background(), "rank")
}

// SetRankLog ...
func (s *Service) SetRankLog() (err error) {
	rankV3Init()
	// 查询rank
	err = s.rankV3Svr.SetRankLog(rankV3Ctx)
	if err != nil {
		log.Errorc(rankV3Ctx, " s.rankV3Svr.SetRankLog err(%v)", err)
		return err
	}
	return nil
}

// SetRankLogCron 排行榜
func (s *Service) SetRankLogCron() {
	s.rankRunning.Lock()
	defer s.rankRunning.Unlock()

	s.SetRankLog()

}
