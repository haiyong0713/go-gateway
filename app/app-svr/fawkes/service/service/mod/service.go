package mod

import (
	"context"

	"go-common/library/database/boss"
	"go-common/library/sync/pipeline/fanout"

	"github.com/asaskevich/EventBus"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// Service service struct info.
type Service struct {
	c     *conf.Config
	fkDao *fawkes.Dao
	boss  *boss.Boss
	cache *fanout.Fanout
	event EventBus.Bus
}

// New service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		fkDao: fawkes.New(c),
		boss:  boss.New(c.BossConfig),
		cache: fanout.New("cache"),
		event: EventBus.New(),
	}
	s.EventInit()
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
	s.cache.Close()
	s.fkDao.Close()
}
