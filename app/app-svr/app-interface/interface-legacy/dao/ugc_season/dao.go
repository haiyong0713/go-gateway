package ugc_season

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	ugcSeasonGrpc "go-gateway/app/app-svr/ugc-season/service/api"

	"github.com/pkg/errors"
)

// Dao is archive dao.
type Dao struct {
	c *conf.Config
	//grpc
	rpcClient ugcSeasonGrpc.UGCSeasonClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	//grpc
	var err error
	if d.rpcClient, err = ugcSeasonGrpc.NewClient(c.UGCSeasonGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}

func (d *Dao) UpperList(c context.Context, vmid, pn, ps int64) (res *ugcSeasonGrpc.UpperListReply, err error) {
	if res, err = d.rpcClient.UpperList(c, &ugcSeasonGrpc.UpperListRequest{Mid: vmid, PageNum: pn, PageSize: ps}); err != nil {
		if ecode.EqualError(ecode.NothingFound, err) {
			err = nil
			return
		}
		log.Error("%v", err)
	}
	return
}

func (d *Dao) SeasonView(ctx context.Context, req *ugcSeasonGrpc.ViewRequest) (*ugcSeasonGrpc.ViewReply, error) {
	reply, err := d.rpcClient.View(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "req=%+v", req)
	}
	return reply, nil
}

func (d *Dao) Seasons(ctx context.Context, req *ugcSeasonGrpc.SeasonsRequest) (*ugcSeasonGrpc.SeasonsReply, error) {
	reply, err := d.rpcClient.Seasons(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "req=%+v", req)
	}
	return reply, nil
}
