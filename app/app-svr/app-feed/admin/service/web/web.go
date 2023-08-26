package web

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"
	arcdao "go-gateway/app/app-svr/app-feed/admin/dao/archive"
	"go-gateway/app/app-svr/app-feed/admin/dao/game"
	"go-gateway/app/app-svr/app-feed/admin/dao/manager"
	"go-gateway/app/app-svr/app-feed/admin/dao/search"
	"go-gateway/app/app-svr/app-feed/admin/dao/show"

	"github.com/robfig/cron"
)

// Service is search service
type Service struct {
	dao        *search.Dao
	showDao    *show.Dao
	cronHot    *cron.Cron
	HotFre     string
	cronDark   *cron.Cron
	DarkFre    string
	GameDao    *game.Dao
	arcDao     *arcdao.Dao
	managerDao *manager.Dao
}

// New new a search service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao:        search.New(c),
		showDao:    show.New(c),
		cronHot:    cron.New(),
		cronDark:   cron.New(),
		HotFre:     c.Cfg.HotCroFre,
		DarkFre:    c.Cfg.DarkCroFre,
		GameDao:    game.New(c),
		arcDao:     arcdao.New(c),
		managerDao: manager.New(c),
	}
	return
}
