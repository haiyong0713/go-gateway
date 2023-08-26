package tunnel

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/app-feed/interface/conf"

	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"

	"google.golang.org/grpc"
)

// appID ...
const appID = "main.community.tunnel-service"

// newClient new xfansmedal grpc client
func newClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (tunnelgrpc.TunnelClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+appID)
	if err != nil {
		return nil, err
	}
	return tunnelgrpc.NewTunnelClient(conn), nil
}

type Dao struct {
	// grpc
	rpcClient tunnelgrpc.TunnelClient
}

func New(c *conf.Config) (d *Dao) {
	var err error
	d = &Dao{}
	if d.rpcClient, err = newClient(c.TunnelGRPC); err != nil {
		panic(fmt.Sprintf("tunnel grpc newClient error (%+v)", err))
	}
	return
}

// FeedCards list
func (d *Dao) FeedCards(c context.Context, mobiApp string, mid, build int64, gatherOids [][]int64) (res map[int64]*tunnelgrpc.FeedCard, err error) {
	var (
		arg = &tunnelgrpc.FeedCardsReq{
			Platform:   mobiApp,
			Mid:        mid,
			Build:      build,
			MobiApp:    mobiApp,
			GatherOids: constructGatherOids(gatherOids),
		}
		resp *tunnelgrpc.FeedCardsReply
	)
	if resp, err = d.rpcClient.FeedCards(c, arg); err != nil || resp == nil {
		log.Error("tunnel grpc FeedCards error(%v) or is resp null", err)
		return
	}
	res = resp.FeedCards
	return
}

func constructGatherOids(gatherOids [][]int64) []*tunnelgrpc.FeedCardsReqGather {
	out := make([]*tunnelgrpc.FeedCardsReqGather, 0, len(gatherOids))
	for _, oids := range gatherOids {
		out = append(out, &tunnelgrpc.FeedCardsReqGather{Oids: oids})
	}
	return out
}
