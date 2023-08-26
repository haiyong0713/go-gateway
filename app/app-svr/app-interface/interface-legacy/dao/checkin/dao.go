package checkin

import (
	"context"
	checkinclient "git.bilibili.co/bapis/bapis-go/platform/interface/checkin-plat"
	"github.com/pkg/errors"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
)

type Dao struct {
	// grpc
	checkinClient checkinclient.CheckinPlatInterfaceV1Client
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.checkinClient, err = checkinclient.NewClientCheckinPlatInterfaceV1(c.CheckinClient); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) CheckinCount(c context.Context, mid int64) (int32, error) {
	reply, err := d.checkinClient.FavorShowCheckinTab(c, &checkinclient.MidReq{Mid: mid})
	if err != nil {
		return 0, errors.Wrapf(err, "CardCount mid(%d)", mid)
	}
	return reply.Status, nil
}
