package vip

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/interface/conf"

	viprpc "git.bilibili.co/bapis/bapis-go/vip/service"
)

type Dao struct {
	// grpc
	rpcClient viprpc.VipClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.rpcClient, err = viprpc.NewClient(c.VipGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}

func (d *Dao) TipsRenew(c context.Context, build int, platform, mid int64) (res *viprpc.TipsRenewReply, err error) {
	arg := &viprpc.TipsRenewReq{
		Position: 7,
		Mid:      mid,
		Platform: platform,
		Build:    strconv.Itoa(build),
	}
	if res, err = d.rpcClient.TipsRenew(c, arg); err != nil {
		log.Error("tipsRenew mid(%d) error(%v)", mid, err)
		return
	}
	b, _ := json.Marshal(&res)
	log.Info("tipsRenew mid(%d) list(%s)", mid, b)
	return
}
