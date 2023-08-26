package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"go-common/library/net/rpc/warden"
	xtime "go-common/library/time"
	sdkwarden "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type Config struct {
	DynService []*ServiceMeta
}

// Matcher is
type Matcher interface {
	Name() string
	RawPattern() string
	String() string
	Match(string) bool
	Len() int
	Priority() int
}

type exactlyMatcher struct{ path string }

func (em *exactlyMatcher) Name() string                    { return "exactlyMatcher" }
func (em *exactlyMatcher) RawPattern() string              { return em.path }
func (em *exactlyMatcher) String() string                  { return "= " + em.path }
func (em *exactlyMatcher) Match(serviceMethod string) bool { return serviceMethod == em.path }
func (em *exactlyMatcher) Len() int                        { return len(em.path) }
func (em *exactlyMatcher) Priority() int                   { return 0 }

type prefixMatcher struct{ prefix string }

func (pm *prefixMatcher) Name() string       { return "prefixMatcher" }
func (pm *prefixMatcher) RawPattern() string { return pm.prefix }
func (pm *prefixMatcher) String() string     { return pm.prefix }
func (pm *prefixMatcher) Match(serviceMethod string) bool {
	return strings.HasPrefix(serviceMethod, pm.prefix)
}
func (pm *prefixMatcher) Len() int      { return len(pm.prefix) }
func (pm *prefixMatcher) Priority() int { return 1 }

type regexMatcher struct{ exp *regexp.Regexp }

func (rm *regexMatcher) Name() string                    { return "regexMatcher" }
func (rm *regexMatcher) RawPattern() string              { return rm.exp.String() }
func (rm *regexMatcher) String() string                  { return rm.exp.String() }
func (rm *regexMatcher) Match(serviceMethod string) bool { return rm.exp.MatchString(serviceMethod) }
func (rm *regexMatcher) Len() int                        { return len(rm.exp.String()) }
func (rm *regexMatcher) Priority() int                   { return 2 }
func createRegexMatcher(exp string) (*regexMatcher, error) {
	p, err := regexp.Compile(exp)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &regexMatcher{exp: p}, nil
}

type ServiceMeta struct {
	Pattern     string
	matcher     Matcher
	ServiceName string // example: `account.service.Account`; prefer using `Pattern`.

	Target string // resolvable service target: `appid` or `discovery_id`

	ClientSDKConfig sdkwarden.ClientSDKConfig
	ClientConfig    warden.ClientConfig
}

func splitPackageService(serviceName string) (string, string) {
	pos := strings.LastIndex(serviceName, ".")
	if pos == -1 {
		return "", ""
	}
	packageName := serviceName[:pos]
	pureService := serviceName[pos+1:]
	return packageName, pureService
}

func (sm *ServiceMeta) Init() error {
	sm.ClientSDKConfig.AppID = sm.resolvableID()
	if err := sm.ClientSDKConfig.Init(); err != nil {
		return err
	}

	// 优先使用pattern
	if sm.Pattern == "" && sm.ServiceName != "" {
		sm.Pattern = fmt.Sprintf("/%s/", strings.Trim(sm.ServiceName, "/"))
	}
	if sm.Pattern == "" {
		return errors.Errorf("empty path pattern")
	}
	if strings.HasPrefix(sm.Pattern, "/") {
		sm.matcher = &prefixMatcher{prefix: sm.Pattern}
		return nil
	}
	if strings.HasPrefix(sm.Pattern, "~ ") {
		rawExp := strings.TrimPrefix(sm.Pattern, "~ ")
		m, err := createRegexMatcher(rawExp)
		if err != nil {
			return err
		}
		sm.matcher = m
		return nil
	}
	if strings.HasPrefix(sm.Pattern, "= ") {
		extPath := strings.TrimPrefix(sm.Pattern, "= ")
		sm.matcher = &exactlyMatcher{path: extPath}
		return nil
	}
	return errors.Errorf("invalid path pattern: %s", sm.Pattern)
}

func (sm ServiceMeta) resolvableID() string {
	resolveTarget, _ := splitPackageService(sm.ServiceName)
	if sm.Target != "" {
		resolveTarget = sm.Target
	}
	return resolveTarget
}

func (sm ServiceMeta) ResolvableTarget() string {
	return fmt.Sprintf("discovery://default/%s", sm.resolvableID())
}

func (sm ServiceMeta) FixedClientConfig() *warden.ClientConfig {
	clientCfg := sm.ClientConfig
	if clientCfg.Dial == xtime.Duration(0) {
		clientCfg.Dial = xtime.Duration(time.Second * 10)
	}
	if clientCfg.Timeout == xtime.Duration(0) {
		clientCfg.Timeout = xtime.Duration(2500 * time.Millisecond)
	}
	if clientCfg.Subset == 0 {
		clientCfg.Subset = 50
	}
	if clientCfg.KeepAliveInterval == xtime.Duration(0) {
		clientCfg.KeepAliveInterval = xtime.Duration(time.Second * 60)
	}
	if clientCfg.KeepAliveTimeout == xtime.Duration(0) {
		clientCfg.KeepAliveTimeout = xtime.Duration(time.Second * 20)
	}
	return &clientCfg
}

func (cfg Config) Digest() string {
	pm := make([]*ServiceMeta, len(cfg.DynService))
	copy(pm, cfg.DynService)
	sort.Slice(pm, func(i, j int) bool {
		l, r := pm[i], pm[j]
		return l.ServiceName < r.ServiceName
	})

	raw := struct {
		ProxyConfig Config
	}{}
	dup := cfg
	dup.DynService = pm
	raw.ProxyConfig = dup
	buf := &bytes.Buffer{}
	//nolint:errcheck
	toml.NewEncoder(buf).Encode(raw)
	digest := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(digest[:])
}
