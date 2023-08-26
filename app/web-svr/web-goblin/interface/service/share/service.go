package share

import (
	"context"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	usersuitgrpc "git.bilibili.co/bapis/bapis-go/account/service/usersuit"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/web-goblin/interface/conf"
	"go-gateway/app/web-svr/web-goblin/interface/dao/share"
)

// Service service struct.
type Service struct {
	c   *conf.Config
	dao *share.Dao
	// cache proc
	cache    *fanout.Fanout
	suit     usersuitgrpc.UsersuitClient
	accGRPC  accgrpc.AccountClient
	Pendants map[int64]int64
}

// New new service.
func New(c *conf.Config) *Service {
	s := &Service{
		c:        c,
		dao:      share.New(c),
		cache:    fanout.New("goblin cache", fanout.Worker(1), fanout.Buffer(1024)),
		Pendants: make(map[int64]int64),
	}
	var err error
	if s.accGRPC, err = accgrpc.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.suit, err = usersuitgrpc.NewClient(nil); err != nil {
		panic(err)
	}
	s.loadPendant()
	return s
}

// Ping ping service.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.dao.Ping(c); err != nil {
		log.Error("s.dao.Ping error(%v)", err)
	}
	return
}
