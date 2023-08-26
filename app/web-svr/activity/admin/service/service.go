package service

import (
	"context"
	actapi "go-gateway/app/web-svr/activity/interface/api"

	acccli "git.bilibili.co/bapis/bapis-go/account/service"
	arcclient "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/admin/conf"
	"go-gateway/app/web-svr/activity/admin/dao"

	tagrpcBapis "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	actPlatCli "git.bilibili.co/bapis/bapis-go/platform/admin/act-plat"
	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	tunnelapi "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"

	"github.com/jinzhu/gorm"
)

// Service biz service def.
type Service struct {
	c         *conf.Config
	dao       *dao.Dao
	DB        *gorm.DB
	accClient acccli.AccountClient
	tagGRPC   tagrpcBapis.TagRPCClient

	artClient       artapi.ArticleGRPCClient
	arcClient       arcclient.ArchiveClient
	platAdminClient actPlatCli.ActPlatAdminClient
	tunnelClient    tunnelapi.TunnelClient
	actPlatClient   actplatapi.ActPlatClient
	actClient       actapi.ActivityClient
}

// New new a Service and return.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: dao.New(c),
	}
	s.DB = s.dao.DB
	UpdateCropWeChat(c.Notifier)
	var err error
	if s.arcClient, err = arcclient.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.accClient, err = acccli.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.artClient, err = artapi.NewClient(c.ArtClient); err != nil {
		panic(err)
	}
	if s.platAdminClient, err = actPlatCli.NewClient(c.ActPlatAdminClient); err != nil {
		panic(err)
	}
	if s.tunnelClient, err = tunnelapi.NewClient(c.TunnelClient); err != nil {
		panic(err)
	}
	if s.actPlatClient, err = actplatapi.NewClient(c.ActPlatClient); err != nil {
		panic(err)
	}
	if s.tagGRPC, err = tagrpcBapis.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	if s.actClient, err = actapi.NewClient(c.ActClient); err != nil {
		panic(err)
	}
	return s
}

// Ping check dao health.
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}

// Wait wait all closed.
func (s *Service) Wait() {}

// Close close all dao.
func (s *Service) Close() {
	s.dao.Close()
}
