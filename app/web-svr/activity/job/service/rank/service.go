package rank

import (
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/dao/like"
	rank "go-gateway/app/web-svr/activity/job/dao/rank_v2"
	"go-gateway/app/web-svr/activity/job/service/source"

	"github.com/robfig/cron"
)

// Service service
type Service struct {
	c       *conf.Config
	dao     *like.Dao
	rankDao rank.Dao
	source  *source.Service
	cron    *cron.Cron
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		source:  source.New(c),
		dao:     like.New(c),
		rankDao: rank.New(c),
		cron:    cron.New(),
	}

	return s
}
