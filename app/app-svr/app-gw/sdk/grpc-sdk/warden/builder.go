package warden

import (
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

type SDKBuilderConfig struct {
	ClientSDK []*ClientSDKConfig
	mapped    atomic.Value
}

type InterceptorBuilder struct {
	sync.Mutex
	cfg         SDKBuilderConfig
	interceptor map[string]*ClientInterceptor
}

func (sbc *SDKBuilderConfig) Init() error {
	mapped := make(map[string]*ClientSDKConfig, len(sbc.ClientSDK))
	for i, csc := range sbc.ClientSDK {
		if err := csc.Init(); err != nil {
			return errors.Wrapf(err, "invalid client sdk config in client config with index: %d", i)
		}
		mapped[csc.AppID] = csc
	}
	sbc.mapped.Store(mapped)
	return nil
}

func (sbc *SDKBuilderConfig) Fetch(appID string) ClientSDKConfig {
	mapped := sbc.mapped.Load().(map[string]*ClientSDKConfig)
	csc, ok := mapped[appID]
	if !ok {
		return ClientSDKConfig{AppID: appID}
	}
	return *csc
}

func NewBuilder(cfg SDKBuilderConfig) *InterceptorBuilder {
	builder := &InterceptorBuilder{
		interceptor: map[string]*ClientInterceptor{},
	}
	if err := builder.Reload(cfg); err != nil {
		panic(err)
	}
	return builder
}

func (ib *InterceptorBuilder) Build(appID string) *ClientInterceptor {
	ib.Lock()
	defer ib.Unlock()
	ci, ok := ib.interceptor[appID]
	if ok {
		return ci
	}
	sdkCfg := ib.cfg.Fetch(appID)
	ci = New(sdkCfg)
	ib.interceptor[appID] = ci
	return ci
}

func (ib *InterceptorBuilder) Reload(cfg SDKBuilderConfig) error {
	if err := cfg.Init(); err != nil {
		return err
	}
	ib.Lock()
	defer ib.Unlock()
	ib.cfg = cfg

	for appID, interceptor := range ib.interceptor {
		sdkCfg := ib.cfg.Fetch(appID)
		if err := interceptor.Reload(sdkCfg); err != nil {
			return err
		}
	}
	return nil
}
