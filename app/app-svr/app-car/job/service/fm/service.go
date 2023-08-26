package fm

import (
	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-car/job/conf"
	arcDao "go-gateway/app/app-svr/app-car/job/dao/archive"
	fmdao "go-gateway/app/app-svr/app-car/job/dao/fm"
)

type Service struct {
	c           *conf.Config
	arc         *arcDao.Dao
	fmSeasonGun *railgun.Railgun
	fmDao       *fmdao.Dao
}

func New(c *conf.Config) *Service {
	s := &Service{
		c:     c,
		arc:   arcDao.New(c),
		fmDao: fmdao.New(c),
	}
	// 接收算法侧 FM特殊聚合（合集）的稿件推送
	s.initFmSeasonInputer(c)
	return s
}

func (s *Service) Close() {
	s.fmSeasonGun.Close()
	log.Info("s.fmSeasonGun Closed.")
}
