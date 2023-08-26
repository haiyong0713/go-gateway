package rank

import (
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/admin/conf"
	rankDao "go-gateway/app/web-svr/activity/admin/dao/rank"
	"go-gateway/app/web-svr/activity/admin/service/account"
	"go-gateway/app/web-svr/activity/admin/service/archive"
)

// Service struct
type Service struct {
	c       *conf.Config
	dao     rankDao.Dao
	account *account.Service
	archive *archive.Service
	cache   *fanout.Fanout
}

// Close service
func (s *Service) Close() {
	if s.dao != nil {
		s.dao.Close()
	}
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		dao:     rankDao.New(c),
		account: account.New(c),
		archive: archive.New(c),
		cache:   fanout.New("rank_v1", fanout.Worker(1), fanout.Buffer(1024)),
	}
	return
}
