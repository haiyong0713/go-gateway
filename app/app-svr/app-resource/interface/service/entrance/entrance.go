package entrance

import (
	"context"

	"go-common/library/log"
	resApi "go-gateway/app/app-svr/app-resource/interface/api/v1"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	entranceDao "go-gateway/app/app-svr/app-resource/interface/dao/entrance"
	model "go-gateway/app/app-svr/app-resource/interface/model/entrance"
)

type Service struct {
	c   *conf.Config
	dao *entranceDao.Dao
}

func New(c *conf.Config) *Service {
	return &Service{
		c:   c,
		dao: entranceDao.New(c),
	}
}

// BusinessInfoc set entrance info from business in redis
func (s *Service) BusinessInfoc(c context.Context, req *model.BusinessInfocReq) error {
	if err := s.dao.AddBusinessInfocCache(c, req); err != nil {
		// 不下发错误，由日志处理
		log.Errorc(c, "s.dao.AddBusinessInfocCache req(%+v), error(%+v)", req, err)
	}
	return nil
}

// CheckEntranceInfoc check if entrance infoc key in redis exists
func (s *Service) CheckEntranceInfoc(c context.Context, req *resApi.CheckEntranceInfocRequest) (*resApi.CheckEntranceInfocReply, error) {
	isExisted, err := s.dao.BusinessInfocKeyExists(c, req)
	if err != nil {
		log.Errorc(c, "resource s.dao.BusinessInfocKeyExists error(%+v)", err)
		return nil, err
	}
	reply := &resApi.CheckEntranceInfocReply{
		IsExisted: isExisted,
	}
	return reply, nil
}
