package service

import (
	"context"
	accountGRPC "git.bilibili.co/bapis/bapis-go/account/service"
	relationGRPC "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/space/admin/conf"
	"go-gateway/app/web-svr/space/admin/dao"
)

// Service biz service def.
type Service struct {
	c             *conf.Config
	dao           *dao.Dao
	cache         *fanout.Fanout
	relaGRPC      relationGRPC.RelationClient
	accountClient accountGRPC.AccountClient
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   dao.New(c),
		cache: fanout.New("cache"),
	}
	var err error
	if s.relaGRPC, err = relationGRPC.NewClient(s.c.RelationGRPC); err != nil {
		panic(err)
	}

	if s.accountClient, err = accountGRPC.NewClient(nil); err != nil {
		panic(err)
	}

	if err = s.UpdateTabState(); err != nil {
		log.Error("cron start UpdateTabState error:%v", err)
	}
	if err = s.UpdateWhitelistState(); err != nil {
		log.Error("cron start UpdateWhitelistState error:%v", err)
	}
	return s
}

// Ping .
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}
