package conf

import (
	"errors"
	"flag"

	"github.com/BurntSushi/toml"
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/database/sql"
	"go-common/library/log"
	infoc2 "go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/time"
)

var (
	confPath string
	client   *conf.Client
	// Conf conf
	Conf = &Config{}
)

// Config config
type Config struct {
	Rule   *Rule
	XLog   *log.Config
	Tracer *trace.Config
	Auth   *auth.Config
	Verify *verify.Config
	// mysql
	Mysql *sql.Config
	// bm
	BM *bm.ServerConfig
	// redis
	Redis *Redis

	TaskPub      *databus.Config
	BuvidTaskPub *databus.Config

	// GRPC Server
	GRPC *warden.ServerConfig
	// 泰山RPC
	TaishanRPC *warden.ClientConfig
	// localcache
	Localcache *Localcache

	Taishan *Taishan

	InfocLogStream *infoc2.Config
}

type Taishan struct {
	Table string
	Token string
}

// Rule .
type Rule struct {
	DocLimit int
}

// Localcache cache stored in local
type Localcache struct {
	Max        int
	BucketSize int
}

// KvoMemcache memcache config
type KvoMemcache struct {
	Kvo    *memcache.Config
	Expire time.Duration
}

type Redis struct {
	Redis       *redis.Config
	Expire      time.Duration
	IncrExpire  time.Duration
	UcDocExpire time.Duration
}

func init() {
	flag.StringVar(&confPath, "conf", "", "config path")
}

// Init init conf
func Init() (err error) {
	if confPath != "" {
		_, err = toml.DecodeFile(confPath, &Conf)
		return
	}
	err = remote()
	return
}

func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
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
