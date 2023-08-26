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
	"go-common/library/log/infoc"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	xtime "go-common/library/time"

	infocV2 "go-common/library/log/infoc.v2"

	"github.com/BurntSushi/toml"
)

var (
	// confPath is.
	confPath string
	// Conf is.
	Conf = &Config{}
	// client is.
	client *conf.Client
)

// Config struct
type Config struct {
	IntlShowInfoc *InfocConf
	RedirectInfoc *infoc.Config
	CoinInfoc     *infoc.Config
	RelateInfocv2 *InfocConf
	ViewInfocv2   *InfocConf
	FeedInfocv2   *InfocConf
	// show  XLog
	XLog *log.Config
	// tick time
	Tick xtime.Duration
	// tracer
	Tracer *trace.Config
	// httpClinet
	HTTPClient *bm.ClientConfig
	// httpAsyn
	HTTPClientAsyn *bm.ClientConfig
	// httpData
	HTTPData *bm.ClientConfig
	// httpTag
	HTTPTag *bm.ClientConfig
	// httpBangumi
	HTTPBangumi *bm.ClientConfig
	// HTTPSearch
	HTTPSearch *bm.ClientConfig
	// HTTPAudio
	HTTPAudio *bm.ClientConfig
	// HTTPWrite
	HTTPWrite *bm.ClientConfig
	// http
	BM *HTTPServers
	// host
	Host *Host
	// http discovery
	HostDiscovery *HostDiscovery
	// db
	MySQL *MySQL
	// redis
	Redis *Redis
	// mc
	Memcache *Memcache
	// rpc client
	AccClient   *warden.ClientConfig
	CoinClient  *warden.ClientConfig
	RelationRPC *rpc.ClientConfig
	FavoriteRPC *rpc.ClientConfig
	AssistRPC   *rpc.ClientConfig
	ResourceRPC *rpc.ClientConfig
	ArticleRPC  *rpc.ClientConfig
	ActivityRPC *rpc.ClientConfig
	// BroadcastRPC grpc
	PGCRPC        *warden.ClientConfig
	ThumbupClient *warden.ClientConfig
	// ecode
	Ecode *ecode.Config
	// feed
	Feed *Feed
	// view
	View *View
	// search
	Search *Search
	// play icon
	PlayIcon *PlayIcon
	// RelationGRPC grpc
	RelationGRPC *warden.ClientConfig
	// grpc Archive
	ArchiveGRPC       *warden.ClientConfig
	ArticleGRPC       *warden.ClientConfig
	SteinClient       *warden.ClientConfig
	UpClient          *warden.ClientConfig
	PlayURLClient     *warden.ClientConfig
	AssistClient      *warden.ClientConfig
	LocationClient    *warden.ClientConfig
	HistoryGRPC       *warden.ClientConfig
	UGCSeasonClient   *warden.ClientConfig
	VideoupClient     *warden.ClientConfig
	DMClient          *warden.ClientConfig
	ChannelClient     *warden.ClientConfig
	ActivityClient    *warden.ClientConfig
	TagClient         *warden.ClientConfig
	FlowControlClient *warden.ClientConfig
	// Custom
	Custom *Custom
	// view config
	ViewConfig *ViewConfig
	// custom config
	Cfg *Cfg
	// tag config
	TagConfig *TagConfig
	// content.flow.control gRPC config
	CfcSvrConfig *CfcSvrConfig
	// view build limit
	ViewBuildLimit *ViewBuildLimit
	// search build limit
	SearchBuildLimit *SearchBuildLimit
	// build limit
	BuildLimit *BuildLimit
}

// CfcSvrConfig content.flow.control.service config
type CfcSvrConfig struct {
	BusinessID int64
	Secret     string // 由服务方下发
	Source     string
}

// HostDiscovery Http Discovery
type HostDiscovery struct {
	Data string
}

type InfocConf struct {
	LogID string
	Conf  *infocV2.Config
}

type BuildLimit struct {
	NewActiveTabIOS     int
	NewActiveTabAndroid int
	NewChannelIOS       int
	NewChannelAndroid   int
}

// Cfg def.
type Cfg struct {
	PgcSearchCard *PgcSearchCard
}

// PgcSearchCard def.
type PgcSearchCard struct {
	Epsize            int
	IpadEpSize        int
	IpadCheckMoreSize int
	OfflineWatch      string
	OnlineWatch       string
	CheckMoreContent  string
	CheckMoreSchema   string
	EpLabel           string
	// 宫格样式是否出角标
	GridBadge bool
}

// ViewConfig is
type ViewConfig struct {
	RelatesTitle      string
	AutoplayDesc      string
	AutoplayCountdown int
}

// Custom is
type Custom struct {
	SteinsSeasonBuild int
	HotAidsTick       xtime.Duration
	// fawkes tick
	FawkesTick       xtime.Duration
	Tick             xtime.Duration
	RecommendTimeout xtime.Duration
}

// HTTPServers Http Servers
type HTTPServers struct {
	Outer *bm.ServerConfig
}

// Host struct
type Host struct {
	Bangumi   string
	Data      string
	Hetongzi  string
	APICo     string
	Rank      string
	BigData   string
	Search    string
	AI        string
	Bvcvod    string
	Playurl   string
	PlayurlBk string
	Black     string
	Fawkes    string
}

// MySQL struct
type MySQL struct {
	Show *sql.Config
}

// Redis struct
type Redis struct {
	Feed *struct {
		*redis.Config
		ExpireRecommend xtime.Duration
		ExpireBlack     xtime.Duration
	}
}

// Memcache struct
type Memcache struct {
	Cache *struct {
		*memcache.Config
	}
}

// Feed struct
type Feed struct {
	// index
	Index *Index
	// ad
	CMResource map[string]int64
}

// Index struct
type Index struct {
	Count          int
	IPadCount      int
	MoePosition    int
	FollowPosition int
	// only archive for data disaster recovery
	Abnormal bool
}

// View struct
type View struct {
	// 相关推荐秒开个数
	RelateCnt int
}

// Search struct
type Search struct {
	SeasonNum          int
	MovieNum           int
	SeasonMore         int
	MovieMore          int
	UpUserNum          int
	UVLimit            int
	UserNum            int
	UserVideoLimit     int
	BiliUserNum        int
	BiliUserVideoLimit int
	OperationNum       int
	IPadSearchBangumi  int
	IPadSearchFt       int
}

// PlayIcon struct
type PlayIcon struct {
	STime int64
	ETime int64
	Tids  []int64
	URL1  string
	Hash1 string
	URL2  string
	Hash2 string
}

type TagConfig struct {
	OpenIcon bool
	ActIcon  string
	NewIcon  string
}

type ViewBuildLimit struct {
	SteinsSeasonBuildAndroid int
	SteinsSeasonBuildIOS     int
	ArcWithPlayerAndroid     int
	ArcWithPlayerIOS         int
}

type SearchBuildLimit struct {
	ArcWithPlayerAndroid int
	ArcWithPlayerIOS     int
}

// init is.
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

// local is.
func local() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

// reomte is.
// nolint:biligowordcheck
func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	client.Watch("app-intl.toml")
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

// load is.
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
