package v1

import (
	"context"

	"go-common/library/net/rpc/warden"

	livexfans "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	pgcstat "git.bilibili.co/bapis/bapis-go/pgc/service/stat/v1"

	"google.golang.org/grpc"
)

const (
	pgcAppID  = "pgc.service.card"
	statAppID = "pgc.stat.service"
)

func newPgcClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (pgcsearch.SearchClient, pgcinline.InlineCardClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+pgcAppID)
	if err != nil {
		return nil, nil, err
	}
	return pgcsearch.NewSearchClient(conn), pgcinline.NewInlineCardClient(conn), nil
}

// newStatClient new xfansmedal grpc client
func newStatClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (pgcstat.StatServiceClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+statAppID)
	if err != nil {
		return nil, err
	}
	return pgcstat.NewStatServiceClient(conn), nil
}

const (
	appID     = "live.xfansmedal"
	roomAppID = "live.xroom"
)

func newLiveClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (livexfans.AnchorClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+appID)
	if err != nil {
		return nil, err
	}
	return livexfans.NewAnchorClient(conn), nil
}

func newLiveRoomClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (livexroom.RoomClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+roomAppID)
	if err != nil {
		return nil, err
	}
	return livexroom.NewRoomClient(conn), nil
}
