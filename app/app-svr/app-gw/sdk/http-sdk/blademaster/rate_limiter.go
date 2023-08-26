package blademaster

import (
	"fmt"
	"net/url"
	"strings"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	quota2 "go-common/library/rate/limit/quota"
	"go-common/library/stat/metric"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/request"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/sdkerr"

	"github.com/pkg/errors"
)

var (
	ServerQuotaLimited = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "http_server",
		Subsystem: "",
		Name:      "proxy_quota_limit_total",
		Help:      "http server quota limit total.",
		Labels:    []string{"appid", "method", "extra"},
	})
)

const (
	ErrCodeRequestRateLimited = "RequestRateLimited"
	RefererBlankDefault       = "referer:_blank"
	LimiterIteratorMax        = 50
	TotalRule                 = "total"
	RefererRule               = "referer"
)

type RateLimiterOption struct {
	Rule string
	rule RateShardingRule

	Preflight bool   // 只打日志不拦截
	DeployEnv string // 环境 默认从环境变量中获取
	Zone      string // 机房地址 默认从环境变量中获取
	AppID     string // 当前应用的 appid 默认从环境变量中获取
}

type RateShardingRule interface {
	Name() string
	Key(r *request.Request) string
}

type RuleTotal struct{}

func (RuleTotal) Name() string                  { return "total" }
func (RuleTotal) Key(_ *request.Request) string { return "total" }

type RuleReferer struct{}

func (RuleReferer) Name() string { return "referer" }
func (RuleReferer) Key(req *request.Request) string {
	referer := req.HTTPRequest.Header.Get("Referer")
	if referer == "" {
		return RefererBlankDefault
	}
	u, err := url.Parse(referer)
	if err != nil {
		log.Error("Failed to parse referer: %+v", err)
		return RefererBlankDefault
	}
	return fmt.Sprintf("referer:%s", u.Host)
}

var (
	_staticRuleTotal    = &RuleTotal{}
	_dynamicRuleReferer = &RuleReferer{}
)

func resolveRateShardingRule(in string) (RateShardingRule, error) {
	switch findPrefixRule(in) {
	case _staticRuleTotal.Name():
		return _staticRuleTotal, nil
	case _dynamicRuleReferer.Name():
		return _dynamicRuleReferer, nil
	default:
		return nil, errors.Errorf("unrecognized rule: %q", in)
	}
}

func findPrefixRule(in string) string {
	list := strings.Split(in, ":")
	return list[0]
}

// RateLimiterPatcher is
type RateLimiterPatcher struct {
	option   []*RateLimiterOption
	pathMeta *PathMeta
}

func NewRateLimiterPatcher(pathMeta *PathMeta, option []*RateLimiterOption) *RateLimiterPatcher {
	return &RateLimiterPatcher{
		option:   option,
		pathMeta: pathMeta,
	}
}

// Name is
func (rp *RateLimiterPatcher) Name() string {
	return "RateLimiterPatcher"
}

// Matched is
func (rp *RateLimiterPatcher) Matched(r *request.Request) bool {
	for _, option := range rp.option {
		if option.rule != nil {
			return true
		}
	}
	return false
}

// Patch is
func (rp *RateLimiterPatcher) Patch(in request.Handlers) request.Handlers {
	out := in.Copy()
	out.Validate.SetFrontNamed(request.NamedHandler{
		Name: "appgwsdk.blademaster.RateLimiterPatcherValidateHandler",
		Fn: func(r *request.Request) {
			pattern := rp.pathMeta.matcher.RawPattern()
			rules := make(map[string][]*RateLimiterOption)
			list := rp.option
			if len(rp.option) > LimiterIteratorMax {
				list = rp.option[0:LimiterIteratorMax]
			}
			for _, option := range list {
				rules[ParseRuleType(option.Rule)] = append(rules[ParseRuleType(option.Rule)], option)
			}
			for _, option := range rules[TotalRule] {
				allowed, matched, _ := optionAllow(r, pattern, option)
				if !matched {
					continue
				}
				if !allowed {
					return
				}
			}
			isMatchReferer := false
			for _, option := range rules[RefererRule] {
				allowed, matched, _ := optionAllow(r, pattern, option)
				if !matched {
					continue
				}
				isMatchReferer = true
				if !allowed {
					return
				}
			}
			if !isMatchReferer && len(rules[RefererRule]) > 0 {
				fullkey := fmt.Sprintf("%s.%s.%s|http|%s|%s", env.DeployEnv,
					env.Zone, env.AppID, pattern, RefererBlankDefault)
				limiter := quota2.NewAllower(&quota2.AllowerConfig{ID: fullkey, NotAllowStranger: false})
				if !limiter.Allow() {
					ServerQuotaLimited.Inc(env.AppID, pattern, RefererBlankDefault)
					setupRequestError(r, rp.pathMeta.Pattern, fullkey)
					return
				}
			}
		},
	})
	return out
}

func setupRateLimiter(r *request.Request, pm *PathMeta, option []*RateLimiterOption) {
	if pm.matcher.Name() != "exactlyMatcher" {
		// rate limit is only be supported on exactlyMatcher now.
		return
	}
	r.ApplyOptions(request.WithHandlerPatchers(NewRateLimiterPatcher(pm, option)))
}

func ParseRuleType(rule string) string {
	if strings.HasPrefix(rule, RefererRule) {
		return RefererRule
	}
	return TotalRule
}

//nolint:unparam
func optionAllow(r *request.Request, pattern string, option *RateLimiterOption) (isAllow, isMatch, isPreflight bool) {
	key := option.rule.Key(r)
	if option.Rule != key {
		return true, false, false
	}
	fullkey := fmt.Sprintf("%s.%s.%s|http|%s|%s",
		env.DeployEnv, env.Zone, env.AppID, pattern, key)
	limiter := quota2.NewAllower(&quota2.AllowerConfig{ID: fullkey, NotAllowStranger: false})
	if limiter.Allow() {
		return true, true, false
	}
	ServerQuotaLimited.Inc(env.AppID, pattern, key)
	if option.Preflight {
		log.Warn("quota: key: %q is running out of quota within permissive mode", fullkey)
		return true, true, true
	}
	setupRequestError(r, pattern, fullkey)
	return false, true, false
}

func setupRequestError(r *request.Request, pattern, fullkey string) {
	msg := fmt.Sprintf("request to pattern: %q is running out of quota on key: %q", pattern, fullkey)
	r.Error = sdkerr.New(ErrCodeRequestRateLimited, msg, errors.WithStack(ecode.Error(ecode.LimitExceed, msg)))
}
