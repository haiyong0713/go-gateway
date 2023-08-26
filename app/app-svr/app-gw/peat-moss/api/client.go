package api

import (
	"context"
	"fmt"

	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"
)

// AppID .
const AppID = "main.app-svr.app-gateway-grpc-proxy"

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (AppGatewayGRPCProxyClient, error) {
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", AppID))
	if err != nil {
		return nil, err
	}
	return NewAppGatewayGRPCProxyClient(cc), nil
}

// 生成 gRPC 代码
//go:generate kratos tool protoc --bm --grpc
