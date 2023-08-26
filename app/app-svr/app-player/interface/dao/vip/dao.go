package vip

import (
	"context"
	"fmt"
	"go-gateway/app/app-svr/app-player/interface/conf"
	"go-gateway/app/app-svr/app-player/interface/model"

	vipApi "git.bilibili.co/bapis/bapis-go/vip/service"
	"github.com/pkg/errors"
)

// Dao is vip dao.
type Dao struct {
	// rpc
	vipClient vipApi.VipClient
}

// New new a vip dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	d.vipClient, err = vipApi.NewClient(c.VipClient)
	if err != nil {
		panic(fmt.Sprintf("vip NewClient error(%v)", err))
	}
	return
}

// ReportOfflineDownloadNum is
func (d *Dao) ReportOfflineDownloadNum(c context.Context, mid int64, param *model.DlNumParam) (err error) {
	req := &vipApi.ReportOfflineDownloadNumReq{
		Mid:     mid,
		Num:     param.Num,
		Buvid:   param.Buvid,
		Build:   param.Build,
		MobiApp: param.MobiApp,
		Device:  param.Device,
	}
	if _, err = d.vipClient.ReportOfflineDownloadNum(c, req); err != nil {
		err = errors.Wrapf(err, "req(%+v)", req)
		return
	}
	return
}
