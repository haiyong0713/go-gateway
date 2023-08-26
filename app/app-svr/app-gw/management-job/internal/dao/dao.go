package dao

import (
	"context"
	"encoding/json"
	"os"

	"go-common/library/conf/env"
	"go-common/library/conf/paladin.v2"
	"go-common/library/database/elastic"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/ip"
	gwconfig "go-gateway/app/app-svr/app-gw/management-job/internal/model/gateway-config"
	pb "go-gateway/app/app-svr/app-gw/management/api"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewKV)

// Dao dao interface
//
//go:generate kratos tool btsgen
type Dao interface {
	Close()
	Ping(ctx context.Context) error
	ListBreakerAPI(ctx context.Context, node string, gateway string) ([]*pb.BreakerAPI, error)
	PushConfigs(ctx context.Context, req *gwconfig.PushConfigReq) error
	ListGateway(ctx context.Context) ([]*pb.Gateway, error)
	ListDynPath(ctx context.Context, node string, gateway string) ([]*pb.DynPath, error)
	RawConfigs(ctx context.Context, req *gwconfig.RawConfigReq) ([]byte, error)
	RawTaskLog(ctx context.Context, req *gwconfig.RawLogReq) (*gwconfig.LogReply, error)
	Gateway(ctx context.Context, node, gateway string) (*pb.Gateway, error)
	GetTaishan() *Taishan
	GetQuotaMethods(ctx context.Context, node, gateway string) ([]*pb.QuotaMethod, error)
	GRPCListBreakerAPI(ctx context.Context, node string, gateway string) ([]*pb.BreakerAPI, error)
	GRPCListDynService(ctx context.Context, node string, gateway string) ([]*pb.DynPath, error)
}

// dao dao.
type dao struct {
	taishan    *Taishan
	httpClient *bm.Client
	es         *elastic.Elastic
}

// New new a dao and return.
func New(taishan *Taishan) (d Dao, cf func(), err error) {
	return newDao(taishan)
}

func newDao(taishan *Taishan) (d *dao, cf func(), err error) {
	var cfg struct {
		HTTPClient *bm.ClientConfig
		ES         *elastic.Config
	}
	if err = paladin.Get("http.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	d = &dao{
		taishan:    taishan,
		httpClient: bm.NewClient(cfg.HTTPClient, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		es:         elastic.NewElastic(cfg.ES),
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}

func (d *dao) GetTaishan() *Taishan {
	return d.taishan
}

type Instance struct {
	HostName string `json:"hostname"`
	IP       string `json:"ip"`
	Port     string `json:"port"`
}

func InstanceValue() []byte {
	hostname, _ := os.Hostname()
	advertisedIP := ip.InternalIP()
	if env.IP != "" {
		advertisedIP = env.IP
	}
	port := "8000"
	if env.HTTPPort != "" {
		port = env.HTTPPort
	}
	i := &Instance{HostName: hostname, IP: advertisedIP, Port: port}
	out, _ := json.Marshal(i)
	return out
}
