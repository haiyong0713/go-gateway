package conf

import (
	"errors"
	"flag"

	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/railgun"

	"github.com/BurntSushi/toml"
)

// is
var (
	confPath string
	client   *conf.Client
	Conf     = &Config{}
)

// Config is
type Config struct {
	// interface XLog
	XLog *log.Config
	// http
	BM *bm.ServerConfig
	// redis
	MixRedis *redis.Config
	// Custom 自定义启动参数
	ArcControlRailGun *ArcControlRailGun
	Custom            *Custom
	Broadcast         *warden.ClientConfig
}

type ArcControlRailGun struct {
	Cfg          *railgun.Config
	Databus      *railgun.DatabusV1Config
	SingleConfig *railgun.SingleConfig
}

// Custom is
type Custom struct {
	SmoothThreshold int64
	ExTime          int64
}

func init() {
	flag.StringVar(&confPath, "conf", "", "config path")
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
	client.Watch("app-player-job.toml")
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
