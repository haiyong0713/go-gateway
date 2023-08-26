package fission

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	fissiDao "go-gateway/app/app-svr/app-resource/interface/dao/fission"
	fissiMdl "go-gateway/app/app-svr/app-resource/interface/model/fission"

	fissiGrpc "git.bilibili.co/bapis/bapis-go/account/service/fission"
)

type Service struct {
	dao *fissiDao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao: fissiDao.New(c),
	}
	return
}

// CheckNew fission check new.
func (s *Service) CheckNew(c context.Context, param *fissiMdl.ParamCheck) (rs *fissiGrpc.CheckNewResp, err error) {
	if rs, err = s.dao.CheckNew(c, param); err != nil {
		log.Error("s.dao.CheckNew param(%+v) error(%v)", param, err)
	}
	return
}

// CheckDevice fission check device.
func (s *Service) CheckDevice(c context.Context, param *fissiMdl.ParamCheck) (rs *fissiGrpc.CheckNewResp, err error) {
	if rs, err = s.dao.CheckDevice(c, param); err != nil {
		log.Error("s.dao.CheckDevice param(%+v) error(%v)", param, err)
	}
	return
}
