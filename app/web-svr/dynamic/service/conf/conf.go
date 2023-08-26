package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-common/library/time"

	"github.com/BurntSushi/toml"
)

// global var
var (
	// // confPath string
	// client *conf.Client
	// Conf config
	Conf = &Config{}
)

// Config .
type Config struct {
	// elk
	Log *log.Config
	// Verify
	Verify *verify.Config
	// app
	App *bm.App
	// http client
	HTTPClient httpClient
	// rpc server
	RPCServer *rpc.ServerConfig
	// grpc
	ArcClient    *warden.ClientConfig
	SeasonClient *warden.ClientConfig
	CfcGRPC      *warden.ClientConfig
	// tracer
	Tracer *trace.Config
	// mc
	Memcache *Memcache
	// redis
	Redis *Redis
	// Rule
	Rule *Rule
	// Host
	Host *Host
	// databus
	ArchiveNotifySub *databus.Config
	// db
	DB *DB
	// http
	BM          *BM
	GRPC        *warden.ServerConfig
	LandingPage map[string]*landingPage
	Cron        *cron
	// region filter
	FilterRids []int32
	// content.flow.control.service gRPC config
	CfcSvrConfig *CfcSvrConfig
	// ArchiveFlowControl databus
	ArcFlowControlCfg     *railgun.SingleConfig
	ArchiveFlowControlSub *databus.Config
}

type landingPage struct {
	TagID []int64
	Rid   []int64
}

type cron struct {
	LoadBusinessRegion string
}

// BM http .
type BM struct {
	Inner *bm.ServerConfig
	Local *bm.ServerConfig
}

// DB .
type DB struct {
	ArcResult *sql.Config
}

// Redis .
type Redis struct {
	Archive *struct {
		*redis.Config
	}
}

// Host hosts.
type Host struct {
	BigDataURI string
	// TODO del
	LpBigDataURI string
	LiveURI      string
	APIURI       string
}

// Rule config.
type Rule struct {
	// region tick.
	TickRegion time.Duration
	// tag tick.
	TickTag time.Duration
	// default num of dynamic archives.
	NumArcs int
	// default num of index dynamic archives.
	NumIndexArcs int
	// min region count.
	MinRegionCount int
	WeChatToken    string
	WeChatSecret   string
	WeChantUsers   string
	WeChanURI      string
	AddArcNum      int
	PermInit       []string
	InitArc        time.Duration
	// init region redis from db
	InitRegStart int
	InitRegEnd   int
}

type httpClient struct {
	Read *bm.ClientConfig
}

// Memcache memcache .
type Memcache struct {
	*memcache.Config
	Expire time.Duration
}

type CfcSvrConfig struct {
	BusinessID int64
	Secret     string // 由服务方下发
	Source     string
}

func Init() (err error) {
	err = paladin.Init()
	if err != nil {
		return
	}
	return remote()
}

func remote() (err error) {
	if err = load(); err != nil {
		return err
	}
	err = paladin.Watch("dynamic-service.toml", Conf)
	if err != nil {
		return err
	}

	return
}

func load() (err error) {
	err = paladin.Get("dynamic-service.toml").UnmarshalTOML(Conf)
	if err != nil {
		return
	}
	return
}

func Close() {
	paladin.Close()
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	log.Info("dynamic-service-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}
