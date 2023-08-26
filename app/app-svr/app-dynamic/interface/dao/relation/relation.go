package relation

import (
	"context"
	"sync"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"

	"github.com/pkg/errors"
)

// StatsGRPC fids stats
func (d *Dao) StatsSlice(ctx context.Context, mids []int64) (res map[int64]*relationgrpc.StatReply, err error) {
	var (
		arg        = &relationgrpc.MidsReq{Mids: mids}
		statsReply *relationgrpc.StatsReply
	)
	if statsReply, err = d.relGRPC.Stats(ctx, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = statsReply.StatReplyMap
	return
}

func (d *Dao) Stats(c context.Context, uids []int64) (map[int64]*relationgrpc.StatReply, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*relationgrpc.StatReply)
	for i := 0; i < len(uids); i += max50 {
		var partUids []int64
		if i+max50 > len(uids) {
			partUids = uids[i:]
		} else {
			partUids = uids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			ss, err := d.StatsSlice(ctx, partUids)
			if err != nil {
				return err
			}
			mu.Lock()
			for uid, s := range ss {
				res[uid] = s
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("Stats uids(%+v) eg.wait(%+v)", uids, err)
		return nil, err
	}
	return res, nil
}
