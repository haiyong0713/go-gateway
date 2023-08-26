package v2

import (
	"context"

	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"
)

// to suppressed 'imported but not used warning'
var _ context.Context
var _ *warden.Client
var _ *grpc.ClientConn

// appid from package name
const appID = "app.dynamic"

// NewClient new a app.dynamic grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (DynamicClient, error) {
	client := warden.NewClient(cfg, opts...)
	conn, err := client.Dial(context.Background(), "discovery://default/"+appID)
	if err != nil {
		return nil, err
	}
	return NewDynamicClient(conn), nil
}
