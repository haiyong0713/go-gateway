package conf

import (
	"errors"
	"flag"

	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	Conf     = &Config{}
	client   *conf.Client
)

type Config struct {
	// Env
	Env string
	// show  XLog
	Log *log.Config
	// tick time
	Tick xtime.Duration
	//quicker time
	QuickerTick xtime.Duration
	// tracer
	Tracer *trace.Config
	// xlog
	Xlog *log.Config
	// http
	BM *bm.ServerConfig

	// Redis
	Redis *Redis

	// BroadcastOnlineRPC grpc
	BroadcastOnlineRPC *warden.ClientConfig

	//在看人数，文案与场景控制
	Online *Online
	//左下角常驻，灰度控制
	OnlineBottom *OnlineBottom
	//特殊弹幕，灰度控制
	OnlineSpecialDM *OnlineSpecialDM

	// 自定义配置
	Custom *Custom

	// grpc
	PgcRPC     *warden.ClientConfig
	ArchiveRpc *warden.ClientConfig
}

type OnlineSpecialDM struct {
	SwitchOn bool
	Gray     int64
	Mid      map[string]int64
	//特殊弹幕展示次数
	DmCount int64
	//特殊弹幕出现，在线人数阈值
	ShowCount int64
	//弹幕文案
	Text string
}

type OnlineBottom struct {
	SwitchOn bool
	Gray     int64
	Mid      map[string]int64
}

type Online struct {
	//在看文案
	Text string
	//客户端下次请求的时间间隔
	SecNext int64
	//功能开关
	SwitchOn bool
}

type Custom struct {
	NoLoginAvatarAll bool
	NoLoginAvatar    map[string]struct {
		URL  string
		Type int
	}
	TopActivityInterval      int64
	TopActivityMngSwitch     bool
	TopActivityFissionSwitch bool
	PopUpAutoClose           bool
	PopUpAutoCloseTime       int
}

type Redis struct {
	Online *redis.Config
}

func init() {
	flag.StringVar(&confPath, "conf", "", "config path")
}

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
	client.Watch("player-online.toml")
	//nolint:biligowordcheck
	go func() {
		for range client.Event() {
			log.Info("config reload")
			if load() != nil {
				log.Error("config reload error(%v)", err)
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
