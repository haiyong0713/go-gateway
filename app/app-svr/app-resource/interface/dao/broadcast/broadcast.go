package broadcast

import (
	"context"
	"fmt"

	pb "git.bilibili.co/bapis/bapis-go/infra/service/broadcast"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/conf"
)

type Dao struct {
	c *conf.Config
	// grpc
	rpcClient pb.ZergClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.rpcClient, err = pb.NewClient(c.BroadcastRPC); err != nil {
		panic(fmt.Sprintf("BroadcastRPC warden.NewClient error (%+v)", err))
	}
	return
}

// ServerList warden server list
func (d *Dao) ServerList(ctx context.Context, platform string) (res *pb.ServerListReply, err error) {
	arg := &pb.ServerListReq{
		Platform: platform,
	}
	if res, err = d.rpcClient.ServerList(ctx, arg); err != nil {
		log.Error("d.rpcClient.ServerList error(%v)", err)
		return
	}
	return
}
