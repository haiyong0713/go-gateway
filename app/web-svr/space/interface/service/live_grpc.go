package service

import (
	"context"

	livexfans "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"
)

// appID ...
const appID = "live.xfansmedal"

// NewLiveClient new xfansmedal grpc client
func NewLiveClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (livexfans.AnchorClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+appID)
	if err != nil {
		return nil, err
	}
	return livexfans.NewAnchorClient(conn), nil
}
