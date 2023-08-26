package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	xlog "go-common/library/log"
	infocv2 "go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	// Env
	Env string
	// show log
	ShowLog string
	// show  XLog
	XLog *xlog.Config
	// tracer
	Tracer *trace.Config
	// httpClinet
	HTTPClient *bm.ClientConfig
	// httpClinetAsyn
	HTTPClientAsyn *bm.ClientConfig
	// httpData
	HTTPData *bm.ClientConfig
	// httpHotData
	HTTPHotData *bm.ClientConfig
	// httpdynamic
	HTTPDynamic  *bm.ClientConfig
	HTTPBusiness *bm.ClientConfig
	HTTPGameCo   *bm.ClientConfig
	HTTPShowCo   *bm.ClientConfig
	HTTPMangaCo  *bm.ClientConfig
	// bm http
	BM        *HTTPServers
	RpcServer *warden.ServerConfig
	// host
	Host *Host
	// db
	MySQL *MySQL
	// redis
	Redis *Redis
	// mc
	Memcache *Memcache
	// rpc client
	ArchiveRPC *rpc.ClientConfig
	// dynamicRPC client
	DynamicRPC *rpc.ClientConfig
	// resource
	ResourceRPC *rpc.ClientConfig
	// relationRPC
	RelationRPC *rpc.ClientConfig
	// rec host
	Recommend *Recommend
	// Infoc2
	Infocv2        *Infocv2
	FeedInfocv2    *Infocv2
	FeedTabInfocv2 *Infocv2
	// databus
	DislikeDataBus *databus.Config
	// duration
	Duration *Duration
	// BroadcastRPC grpc
	PGCRPC  *warden.ClientConfig
	FlowRPC *warden.ClientConfig
	// show hot config
	ShowHotConfig *ShowHotConfig
	// show selected cfg
	ShowSelectedCfg *ShowSelectedCfg
	// CustomTick is
	CustomTick *CustomTick
	// AccountGRPC grpc
	AccountGRPC *warden.ClientConfig
	// RelationGRPC grpc
	RelationGRPC  *warden.ClientConfig
	PgcFollowGRPC *warden.ClientConfig
	PgcAppGRPC    *warden.ClientConfig
	// grpc Archive
	ArchiveGRPC *warden.ClientConfig
	ArticleGRPC *warden.ClientConfig
	// grpc Dynamice
	DynamicGRPC *warden.ClientConfig
	// Favorite GRPC
	FavoriteGRPC   *warden.ClientConfig
	HmtChannelGRPC *warden.ClientConfig
	// Activity GRPC
	ActivityGRPC *warden.ClientConfig
	PlatGRPC     *warden.ClientConfig
	LocationGRPC *warden.ClientConfig
	TagGRPC      *warden.ClientConfig
	DynGRPC      *warden.ClientConfig
	EsportsGRPC  *warden.ClientConfig
	// media GRPC
	CharGRPC *warden.ClientConfig
	// PreciousInfo .
	PreciousInfo *PreciousInfo
	Aggregation  *Aggregation
	// PreciousFeed ABTest white list
	PreciousFeed *PreciousFeed
	// Channel GRPC
	ChannelGRPC *warden.ClientConfig
	PopularGRPC *warden.ClientConfig
	// Live GRPC
	LiveGRPC     *warden.ClientConfig
	LiveFeedGRPC *warden.ClientConfig
	RoomGateGRPC *warden.ClientConfig
	Share        map[string]bool
	LiveShare    map[string]bool
	// Custom
	Custom *Custom
	// BuildLimit
	BuildLimit *BuildLimit
	// intl
	Intl *Intl
	// click tip
	ClickSpecialTip *ClickSpecialTip
	// feature平台
	Feature *Feature
	// infoc
	NaInfoc *NaInfoc
	// GRPC
	DynvoteGRPC *warden.ClientConfig
	ScoreGRPC   *warden.ClientConfig
	BGroupGRPC  *warden.ClientConfig
	// s11
	S11Cfg *S11Cfg
	// 冬奥
	WinterOlyMedal *WinterOlyMedal
	WinterOlyEvent *WinterOlyEvent
}

type WinterOlyMedal struct {
	TitleColor       string
	HeaderBgColor    string
	DefaultBgColor   string
	IntervalBgColor  string
	SpecialBgColor   string
	SpecialFontColor string
	RankColor        map[string]string
}

type WinterOlyEvent struct {
	TitleColor         string
	TitleBgColor       string
	DefaultStatusColor string
	RunningStatusColor string
}

type S11Cfg struct {
	PageID int64
}

type Infocv2 struct {
	LogID string
	Infoc *infocv2.Config
}

type NaInfoc struct {
	PageViewLogID   string
	ModuleViewLogID string
	Infoc           *infocv2.Config
}

type ClickSpecialTip struct {
	Sid       map[string]int
	Msg       string
	SureMsg   string
	ThinkMsg  string
	CancelMsg string
}

// Aggregation .
type Aggregation struct {
	Icon   string
	IconV2 string
}

// PreciousInfo town station treasure .
type PreciousInfo struct {
	SubTagID int64
}

// CustomTick is
type CustomTick struct {
	Rank xtime.Duration
	Tick xtime.Duration
}

// Duration is
type Duration struct {
	// splash
	Splash string
	// search time_from
	Search       string
	SearchDay    int
	PGCSearchDay map[string]int
}

type Custom struct {
	LiveRcmdClose     bool
	HotContinuousPlay int
	HotContinuousGray int64
	AIHotAbnormal     bool
	AIHotMid          map[string]int
	AIGroupMid        map[string]int
	AIGroupBuvid      map[string]int
	Hit               int64
	SelectedTid       int64
	TagSwitchOn       bool
	FavSwitchOn       bool
	RecommendTimeout  xtime.Duration
	ShareIcon         string
	PlayerArgs        bool
	PartImage         string
	PartUnImage       string
	//广告资源位
	PopularAdResourceIOS     string
	PopularAdResourceAndroid string
	FlowSecret               string
}

// Host is
type Host struct {
	ApiLiveCo    string
	Bangumi      string
	Hetongzi     string
	HetongziRank string
	Data         string
	ApiCo        string
	ApiCoX       string
	Ad           string
	Search       string
	Activity     string
	Dynamic      string
	Black        string
	WWW          string
	Business     string
	GameCo       string
	ShowCo       string
	MangaCo      string
}

// HTTPServers is
type HTTPServers struct {
	Outer *bm.ServerConfig
}

// MySQL is
type MySQL struct {
	Show     *sql.Config
	Resource *sql.Config
}

// Redis is
type Redis struct {
	Recommend *struct {
		*redis.Config
		Expire      xtime.Duration
		ExpireSerie xtime.Duration
	}
	Entrance *redis.Config
}

// Memcache is
type Memcache struct {
	Cards *struct {
		*memcache.Config
		Expire            xtime.Duration
		ExpireAggregation xtime.Duration
	}
}

// Recommend is
type Recommend struct {
	Host  map[string][]string
	Group map[string]int
}

// ShowHotConfig is
type ShowHotConfig struct {
	ItemTitle       string
	BottomText      string
	BottomTextCover string
	BottomTextURL   string
	ShareDesc       string
	ShareTitle      string
	ShareSubTitle   string
	ShareIcon       string
}

// ShowSelectedCfg def.
type ShowSelectedCfg struct {
	Reminder    string
	MediaMID    int64 // 播单mid
	DisasterMax int   // disaster recovery mode max number of cards}
}

type H5Entrance struct {
	Icon     string
	Title    string
	ModuleID string
	URI      string
}

type PreciousFeed struct {
	WhiteListMid   []int64
	WhiteListBuvid []string
}

type BuildLimit struct {
	StarsSingleAndroid     int
	StarsSingleIOS         int
	HotCardOptimizeAndroid int
	HotCardOptimizeIPhone  int
	HotCardOptimizeIPad    int
	SVideoAndroid          int
	SVideoIOS              int
}

type Intl struct {
	MidControl []int64
}

type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}

type FeatureBuildLimit struct {
	Switch               bool
	InlineLow            string
	NewFeed              string
	SelectLow            string
	Version615           string
	Version618           string
	Miaokai              string
	MiaokaiAndroid       string
	ClickArea            string
	DynamicCard549       string
	ResourceSmallCard557 string
	CommonPage           string
	BgmBanners           string
	RegionX              string // del
	BangumiType1         string // del
	BangumiTypeTv        string // del
	BangumiTypeX         string
	ResBanners           string // del
	PickEntrances        string
	StarCard             string // del
	SVideo               string // del
	LiveTabParticipation string
	PartiUseNewTopic     string
	CompatIosHover       string
	CompatIpadBPU2021    string
	LiveWatched          string
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	log.Info("progress-service-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}
