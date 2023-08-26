package channel

import (
	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/dao"
	chDao "go-gateway/app/web-svr/web/interface/dao/channel"
)

type Service struct {
	c     *conf.Config
	chDao *chDao.Dao
	dao   *dao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		chDao: chDao.New(c),
		dao:   dao.New(c),
	}
	return
}
