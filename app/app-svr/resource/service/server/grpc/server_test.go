package grpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
	xtime "go-common/library/time"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/conf"
	"go-gateway/app/app-svr/resource/service/service"
)

// rpc server const
const (
	addr = "10.23.167.167:9000"
)

// TestResource test rpc server
func TestResource(t *testing.T) {
	err := paladin.Init()
	if err != nil {
		t.Errorf("conf.Init() error(%v)", err)
		t.FailNow()
	}
	defer paladin.Close()
	cfgu := &conf.Config{}
	if err = paladin.Get("resource-service.toml").UnmarshalTOML(&cfgu); err != nil {
		t.Errorf("conf.Init() error(%v)", err)
		t.FailNow()
	}
	svr := service.New(cfgu)
	New(nil, svr)
	time.Sleep(time.Second * 3)
	cfg := &warden.ClientConfig{
		Dial:    xtime.Duration(time.Second * 3),
		Timeout: xtime.Duration(time.Second * 3),
	}
	cc, err := warden.NewClient(cfg).Dial(context.Background(), addr)
	if err != nil {
		t.Errorf("rpc.Dial(tcp, \"%s\") error(%v)", addr, err)
		t.FailNow()
	}
	client := pb.NewResourceClient(cc)
	WebRcmdGRPC(client, t)
	Banners2(client, t)
}

func WebRcmdGRPC(client pb.ResourceClient, t *testing.T) {
	arg := &pb.NoArgRequest{}
	res, err := client.WebRcmd(context.TODO(), arg)
	if err != nil {
		t.Error(err)
	} else {
		result("web rcmd", t, res)
	}
}

func Banners2(client pb.ResourceClient, t *testing.T) {
	arg := &pb.BannersRequest{ResIDs: "3391"}
	res, err := client.Banners2(context.TODO(), arg)
	if err != nil {
		t.Error(err)
	} else {
		result("banners2", t, res)
	}
}

func result(name string, t *testing.T, res interface{}) {
	fmt.Printf("res : %+v \n", res)
	t.Log("[==========" + name + "单元测试结果==========]")
	t.Log(res)
	t.Log("[↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑]\r\n")
}
