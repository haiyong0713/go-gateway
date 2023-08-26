package conf

import (
	"errors"
	"flag"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/railgun"
	"go-common/library/time"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	client   *conf.Client
	// Conf config
	Conf = &Config{}
)

// Config .
type Config struct {
	Log      *log.Config
	BM       *bm.ServerConfig
	Tracer   *trace.Config
	Redis    *redis.Config
	Memcache *memcache.Config
	Ecode    *ecode.Config
	// rpc
	DynamicRPC *rpc.ClientConfig
	// groc
	ArchiveGRPC *warden.ClientConfig
	// Host
	Host Host
	// HTTP client
	HTTPClient   *bm.ClientConfig
	XiaomiClient *bm.ClientConfig
	// Rule
	Rule *Rule
	// App
	App *bm.App
	// databus
	ArchiveNotifySub     *railgun.DatabusV1Config
	ArchiveNotifyRailgun *railgun.SingleConfig
	OutArcSub            *railgun.DatabusV1Config
	OutArcRailgun        *railgun.SingleConfig
	// mysql
	Mysql *sql.Config
	// bfs
	BFS *bfs
	// cron
	Cron *cron
	// xiaomi
	Xiaomi xiaomi
}

type xiaomi struct {
	Appkey string
	AppID  int64
}

type cron struct {
	LoadArcTypes     string
	LoadChangeOurArc string
}

type bfs struct {
	Bucket string
	Dir    string
	AppKey string
	Secret string
}

// Rule .
type Rule struct {
	BroadFeed       int
	ReadTimeout     time.Duration
	PushArcBfsURL   string
	PushArcFileName string
	DelArcBfsURL    string
	DelArcFileName  string
}

// Host remote host
type Host struct {
	API    string
	Xiaomi string
}

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
}

// Init init conf
func Init() error {
	if confPath != "" {
		return local()
	}
	return remote()
}

func local() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	// nolint:biligowordcheck
	go func() {
		for range client.Event() {
			log.Info("config reload")
			if load() != nil {
				log.Error("config reload error (%v)", err)
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
