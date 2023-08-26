package relation

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/interface/conf"

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

// RelationsInterrelations
func (d *Dao) RelationsInterrelations(ctx context.Context, mid int64, fids []int64) (res map[int64]*relationgrpc.InterrelationReply, err error) {
	arg := &relationgrpc.RelationsReq{
		Mid: mid,
		Fid: fids,
	}
	reply, err := d.relGRPC.Interrelations(ctx, arg)
	if err != nil {
		log.Error("d.relGRPC.Interrelations(%v) error(%v)", arg, err)
		return
	}
	if reply == nil {
		return
	}
	res = reply.InterrelationMap
	return
}
