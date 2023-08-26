package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/conf"
	"time"
)

var (
	_defaultRefreshTicker = 60
)

func (s *Service) RefreshBindConfig(ctx context.Context) (err error) {
	duration := time.Duration(_defaultRefreshTicker) * time.Second
	if conf.Conf.BindConfig != nil && conf.Conf.BindConfig.RefreshTickerSecond != 0 {
		duration = time.Duration(conf.Conf.BindConfig.RefreshTickerSecond) * time.Second
	}
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			_, errG := client.ActivityClient.RefreshBindConfigCache(ctx, &api.NoReply{})
			if errG != nil {
				log.Errorc(ctx, "[RefreshBindConfig][ActivityClient][Error], err:%+v", errG)
			}
		case <-ctx.Done():
			return
		}
	}
}
