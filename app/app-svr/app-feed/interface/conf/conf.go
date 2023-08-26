package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/log/infoc"
	infocV2 "go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/sample"

	"github.com/BurntSushi/toml"
)

// Config is
type Config struct {
	// Custom config
	Custom *Custom
	// infoc log2
	RedirectInfoc2 *infoc.Config
	// show  XLog
	XLog *log.Config
	// tracer
	Tracer *trace.Config
	// httpClinet
	HTTPClient *bm.ClientConfig
	// httpAsyn
	HTTPClientAsyn *bm.ClientConfig
	// httpData
	HTTPData *bm.ClientConfig
	// httpData
	HTTPDataAd *bm.ClientConfig
	// httpTag
	HTTPTag *bm.ClientConfig
	// httpAd
	HTTPAd *bm.ClientConfig
	// httpActivity
	HTTPActivity *bm.ClientConfig
	// httpBangumi
	HTTPBangumi *bm.ClientConfig
	// httpShow
	HTTPShow *bm.ClientConfig
	// httpDynamic
	HTTPDynamic *bm.ClientConfig
	// httpClinet
	HTTPSearch *bm.ClientConfig
	// httpNg
	HTTPNg   *bm.ClientConfig
	HTTPGame *bm.ClientConfig
	// http
	BM *HTTPServers
	// host
	Host *Host
	// redis
	Redis *Redis
	// mc
	Memcache *Memcache
	// rpc client
	ResourceRPC *rpc.ClientConfig
	CoinClient  *warden.ClientConfig
	// databus
	DislikeDatabus *databus.Config
	// ecode
	Ecode *ecode.Config
	// feed
	Feed *Feed
	// bnj2018
	Bnj *BnjConfig
	// BroadcastRPC grpc
	PGCRPC *warden.ClientConfig
	// pgc inline grpc
	PGCInline *warden.ClientConfig
	// AccountGRPC grpc
	AccountGRPC *warden.ClientConfig
	// ActivityGRPC grpc
	ActivityClient *warden.ClientConfig
	// RelationGRPC grpc
	RelationGRPC *warden.ClientConfig
	// grpc Archive
	ArchiveGRPC *warden.ClientConfig
	// grpc vip
	VipGRPC     *warden.ClientConfig
	ThumbupGRPC *warden.ClientConfig
	// grpc article
	ArticleGRPC *warden.ClientConfig
	// grpc location
	LocationGRPC *warden.ClientConfig
	// grpc tunnel
	TunnelGRPC *warden.ClientConfig
	// grpc FavClient
	FavClient *warden.ClientConfig
	// grpc live
	LiveGRPC     *warden.ClientConfig
	LiveRankGRPC *warden.ClientConfig
	// grpc tag
	TagClient *warden.ClientConfig
	// grpc pgc app card
	PgcClient *warden.ClientConfig
	// grpc pgc card card
	PgcCardClient    *warden.ClientConfig
	PgcStoryClient   *warden.ClientConfig
	PgcFollowClient  *warden.ClientConfig
	DynamicGRPC      *warden.ClientConfig
	FeedRPC          *warden.ClientConfig
	VideoOpenClient  *warden.ClientConfig
	TopicClient      *warden.ClientConfig
	MaterialClient   *warden.ClientConfig
	OpenCourseClient *warden.ClientConfig
	DeliveryClient   *warden.ClientConfig
	// databus
	CardDatabus          *databus.Config
	CardAdFeedDatabus    *databus.Config
	SessionRecordDatabus *databus.Config
	// grpc Channel
	ChannelGRPC *warden.ClientConfig
	// build limit
	BuildLimit *BuildLimit
	// mogul config
	MogulDatabus *databus.Config
	// http discovery
	HostDiscovery *HostDiscovery
	// grpc resource
	ResourceGRPC *warden.ClientConfig
	// grpc resource v2
	ResourceV2GRPC *warden.ClientConfig
	// app-resource client
	AppResourceClient *warden.ClientConfig
	TunnelV2Client    *warden.ClientConfig
	// sample
	SampleConfig *sample.Config
	// ng switch
	NgSwitch *NgSwitch
	// Cron
	Cron           *Cron
	FeatureControl *FeatureControl
	V9Custom       *SmallCoverV9Custom
	// feature平台
	Feature *Feature
	// databus collect
	Databus *Databus
	InfocV2 *infocV2.Config
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

type Databus struct {
	Group string
	AppID string
	Token string
	Topic string
}

type SmallCoverV9Custom struct {
	LeftBottomBadgeKey   string
	LeftBottomBadgeStyle map[string]*operate.LiveBottomBadge
	LeftCoverBadgeStyle  []*operate.V9LiveLeftCoverBadge
}

type ExpMidGroup struct {
	Sharding []string
}

type NgDispatchSwitch struct {
	DisableAll bool // 总开关
	Sharding   []string
}

type NgSwitch struct {
	DisableAll   bool // 总开关
	CardSharding map[string][]string
}

type FeatureControl struct {
	DisableAll bool // 总开关
	Feature    map[string][]string
}

// Custom config
type Custom struct {
	AutoPlayMids              []int64
	Prefer4GAutoPlay          bool
	SingleAutoPlayForce       int8
	TransferSwitch            bool
	TransferGroup             uint32
	BannerAbtestTime          int64
	NewDoubleMids             []int64
	GifSwitch                 int // 0 运营gif优先、1 广告gif优先
	NewAdAbtest               int
	DaLaoMids                 []int64
	WhiteHideBannerAd         map[string]int
	FeedLiveMids              map[string]int
	FeedLiveSwitch            bool
	AutoRefreshTime           xtime.Duration
	AIBannerGroupMid          map[string]int // ai banner 分组实验
	AIBannerGroupBuvid        map[string]int // ai banner 分组实验
	AIBannerGroupAll          bool           // ai banner 分组实验全量
	GuidenceBuvid             map[string]int
	AIAdMid                   map[string]int // ai 广告白名单
	AIAdGroupMid              map[string]int // ai 广告 分组实验mid
	AIAdGroupBuvid            map[string]int // ai 广告 分组实验buvid
	StoryThreePoint           bool           // story三点不感兴趣开关
	RecommendTimeout          xtime.Duration
	GotoStoryDislikeReason    string // story不感兴趣配置字段
	StoryUISwitch             bool
	SingleInlineAutoPlay      int8
	Resetting                 *Resetting
	DoubleInlineLike          int
	SingleInlineLike          int
	AndroidBAutoplaySwitch    bool
	AndroidBAutoplayTimestamp int64
	AllowGameBadge            bool
	PrivacyModeAid            []int64
	ResourceDegradeSwitch     bool
	ClassBadgeGroup           int
	DisableLikeStat           bool
	PGCMaterialDegradeSwitch  bool
}

type Resetting struct {
	ColumnOpen        bool
	AutoplayOpen      bool
	ColumnTimestamp   int64
	AutoplayTimestamp int64
	Column            int64
	Autoplay          int64
}

type Cron struct {
	LoadCache     string
	LoadLiveCache string
}

// HTTPServers Http Servers
type HTTPServers struct {
	Outer *bm.ServerConfig
}

// BnjConfig 2018拜年祭配置
type BnjConfig struct {
	TabImg    string
	TabID     int64
	BeginTime string
}

type Host struct {
	LiveAPI    string
	Bangumi    string
	Data       string
	Hetongzi   string
	APICo      string
	Ad         string
	Activity   string
	Rank       string
	Show       string
	Dynamic    string
	DynamicCo  string
	BigData    string
	Search     string
	Fawkes     string
	Black      string
	GameCenter string
}

// HostDiscovery Http Discovery
type HostDiscovery struct {
	Ad   string
	Data string
	Live string
}

type Redis struct {
	Feed *struct {
		*redis.Config
		ExpireBlack xtime.Duration
	}
	Upper *struct {
		*redis.Config
		ExpireUpper xtime.Duration
	}
}

type Memcache struct {
	Cache *struct {
		*memcache.Config
		ExpireCache xtime.Duration
	}
}

type Feed struct {
	// feed
	FeedCacheCount int
	LiveFeedCount  int
	// index
	Index *Index
	// inline相关配置
	Inline *Inline
	// story相关配置
	StoryIcon   map[string]*model.GotoIcon
	FeedInfocID string
}

type Inline struct {
	ShowInlineDanmaku   int
	LikeButtonShowCount bool
	// 点赞按钮资源
	LikeResource             string
	LikeResourceHash         string
	DisLikeResource          string
	DisLikeResourceHash      string
	LikeNightResource        string
	LikeNightResourceHash    string
	DisLikeNightResource     string
	DisLikeNightResourceHash string
	IconDrag                 string
	IconDragHash             string
	IconStop                 string
	IconStopHash             string
	ThreePointPanelType      int
}

type Index struct {
	Count          int
	IPadCount      int
	MoePosition    int
	FollowPosition int
	// only archive for data disaster recovery
	Abnormal    bool
	Interest    []string
	FollowMode  *feed.FollowMode
	NewInterest *Interest
	// teenagers special_s
	TeenagersSpecialCard   *TeenagersSpecialCard
	IpadHDThreeColumnCount int64
}

type BuildLimit struct {
	NewActiveTabIOS     int
	NewActiveTabAndroid int
	NewChannelIOS       int
	NewChannelAndroid   int
}

type Interest struct {
	TitleHide string
	DescHide  string
	TitleShow string
	DescShow  string
	Message   string
}

type TeenagersSpecialCard struct {
	ID    int64
	Title string
	Cover string
	URL   string
}

type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}

type FeatureBuildLimit struct {
	Switch                   bool
	FeedTunnel               string
	BanBannerHash            string // del
	DislikeReasonName        string // del
	Index2AbtestNewBanner    string
	Index2AbtestRcmdReason   string
	Index2AbtestRcmdReasonV2 string
	Index2AbtestStoryTP      string
	AdResource               string // del
	CanEnable4GWiFiAutoPlay  string
	AutoplayCard             string // del
	HomeTransferTest         string
	IndexConfigColumn        string
	StoryAids                string
	PgcRemind                string // del
	PicSwitchStyle           string
	PicIsNewChannel          string
	OnePicV3                 string
	OnePicV2                 string
	ThreePicV3               string
	LiveV9Custom             string
	InlineAV2                string
	Storyamplayer            string // del
	SwitchCooperation        string // del
	SpecialSwitchStyle       string
	FeedChannelDetailm       string // del
	TagEp                    string // del
	TagPic                   string // del
	TagThreePoint            string // del
	TagNewBanner             string
	UpperFeed                string // del
	UpperPullSeasons         string // del
	UpperWithoutBangumi      string // del
}
