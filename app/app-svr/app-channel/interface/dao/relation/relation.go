package relation

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-channel/interface/conf"

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
