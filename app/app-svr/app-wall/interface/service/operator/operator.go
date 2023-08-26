package operator

import (
	"time"

	"go-gateway/app/app-svr/app-wall/interface/conf"
	"go-gateway/app/app-svr/app-wall/interface/model/operator"
)

// Service reddot service
type Service struct {
	c     *conf.Config
	cache *operator.Reddot
}

// New reddot new
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		cache: &operator.Reddot{},
	}
	s.loadCache(c)
	return
}

// Reddot get reddot
func (s *Service) Reddot(now time.Time) (res *operator.Reddot) {
	res = s.cache
	if res != nil {
		current := now.Unix()
		if current > int64(res.EndTime) || current < int64(res.StartTime) {
			res = &operator.Reddot{}
		}
	}
	return
}

func (s *Service) loadCache(c *conf.Config) {
	tmp := &operator.Reddot{}
	tmp.ReddotChange(c.Reddot.StartTime, c.Reddot.EndTime)
	s.cache = tmp
}
