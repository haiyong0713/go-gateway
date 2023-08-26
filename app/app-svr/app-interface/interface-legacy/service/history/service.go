package history

import (
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	accdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/account"
	archivedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/archive"
	artdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/article"
	bangumidao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/bangumi"
	cheesedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/cheese"
	gamedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/game"
	historydao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/history"
	livedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/live"
	reldao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/relation"
)

// Service service struct
type Service struct {
	c          *conf.Config
	historyDao *historydao.Dao
	liveDao    *livedao.Dao
	artDao     *artdao.Dao
	accDao     *accdao.Dao
	cheeseDao  *cheesedao.Dao
	bangumiDao *bangumidao.Dao
	relDao     *reldao.Dao
	arcDao     *archivedao.Dao
	gameDao    *gamedao.Dao
}

// New new service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		historyDao: historydao.New(c),
		liveDao:    livedao.New(c),
		artDao:     artdao.New(c),
		accDao:     accdao.New(c),
		cheeseDao:  cheesedao.New(c),
		bangumiDao: bangumidao.New(c),
		relDao:     reldao.New(c),
		arcDao:     archivedao.New(c),
		gameDao:    gamedao.New(c),
	}
	return
}
