package grpc

import (
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	pb "go-gateway/app/app-svr/app-gw/gateway/api"
	sdkwardenserver "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden/server"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type proxyCfgSetter struct {
	*sdkwardenserver.ProxyPass
}

func (pcs proxyCfgSetter) Set(rawCfg string) error {
	raw := struct {
		ProxyConfig *sdkwardenserver.Config
	}{}
	if err := toml.Unmarshal([]byte(rawCfg), &raw); err != nil {
		return err
	}
	if raw.ProxyConfig != nil {
		if err := pcs.Reload(*raw.ProxyConfig); err != nil {
			return err
		}
	}
	return nil
}

func parseProxyTOML(ct *paladin.TOML) error {
	if err := paladin.Get("grpc-proxy-config.toml").Unmarshal(ct); err != nil {
		if errors.Cause(err) == paladin.ErrNotExist {
			_ = ct.UnmarshalText([]byte("[ProxyConfig]"))
			return nil
		}
		return err
	}
	return nil
}

func newProxyPass() (*sdkwardenserver.ProxyPass, error) {
	pct := paladin.TOML{}
	if err := parseProxyTOML(&pct); err != nil {
		return nil, err
	}
	proxyCfg := sdkwardenserver.Config{}
	if err := pct.Get("ProxyConfig").UnmarshalTOML(&proxyCfg); err != nil {
		return nil, err
	}
	proxy := sdkwardenserver.New(proxyCfg)
	if err := paladin.Watch("grpc-proxy-config.toml", &proxyCfgSetter{ProxyPass: proxy}); err != nil {
		if errors.Cause(err) == paladin.ErrNotExist {
			log.Warn("No `grpc-proxy-config.toml` file detected, disabling grpc proxy reload feature.")
			return proxy, nil
		}
		return nil, err
	}
	return proxy, nil
}

// NewGrpcProxy new a grpc server.
func NewGRPCProxy(svc pb.AppGatewayServer, engine *bm.Engine) (*warden.Server, error) {
	var (
		cfg warden.ServerConfig
		ct  paladin.TOML
	)
	if err := paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return nil, err
	}
	if err := ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	proxy, err := newProxyPass()
	if err != nil {
		return nil, err
	}
	handler, interceptorSetter := proxy.WrappedHandler()
	wardenServer := warden.NewServer(&cfg, grpc.UnknownServiceHandler(handler))
	interceptorSetter(wardenServer.Interceptor)
	pb.RegisterAppGatewayServer(wardenServer.Server(), svc)
	wardenServer, err = wardenServer.Start()
	if err != nil {
		return nil, err
	}
	proxy.SetupRouter(engine)
	return wardenServer, nil
}
