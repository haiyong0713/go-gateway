package service

import (
	"context"
	"time"

	"go-gateway/app/app-svr/newmont/service/api"

	"go-common/library/log"

	opIcon "git.bilibili.co/bapis/bapis-go/manager/operation/icon"
)

type sidebarIconLoader struct {
	IconCache     map[int64]*api.MngIcon
	operationIcon map[int64]string
	preIconCache  map[int64]*api.MngIcon
	load          func() error
}

func (s *sidebarIconLoader) Timer() string {
	return "@every 5s"
}

func (s *sidebarIconLoader) Load() error {
	return s.load()
}

func (s *Service) loadIconCache() error {
	ics, err := s.sectionDao.Icons(context.Background(), time.Now(), time.Now())
	if err != nil {
		return err
	}
	s.IconCache = ics
	log.Info("loadIconCache success")

	preIcs, err := s.sectionDao.Icons(context.Background(),
		time.Now(),
		time.Now().Add(time.Duration(s.c.IconCacheConfig.PreloadDuration)*time.Hour))
	if err != nil {
		return err
	}
	s.preIconCache = preIcs

	res, err := s.opIconClient.List(context.Background(), &opIcon.ListReq{})
	if err != nil {
		return err
	}
	tempOperationIcon := make(map[int64]string, len(res.Icon))
	for _, v := range res.Icon {
		tempOperationIcon[int64(v.Id)] = v.Picture
	}
	s.operationIcon = tempOperationIcon
	return nil
}
