package mobile

import (
	"go-common/library/stat/prom"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-wall/interface/conf"
	mobileDao "go-gateway/app/app-svr/app-wall/interface/dao/mobile"
)

type Service struct {
	c   *conf.Config
	dao *mobileDao.Dao
	// prom
	pHit  *prom.Prom
	pMiss *prom.Prom
	// cache
	cache *fanout.Fanout
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: mobileDao.New(c),
		// prom
		pHit:  prom.CacheHit,
		pMiss: prom.CacheMiss,
		// cache
		cache: fanout.New("cache", fanout.Buffer(10240)),
	}
	return
}
