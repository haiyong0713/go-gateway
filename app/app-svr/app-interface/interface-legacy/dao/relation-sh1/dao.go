package relation_sh1

import (
	"context"
	"fmt"

	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"github.com/pkg/errors"
)

type Dao struct {
	// grpc
	relGRPC relationgrpc.RelationClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.relGRPC, err = relationgrpc.NewClient(c.RelationSh1GRPC); err != nil {
		panic(fmt.Sprintf("relationgrpc NewClientt error (%+v)", err))
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
