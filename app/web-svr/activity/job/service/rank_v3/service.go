package rank

import (
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/conf"
	rank "go-gateway/app/web-svr/activity/job/dao/rank_v3"
	wechat "go-gateway/app/web-svr/activity/job/dao/wechat"
)

// Service service
type Service struct {
	c         *conf.Config
	dao       rank.Dao
	wechatDao wechat.Dao
	client    *httpx.Client
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:         c,
		dao:       rank.New(c),
		wechatDao: wechat.New(c),
		client:    httpx.NewClient(c.HTTPClient),
	}
	return s
}
