package relation

import (
	"context"
	"fmt"
	"math"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"github.com/pkg/errors"
)

// Dao is rpc dao.
type Dao struct {
	// grpc
	relGRPC relationgrpc.RelationClient
}

// New new a relation dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.relGRPC, err = relationgrpc.NewClient(c.RelationGRPC); err != nil {
		panic(fmt.Sprintf("relationgrpc NewClientt error (%+v)", err))
	}
	return
}

// RelationsGRPC fids relations
func (d *Dao) RelationsGRPC(ctx context.Context, mid int64, fids []int64) (res map[int64]*relationgrpc.FollowingReply, err error) {
	var (
		arg = &relationgrpc.RelationsReq{
			Mid: mid,
			Fid: fids,
		}
		followingMapReply *relationgrpc.FollowingMapReply
	)
	if followingMapReply, err = d.relGRPC.Relations(ctx, arg); err != nil {
		log.Error("d.relGRPC.Relations(%v) error(%v)", arg, err)
		res = nil
		return
	}
	res = followingMapReply.FollowingMap
	return
}

func (d *Dao) RelationsInterrelations(ctx context.Context, mid int64, fids []int64) (res map[int64]*relationgrpc.InterrelationReply, err error) {
	const (
		_max = 20
	)
	var (
		forNum     = int(math.Ceil(float64(len(fids)) / float64(_max)))
		mutex      = sync.Mutex{}
		start, end int
	)
	res = map[int64]*relationgrpc.InterrelationReply{}
	g := errgroup.WithContext(ctx)
	for i := 0; i < forNum; i++ {
		start = i * _max
		end = start + _max
		var (
			tmpfids []int64
		)
		if len(fids) >= end {
			tmpfids = fids[start:end]
		} else if len(fids) < end {
			tmpfids = fids[start:]
		} else if len(fids) < start {
			break
		}
		g.Go(func(cc context.Context) (err error) {
			var (
				reply *relationgrpc.InterrelationMapReply
				arg   = &relationgrpc.RelationsReq{
					Mid: mid,
					Fid: tmpfids,
				}
			)
			if reply, err = d.relGRPC.Interrelations(cc, arg); err != nil {
				log.Error("d.relGRPC.Interrelations(%v) error(%v)", arg, err)
				return
			}
			if reply != nil {
				mutex.Lock()
				for k, v := range reply.InterrelationMap {
					res[k] = v
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

// AddFollowing .
func (d *Dao) AddFollowing(ctx context.Context, fid, mid int64, spmid string) (err error) {
	if _, err = d.relGRPC.AddFollowing(ctx, &relationgrpc.FollowingReq{Mid: mid, Fid: fid, Spmid: spmid}); err != nil {
		log.Error("d.relGRPC.AddFollowing(%d,%d) error(%v)", mid, fid, err)
	}
	return
}

// AddFollowing .
func (d *Dao) DelFollowing(ctx context.Context, fid, mid int64, spmid string) (err error) {
	if _, err = d.relGRPC.DelFollowing(ctx, &relationgrpc.FollowingReq{Mid: mid, Fid: fid, Spmid: spmid}); err != nil {
		log.Error("d.relGRPC.DelFollowing(%d,%d) error(%v)", mid, fid, err)
	}
	return
}

// StatsGRPC fids stats
func (d *Dao) StatsGRPC(ctx context.Context, mids []int64) (res map[int64]*relationgrpc.StatReply, err error) {
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

// Attentions .
func (d *Dao) Attentions(c context.Context, mid int64) (list []*relationgrpc.FollowingReply, err error) {
	var (
		rely *relationgrpc.FollowingsReply
	)
	if rely, err = d.relGRPC.Attentions(c, &relationgrpc.MidReq{Mid: mid}); err != nil {
		err = errors.Wrapf(err, "%d", mid)
		return
	}
	if rely != nil {
		list = rely.FollowingList
	}
	return
}
