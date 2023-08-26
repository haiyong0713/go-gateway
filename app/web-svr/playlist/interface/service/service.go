package service

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"

	accclient "git.bilibili.co/bapis/bapis-go/account/service"
	accwarden "git.bilibili.co/bapis/bapis-go/account/service"
	favclient "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	arcclient "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/playlist/interface/conf"
	"go-gateway/app/web-svr/playlist/interface/dao"
)

// Service service struct.
type Service struct {
	c   *conf.Config
	dao *dao.Dao
	// cache proc
	cache *fanout.Fanout
	// playlist power mids
	allowMids map[int64]struct{}
	maxSort   int64
	arcClient arcclient.ArchiveClient
	accClient accwarden.AccountClient
	favClient favclient.FavoriteClient
}

// New new service.
func New(c *conf.Config) *Service {
	s := &Service{
		c:       c,
		dao:     dao.New(c),
		cache:   fanout.New("playlist cache", fanout.Buffer(1024)),
		maxSort: c.Rule.MinSort + 4*c.Rule.SortStep*int64(c.Rule.MaxVideoCnt),
	}
	var err error
	if s.arcClient, err = arcclient.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.accClient, err = accclient.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.favClient, err = favclient.NewClient(c.FavClient); err != nil {
		panic(err)
	}
	s.initMids()
	return s
}

func (s *Service) initMids() {
	tmp := make(map[int64]struct{}, len(s.c.Rule.PowerMids))
	for _, id := range s.c.Rule.PowerMids {
		tmp[id] = struct{}{}
	}
	s.allowMids = tmp
}

// Ping ping service.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.dao.Ping(c); err != nil {
		log.Error("s.dao.Ping error(%v)", err)
	}
	return
}
