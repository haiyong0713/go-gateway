package ipdisplay

import (
	"context"

	"go-common/library/log"

	ipDisplay "git.bilibili.co/bapis/bapis-go/manager/operation/ip-display"

	"go-gateway/app/app-svr/archive/service/conf"
)

type Dao struct {
	c               *conf.Config
	ipDisplayClient ipDisplay.OperationItemIpDisplayV1Client
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.ipDisplayClient, err = ipDisplay.NewClientOperationItemIpDisplayV1(c.IPDisplayClient); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) IpDisplay(c context.Context) (res *ipDisplay.IpDisplayRecordsResp, err error) {
	req := &ipDisplay.IpDisplayRecordsReq{Tp: ipDisplay.IpDisplayTp_IpDisplayTpBv}
	if res, err = d.ipDisplayClient.IpDisplayRecords(c, req); err != nil {
		log.Error("d.ipDisplayClient.IpDisplayRecords error req(%+v) err(%+v)", req, err)
		return
	}
	return
}
