package mission

import (
	"context"
	api "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/client"
)

func (d *Dao) GetUserPhoneBind(ctx context.Context, mid int64) (isBind int32, err error) {
	resp, err := client.AccountClient.Profile3(ctx, &api.MidReq{
		Mid: mid,
	})
	if err != nil || resp == nil || resp.Profile == nil {
		log.Errorc(ctx, "[GetUserPhoneBind][Profile3][Error], err:%+v, resp:%+v", err, resp)
		return
	}
	isBind = resp.Profile.TelStatus
	return
}
