package system

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/system"
)

type Service struct {
	c   *conf.Config
	dao *system.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: system.New(c),
	}

	// 拉取OA员工全量数据
	if err := s.InitEmployeesInfo(context.Background()); err != nil {
		panic(err)
	}

	return s
}
