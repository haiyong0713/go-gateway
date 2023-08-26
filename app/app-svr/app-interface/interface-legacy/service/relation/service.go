package relation

import (
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	accdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/account"
	livedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/live"
	matchdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/match"
	reldao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/relation"
)

// Service is favorite.
type Service struct {
	c *conf.Config
	// dao
	accDao   *accdao.Dao
	relDao   *reldao.Dao
	liveDao  *livedao.Dao
	matchDao *matchdao.Dao
}

// New new favorite。
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
		// dao
		accDao:   accdao.New(c),
		relDao:   reldao.New(c),
		liveDao:  livedao.New(c),
		matchDao: matchdao.New(c),
	}
	return s
}
