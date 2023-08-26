package archive

import (
	"context"
	"fmt"
	"sync"

	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/job/conf"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	flowctrlgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
)

// Dao is archive dao.
type Dao struct {
	c *conf.Config
	// grpc
	rpcClient         arcgrpc.ArchiveClient
	flowControlClient flowctrlgrpc.FlowControlClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	// grpc
	var err error
	if d.rpcClient, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.flowControlClient, err = flowctrlgrpc.NewClient(c.FlowControlGRPC); err != nil {
		panic(fmt.Sprintf("flowControlClient NewClient error (%+v)", err))
	}
	return
}

// Archives get archive.
func (d *Dao) Archive(c context.Context, aid int64) (*api.Arc, error) {
	reply, err := d.rpcClient.Arc(c, &arcgrpc.ArcRequest{Aid: aid})
	if err != nil {
		return nil, err
	}
	return reply.GetArc(), nil
}

func (d *Dao) Archives(c context.Context, aids []int64) (map[int64]*api.Arc, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*arcgrpc.Arc)
	for i := 0; i < len(aids); i += max50 {
		var partAids []int64
		if i+max50 > len(aids) {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			arg := &arcgrpc.ArcsRequest{Aids: partAids}
			archives, err := d.rpcClient.Arcs(c, arg)
			if err != nil {
				return err
			}
			for k, v := range archives.GetArcs() {
				if v == nil {
					continue
				}
				if !v.IsNormal() {
					continue
				}
				mu.Lock()
				res[k] = v
				mu.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
