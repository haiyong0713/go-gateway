package task

import (
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
	act "go-gateway/app/web-svr/activity/interface/dao/actplat"
	"go-gateway/app/web-svr/activity/interface/dao/wechat"
)

// Service ...
type Service struct {
	c         *conf.Config
	actDao    *act.Dao
	cache     *fanout.Fanout
	wechatdao *wechat.Dao
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:         c,
		cache:     fanout.New("activity_task", fanout.Worker(1), fanout.Buffer(1024)),
		actDao:    act.New(c),
		wechatdao: wechat.New(c),
	}

	return s
}

// Close ...
func (s *Service) Close() {
	if s.actDao != nil {
		s.actDao.Close()
	}

}
