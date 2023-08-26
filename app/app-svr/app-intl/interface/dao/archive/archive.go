package archive

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

func (d *Dao) ArcsPlayer(c context.Context, aids []*arcgrpc.PlayAv) (res map[int64]*arcgrpc.ArcPlayer, err error) {
	batchArg, _ := arcmid.FromContext(c)
	arg := &arcgrpc.ArcsPlayerRequest{
		PlayAvs:      aids,
		BatchPlayArg: batchArg,
	}
	info, err := d.rpcClient.ArcsPlayer(c, arg)
	if err != nil {
		return nil, err
	}
	return info.GetArcsPlayer(), nil
}

// Arc multi get archives.
func (d *Dao) Arc(c context.Context, aid int64) (res *api.Arc, err error) {
	info, err := d.rpcClient.Arc(c, &arcgrpc.ArcRequest{Aid: aid})
	if err != nil {
		return nil, err
	}
	return info.Arc, nil
}

// Archives multi get archives.
func (d *Dao) Arcs(c context.Context, aids []int64) (res map[int64]*api.ArcPlayer, err error) {
	if len(aids) == 0 {
		return
	}
	var am = map[int64]*api.Arc{}
	mutexArcs := sync.Mutex{}
	g := errgroup.WithContext(c)
	if arcsLen := len(aids); arcsLen > 0 {
		for i := 0; i < arcsLen; i += _maxAids {
			var partAids []int64
			if i+_maxAids > arcsLen {
				partAids = aids[i:]
			} else {
				partAids = aids[i : i+_maxAids]
			}
			g.Go(func(ctx context.Context) (err error) {
				var (
					tmpRes *api.ArcsReply
					arg    = &api.ArcsRequest{Aids: partAids}
				)
				if tmpRes, err = d.rpcClient.Arcs(ctx, arg); err != nil {
					log.Error("Archives aids(%v) d.rpcClient.Arcs error(%v)", aids, err)
					return
				}
				if len(tmpRes.GetArcs()) > 0 {
					mutexArcs.Lock()
					for aid, stat := range tmpRes.GetArcs() {
						am[aid] = stat
					}
					mutexArcs.Unlock()
				}
				return
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, aid := range aids {
		if a, ok := am[aid]; ok {
			if res == nil {
				res = map[int64]*arcgrpc.ArcPlayer{
					aid: {
						Arc: a,
					},
				}
			} else {
				res[aid] = &arcgrpc.ArcPlayer{
					Arc: a,
				}
			}
		}
	}
	return res, err
}
