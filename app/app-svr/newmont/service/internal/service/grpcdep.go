package service

import (
	"context"
	"fmt"
	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"

	"github.com/pkg/errors"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	tus "git.bilibili.co/bapis/bapis-go/datacenter/service/titan"
	hmtchannelgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	opIcon "git.bilibili.co/bapis/bapis-go/manager/operation/icon"
	display "git.bilibili.co/bapis/bapis-go/platform/interface/display"
	resV2 "git.bilibili.co/bapis/bapis-go/resource/service/v2"
)

type grpcDep struct {
	resClentV2       resV2.ResourceClient
	displayClient    display.DisplayClient
	locClient        locgrpc.LocationClient
	hmtChannelClient hmtchannelgrpc.ChannelRPCClient
	opIconClient     opIcon.OperationItemIconV1Client
	tusClient        tus.TitanUserServerClient
}

func initDep() *grpcDep {
	var grpcCfg struct {
		ResV2Client      *warden.ClientConfig
		DisplayClient    *warden.ClientConfig
		LocClient        *warden.ClientConfig
		HmtchannelClient *warden.ClientConfig
		OpIconClient     *warden.ClientConfig
		TusClient        *warden.ClientConfig
	}
	if err := paladin.Get("grpc.toml").UnmarshalTOML(&grpcCfg); err != nil {
		panic(err)
	}
	var (
		g   = &grpcDep{}
		err error
	)
	if g.resClentV2, err = resV2.NewClient(grpcCfg.ResV2Client); err != nil {
		panic(err)
	}
	if g.displayClient, err = display.NewClient(grpcCfg.DisplayClient); err != nil {
		panic(err)
	}
	if g.locClient, err = locgrpc.NewClient(grpcCfg.LocClient); err != nil {
		panic(err)
	}
	if g.hmtChannelClient, err = hmtchannelgrpc.NewClient(grpcCfg.HmtchannelClient); err != nil {
		panic(err)
	}
	if g.opIconClient, err = opIcon.NewClientOperationItemIconV1(grpcCfg.OpIconClient); err != nil {
		panic(err)
	}
	if g.tusClient, err = func(cfg *warden.ClientConfig) (tus.TitanUserServerClient, error) {
		client := warden.NewClient(cfg)
		conn, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", "datacenter.titan.tus"))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return tus.NewTitanUserServerClient(conn), nil
	}(grpcCfg.TusClient); err != nil {
		panic(err)
	}
	return g
}
