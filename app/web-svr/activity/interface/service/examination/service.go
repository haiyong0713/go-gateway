package examination

import (
	"sync/atomic"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/service/account"
	likesvr "go-gateway/app/web-svr/activity/interface/service/like"
)

// Service ...
type Service struct {
	c       *conf.Config
	config  *atomic.Value
	account *account.Service
	likeSvr *likesvr.Service
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		config:  &atomic.Value{},
		account: account.New(c),
		likeSvr: likesvr.New(c),
	}

	return s
}

// Close ...
func (s *Service) Close() {

}
