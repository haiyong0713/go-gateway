package conf

import (
	"errors"
	"flag"

	"go-common/library/conf"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	Conf     = &Config{}
	client   *conf.Client
)

type Config struct {
	// show  XLog
	Log *log.Config
	// bm http
	HTTPServers *HTTPServers
	// db
	DB *DB
	// 交互开关
	Degrade *Degrade
}

type Degrade struct {
	// feature清晰度3期-所有清晰度列表
	FeatureQnMap map[string]int8
	// 交互配置
	Cfg *DegradeCfg
}
type DegradeCfg struct {
	Cron              string
	DefaultDecode     int64
	DefaultEnlarge    float32
	ChDefaults        []*DegradeDefault
	DefaultLogLevel   int32
	DefaultAutoLaunch int32
}

type DegradeDefault struct {
	Code       string
	DecodeType int64
}

// BM http
type HTTPServers struct {
	Inner *bm.ServerConfig
}

// DB define MySQL config
type DB struct {
	Feature *xsql.Config
	TV      *xsql.Config
}

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
}

// Init init config.
func Init() (err error) {
	if confPath != "" {
		_, err = toml.DecodeFile(confPath, &Conf)
		return
	}
	err = remote()
	return
}

// nolint:biligowordcheck
func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	client.Watch("feature-service.toml")
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
