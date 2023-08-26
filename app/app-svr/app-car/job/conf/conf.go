package conf

import (
	"errors"
	"flag"

	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/database/sql"
	xlog "go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/railgun"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	Conf     = &Config{}
	client   *conf.Client
)

type Config struct {
	// tracer
	Tracer *trace.Config
	// bm http
	BM *HTTPServers
	// httpPgcClient
	HTTPPGC *bm.ClientConfig
	// httpDuertv
	HTTPDuertv *bm.ClientConfig
	// host
	Host *Host
	// Custom
	Custom *Custom
	// Cron
	Cron *Cron
	// Duertv
	Duertv *Duertv
	// ArchiveGRPC grpc
	ArchiveGRPC     *warden.ClientConfig
	FlowControlGRPC *warden.ClientConfig
	// AccountGRPC grpc
	AccountGRPC *warden.ClientConfig
	// MySQL
	MySQL *MySQL
	// Redis
	Redis *Redis
	// ArchiveRailGun
	ArchiveRailGun *ArchiveRailGun
	// DuertvBangumiGun
	DuertvBangumiGun *DuertvBangumiGun
	// PGCRailGun
	PGCRailGun *PGCRailGun
	// RegionGun
	RegionGun *RegionGun
	// DuertvUGCGun
	DuertvUGCGun *DuertvUGCGun
	// FmSeasonGun
	FmSeasonGun *FmSeasonGun
	// httpData
	HTTPData *bm.ClientConfig
	// CreativeClient
	CreativeClient *warden.ClientConfig
	// FlowControl
	FlowControl *FlowControl
}

type FlowControl struct {
	BusinessID int
	Source     string
	Secret     string
}

type ArchiveRailGun struct {
	Cfg          *railgun.Config
	Databus      *railgun.DatabusV1Config
	SingleConfig *railgun.SingleConfig
}

type PGCRailGun struct {
	Cfg          *railgun.Config
	Databus      *railgun.DatabusV2Config
	SingleConfig *railgun.SingleConfig
}

type DuertvUGCGun struct {
	Cfg           *railgun.Config
	CronInputer   *railgun.CronInputerConfig
	CronProcessor *railgun.CronProcessorConfig
}

type DuertvBangumiGun struct {
	Cfg           *railgun.Config
	CronInputer   *railgun.CronInputerConfig
	CronProcessor *railgun.CronProcessorConfig
}

type RegionGun struct {
	Cfg           *railgun.Config
	CronInputer   *railgun.CronInputerConfig
	CronProcessor *railgun.CronProcessorConfig
}

type FmSeasonGun struct {
	Cfg          *railgun.Config
	KafkaCfg     *railgun.KafkaConfig
	SingleConfig *railgun.SingleConfig
}

// MySQL is
type MySQL struct {
	Show *sql.Config
	Car  *sql.Config
}

type Redis struct {
	Entrance   *redis.Config
	EntranceJd *redis.Config
}

// Host is
type Host struct {
	Bangumi string
	Duertv  string
	Data    string
}

// HTTPServers is
type HTTPServers struct {
	Inner *bm.ServerConfig
}

// Custom config
type Custom struct {
	PushBagumiAll       bool
	PushBagumiOffshelve bool
}

type Cron struct {
	PushBagumiCron string
	RegionCron     string
}

type Duertv struct {
	Key     string
	Partner string
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

func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	client.Watch("app-car-job.toml")
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
