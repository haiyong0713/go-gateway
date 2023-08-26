package archive

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/conf"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

// Dao is archive dao.
type Dao struct {
	c *conf.Config
	//grpc
	rpcClient arcgrpc.ArchiveClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	//grpc
	var err error
	if d.rpcClient, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	return
}

func (d *Dao) Arc(c context.Context, aid int64) (arc *arcgrpc.Arc, err error) {
	arg := &arcgrpc.ArcRequest{
		Aid: aid,
	}
	info, err := d.rpcClient.Arc(c, arg)
	if err != nil {
		log.Error("%v", err)
		return
	}
	arc = info.GetArc()
	return
}

func (d *Dao) Arcs(c context.Context, aids []int64) (res map[int64]*arcgrpc.Arc, err error) {
	var (
		args   = &arcgrpc.ArcsRequest{Aids: aids}
		resTmp *arcgrpc.ArcsReply
	)
	if resTmp, err = d.rpcClient.Arcs(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = resTmp.GetArcs()
	return
}
