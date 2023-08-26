package relation

import (
	"context"
	"fmt"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	relV2 "git.bilibili.co/bapis/bapis-go/account/service/relation/v2"
	"github.com/pkg/errors"
)

type Dao struct {
	// grpc
	relGRPC   relationgrpc.RelationClient
	relV2GRPC relV2.RelationClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.relGRPC, err = relationgrpc.NewClient(c.RelationGRPC); err != nil {
		panic(fmt.Sprintf("relationgrpc NewClientt error (%+v)", err))
	}
	if d.relV2GRPC, err = relV2.NewClientRelation(c.RelationGRPC); err != nil {
		panic(fmt.Sprintf("relV2 NewClientt error (%+v)", err))
	}
	return
}

// Stat get mid relation stat
func (d *Dao) Stat(c context.Context, mid int64) (stat *relationgrpc.StatReply, err error) {
	stat, err = d.relGRPC.Stat(c, &relationgrpc.MidReq{Mid: mid})
	if err != nil {
		err = errors.Wrapf(err, "%v", mid)
		return
	}
	return
}

func (d *Dao) FollowersUnread(c context.Context, vmid int64) (res bool, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &relationgrpc.MidReq{Mid: vmid, RealIp: ip}
	var rly *relationgrpc.FollowersUnreadReply
	if rly, err = d.relGRPC.FollowersUnread(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if rly != nil {
		res = rly.HasUnread
	}
	return
}

func (d *Dao) Followings(c context.Context, vmid int64) (res []*relationgrpc.FollowingReply, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &relationgrpc.MidReq{Mid: vmid, RealIp: ip}
	var rly *relationgrpc.FollowingsReply
	if rly, err = d.relGRPC.Followings(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if rly != nil {
		res = rly.FollowingList
	}
	return
}

func (d *Dao) Relations(c context.Context, mid int64, fids []int64) (res map[int64]*relationgrpc.FollowingReply, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &relationgrpc.RelationsReq{Mid: mid, Fid: fids, RealIp: ip}
	var rly *relationgrpc.FollowingMapReply
	if rly, err = d.relGRPC.Relations(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if rly != nil {
		res = rly.FollowingMap
	}
	return
}

func (d *Dao) Tag(c context.Context, mid, tid int64) (res []int64, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &relationgrpc.TagIdReq{Mid: mid, TagId: tid, RealIp: ip}
	var rly *relationgrpc.TagReply
	if rly, err = d.relGRPC.Tag(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if rly != nil {
		res = rly.Mids
	}
	return
}

func (d *Dao) FollowersUnreadCount(c context.Context, mid int64) (res int64, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	arg := &relationgrpc.MidReq{Mid: mid, RealIp: ip}
	var rly *relationgrpc.FollowersUnreadCountReply
	if rly, err = d.relGRPC.FollowersUnreadCount(c, arg); err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	if rly != nil {
		res = rly.UnreadCount
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

// Relation .
func (d *Dao) Relation(ctx context.Context, mid, fid int64) (relReply *relationgrpc.FollowingReply, err error) {
	var (
		arg = &relationgrpc.RelationReq{Mid: mid, Fid: fid}
	)
	if relReply, err = d.relGRPC.Relation(ctx, arg); err != nil {
		log.Error("d.relGRPC.Relation(%v) error(%v)", arg, err)
	}
	return
}

// SpecialEffect .
func (d *Dao) SpecialEffect(c context.Context, mid, fid int64, buvid string) (*relationgrpc.SpecialEffectReply, error) {
	return d.relGRPC.SpecialEffect(c, &relationgrpc.SpecialEffectReq{Mid: mid, Fid: fid, Buvid: buvid})
}

// Interrelations
func (d *Dao) Interrelations(ctx context.Context, mid int64, owners []int64) (res map[int64]*relationgrpc.InterrelationReply, err error) {
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
			if reply, err = d.relGRPC.Interrelations(ctx, arg); err != nil {
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

func (d *Dao) PeakStats(c context.Context, mid int64) (int64, error) {
	reply, err := d.relGRPC.PeakStats(c, &relationgrpc.MidsReq{Mids: []int64{mid}})
	if err != nil {
		return 0, errors.Wrapf(err, "d.relGRPC.PeakStats mid(%d)", mid)
	}
	if reply == nil || len(reply.StatReplyMap) == 0 {
		return 0, nil
	}
	return reply.StatReplyMap[mid].GetFollower(), nil
}

func (d *Dao) FetchLastFollowingTime(c context.Context, mid int64) (int64, error) {
	reply, err := d.relV2GRPC.ListFollowing(c, &relV2.ListMidReq{
		Mid: mid,
		Page: relV2.PageRequest{
			Limit: 1,
		},
	})
	if err != nil {
		return 0, err
	}
	if len(reply.List) == 0 {
		return 0, ecode.NothingFound
	}
	return reply.List[0].GetTs(), nil
}
