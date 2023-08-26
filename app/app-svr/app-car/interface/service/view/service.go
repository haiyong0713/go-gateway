package view

import (
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-car/interface/conf"
	arcdao "go-gateway/app/app-svr/app-car/interface/dao/archive"
	bgmdao "go-gateway/app/app-svr/app-car/interface/dao/bangumi"
	favdao "go-gateway/app/app-svr/app-car/interface/dao/favorite"
	fmdao "go-gateway/app/app-svr/app-car/interface/dao/fm"
	historydao "go-gateway/app/app-svr/app-car/interface/dao/history"
	reldao "go-gateway/app/app-svr/app-car/interface/dao/relation"
	serialDao "go-gateway/app/app-svr/app-car/interface/dao/serial"
	silverdao "go-gateway/app/app-svr/app-car/interface/dao/silverbullet"
	thumbupdao "go-gateway/app/app-svr/app-car/interface/dao/thumbup"
)

type Service struct {
	c          *conf.Config
	arc        *arcdao.Dao
	his        *historydao.Dao
	bgm        *bgmdao.Dao
	reldao     *reldao.Dao
	fav        *favdao.Dao
	silverDao  *silverdao.Dao
	thumbupDao *thumbupdao.Dao
	fmDao      *fmdao.Dao
	serialDao  *serialDao.Dao
	fmReport   *fanout.Fanout
}

func New(c *conf.Config) *Service {
	s := &Service{
		c:          c,
		arc:        arcdao.New(c),
		his:        historydao.New(c),
		bgm:        bgmdao.New(c),
		reldao:     reldao.New(c),
		fav:        favdao.New(c),
		silverDao:  silverdao.New(c),
		thumbupDao: thumbupdao.New(c),
		fmDao:      fmdao.New(c),
		serialDao:  serialDao.New(c),
	}
	s.fmReport = fanout.New("fmReport", fanout.Worker(2))
	return s
}
