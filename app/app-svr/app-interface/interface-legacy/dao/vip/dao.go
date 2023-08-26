package vip

import (
	"context"
	"strconv"

	"go-common/component/tinker/env"
	"go-common/library/exp/ab"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	vipclient "git.bilibili.co/bapis/bapis-go/vip/service"
	vipinfogrpc "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"

	"github.com/pkg/errors"
)

// Dao is coin dao
type Dao struct {
	vipClient     vipclient.VipClient
	vipInfoClient vipinfogrpc.VipInfoClient
}

// New initial coin dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.vipClient, err = vipclient.NewClient(c.VipGRPC); err != nil {
		panic(err)
	}
	if d.vipInfoClient, err = vipinfogrpc.NewClient(c.VipInfoGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) VipSpaceLabel(ctx context.Context, req *vipinfogrpc.SpaceLabelReq) (*vipinfogrpc.SpaceLabelReply, error) {
	res, err := d.vipInfoClient.SpaceLabel(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "d.vipInfoClient.SpaceLabel req=%+v", req)
	}
	return res, nil
}

func (d *Dao) VipTips(c context.Context, mid int64, platfrom, mobiApp, device string, build int, buvid string) (res *vipclient.TipsVipReply, err error) {
	in := &vipclient.TipsVipReq{
		Position: []int64{3},
		Mid:      mid,
		Platform: platfrom,
		MobiApp:  mobiApp,
		Build:    strconv.Itoa(build),
		Device:   device,
		Buvid:    buvid,
	}
	//ab test
	t := ab.New(env.Extract(c)...)
	ctx := ab.NewContext(c, t)
	if res, err = d.vipClient.TipsVip(ctx, in); err != nil {
		log.Error("%v", err)
	}
	return
}
