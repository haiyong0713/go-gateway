package dubbing

import (
	"git.bilibili.co/bapis/bapis-go/account/service"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/dubbing"
	"go-gateway/app/web-svr/activity/interface/dao/rank"
	"go-gateway/app/web-svr/activity/interface/dao/remix"
	"go-gateway/app/web-svr/activity/interface/dao/task"
	"go-gateway/app/web-svr/activity/interface/service/account"
	"go-gateway/app/web-svr/activity/interface/service/archive"
	"go-gateway/app/web-svr/activity/interface/service/like"
)

// Service ...
type Service struct {
	c         *conf.Config
	remix     remix.Dao
	rank      rank.Dao
	dubbing   dubbing.Dao
	task      *task.Dao
	accClient api.AccountClient
	like      *like.Service
	account   *account.Service
	archive   *archive.Service
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		remix:   remix.New(c),
		rank:    rank.New(c),
		dubbing: dubbing.New(c),
		like:    like.New(c),
		account: account.New(c),
		archive: archive.New(c),
		task:    task.New(c),
	}
	var err error
	if s.accClient, err = api.NewClient(c.AccClient); err != nil {
		panic(err)
	}

	return s
}

// Close ...
func (s *Service) Close() {
	s.remix.Close()
	s.rank.Close()
}
