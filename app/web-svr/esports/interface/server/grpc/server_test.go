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
	pb "go-gateway/app/web-svr/esports/interface/api/v1"
)

var client pb.EsportsClient

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
	client = pb.NewEsportsClient(cc)
}

// TestLiveContests test rpc server
func TestLiveContests(t *testing.T) {
	arg := &pb.LiveContestsRequest{Cids: []int64{2, 23}}
	resp, err := client.LiveContests(context.Background(), arg)
	if err != nil {
		log.Error("get value failed!err:=%v", err)
		return
	}
	b, _ := json.Marshal(resp.Contests)
	fmt.Printf("get SpecialReply:%+v", string(b))
}

// TestLiveAddFav .
func TestLiveAddFav(t *testing.T) {
	arg := &pb.FavRequest{Cid: 499, Mid: 100}
	_, err := client.LiveAddFav(context.Background(), arg)
	if err != nil {
		log.Error("LiveAddFav err:=%v", err)
		return
	}
	fmt.Printf("LiveAddFav success!")
}

// TestLiveDelFavtest .
func TestLiveDelFav(t *testing.T) {
	arg := &pb.FavRequest{Cid: 499, Mid: 100}
	_, err := client.LiveDelFav(context.Background(), arg)
	if err != nil {
		log.Error("get value failed!err:=%v", err)
		return
	}
	fmt.Printf("LiveDelFav success!")
}
