package grpc

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"testing"
	"time"

	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	xtime "go-common/library/time"
	pb "go-gateway/app/web-svr/space/interface/api/v1"
)

var client pb.SpaceClient

func init() {
	flag.Parse()
	cfg := &warden.ClientConfig{
		Dial:    xtime.Duration(time.Second * 3),
		Timeout: xtime.Duration(time.Second * 3),
	}
	cc, err := warden.NewClient(cfg).Dial(context.Background(), "127.0.0.1:9000")
	if err != nil {
		log.Error("new client failed!err:=%v", err)
		return
	}
	client = pb.NewSpaceClient(cc)
}

// TestOfficial .
func TestOfficial(t *testing.T) {
	arg := &pb.OfficialRequest{Mid: 1}
	resp, err := client.Official(context.Background(), arg)
	if err != nil {
		log.Error("get value failed!err:=%v", err)
		return
	}
	b, _ := json.Marshal(resp)
	fmt.Printf("get Official:%+v", string(b))
}
