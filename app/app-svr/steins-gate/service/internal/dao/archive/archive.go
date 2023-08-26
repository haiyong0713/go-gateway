package archive

import (
	"context"
	"sync"

	"go-common/library/sync/errgroup.v2"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

const (
	_arcServiceNb = 50
)

// Arc int
func (d *Dao) Arc(ctx context.Context, aid int64) (arc *arcgrpc.Arc, err error) {
	var res *arcgrpc.ArcReply
	if res, err = d.arcClient.Arc(ctx, &arcgrpc.ArcRequest{
		Aid: aid,
	}); err != nil {
		err = errors.Wrapf(err, "%v", aid)
		return
	}
	arc = res.Arc
	return
}

// ArcView int
func (d *Dao) ArcView(ctx context.Context, aid int64) (res *arcgrpc.SteinsGateViewReply, err error) {
	if res, err = d.arcClient.SteinsGateView(ctx, &arcgrpc.SteinsGateViewRequest{
		Aid: aid,
	}); err != nil {
		err = errors.Wrapf(err, "%v", aid)
		return
	}
	return
}

func splitIDs(ids []int64, ps int) (pces [][]int64) {
	if len(ids) == 0 {
		return
	}
	var nbPce int
	if len(ids)%ps == 0 {
		nbPce = len(ids) / ps
	} else {
		nbPce = len(ids)/ps + 1
	}
	for i := 0; i < nbPce; i++ {
		if end := (i + 1) * ps; end > len(ids) {
			pces = append(pces, ids[i*ps:])
		} else {
			pces = append(pces, ids[i*ps:(i+1)*ps])
		}
	}
	return
}

// Arcs multi get archives
func (d *Dao) Arcs(c context.Context, aids []int64) (res map[int64]*arcgrpc.Arc, err error) {
	var (
		mutex   = sync.Mutex{}
		aidPces = splitIDs(aids, _arcServiceNb)
		eg      = errgroup.WithContext(c)
	)
	res = make(map[int64]*arcgrpc.Arc, len(aids))
	for _, pce := range aidPces {
		tmp := pce
		eg.Go(func(c context.Context) (err error) {
			req := &arcgrpc.ArcsRequest{Aids: tmp}
			var reply *arcgrpc.ArcsReply
			if reply, err = d.arcClient.Arcs(c, req); err != nil {
				err = errors.Wrapf(err, "Arcs Pce %v", tmp)
				return
			}
			mutex.Lock()
			for _, a := range reply.Arcs {
				res[a.Aid] = a
			}
			mutex.Unlock()
			return
		})
	}
	err = eg.Wait()
	return
}

// ArcViews int
func (d *Dao) ArcViews(ctx context.Context, aids []int64) (res map[int64]*arcgrpc.SteinsGateViewReply, err error) {
	var (
		req = &arcgrpc.SteinsGateViewsRequest{
			Aids: aids,
		}
		reply *arcgrpc.SteinsGateViewsReply
	)
	if reply, err = d.arcClient.SteinsGateViews(ctx, req); err != nil {
		err = errors.Wrapf(err, "%v", req)
		return
	}
	if reply != nil {
		res = reply.Views
	}
	return

}
