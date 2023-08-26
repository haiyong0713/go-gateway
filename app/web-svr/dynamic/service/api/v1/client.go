package v1

import (
	"context"
	"fmt"

	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"
)

// AppID .
const AppID = "archive.service.dynamic"

// NewClient new grpc client .
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (DynamicClient, error) {
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", AppID))
	if err != nil {
		return nil, err
	}
	return NewDynamicClient(cc), nil
}
