package api

import (
	"context"

	playeronlinegrpc "git.bilibili.co/bapis/bapis-go/bilibili/app/playeronline/v1"
	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"
)

// appid from package name
const appID = "main.app-svr.player-online"

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (playeronlinegrpc.PlayerOnlineClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+appID)
	if err != nil {
		return nil, err
	}
	return playeronlinegrpc.NewPlayerOnlineClient(conn), nil
}
