package dynamic

import (
	"context"

	"go-common/component/tinker"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	accdao "go-gateway/app/app-svr/app-dynamic/interface/dao/account"
	arcdao "go-gateway/app/app-svr/app-dynamic/interface/dao/archive"
	dyndao "go-gateway/app/app-svr/app-dynamic/interface/dao/dynamic"
	geodao "go-gateway/app/app-svr/app-dynamic/interface/dao/geo"
	pgcdao "go-gateway/app/app-svr/app-dynamic/interface/dao/pgc"

	"github.com/robfig/cron"
)

type Service struct {
	c      *conf.Config
	arcDao *arcdao.Dao
	dynDao *dyndao.Dao
	accDao *accdao.Dao
	pgcDao *pgcdao.Dao
	GeoDao *geodao.Dao
	cron   *cron.Cron
	// bottom config map
	bottomMap map[string]*conf.BottomItem
	// Dynamic List Res
	resRcmd map[int64]struct{}
	// infoc
	svideoInfoc infocV2.Infoc
	// abtest
	tinker *tinker.ABTest
}

func New(c *conf.Config, infoc infocV2.Infoc) (s *Service) {
	s = &Service{
		c:           c,
		arcDao:      arcdao.New(c),
		dynDao:      dyndao.New(c),
		cron:        cron.New(),
		accDao:      accdao.New(c),
		pgcDao:      pgcdao.New(c),
		GeoDao:      geodao.New(c),
		resRcmd:     make(map[int64]struct{}),
		svideoInfoc: infoc,
	}
	s.bottomMap = makeBottomMap(c)
	s.tinker = tinker.Init(s.svideoInfoc, nil)
	s.loadRecommend()
	s.loadCron()
	return
}

func (s *Service) loadRecommend() {
	var (
		rcmd map[int64]struct{}
		err  error
	)
	if rcmd, err = s.dynDao.Recommend(context.Background()); err != nil {
		log.Error("cron Recommend error %v", err)
		return
	}
	s.resRcmd = rcmd
}

func (s *Service) loadCron() {
	if err := s.cron.AddFunc(s.c.Tick.RcmdCron, s.loadRecommend); err != nil {
		panic(err)
	}
	s.cron.Start()
}

func makeBottomMap(c *conf.Config) map[string]*conf.BottomItem {
	if c.BottomConfig == nil || c.BottomConfig.TopicJumpLinks == nil {
		return nil
	}
	var res = make(map[string]*conf.BottomItem)
	for _, bottom := range c.BottomConfig.TopicJumpLinks {
		btmTmp := bottom
		if len(bottom.RelatedTopic) == 0 {
			continue
		}
		for _, topic := range bottom.RelatedTopic {
			res[topic] = &btmTmp
		}
	}
	return res
}
