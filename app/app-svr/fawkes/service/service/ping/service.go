package ping

import (
	"context"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// Service service struct info.
type Service struct {
	c     *conf.Config
	fkDao *fawkes.Dao
}

// New service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		fkDao: fawkes.New(c),
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
