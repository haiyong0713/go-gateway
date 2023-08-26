package archive

import (
	"context"
	"fmt"
	"sync"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-channel/interface/conf"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/sync/errgroup.v2"

	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	"github.com/pkg/errors"
)

// Dao is archive dao.
type Dao struct {
	//grpc
	rpcClient arcgrpc.ArchiveClient
	// content.flow.control.service gRPC
	cfcGRPC cfcgrpc.FlowControlClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
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

// UpCount get upper count.
func (d *Dao) UpCount(c context.Context, mid int64) (cnt int, err error) {
	var (
		arg   = &arcgrpc.UpCountRequest{Mid: mid}
		reply *arcgrpc.UpCountReply
	)
	if reply, err = d.rpcClient.UpCount(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if reply != nil {
		cnt = int(reply.Count)
	}
	return
}

// ArcsPlayer get archive player.
func (d *Dao) ArcsPlayer(c context.Context, aids []*arcgrpc.PlayAv, showPgcPlayurl bool) (map[int64]*arcgrpc.ArcPlayer, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*arcgrpc.ArcPlayer)
	batchArg, _ := arcmid.FromContext(c)
	if batchArg != nil {
		batchArg.ShowPgcPlayurl = showPgcPlayurl
	}
	for i := 0; i < len(aids); i += max50 {
		var partAids []*arcgrpc.PlayAv
		if i+max50 > len(aids) {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			arg := &arcgrpc.ArcsPlayerRequest{
				PlayAvs:      partAids,
				BatchPlayArg: batchArg,
			}
			archives, err := d.rpcClient.ArcsPlayer(ctx, arg)
			if err != nil {
				log.Error("ArcsPlayer partAids(%+v) err(%v)", partAids, err)
				return err
			}
			mu.Lock()
			for aid, arc := range archives.GetArcsPlayer() {
				if arc == nil {
					continue
				}
				if !arc.Arc.IsNormal() {
					continue
				}
				res[aid] = arc
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

// Arcs get archive player.
func (d *Dao) Arcs(c context.Context, aids []int64) (map[int64]*arcgrpc.Arc, error) {
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
			arg := &arcgrpc.ArcsRequest{
				Aids: partAids,
			}
			archives, err := d.rpcClient.Arcs(ctx, arg)
			if err != nil {
				log.Error("Arcs partAids(%+v) err(%v)", partAids, err)
				return err
			}
			mu.Lock()
			for aid, arc := range archives.GetArcs() {
				if arc == nil {
					continue
				}
				if !arc.IsNormal() {
					continue
				}
				res[aid] = arc
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
