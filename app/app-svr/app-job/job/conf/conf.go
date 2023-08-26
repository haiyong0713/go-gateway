package conf

import (
	"errors"
	"flag"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/database/bfs"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-job/job/model/show"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	client   *conf.Client
	Conf     = &Config{}
)

type Config struct {
	WeChantUsers string
	WeChatToken  string
	WeChatSecret string
	// host
	Host *Host
	// interface XLog
	XLog            *log.Config
	ContributeSub   *databus.Config
	AISelectedSub   *databus.Config
	SelResBinlogSub *databus.Config
	ArchiveHonorPub *databus.Config
	AggregationSub  *databus.Config
	OTTSeriesPub    *databus.Config
	ResourceMngSub  *databus.Config
	// http
	BM *HTTPServers
	// httpClinet
	HTTPClient     *bm.ClientConfig
	HTTPClientAsyn *bm.ClientConfig
	// mc
	Memcache *Memcache
	// rpc client
	ArchiveRPC      *rpc.ClientConfig
	AccountRPC      *rpc.ClientConfig
	ArticleRPC      *rpc.ClientConfig
	FavClient       *warden.ClientConfig
	ShowClient      *warden.ClientConfig
	TagClient       *warden.ClientConfig
	CreativeClient  *warden.ClientConfig
	BroadCastClient *warden.ClientConfig
	// db
	MySQL *MySQL
	// redis
	Redis      *Redis
	Contribute *Contribute
	// up service
	UpArcClient *warden.ClientConfig
	ArtClient   *warden.ClientConfig
	// Wechat alert config for weekly selected in popular index
	WechatAlert *WechatAlert
	WeeklySel   *SerieCfg
	// databus
	CardDatabus *databus.Config
	// fawkes laser
	FawkesLaser bool
	Aggregation *Aggregation
	//bfs config
	BFS *bfs.Config
	// grpc Archive
	ArchiveGRPC *warden.ClientConfig
	// hot labels
	HotLabels *HotLabels
	// good history cfg
	GoodHis *GoodHisCfg
	Popular *Popular
	// Push
	Push      *Push
	Custom    *Custom
	Broadcast *Broadcast
}

// Broadcast
type Broadcast struct {
	ResourceToken string
}

// Custom
type Custom struct {
	TopActivityInterval int64
	SelectedTid         int64
	TagSwitchOn         bool
	FavSwitchOn         bool
	RefreshSwitchOn     bool
	RecommendTimeout    xtime.Duration
}

// aggregation def .
type Aggregation struct {
	Image string
}

// SerieCfg def.
type SerieCfg struct {
	NewSerieCron     string
	PublishCron      string
	RollBackRankCron string
	UpdateTime       xtime.Duration
	RecoveryNb       int
	Push             *SeriePush
	PlaylistMid      int64
	MaxNumber        int    // 兜底状态最大卡片数
	HonorLink        string // 稿件荣誉链接
	HonorLinkV2      string // 支持native页的稿件荣誉链接
	RankId           int
	RankIndex        int
	MaxSerieNumber   int64
	MinSerieNumber   int64
}

// GoodHisCfg def
type GoodHisCfg struct {
	FID int64 // 收藏夹id
	MID int64
	URL string
}

// SeriePush is push cfg for serie
type SeriePush struct {
	Token      string
	BusinessID string
	Title      string
	Link       string
}

// WechatAlert cfg.
type WechatAlert struct {
	Host        string
	Key         string
	Secret      string
	AI          *show.Merak
	Audit       *show.Merak
	Aggregation *show.Merak
}

// HTTPServers Http Servers
type HTTPServers struct {
	Inner *bm.ServerConfig
}

type Host struct {
	APP      string
	Config   string
	Hetongzi string
	APICo    string
	VC       string
	Fawkes   string
	Manga    string
	Manager  string
	Data     string
	Bap      string
}

type Memcache struct {
	Feed *struct {
		ExpireCache xtime.Duration // cache expiration for disaster recovery
	}
	Cache *struct {
		*memcache.Config
	}
	Cards *struct {
		*memcache.Config
		ExpireAggregation xtime.Duration
	}
	Aggregation *struct {
		*memcache.Config
		ExpireCache xtime.Duration
	}
}

type MySQL struct {
	Show    *sql.Config
	Manager *sql.Config
}

type Redis struct {
	Feed *struct {
		*redis.Config
	}
	Contribute *struct {
		*redis.Config
		ExpireContribute xtime.Duration
	}
	Interface *struct {
		*redis.Config
		ExpireContribute xtime.Duration
	}
	Entrance *struct {
		*redis.Config
	}
	Recommend *struct {
		*redis.Config
		ExpireRank xtime.Duration
	}
	DynamicSchool *struct {
		*redis.Config
	}
}

type Contribute struct {
	Cluster bool
}

type HotLabels struct {
	IsDiff   bool
	Bucket   string
	Dir      string
	TopLeft  *WaterMark
	TopRight *WaterMark
	Bottom   *WaterMark
}

type WaterMark struct {
	Suffix         string
	WMKey          string
	WMPaddingX     uint32
	WMPaddingY     uint32
	WMScale        float64
	WMPos          string
	WMTransparency float64
}

type Popular struct {
	PopularCardCron string
}

type Push struct {
	FawkesLaser *struct {
		AppID      int64
		BusinessID int64
		LinkType   int64
		Token      string
	}
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

// nolint:biligowordcheck
func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	client.Watch("app-job.toml")
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
