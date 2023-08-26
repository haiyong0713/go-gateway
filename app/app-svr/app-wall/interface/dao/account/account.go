package account

import (
	"context"
	"strconv"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-wall/interface/conf"

	vip "git.bilibili.co/bapis/bapis-go/vip/resource/service"
)

// Dao account dao
type Dao struct {
	accGRPC account.AccountClient
	vipGRPC vip.ResourceClient
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.accGRPC, err = account.NewClient(c.AccountGRPC); err != nil {
		panic(err)
	}
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

// Info user info
func (d *Dao) Info(c context.Context, mid int64) (*account.Info, error) {
	arg := &account.MidReq{Mid: mid}
	reply, err := d.accGRPC.Info3(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	info := reply.GetInfo()
	if info == nil {
		return nil, ecode.NothingFound
	}
	return info, nil
}
