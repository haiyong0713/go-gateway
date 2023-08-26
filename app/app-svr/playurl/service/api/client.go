package api

import (
	"context"
	"fmt"

	"go-common/library/net/rpc/warden"

	"google.golang.org/grpc"
)

// AppID .
const (
	AppID      = "playurl.service"
	WindowSize = int32(65535000)
)

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (PlayURLClient, error) {
	opts = append(opts, grpc.WithInitialWindowSize(WindowSize))
	opts = append(opts, grpc.WithInitialConnWindowSize(WindowSize))
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", AppID))
	if err != nil {
		return nil, err
	}
	return NewPlayURLClient(cc), nil
}
