package moss

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"go-common/library/conf/env"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

const (
	MatchTypeExact  = 1
	MatchTypePrefix = 2
	MatchTypeRegex  = 3
)

type Namespace struct {
	Namespace string `json:"namespace"`
	Env       string `json:"env"`
	Zone      string `json:"zone"`
}

type NamespaceReply struct {
	Namespaces []*Namespace `json:"namespaces"`
}

type Route struct {
	AppID     string `json:"app_id"`
	MatchStr  string `json:"match_str"`
	MatchType int64  `json:"match_type"`
	Upstream  struct {
		AppID string `json:"app_id"`
		Zone  string `json:"zone"`
	} `json:"upstream"`
}

type RouteReply struct {
	Routes []*Route `json:"routes"`
}

type MossLoader interface {
	ALLNamespaces(ctx context.Context) ([]*Namespace, error)
	ALLRoutes(ctx context.Context, namespace string) ([]*Route, error)
}

type mossLoader struct {
	cfg    Config
	client *bm.Client
}

type Config struct {
	HTTPClient *bm.ClientConfig
	Token      string
	Host       string
	Enable     bool
}

type fakeLoader struct{}

func (fakeLoader) ALLNamespaces(ctx context.Context) ([]*Namespace, error) {
	return nil, nil
}
func (fakeLoader) ALLRoutes(ctx context.Context, namespace string) ([]*Route, error) {
	return nil, nil
}

func New(cfg *Config) MossLoader {
	if !cfg.Enable {
		log.Warn("Disabling moss config loader: %+v", cfg)
		return &fakeLoader{}
	}
	return &mossLoader{
		cfg:    *cfg,
		client: bm.NewClient(cfg.HTTPClient),
	}
}

func (m *mossLoader) ALLNamespaces(ctx context.Context) ([]*Namespace, error) {
	reply := &struct {
		Code int64          `json:"code"`
		Data NamespaceReply `json:"data"`
	}{}
	req, err := m.client.NewRequest("GET", m.cfg.Host+"/api/v1/moss/open/namespaces", "", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Token", m.cfg.Token)
	if err := m.client.Do(ctx, req, reply); err != nil {
		return nil, err
	}
	if reply.Code != 0 {
		return nil, errors.Errorf("Failed to get moss all namespace with code: %d: %+v", reply.Code, reply)
	}
	return reply.Data.Namespaces, nil
}

func (m *mossLoader) ALLRoutes(ctx context.Context, namespace string) ([]*Route, error) {
	reply := &struct {
		Code int64      `json:"code"`
		Data RouteReply `json:"data"`
	}{}
	params := url.Values{}
	params.Set("namespace", namespace)
	req, err := m.client.NewRequest("GET", m.cfg.Host+"/api/v1/moss/open/routes", "", params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Token", m.cfg.Token)
	if err := m.client.Do(ctx, req, reply); err != nil {
		return nil, err
	}
	if reply.Code != 0 {
		return nil, errors.Errorf("Failed to get moss all routes with namespace: %q code: %d: %+v", namespace, reply.Code, reply)
	}
	return reply.Data.Routes, nil
}

func CurrentNamespace() string {
	if env.DeployEnv != "prod" {
		return fmt.Sprintf("%s.%s.%s", env.Region, env.Zone, env.DeployEnv)
	}
	idc := os.Getenv("IDC")
	return fmt.Sprintf("%s.%s.%s", env.Region, env.Zone, idc)
}
