package vas

import (
	"context"
	"fmt"
	"github.com/thoas/go-funk"

	"go-common/library/log"

	vasGrpc "git.bilibili.co/bapis/bapis-go/vas/trans/service"

	"go-gateway/app/app-svr/archive/service/conf"
)

type Dao struct {
	c       *conf.Config
	vasGRPC vasGrpc.VasTransServiceClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.vasGRPC, err = vasGrpc.NewClientVasTransService(c.VasGRPC); err != nil {
		panic(fmt.Sprintf("vasGrpc.NewClientVasTransService error (%+v)", err))
	}
	return d
}

func (d *Dao) SeasonUserVoucherBatch(c context.Context, mid int64, sids []int64) (*vasGrpc.SeasonUserVoucherBatchReply, error) {
	req := &vasGrpc.SeasonUserVoucherBatchReq{
		Mid:       mid,
		SeasonIds: funk.UniqInt64(sids),
	}
	reply, err := d.vasGRPC.SeasonUserVoucherBatch(c, req)
	if err != nil {
		log.Error("d.vasGRPC.SeasonUserVoucherBatch err %v", err)
		return nil, err
	}
	return reply, nil
}
