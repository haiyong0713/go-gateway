package account

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/app-wall/job/conf"

	vip "git.bilibili.co/bapis/bapis-go/vip/resource/service"
)

// Dao account dao
type Dao struct {
	vipGRPC vip.ResourceClient
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.vipGRPC, err = vip.NewClient(c.VIPGRPC); err != nil {
		panic(err)
	}
	return
}

// AddVIP add user vip
func (d *Dao) AddVIP(c context.Context, mid, batchID, orderNo int64, remark, appKey string) error {
	arg := &vip.ResourceUseReq{
		Mid:     mid,
		BatchId: batchID,
		OrderNo: strconv.FormatInt(orderNo, 10),
		Remark:  remark,
		Appkey:  appKey,
	}
	_, err := d.vipGRPC.ResourceUse(c, arg)
	return err
}
