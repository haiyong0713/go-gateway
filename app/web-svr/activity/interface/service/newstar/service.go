package newstar

import (
	"context"

	accAPI "git.bilibili.co/bapis/bapis-go/account/service"
	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/newstar"
	mdlNewstar "go-gateway/app/web-svr/activity/interface/model/newstar"

	relationAPI "git.bilibili.co/bapis/bapis-go/account/service/relation"
	upAPI "git.bilibili.co/bapis/bapis-go/archive/service/up"
)

type Service struct {
	c              *conf.Config
	dao            *newstar.Dao
	cache          *fanout.Fanout
	accClient      accAPI.AccountClient
	upClient       upAPI.UpClient
	memberClient   memberAPI.MemberClient
	relationClient relationAPI.RelationClient
	newstarAwards  map[string][]*mdlNewstar.NewstarAward
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   newstar.New(c),
		cache: fanout.New("cache"),
	}
	var err error
	if s.accClient, err = accAPI.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.upClient, err = upAPI.NewClient(c.UpClientNew); err != nil {
		panic(err)
	}
	if s.memberClient, err = memberAPI.NewClient(c.MemberClient); err != nil {
		panic(err)
	}
	if s.relationClient, err = relationAPI.NewClient(c.RelationClient); err != nil {
		panic(err)
	}
	s.initStarAward()
	return
}

func (s *Service) initStarAward() {
	s.newstarAwards = make(map[string][]*mdlNewstar.NewstarAward, 0)
	res, err := s.dao.RawAwards(context.Background())
	if err != nil {
		log.Error("newstar initStarAward s.dao.RawAwards error(%+v)", err)
		return
	}
	if len(res) == 0 {
		return
	}
	s.newstarAwards = res
}
