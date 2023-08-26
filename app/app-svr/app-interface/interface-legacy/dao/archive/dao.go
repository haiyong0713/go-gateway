package archive

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	"github.com/pkg/errors"
)

// Dao is archive dao.
type Dao struct {
	// http client
	client *bm.Client

	//grpc
	rpcClient arcgrpc.ArchiveClient
	// content.flow.control.service gRPC
	cfcGRPC cfcgrpc.FlowControlClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		client: bm.NewClient(c.HTTPWrite),
	}
	//grpc
	var err error
	if d.rpcClient, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.cfcGRPC, err = cfcgrpc.NewClient(c.CfcGRPC); err != nil {
		panic(fmt.Sprintf("CfcGRPC NewClient error (%+v)", err))
	}
	return
}

// UpArcs3 get upper archives
func (d *Dao) UpArcs3(c context.Context, mid int64, pn, ps int) (as []*api.Arc, err error) {
	var (
		upReply *api.UpArcsReply
	)
	arg := &api.UpArcsRequest{Mid: mid, Pn: int32(pn), Ps: int32(ps)}
	if upReply, err = d.rpcClient.UpArcs(c, arg); err != nil {
		if ecode.Cause(err) == ecode.NothingFound {
			err = nil
		}
		return
	}
	as = upReply.Arcs
	return
}

// UpCount2 get upper count.
func (d *Dao) UpCount2(c context.Context, mid int64) (cnt int, err error) {
	var (
		countReply *api.UpCountReply
	)
	arg := &api.UpCountRequest{Mid: mid}
	if countReply, err = d.rpcClient.UpCount(c, arg); err == nil {
		cnt = int(countReply.Count)
	}
	return
}

// Arcs get archive aids.
func (d *Dao) Arcs(c context.Context, aids []int64, mobiApp, device string, mid int64) (map[int64]*arcgrpc.ArcPlayer, error) {
	if len(aids) == 0 {
		return nil, ecode.RequestErr
	}
	arg := &arcgrpc.ArcsRequest{
		Aids:    aids,
		Mid:     mid,
		MobiApp: mobiApp,
		Device:  device,
	}
	info, err := d.rpcClient.Arcs(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}
	res := make(map[int64]*arcgrpc.ArcPlayer)
	for _, aid := range aids {
		if a, ok := info.Arcs[aid]; ok {
			res[aid] = &arcgrpc.ArcPlayer{Arc: a}
		}
	}
	return res, nil
}
