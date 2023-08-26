package location

import (
	"context"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	locdao "go-gateway/app/app-svr/app-resource/interface/dao/location"
	"go-gateway/app/app-svr/app-resource/interface/model/location"
)

type Service struct {
	dao *locdao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao: locdao.New(c),
	}
	return
}

func (s *Service) Info(c context.Context) (ipInfo *location.Info, err error) {
	var (
		ip = metadata.String(c, metadata.RemoteIP)
	)
	if ipInfo, err = s.dao.Info(c, ip); err != nil {
		log.Error("s.dao.Info ip(%v),err(%v)", ip, err)
	}
	return
}
