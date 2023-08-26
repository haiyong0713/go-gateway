package image

import (
	infocV2 "go-common/library/log/infoc.v2"
	"go-gateway/app/app-svr/hkt-note/interface/conf"
	imgDao "go-gateway/app/app-svr/hkt-note/interface/dao/image"
)

type Service struct {
	c          *conf.Config
	dao        *imgDao.Dao
	infocV2Log infocV2.Infoc
}

func New(c *conf.Config, infoc infocV2.Infoc) (s *Service) {
	s = &Service{
		c:          c,
		dao:        imgDao.New(c),
		infocV2Log: infoc,
	}
	return
}
