package conf

import (
	"errors"
	"flag"

	"go-common/library/cache/memcache"
	"go-common/library/conf"
	"go-common/library/database/sql"
	xlog "go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/time"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	client   *conf.Client
	// Conf is global config
	Conf = &Config{}
)

// Config service config
type Config struct {
	Version     string `toml:"version"`
	Static      string
	LocsDegrade bool
	// reload
	Reload ReloadInterval
	// app
	App *bm.App
	// Auth
	Auth *auth.Config
	// verify
	Verify *verify.Config
	// http
	BM     *HTTPServers
	Tracer *trace.Config
	// db
	MySQL *MySQL
	// rpc
	RPCClient2 *RPCClient2
	// httpClient
	HTTPClient *bm.ClientConfig
	// Host
	Host *Host
	// XLog
	XLog *xlog.Config
	// DegradeConfig
	DegradeConfig *DegradeConfig
	// season grpc
	SeasonGRPC *warden.ClientConfig
	// room grpc
	LiveGRPC *warden.ClientConfig
	// location grpc
	LocationGRPC *warden.ClientConfig
	// archive grpc
	ArchiveGRPC *warden.ClientConfig
	// videoup grpc
	VideoUpGRPC *warden.ClientConfig
	// acc grpc
	AccClient *warden.ClientConfig
	// channel grpc
	ChannelGRPC *warden.ClientConfig
	// 142 switch
	FrontPage *FrontPage
	//Cron
	Cron *cronConf
	// resource grpc
	ResourceGRPC    *warden.ClientConfig
	Resourcev2GRPC  *warden.ClientConfig
	BannerID        []int64
	BanResGRPCToken string
}

type cronConf struct {
	ReloadJob   string
	Notice      string
	LoadRes     string
	LoadVideoAd string
	CheckDiff   string
}

// Host defeine host info
type Host struct {
	Bangumi     string
	Ad          string
	AdDiscovery string
	Related     string
	APICo       string
}

// HTTPServers Http Servers
type HTTPServers struct {
	Outer *bm.ServerConfig
	Inner *bm.ServerConfig
	Local *bm.ServerConfig
}

// MySQL define MySQL config
type MySQL struct {
	Operation *sql.Config
	Ads       *sql.Config
	Res       *sql.Config
	Cpt       *sql.Config
}

// ReloadInterval define reolad config
type ReloadInterval struct {
	Jobs   time.Duration
	Notice time.Duration
	Ad     time.Duration
}

// RPCClient2 define RPC client config
type RPCClient2 struct {
	Archive  *rpc.ClientConfig
	Account  *rpc.ClientConfig
	Resource *rpc.ClientConfig
}

// DegradeConfig .
type DegradeConfig struct {
	Expire   int32
	Memcache *memcache.Config
}

type FrontPage struct {
	ETime  int64
	Name   string
	Pic    string
	LitPic string
	URL    string
}

func init() {
	flag.StringVar(&confPath, "conf", "", "config path")
}

func local() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

// Init int config
func Init() error {
	if confPath != "" {
		return local()
	}
	return remote()
}

func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	client.Watch("web-show.toml")
	// nolint:biligowordcheck
	go func() {
		for range client.Event() {
			xlog.Info("config reload")
			if load() != nil {
				xlog.Error("config reload error (%v)", err)
			}
		}
	}()
	return
}

func load() (err error) {
	var (
		s       string
		ok      bool
		tmpConf *Config
	)
	if s, ok = client.Toml2(); !ok {
		return errors.New("load config center error")
	}
	if _, err = toml.Decode(s, &tmpConf); err != nil {
		return errors.New("could not decode config")
	}
	*Conf = *tmpConf
	return
}
