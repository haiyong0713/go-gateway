package archive

import (
	"context"
	"errors"
	"sync"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	arcmid "go-gateway/app/app-svr/archive/middleware"
	"go-gateway/app/app-svr/archive/service/api"
)

const (
	_max   = 100
	_max50 = 50
)

// Archives is
func (d *Dao) Archives(c context.Context, aids []int64, mobiApp, device string, mid int64) (map[int64]*api.Arc, error) {
	if len(aids) == 0 {
		return nil, errors.New("empty aids")
	}
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	arcs := make(map[int64]*api.Arc)
	for i := 0; i < len(aids); i += _max {
		var partAids []int64
		if i+_max > len(aids) {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_max]
		}
		g.Go(func(ctx context.Context) (err error) {
			var res *api.ArcsReply
			arg := &api.ArcsRequest{Aids: partAids, Mid: mid, MobiApp: mobiApp, Device: device}
			if res, err = d.rpcClient.Arcs(ctx, arg); err != nil {
				return err
			}
			mu.Lock()
			for aid, arc := range res.Arcs {
				arcs[aid] = arc
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return arcs, nil
}

func dupBatchArg(in *api.BatchPlayArg) *api.BatchPlayArg {
	out := *in
	return &out
}

// ArcsPlayer get archive player.
func (d *Dao) ArcsPlayer(c context.Context, playAvs []*api.PlayAv, autoplayAreaValidate bool) (map[int64]*api.ArcPlayer, error) {
	g := errgroup.WithContext(c)
	batchArg, _ := arcmid.FromContext(c)
	mu := sync.Mutex{}
	arcs := make(map[int64]*api.ArcPlayer)
	if batchArg != nil {
		batchArg = dupBatchArg(batchArg)
		batchArg.AutoplayAreaValidate = autoplayAreaValidate
	}
	for i := 0; i < len(playAvs); i += _max50 {
		var partPlayAvs []*api.PlayAv
		if i+_max50 > len(playAvs) {
			partPlayAvs = playAvs[i:]
		} else {
			partPlayAvs = playAvs[i : i+_max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			arg := &api.ArcsPlayerRequest{
				PlayAvs:      partPlayAvs,
				BatchPlayArg: batchArg,
			}
			res, err := d.rpcClient.ArcsPlayer(ctx, arg)
			if err != nil {
				return err
			}
			mu.Lock()
			for aid, arc := range res.GetArcsPlayer() {
				arcs[aid] = arc
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("ArcsPlayer aids(%+v) eg.wait(%+v)", playAvs, err)
		return nil, err
	}
	return arcs, nil
}
