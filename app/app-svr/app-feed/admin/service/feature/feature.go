package feature

import (
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	dao "go-gateway/app/app-svr/app-feed/admin/dao/feature"
)

type Service struct {
	dao       *dao.Dao
	logWorker *fanout.Fanout
}

func New(c *conf.Config) *Service {
	s := &Service{
		dao:       dao.New(c),
		logWorker: fanout.New("log_worker"),
	}
	return s
}
