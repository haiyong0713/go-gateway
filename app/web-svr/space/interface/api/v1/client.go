package v1

import (
	fmt "fmt"

	"go-common/library/net/rpc/warden"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

// AppID .
const AppID = "main.web-svr.space-interface"

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (SpaceClient, error) {
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", AppID))
	if err != nil {
		return nil, err
	}
	return NewSpaceClient(cc), nil
}

//go:generate kratos tool protoc --grpc api.proto
