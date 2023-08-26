package service

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/web/job/internal/model"
)

const _onlineListNum = 50

func (s *Service) setOnlineAids() {
	ctx := context.Background()
	var onlineData []*model.OnlineAid
	if err := retry(func() (err error) {
		onlineData, err = s.dao.OnlineAids(ctx, _onlineListNum)
		return err
	}); err != nil {
		log.Error("日志告警 OnlineAids error:%+v", err)
		return
	}
	if err := retry(func() (err error) {
		return s.dao.AddCacheOnlineAids(ctx, onlineData)
	}); err != nil {
		log.Error("日志告警 AddCacheOnlineAids error:%+v", err)
		return
	}
}
