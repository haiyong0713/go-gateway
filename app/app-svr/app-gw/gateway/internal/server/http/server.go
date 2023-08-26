package http

import (
	"encoding/json"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	gotime "go-common/library/time"
	pb "go-gateway/app/app-svr/app-gw/gateway/api"
	httpsdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	sdkbm "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// nolint:unused
var svc pb.AppGatewayServer

func jsonify(in interface{}) string {
	out, _ := json.Marshal(in)
	return string(out)
}

// NewHttpProxy new a bm server.
func NewHTTPProxy(s pb.AppGatewayServer) (engine *bm.Engine, err error) {
	var (
		cfg bm.ServerConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	svc = s
	engine = GatewayServer(&cfg)
	proxyPass, err := newProxyPass()
	if err != nil {
		return
	}
	initRouter(engine)
	proxyPass.SetupRouter(engine)
	err = engine.Start()
	return
}

// GatewayServer returns an Engine instance with the Recovery, Logger and Mobile already attached.
func GatewayServer(conf *bm.ServerConfig) *bm.Engine {
	engine := bm.NewServer(conf)
	engine.Use(bm.Recovery(), bm.Trace(), bm.Logger(), bm.Mobile())
	return engine
}

type proxyCfgSetter struct {
	*sdkbm.ProxyPass
}

func (pcs proxyCfgSetter) Set(rawCfg string) error {
	raw := struct {
		ProxyConfig *sdkbm.Config
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

func parseClientResolver(ct paladin.TOML, dst *resolver.Config) error {
	if err := ct.Get("ClientResolver").UnmarshalTOML(dst); err != nil {
		if errors.Cause(err) == paladin.ErrNotExist {
			*dst = resolver.Config{}
			log.Warn("No `ClientResolver` config found, using default config: %+v", jsonify(dst))
			return nil
		}
		return err
	}
	return nil
}

func parseClient(ct paladin.TOML, dst *bm.ClientConfig) error {
	if err := ct.Get("Client").UnmarshalTOML(dst); err != nil {
		if errors.Cause(err) == paladin.ErrNotExist {
			*dst = bm.ClientConfig{
				App: &bm.App{
					Key:    "4db0b87a307cad39",
					Secret: "54af8823069524ca66ba1b9cae12ee4d",
				},
				Dial:      gotime.Duration(200 * time.Millisecond),
				Timeout:   gotime.Duration(time.Second),
				KeepAlive: gotime.Duration(time.Second * 60),
			}
			log.Warn("No `Client` config found, using default config: %+v", jsonify(dst))
			return nil
		}
		return err
	}
	return nil
}

func parseHttpProxy(ct paladin.TOML, dst *httpsdk.Config) error {
	if err := ct.Get("HttpProxy").UnmarshalTOML(dst); err != nil {
		if errors.Cause(err) == paladin.ErrNotExist {
			*dst = httpsdk.Config{
				Key:    "4db0b87a307cad39",
				Secret: "eeafadc0a6077cad3ec386e5ef33addd",
				Debug:  false,
			}
			log.Warn("No `HttpProxy` config found, using default config: %+v", jsonify(dst))
			return nil
		}
		return err
	}
	return nil
}

func parseClientInfo(ct paladin.TOML, dst *metadata.ClientInfo) error {
	if err := ct.Get("ClientInfo").UnmarshalTOML(dst); err != nil {
		if errors.Cause(err) == paladin.ErrNotExist {
			*dst = metadata.ClientInfo{
				AppID:    "main.web-svr.www-old",
				Endpoint: "discovery://main.web-svr.www-old",
			}
			log.Warn("No `ClientInfo` config found, using default config: %+v", jsonify(dst))
			return nil
		}
		return err
	}
	return nil
}

func parseProxyTOML(ct *paladin.TOML) error {
	if err := paladin.Get("proxy.toml").Unmarshal(ct); err != nil {
		if errors.Cause(err) == paladin.ErrNotExist {
			_ = ct.UnmarshalText([]byte(""))
			return nil
		}
		return err
	}
	return nil
}

func newProxyPass() (*sdkbm.ProxyPass, error) {
	ct := paladin.TOML{}
	if err := parseProxyTOML(&ct); err != nil {
		return nil, err
	}

	resolverCfg := resolver.Config{}
	if err := parseClientResolver(ct, &resolverCfg); err != nil {
		return nil, err
	}
	dynResolver := resolver.New(&resolverCfg, discovery.Builder())

	cliCfg := bm.ClientConfig{}
	if err := parseClient(ct, &cliCfg); err != nil {
		return nil, err
	}
	bmCli := bm.NewClient(&cliCfg, bm.SetResolver(dynResolver))

	cfg := httpsdk.Config{}
	if err := parseHttpProxy(ct, &cfg); err != nil {
		return nil, err
	}
	cfg.Client = sdkbm.WrapClient(bmCli)

	info := metadata.ClientInfo{}
	if err := parseClientInfo(ct, &info); err != nil {
		return nil, err
	}

	pct := paladin.TOML{}
	if err := paladin.Get("proxy-config.toml").Unmarshal(&pct); err != nil {
		return nil, err
	}
	proxyCfg := sdkbm.Config{}
	if err := pct.Get("ProxyConfig").UnmarshalTOML(&proxyCfg); err != nil {
		return nil, err
	}
	proxy := sdkbm.New(proxyCfg, cfg, info)
	if err := paladin.Watch("proxy-config.toml", &proxyCfgSetter{ProxyPass: proxy}); err != nil {
		return nil, err
	}
	return proxy, nil
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
}

func ping(ctx *bm.Context) {}
