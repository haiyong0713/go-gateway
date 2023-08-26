//nolint:biliautomaxprocs
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"go-common/library/conf/paladin"
	"go-common/library/net/rpc/warden"
	wardensdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden"

	vipInforpc "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
)

const appID = "vipinfo.service"

type configSetter struct {
	*wardensdk.InterceptorBuilder
}

func (cs configSetter) Set(rawCfg string) error {
	raw := struct {
		SDKBuilderConfig *wardensdk.SDKBuilderConfig
	}{}
	if err := toml.Unmarshal([]byte(rawCfg), &raw); err != nil {
		return err
	}
	if raw.SDKBuilderConfig != nil {
		if err := cs.Reload(*raw.SDKBuilderConfig); err != nil {
			return err
		}
	}
	return nil
}

func NewBuilder() (*wardensdk.InterceptorBuilder, error) {
	ct := paladin.TOML{}
	if err := paladin.Get("grpc-client-sdk.toml").Unmarshal(&ct); err != nil {
		return nil, err
	}

	cfg := wardensdk.SDKBuilderConfig{}
	if err := ct.Get("SDKBuilderConfig").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	if err := cfg.Init(); err != nil {
		return nil, err
	}
	builder := wardensdk.NewBuilder(cfg)
	if err := paladin.Watch("grpc-client-sdk.toml", &configSetter{InterceptorBuilder: builder}); err != nil {
		return nil, err
	}
	return builder, nil
}

func newBuilder() *wardensdk.InterceptorBuilder {
	builder, err := NewBuilder()
	if err != nil {
		panic(err)
	}
	return builder
}

var builder *wardensdk.InterceptorBuilder

func NewClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) (vipInforpc.VipInfoClient, error) {
	client := warden.NewClient(cfg, opts...)
	clientSDK := builder.Build(appID)
	client.Use(clientSDK.UnaryClientInterceptor())
	conn, err := client.Dial(context.Background(), "discovery://default/"+appID)
	if err != nil {
		return nil, err
	}
	return vipInforpc.NewVipInfoClient(conn), nil
}

func init() {
	flag.Parse()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	builder = newBuilder()
}

func jsonify(in interface{}) string {
	out, _ := json.Marshal(in)
	return string(out)
}

func main() {
	vipClient, err := NewClient(nil)
	if err != nil {
		panic(err)
	}
	reply, err := vipClient.Info(context.TODO(), &vipInforpc.InfoReq{Mid: 2231365})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Reply: %+v\n", jsonify(reply))
}
