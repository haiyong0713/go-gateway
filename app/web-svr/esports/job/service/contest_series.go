package service

import (
	"context"
	"go-common/library/log"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/job/component"
	"go-gateway/app/web-svr/esports/job/conf"
	"go-gateway/app/web-svr/esports/job/tool"
	"time"
)

func (s *Service) RefreshAllContestSeriesInfoLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(conf.Conf.SeriesRefresh.RefreshDuration))
	log.Infoc(ctx, "RefreshAllContestSeriesInfoLoop started")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			seriesList, err := s.dao.GetAllSeriesInfo4Refresh(ctx)
			if err != nil {
				log.Errorc(ctx, "RefreshAllContestSeriesInfoLoop GetAllSeriesInfo4Refresh error: %v", err)
			}
			for _, series := range seriesList {
				if tool.Int64InSlice(series.ID, conf.Conf.SeriesRefresh.RefreshIgnoreIDList) {
					continue
				}
				switch series.Type {
				case 1:
					_, err = component.EspClient.RefreshSeriesPointMatchInfo(ctx, &v1.RefreshSeriesPointMatchInfoReq{SeriesId: series.ID})
				case 2:
					_, err = component.EspClient.RefreshSeriesKnockoutMatchInfo(ctx, &v1.RefreshSeriesKnockoutMatchInfoReq{SeriesId: series.ID})
				}
				if err != nil {
					log.Errorc(ctx, "RefreshAllContestSeriesInfoLoop for series id %v error: %v", series.ID, err)
				}
			}
		}
	}

}
