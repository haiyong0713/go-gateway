package pgc

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	pgcDynGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/dynamic"
)

func (d *Dao) FollowCard(c context.Context, mobiApp, device, platform string, mid int64, build int, aids []int64) (map[int64]*pgcDynGrpc.FollowCardProto, error) {
	var max = 30
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*pgcDynGrpc.FollowCardProto)
	for i := 0; i < len(aids); i += max {
		var partAids []int64
		if i+max > len(aids) {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+max]
		}
		g.Go(func(ctx context.Context) (err error) {
			req := &pgcDynGrpc.FollowCardReq{
				Aid: partAids,
				User: &pgcDynGrpc.UserProto{
					Mid:      mid,
					MobiApp:  mobiApp,
					Device:   device,
					Platform: platform,
					Build:    int32(build),
				},
			}
			pgc, err := d.pgcDynGRPC.FollowCard(ctx, req)
			if err != nil {
				log.Error("FollowCard partAids(%+v) err(%v)", partAids, err)
				return err
			}
			for k, v := range pgc.Card {
				if v == nil {
					continue
				}
				mu.Lock()
				res[k] = v
				mu.Unlock()
			}
			return
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
