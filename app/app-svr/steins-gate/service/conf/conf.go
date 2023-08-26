package conf

import (
	"flag"

	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/log/infoc"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

var (
	confPath string
	Conf     = &Config{}
	client   *conf.Client
)

type Config struct {
	MySQL             *MySQL
	XLog              *log.Config
	Host              *Host
	Redis             *Redis
	VideoClient       *bm.ClientConfig
	WechatClient      *bm.ClientConfig
	Wechat            *Wechat
	Bvc               *Bvc
	Rule              *Rule
	Custom            *Custom
	Interval          *Interval
	ArcInteractivePub *databus.Config
	SteinsGate        *databus.Config
	Node              *infoc.Config
	Mark              *infoc.Config
	Server            *bm.ServerConfig
	Playurl           *warden.ClientConfig
	Filter            *warden.ClientConfig
	Archive           *warden.ClientConfig
	Account           *warden.ClientConfig
	OgvPay            *warden.ClientConfig
	DefaultSkin       *api.Skin
	Quota             *quota.Config
}

// Interval .
type Interval struct {
	SkinInterval xtime.Duration
}

type Custom struct {
}

type Rule struct {
	GraphMids  []int64
	PrivateMsg *PrivateMsg
	ToastMsg   *ToastMsg
}

type PrivateMsg struct {
	MC            string
	PassTitle     string
	PassContent   string
	RejectTitle   string
	RejectContent string
}

type ToastMsg struct {
	NeedLogin   string
	GraphUpdate string
}

type Bvc struct {
	Key string
}

type Wechat struct {
	Wxkey   string
	WxTitle string
	WxUser  string
}

type Host struct {
	Videoup string
	Bvc     string
	Merak   string
	Bfs     string
}

// MySQL def.
type MySQL struct {
	Steinsgate *sql.Config
}

type Redis struct {
	Graph              *redis.Config
	RecordExpiration   xtime.Duration
	HvarExpirationMinH xtime.Duration
	HvarExpirationMaxH xtime.Duration
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
	client.Watch("steins-gate-service.toml")
	//nolint:biligowordcheck
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
