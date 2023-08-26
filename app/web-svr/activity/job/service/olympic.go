package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	"time"
)

const (
	_defaultRefreshOlympicTicker = 20
)

func (s *Service) refreshOlympicContest(ctx context.Context) (err error) {
	duration := time.Duration(_defaultRefreshOlympicTicker) * time.Second
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			err = s.doRefreshOlympicContest(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) doRefreshOlympicContest(ctx context.Context) (err error) {
	resp, err := s.actGRPC.GetOlympicQueryConfig(ctx, &api.GetOlympicQueryConfigReq{
		SkipCache: false,
	})
	if err != nil {
		log.Errorc(ctx, "[doRefreshOlympicContest][GetOlympicQueryConfig][Error], err:%+v", err)
		return
	}
	if resp == nil || resp.QueryConfigs == nil || len(resp.QueryConfigs) == 0 {
		return
	}
	configs := resp.QueryConfigs
	for _, config := range configs {
		_, errG := s.actGRPC.GetOlympicContestDetail(ctx, &api.GetOlympicContestDetailReq{
			Id:        config.ContestId,
			SkipCache: true,
		})
		if errG != nil {
			log.Errorc(ctx, "[doRefreshOlympicContest][GetOlympicContestDetail][Error], err:%+v", errG)
		}
	}
	return
}
