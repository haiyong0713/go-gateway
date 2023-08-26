package blademaster

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"regexp"
	"sort"
	"strings"

	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/ab"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Config is
type Config struct {
	DynPath []*PathMeta
}

// Matcher is
type Matcher interface {
	Name() string
	RawPattern() string
	String() string
	Match(url *url.URL) bool
	Len() int
	Priority() int
}

// Annotation is
type Annotation map[string]string

// PathMeta is
type PathMeta struct {
	Pattern string
	matcher Matcher

	ValidatorDSN string
	validator    Validator

	BackupRetryOption  BackupRetryOption
	RateLimiterOption  RateLimiterOption
	RateLimiterOptions []*RateLimiterOption

	SDKConfig  sdk.Config
	ClientInfo metadata.ClientInfo
	Annotation Annotation
}

func (pm *PathMeta) GetMatcher() Matcher {
	return pm.matcher
}

// InitStatic is
func (pm *PathMeta) InitStatic() error {
	if pm.ValidatorDSN != "" {
		validator, err := BuildValidator(pm.ValidatorDSN)
		if err != nil {
			return err
		}
		pm.validator = validator
	}

	if pm.BackupRetryOption.BackupURL != "" {
		backupURL, err := url.Parse(pm.BackupRetryOption.BackupURL)
		if err != nil {
			return errors.WithStack(err)
		}
		pm.BackupRetryOption.backupURL = backupURL
	}
	pm.BackupRetryOption.forceBackupCondition = ab.ParseCondition(pm.BackupRetryOption.ForceBackupCondition)
	if pm.BackupRetryOption.forceBackupCondition == nil {
		pm.BackupRetryOption.forceBackupCondition = ab.FALSE
	}

	if pm.RateLimiterOption.Rule != "" {
		rule, err := resolveRateShardingRule(pm.RateLimiterOption.Rule)
		if err != nil {
			return err
		}
		pm.RateLimiterOption.rule = rule
	}

	for _, rateLimiter := range pm.RateLimiterOptions {
		if rateLimiter.Rule == "" {
			continue
		}
		rule, err := resolveRateShardingRule(rateLimiter.Rule)
		if err != nil {
			return err
		}
		rateLimiter.rule = rule
	}

	if pm.Pattern == "" {
		return errors.Errorf("empty path pattern")
	}
	if strings.HasPrefix(pm.Pattern, "/") {
		pm.matcher = &prefixMatcher{prefix: pm.Pattern}
		return nil
	}
	if strings.HasPrefix(pm.Pattern, "~ ") {
		rawExp := strings.TrimPrefix(pm.Pattern, "~ ")
		m, err := createRegexMatcher(rawExp)
		if err != nil {
			return err
		}
		pm.matcher = m
		return nil
	}
	if strings.HasPrefix(pm.Pattern, "= ") {
		extPath := strings.TrimPrefix(pm.Pattern, "= ")
		pm.matcher = &exactlyMatcher{path: extPath}
		return nil
	}
	return errors.Errorf("invalid path pattern: %s", pm.Pattern)
}

type exactlyMatcher struct{ path string }

func (em *exactlyMatcher) Name() string            { return "exactlyMatcher" }
func (em *exactlyMatcher) RawPattern() string      { return em.path }
func (em *exactlyMatcher) String() string          { return "= " + em.path }
func (em *exactlyMatcher) Match(url *url.URL) bool { return url.Path == em.path }
func (em *exactlyMatcher) Len() int                { return len(em.path) }
func (em *exactlyMatcher) Priority() int           { return 0 }

type prefixMatcher struct{ prefix string }

func (pm *prefixMatcher) Name() string            { return "prefixMatcher" }
func (pm *prefixMatcher) RawPattern() string      { return pm.prefix }
func (pm *prefixMatcher) String() string          { return pm.prefix }
func (pm *prefixMatcher) Match(url *url.URL) bool { return strings.HasPrefix(url.Path, pm.prefix) }
func (pm *prefixMatcher) Len() int                { return len(pm.prefix) }
func (pm *prefixMatcher) Priority() int           { return 1 }

type regexMatcher struct{ exp *regexp.Regexp }

func (rm *regexMatcher) Name() string            { return "regexMatcher" }
func (rm *regexMatcher) RawPattern() string      { return rm.exp.String() }
func (rm *regexMatcher) String() string          { return rm.exp.String() }
func (rm *regexMatcher) Match(url *url.URL) bool { return rm.exp.MatchString(url.Path) }
func (rm *regexMatcher) Len() int                { return len(rm.exp.String()) }
func (rm *regexMatcher) Priority() int           { return 2 }
func createRegexMatcher(exp string) (*regexMatcher, error) {
	p, err := regexp.Compile(exp)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &regexMatcher{exp: p}, nil
}

func (cfg Config) Digest() string {
	pm := make([]*PathMeta, len(cfg.DynPath))
	copy(pm, cfg.DynPath)
	sort.Slice(pm, func(i, j int) bool {
		l, r := pm[i], pm[j]
		return l.Pattern < r.Pattern
	})

	raw := struct {
		ProxyConfig Config
	}{}
	dup := cfg
	dup.DynPath = pm
	raw.ProxyConfig = dup
	buf := &bytes.Buffer{}
	//nolint:errcheck
	toml.NewEncoder(buf).Encode(raw)
	digest := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(digest[:])
}
