package dao

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"github.com/pkg/errors"
)

func (d *dao) StatsGRPC(ctx context.Context, mids []int64) (res map[int64]*relationgrpc.StatReply, err error) {
	var (
		arg        = &relationgrpc.MidsReq{Mids: mids}
		statsReply *relationgrpc.StatsReply
	)
	if statsReply, err = d.relationClient.Stats(ctx, arg); err != nil {
		err = errors.Wrapf(err, "%+v", arg)
		return
	}
	res = statsReply.StatReplyMap
	return
}

// RelationsInterrelations
func (d *dao) RelationsInterrelations(ctx context.Context, mid int64, fids []int64) (res map[int64]*relationgrpc.InterrelationReply, err error) {
	const _max = 20
	var (
		mutex = sync.Mutex{}
	)
	res = make(map[int64]*relationgrpc.InterrelationReply)
	g := errgroup.WithContext(ctx)
	if fidLen := len(fids); fidLen > 0 {
		for i := 0; i < fidLen; i += _max {
			var partFids []int64
			if i+_max > fidLen {
				partFids = fids[i:]
			} else {
				partFids = fids[i : i+_max]
			}
			g.Go(func(ctx context.Context) (err error) {
				reply, err := d.relationsInterrelations(ctx, mid, partFids)
				if err != nil {
					return
				}
				if len(reply) > 0 {
					mutex.Lock()
					for k, v := range reply {
						res[k] = v
					}
					mutex.Unlock()
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

func (d *dao) relationsInterrelations(ctx context.Context, mid int64, fids []int64) (res map[int64]*relationgrpc.InterrelationReply, err error) {
	arg := &relationgrpc.RelationsReq{
		Mid: mid,
		Fid: fids,
	}
	reply, err := d.relationClient.Interrelations(ctx, arg)
	if err != nil {
		return
	}
	if reply == nil {
		return
	}
	res = reply.InterrelationMap
	return
}
