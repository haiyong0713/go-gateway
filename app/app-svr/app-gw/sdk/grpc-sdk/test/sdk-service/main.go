//nolint:biliautomaxprocs
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/net/rpc/warden"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/playurl/service/api/v2"
	"google.golang.org/grpc"
)

func jsonify(in interface{}) string {
	out, _ := json.Marshal(in)
	return string(out)
}

// NewClient new grpc client
func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (v2.PlayURLClient, error) {
	client := warden.NewClient(cfg, opts...)
	cc, err := client.Dial(context.Background(), "127.0.0.1:9000")
	if err != nil {
		return nil, err
	}
	return v2.NewPlayURLClient(cc), nil
}

func main() {
	client, err := NewClient(&warden.ClientConfig{
		Timeout: xtime.Duration(time.Second),
	})
	if err != nil {
		panic(err)
	}
	resp, err := client.PlayView(context.Background(), &v2.PlayViewReq{
		Aid:       10318716,
		Cid:       10211289,
		Qn:        32,
		Platform:  "ios",
		Fnval:     16,
		Mid:       27515232,
		BackupNum: 5,
		Device:    "pad",
		MobiApp:   "iphone",
		Build:     10045,
		Buvid:     "fce10ba2538621a8e09261e8d3377c7b",
		VerifyVip: 1,
		NetType:   v2.NetworkType_WIFI,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(jsonify(resp))
}
