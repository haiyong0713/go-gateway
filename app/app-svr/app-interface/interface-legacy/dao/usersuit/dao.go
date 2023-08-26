package usersuit

import (
	"context"

	usersuitgrpc "git.bilibili.co/bapis/bapis-go/account/service/usersuit"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
)

// Dao is coin dao
type Dao struct {
	usersuitGRPC usersuitgrpc.UsersuitClient
}

// New initial coin dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.usersuitGRPC, err = usersuitgrpc.NewClient(c.UsersuitGRPC); err != nil {
		panic(err)
	}
	return
}

// InviteCountStat .
func (d *Dao) InviteCountStat(c context.Context, mid int64) (res *usersuitgrpc.InviteCountStatReply, err error) {
	if res, err = d.usersuitGRPC.InviteCountStat(c, &usersuitgrpc.InviteCountStatReq{Mid: mid, Ip: metadata.String(c, metadata.RemoteIP)}); err != nil {
		log.Error("d.usersuitGRPC.InviteCountStat(%d) error(%+v)", mid, err)
	}
	return
}
