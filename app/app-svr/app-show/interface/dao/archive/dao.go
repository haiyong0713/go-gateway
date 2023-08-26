package archive

import (
	"context"
	"fmt"
	"sync"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	dygrpc "go-gateway/app/web-svr/dynamic/service/api/v1"

	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

const _maxAids = 100

// Dao is archive dao.
type Dao struct {
	c *conf.Config
	// grpc
	rpcClient     arcgrpc.ArchiveClient
	dynamicClient dygrpc.DynamicClient
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
	if d.dynamicClient, err = dygrpc.NewClient(c.DynamicGRPC); err != nil {
		panic(err)
	}
	return
}

// Archive get archive by aid.
func (d *Dao) Archive(ctx context.Context, aid int64) (a *api.Arc, err error) {
	var (
		arcReply *arcgrpc.ArcReply
	)
	arg := &arcgrpc.ArcRequest{Aid: aid}
	if arcReply, err = d.rpcClient.Arc(ctx, arg); err != nil {
		log.Error(" d.rpcClient.Arc(%v) error(%v)", arg, err)
		return
	}
	a = arcReply.Arc
	return
}

// ArchivesPB multi get archives.
func (d *Dao) ArchivesPB(ctx context.Context, aids []int64, mid int64, mobiApp, device string) (as map[int64]*api.Arc, err error) {
	if as, err = d.circleReqArcs(ctx, aids, mid, mobiApp, device); err != nil {
		err = errors.Wrapf(err, "ArchivesPB(%v)", aids)
	}
	return
}

// RanksArcs
func (d *Dao) RanksArcs(ctx context.Context, rid, pn, ps int) (res []*api.Arc, aids []int64, err error) {
	arg := &dygrpc.RegAllReq{
		Rid: int64(rid),
		Pn:  int64(pn),
		Ps:  int64(ps),
	}
	var as *dygrpc.RegAllReply
	if as, err = d.dynamicClient.RegAllArcs(ctx, arg); err != nil {
		log.Error("d.arcRpc.RankArcs3(%v) error(%v)", arg, err)
		return
	}
	if as != nil {
		res = as.Archives
		for _, a := range res {
			if a == nil {
				err = errors.New("RanksArcs a is nil")
				return
			}
			aids = append(aids, a.Aid)
		}
	}
	return
}

// RankTopArcs
func (d *Dao) RankTopArcs(ctx context.Context, rid, pn, ps int) (res []*api.Arc, err error) {
	var (
		reply *dygrpc.RecentThrdRegArcReply
		arg   = &dygrpc.RecentThrdRegArcReq{
			Rid: int32(rid),
			Pn:  int64(pn),
			Ps:  int64(ps),
		}
	)
	if reply, err = d.dynamicClient.RecentThrdRegArc(ctx, arg); err != nil {
		log.Error("d.dynamicClient.RecentThrdRegArc(%v) error(%v)", arg, err)
		return
	}
	if reply != nil {
		res = reply.Archives
	}
	return
}

func (d *Dao) circleReqArcs(ctx context.Context, aids []int64, mid int64, mobiApp, device string) (aidMap map[int64]*api.Arc, err error) {
	var (
		aidsLen = len(aids)
		mutex   = sync.Mutex{}
	)
	aidMap = make(map[int64]*api.Arc, aidsLen)
	gp := errgroup.WithContext(ctx)
	for i := 0; i < aidsLen; i += _maxAids {
		var partAids []int64
		if i+_maxAids > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_maxAids]
		}
		gp.Go(func(ctx context.Context) (err error) {
			var tmpRes *api.ArcsReply
			arg := &api.ArcsRequest{Aids: partAids, Mid: mid, MobiApp: mobiApp, Device: device}
			if tmpRes, err = d.rpcClient.Arcs(ctx, arg); err != nil {
				return
			}
			if tmpRes == nil {
				err = errors.New("circleReqArcs is nil")
				return
			}
			if len(tmpRes.Arcs) > 0 {
				mutex.Lock()
				for aid, arc := range tmpRes.Arcs {
					if arc == nil {
						err = errors.New("circleReqArcs is nil")
						return
					}
					aidMap[aid] = arc
				}
				mutex.Unlock()
			}
			return err
		})
	}
	err = gp.Wait()
	return
}
