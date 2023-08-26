package resource

import (
	"context"

	"go-common/component/metadata/device"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/model"

	resSvc "git.bilibili.co/bapis/bapis-go/resource/service/v1"
)

type Dao struct {
	c           *conf.Config
	resourceSvr resSvc.ResourceClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	d.resourceSvr, err = resSvc.NewClient(c.ResourceGRPC)
	if err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) EntrancesIsHidden(ctx context.Context, oids []int64, otype int64) (bool, error) {
	dev, _ := device.FromContext(ctx)
	resp, err := d.resourceSvr.EntrancesIsHidden(ctx, &resSvc.EntrancesIsHiddenRequest{
		OidItems: map[int64]*resSvc.OidList{
			otype: {Oids: oids},
		},
		Build: dev.Build, Plat: int32(model.Plat(dev.RawMobiApp, dev.Device)), Channel: dev.Channel,
	})
	if err != nil {
		return false, err
	}
	return resp.GetHideDynamic(), nil
}
