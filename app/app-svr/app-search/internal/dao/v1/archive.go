package v1

import (
	"context"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcapi "go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

const (
	_max   = 100
	_max50 = 50
)

func (d *dao) ArcsPlayer(c context.Context, playAvs []*arcapi.PlayAv, autoplayAreaValidate bool) (map[int64]*arcapi.ArcPlayer, error) {
	g := errgroup.WithContext(c)
	batchArg, _ := arcmid.FromContext(c)
	mu := sync.Mutex{}
	arcs := make(map[int64]*arcapi.ArcPlayer)
	if batchArg != nil {
		batchArg = dupBatchArg(batchArg)
		batchArg.AutoplayAreaValidate = autoplayAreaValidate
	}
	for i := 0; i < len(playAvs); i += _max50 {
		var partPlayAvs []*arcapi.PlayAv
		if i+_max50 > len(playAvs) {
			partPlayAvs = playAvs[i:]
		} else {
			partPlayAvs = playAvs[i : i+_max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			arg := &arcapi.ArcsPlayerRequest{
				PlayAvs:      partPlayAvs,
				BatchPlayArg: batchArg,
			}
			res, err := d.archiveClient.ArcsPlayer(ctx, arg)
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

func dupBatchArg(in *arcapi.BatchPlayArg) *arcapi.BatchPlayArg {
	out := *in
	return &out
}

func (d *dao) Arcs(c context.Context, aids []int64, mobiApp, device string, mid int64) (map[int64]*arcapi.ArcPlayer, error) {
	if len(aids) == 0 {
		return nil, ecode.RequestErr
	}
	arg := &arcapi.ArcsRequest{
		Aids:    aids,
		Mid:     mid,
		MobiApp: mobiApp,
		Device:  device,
	}
	info, err := d.archiveClient.Arcs(c, arg)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", arg)
	}
	res := make(map[int64]*arcapi.ArcPlayer)
	for _, aid := range aids {
		if a, ok := info.Arcs[aid]; ok {
			res[aid] = &arcapi.ArcPlayer{Arc: a}
		}
	}
	return res, nil
}

func (d *dao) Archives(c context.Context, aids []int64, mobiApp, device string, mid int64) (map[int64]*arcapi.Arc, error) {
	if len(aids) == 0 {
		return nil, errors.New("empty aids")
	}
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	arcs := make(map[int64]*arcapi.Arc)
	for i := 0; i < len(aids); i += _max {
		var partAids []int64
		if i+_max > len(aids) {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_max]
		}
		g.Go(func(ctx context.Context) (err error) {
			var res *arcapi.ArcsReply
			arg := &arcapi.ArcsRequest{Aids: partAids, Mid: mid, MobiApp: mobiApp, Device: device}
			if res, err = d.archiveClient.Arcs(ctx, arg); err != nil {
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
