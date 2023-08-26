package wall

import (
	"context"

	log "go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-wall/interface/conf"
	walldao "go-gateway/app/app-svr/app-wall/interface/dao/wall"
	"go-gateway/app/app-svr/app-wall/interface/model/wall"

	"github.com/robfig/cron"
)

type Service struct {
	c      *conf.Config
	client *httpx.Client
	dao    *walldao.Dao
	cache  []*wall.Wall
	cron   *cron.Cron
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:      c,
		client: httpx.NewClient(c.HTTPClient),
		dao:    walldao.New(c),
		cron:   cron.New(),
	}
	s.load()
	// 间隔2分钟
	if err := s.cron.AddFunc("@every 2m", s.load); err != nil {
		panic(err)
	}
	s.cron.Start()
	return
}

// Wall get all.
func (s *Service) Wall() (res []*wall.Wall) {
	res = s.cache
	return
}

// load WallAll.
func (s *Service) load() {
	res, err := s.dao.WallAll(context.TODO())
	if err != nil {
		log.Error("s.dao.wallAll error(%v)", err)
		return
	}
	s.cache = res
	log.Info("loadWallsCache success")
}

func (s *Service) Close() {
	s.cron.Stop()
}
