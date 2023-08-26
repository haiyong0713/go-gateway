package ogv

import (
	"context"

	ogvgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/pay"

	"github.com/pkg/errors"
)

func (d *Dao) OgvPay(c context.Context, mid, aid int64) (allowPlay bool, err error) {
	var reply *ogvgrpc.UserRightsReply
	if reply, err = d.ogvpayClient.UserRights(c, &ogvgrpc.UserRightsReq{
		Mid: mid,
		Aid: aid,
	}); err != nil {
		err = errors.Wrapf(err, "mid %d aid %d", mid, aid)
		return
	}
	allowPlay = reply.AllowPlay
	return

}
