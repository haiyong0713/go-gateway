package bugly

import (
	"context"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// Service struct.
type Service struct {
	c     *conf.Config
	fkDao *fkdao.Dao
}

// New new service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		fkDao: fkdao.New(c),
	}
	return
}

// Ping dao.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.fkDao.Ping(c); err != nil {
		log.Error("s.dao error(%v)", err)
	}
	return
}

// Close dao.
func (s *Service) Close() {
	s.fkDao.Close()
}
