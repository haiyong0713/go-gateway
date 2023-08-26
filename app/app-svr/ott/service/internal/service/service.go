package service

import (
	"context"

	"go-gateway/app/app-svr/ott/service/conf"
	"go-gateway/app/app-svr/ott/service/internal/dao"

	arccli "git.bilibili.co/bapis/bapis-go/archive/service"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/robfig/cron"
)

// Service service.
type Service struct {
	c        *conf.Config
	dao      *dao.Dao
	ArcTypes map[int32]*arccli.Tp
	pgcTypes map[string]int
	cron     *cron.Cron
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:        c,
		dao:      dao.New(c),
		ArcTypes: make(map[int32]*arccli.Tp),
		pgcTypes: make(map[string]int),
		cron:     cron.New(),
	}
	for _, v := range s.c.Cfg.PGCTypes {
		s.pgcTypes[v] = 1
	}
	s.loadTypes()
	if err := s.cron.AddFunc(s.c.Cfg.PGCCron, s.loadTypes); err != nil {
		panic(err)
	}
	s.cron.Start()
	return s
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}
