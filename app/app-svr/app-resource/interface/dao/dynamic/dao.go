package dynamic

import (
	"context"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-resource/interface/conf"

	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
	dynamicLocgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/localcity-svr"
)

type Dao struct {
	c               *conf.Config
	dynamicLocGRPC  dynamicLocgrpc.LocalCitySvrClient
	dyncampusClient dyncampusgrpc.CampusSvrClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.dynamicLocGRPC, err = dynamicLocgrpc.NewClient(d.c.DynamicLocGRPC); err != nil {
		panic(err)
	}
	if d.dyncampusClient, err = dyncampusgrpc.NewClient(c.DynamicCampusGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) PrivacyConfig(c context.Context, mid int64) (int64, error) {
	var (
		resTmp *dynamicLocgrpc.GetUserPrivacyResp
		err    error
	)
	if resTmp, err = d.dynamicLocGRPC.GetUserPrivacy(c, &dynamicLocgrpc.GetUserPrivacyReq{Uid: mid}); err != nil {
		log.Error("%v", err)
		return 0, err
	}
	return resTmp.GetDynStatus(), nil
}

func (d *Dao) SetPrivacyConfig(c context.Context, mid int64, state int64) (err error) {
	if _, err = d.dynamicLocGRPC.UpdateUserPrivacy(c, &dynamicLocgrpc.UpdateUserPrivacyReq{Uid: mid, DynStatus: state}); err != nil {
		log.Error("%+v", err)
	}
	return err
}

func (d *Dao) FetchUserPrivacy(c context.Context, mid int64) (int64, error) {
	arg := &dyncampusgrpc.FetchUserPrivacyReq{Mid: mid}
	reply, err := d.dyncampusClient.FetchUserPrivacy(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return int64(reply.Status), nil
}

func (d *Dao) UpdateUserPrivacy(c context.Context, mid int64, state int64) error {
	arg := &dyncampusgrpc.UpdateUserPrivacyReq{Mid: mid, Status: dyncommongrpc.UserPrivacyStatus(state)}
	_, err := d.dyncampusClient.UpdateUserPrivacy(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}
