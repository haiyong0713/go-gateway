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
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"

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
	Redis    *Redis
	Memcache *memcache.Config
	Ecode    *ecode.Config
	// HTTP client
	HTTPClient *bm.ClientConfig
	// mysql
	Mysql *sql.Config
	// databus
	DanceBinlogSub *databus.Config
	Cfg            *Cfg
}

type Cfg struct {
	DefaultScore  int        // 默认分数
	StatCron      string     // 结算周期
	StatCurrency  int        // 同时参与结算的最多goroutine
	StatDelay     int64      // 延迟多久进行结算
	MaxScore      float64    // 最大分值
	Deviation     *Deviation // 评分标准
	Normalization bool       // 是否需要归一化，拉开差距
	Boundary      int64      // stat计算边界
	Score         *Score     // 评分配置
}

type Deviation struct {
	Perfect float64
	Super   float64
	Good    float64
	Bad     float64
}

type Score struct {
	Perfect int64
	Super   int64
	Good    int64
	Bad     int64
	Miss    int64
}

type Redis struct {
	*redis.Config
	GameExp xtime.Duration
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
	client.Watch("dance-taiko-job.toml")
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
