package v1

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

func (d *dao) Interrelations(ctx context.Context, mid int64, owners []int64) (res map[int64]*relationgrpc.InterrelationReply, err error) {
	fidsMap := make(map[int64]int64)
	fids := []int64{}
	for _, fid := range owners {
		if _, ok := fidsMap[fid]; ok {
			continue
		}
		fidsMap[fid] = fid
		fids = append(fids, fid)
	}
	const _max = 20
	g := errgroup.WithContext(ctx)
	mu := sync.Mutex{}
	res = make(map[int64]*relationgrpc.InterrelationReply)
	for i := 0; i < len(fids); i += _max {
		var partFids []int64
		if i+_max > len(fids) {
			partFids = fids[i:]
		} else {
			partFids = fids[i : i+_max]
		}
		g.Go(func(ctx context.Context) (err error) {
			var (
				reply *relationgrpc.InterrelationMapReply
				arg   = &relationgrpc.RelationsReq{
					Mid: mid,
					Fid: partFids,
				}
			)
			if reply, err = d.relationClient.Interrelations(ctx, arg); err != nil {
				log.Error("d.relGRPC.Interrelations(%v) error(%v)", arg, err)
				return err
			}
			if reply == nil {
				return nil
			}
			mu.Lock()
			for k, v := range reply.InterrelationMap {
				res[k] = v
			}
			mu.Unlock()
			return nil
		})
	}
	err = g.Wait()
	return
}
