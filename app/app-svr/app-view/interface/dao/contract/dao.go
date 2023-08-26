package contract

import (
	"context"
	"fmt"

	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-common/library/net/metadata"

	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/conf"

	"git.bilibili.co/bapis/bapis-go/community/service/contract"
)

type Dao struct {
	contractClient api.ContractClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.contractClient, err = api.NewClient(c.ContractClient)
	if err != nil {
		panic(fmt.Sprintf("appconf NewClient error(%+v)", err))
	}
	return
}

// AddContract .
func (d *Dao) AddContract(c context.Context, arg *viewApi.AddContractReq, mid int64, dev device.Device) error {
	req := &api.AddContractReq{
		Mid:   mid,
		UpMid: arg.UpMid,
		Aid:   arg.Aid,
		Common: &api.CommonReq{
			Platform: dev.RawPlatform,
			Build:    int32(dev.Build),
			Buvid:    dev.Buvid,
			MobiApp:  dev.RawMobiApp,
			Device:   dev.Device,
			Ip:       metadata.String(c, metadata.RemoteIP),
			Spmid:    arg.Spmid,
		},
	}
	_, err := d.contractClient.AddContract(c, req)
	if err != nil {
		log.Error("AddContract err(%+v) req(%+v)", err, req)
		return err
	}
	return nil
}
