package http

import (
	"context"
	"fmt"
	"sync"

	"github.com/robfig/cron"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-gw/peat-moss/internal/dao/moss"
	sdkwardenserver "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden/server"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type proxyCfgSetter struct {
	*sdkwardenserver.ProxyPass
	mossLoader moss.MossLoader
	cron       *cron.Cron

	setterLock sync.Mutex
}

func mashupMossConfig(mossLoader moss.MossLoader, dst *sdkwardenserver.Config) error {
	routes, err := mossLoader.ALLRoutes(context.Background(), moss.CurrentNamespace())
	if err != nil {
		return err
	}
	for _, r := range routes {
		sm := &sdkwardenserver.ServiceMeta{}
		switch r.MatchType {
		case moss.MatchTypeExact:
			sm.Pattern = fmt.Sprintf("= %s", r.MatchStr)
			sm.Target = r.Upstream.AppID
		case moss.MatchTypePrefix:
			sm.Pattern = r.MatchStr
			sm.Target = r.Upstream.AppID
		case moss.MatchTypeRegex:
			sm.Pattern = fmt.Sprintf("~ %s", r.MatchStr)
			sm.Target = r.Upstream.AppID
		default:
			log.Warn("Unrecognized moss match type: %+v", r)
			continue
		}
		dst.DynService = append(dst.DynService, sm)
	}
	return nil
}

func (pcs *proxyCfgSetter) mashupMossConfig(dst *sdkwardenserver.Config) error {
	return mashupMossConfig(pcs.mossLoader, dst)
}

func (pcs *proxyCfgSetter) Set(rawCfg string) error {
	pcs.setterLock.Lock()
	defer pcs.setterLock.Unlock()

	raw := struct {
		ProxyConfig *sdkwardenserver.Config
	}{}
	if err := toml.Unmarshal([]byte(rawCfg), &raw); err != nil {
		return err
	}
	if err := pcs.mashupMossConfig(raw.ProxyConfig); err != nil {
		return err
	}
	if raw.ProxyConfig != nil {
		if err := pcs.Reload(*raw.ProxyConfig); err != nil {
			return err
		}
	}
	return nil
}

func (pcs *proxyCfgSetter) periodicLoader() func() {
	inner := func() error {
		pct := paladin.TOML{}
		if err := parseProxyTOML(&pct); err != nil {
			return err
		}
		proxyCfg := sdkwardenserver.Config{}
		if err := pct.Get("ProxyConfig").UnmarshalTOML(&proxyCfg); err != nil {
			return err
		}
		if err := pcs.mashupMossConfig(&proxyCfg); err != nil {
			return err
		}
		if err := pcs.Reload(proxyCfg); err != nil {
			return err
		}
		return nil
	}
	return func() {
		if err := inner(); err != nil {
			log.Error("Failed to load config periodically: %+v", err)
			return
		}
		//nolint:gosimple
		return
	}
}

func (pcs *proxyCfgSetter) startCronjob() {
	if err := pcs.cron.AddFunc("@every 1m", pcs.periodicLoader()); err != nil {
		panic(err)
	}
	pcs.cron.Start()
}

func parseProxyTOML(ct *paladin.TOML) error {
	if err := paladin.Get("grpc-proxy-config.toml").Unmarshal(ct); err != nil {
		if errors.Cause(err) == paladin.ErrNotExist {
			//nolint:errcheck
			ct.UnmarshalText([]byte("[ProxyConfig]"))
			return nil
		}
		return err
	}
	return nil
}

func parseMossConfigLoader(ct *paladin.TOML) error {
	if err := paladin.Get("application.toml").Unmarshal(ct); err != nil {
		if errors.Cause(err) == paladin.ErrNotExist {
			//nolint:errcheck
			ct.UnmarshalText([]byte("[MossConfigLoader]"))
			return nil
		}
		return err
	}
	return nil
}

func mossConfigLoader() (moss.MossLoader, error) {
	ct := paladin.TOML{}
	if err := parseMossConfigLoader(&ct); err != nil {
		return nil, err
	}
	mossCfg := moss.Config{}
	if err := ct.Get("MossConfigLoader").UnmarshalTOML(&mossCfg); err != nil {
		return nil, err
	}
	mossLoader := moss.New(&mossCfg)
	return mossLoader, nil
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
	mossLoader, err := mossConfigLoader()
	if err != nil {
		log.Warn("Failed to build moss config loader: %+v", err)
		return nil, err
	}
	if err := mashupMossConfig(mossLoader, &proxyCfg); err != nil {
		log.Warn("Failed to mashup moss config: %+v", err)
		return nil, err
	}
	proxy := sdkwardenserver.New(proxyCfg)
	cfgSetter := &proxyCfgSetter{
		ProxyPass:  proxy,
		mossLoader: mossLoader,
		cron:       cron.New(),
	}
	cfgSetter.startCronjob()
	if err := paladin.Watch("grpc-proxy-config.toml", cfgSetter); err != nil {
		if errors.Cause(err) == paladin.ErrNotExist {
			log.Warn("No `grpc-proxy-config.toml` file detected, disabling grpc proxy reload feature.")
			return proxy, nil
		}
		return nil, err
	}
	return proxy, nil
}
