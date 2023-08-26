package archive

import (
	"context"
	"sync"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/archive/service/api"
)

const (
	_maxAids = 100
)

// Archives multi get archives.
func (d *Dao) Archives(c context.Context, aids []int64, mid int64, mobiApp, device string) (am map[int64]*api.Arc, err error) {
	if len(aids) == 0 {
		return
	}
	var (
		mutexArcs = sync.Mutex{}
	)
	am = map[int64]*api.Arc{}
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
					arg    = &api.ArcsRequest{
						Aids:    partAids,
						Mid:     mid,
						MobiApp: mobiApp,
						Device:  device,
					}
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
	}
	return
}
