package fingerprint

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	fpdao "go-gateway/app/app-svr/app-resource/interface/dao/fingerprint"
	fpmdl "go-gateway/app/app-svr/app-resource/interface/model/fingerprint"
)

// Service module service.
type Service struct {
	dao *fpdao.Dao
}

// New new a module service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao: fpdao.New(c),
	}
	return
}

func (s *Service) Fingerprint(c context.Context, platfrom, buvid string, mid int64, body []byte) (res *fpmdl.Fingerprint, err error) {
	if res, err = s.dao.Fingerprint(c, platfrom, buvid, mid, body); err != nil {
		log.Error("%v", err)
	}
	return
}
