package rank

import (
	"go-gateway/app/web-svr/activity/interface/conf"
	rank "go-gateway/app/web-svr/activity/interface/dao/rank_v2"
	"go-gateway/app/web-svr/activity/interface/service/account"
	"go-gateway/app/web-svr/activity/interface/service/archive"
	"go-gateway/app/web-svr/activity/interface/service/like"
)

// Service ...
type Service struct {
	c       *conf.Config
	rank    rank.Dao
	like    *like.Service
	account *account.Service
	archive *archive.Service
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		rank:    rank.New(c),
		like:    like.New(c),
		account: account.New(c),
		archive: archive.New(c),
	}

	return s
}

// Close ...
func (s *Service) Close() {
	s.rank.Close()
}
