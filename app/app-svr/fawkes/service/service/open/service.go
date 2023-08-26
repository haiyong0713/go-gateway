// Package open 用于管理fawkes的开放接口
package open

import (
	"context"
	"reflect"

	"github.com/asaskevich/EventBus"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// Service struct.
type Service struct {
	c         *conf.Config
	fkDao     *fkdao.Dao
	OpenPaths []reflect.Value
	event     EventBus.Bus
}

// New service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		fkDao: fkdao.New(c),
		event: EventBus.New(),
	}

	if err := s.event.Subscribe(AuthAddEvent, s.PathAuthAddAction); err != nil {
		panic(err)
	}
	if err := s.event.Subscribe(AuthUpdateEvent, s.PathAuthUpdateAction); err != nil {
		panic(err)
	}
	if err := s.event.Subscribe(AuthDeleteEvent, s.PathAuthDeleteAction); err != nil {
		panic(err)
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

func (s *Service) isSupervisor(ctx context.Context, username string) (bool, error) {
	// 超管权限
	if username != "" {
		supervisorRole, err := s.fkDao.AuthSupervisor(ctx, username)
		if err != nil {
			return false, err
		}
		if len(supervisorRole) > 0 {
			return true, nil
		}
	}
	return false, nil
}
