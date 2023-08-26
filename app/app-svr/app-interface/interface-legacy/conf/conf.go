package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/log/infoc"
	infocv2 "go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
)

var (
	Conf = &Config{}
)

// Config struct
type Config struct {
	// Local cache
	LocalCache bool
	// interface XLog
	XLog *log.Config
	// tick time
	Tick xtime.Duration
	// tracer
	Tracer *trace.Config
	// databus
	UseractPub   *databus.Config
	ConfigSetPub *databus.Config
	// httpClinet
	HTTPClient *bm.ClientConfig
	// httpIm9
	HTTPIm9 *bm.ClientConfig
	// httpSearch
	HTTPSearch *bm.ClientConfig
	// httpWrite
	HTTPWrite *bm.ClientConfig
	// httpLive
	HTTPLive *bm.ClientConfig
	// httpbangumi
	HTTPBangumi *bm.ClientConfig
	// httpbplus
	HTTPBPlus *bm.ClientConfig
	// httpgame
	HTTPGame *bm.ClientConfig
	// httpgame.co
	HTTPGameCo *bm.ClientConfig
	// httpAd
	HTTPAd        *bm.ClientConfig
	HTTPAdTopic   *bm.ClientConfig
	HTTPFeedAdmin *bm.ClientConfig
	HTTPEntrance  *bm.ClientConfig
	HTTPRcmdTags  *bm.ClientConfig
	// http
	BM *HTTPServers
	// host
	Host *Host
	// http discovery
	HostDiscovery *HostDiscovery
	// account
	AccountRPC *rpc.ClientConfig
	// rpc client
	ResourceRPC *rpc.ClientConfig
	ArticleRPC  *rpc.ClientConfig
	// db
	MySQL *MySQL
	// ecode
	// ecode
	Ecode *ecode.Config
	// redis
	Redis *Redis
	// search
	Search *Search
	// space
	Space *Space
	// contribute
	ContributePub *databus.Config
	// Report
	Report *databus.Config
	// build limit
	SearchBuildLimit *SearchBuildLimit
	SpaceBuildLimit  *SpaceBuildLimit
	SpaceLikeRule    *SpaceLikeRule
	SpaceSkipRule    *SpaceSkipRule
	// login build
	LoginBuild *LoginBuild
	BuildLimit *BuildLimit
	// infoc
	Infoc   *infoc.Config
	Infocv2 infocv2.Infoc
	// quota
	QuotaConf *quota.Config
	// search dynamic
	SearchDynamicSwitch *SearchDynamicSwitch
	// grpc
	Warden            *warden.ServerConfig
	VipGRPC           *warden.ClientConfig
	VipInfoGRPC       *warden.ClientConfig
	FavClient         *warden.ClientConfig
	RelationGRPC      *warden.ClientConfig
	RelationSh1GRPC   *warden.ClientConfig
	AccountGRPC       *warden.ClientConfig
	UsersuitGRPC      *warden.ClientConfig
	UpClient          *warden.ClientConfig
	PGCRPC            *warden.ClientConfig
	CoinClient        *warden.ClientConfig
	ThumbupClient     *warden.ClientConfig
	VideoupOpenClient *warden.ClientConfig
	ArticleGRPC       *warden.ClientConfig
	AnswerGRPC        *warden.ClientConfig
	ArchiveGRPC       *warden.ClientConfig
	UGCSeasonGRPC     *warden.ClientConfig
	HistoryGRPC       *warden.ClientConfig
	LiveGRPC          *warden.ClientConfig
	CheeseGRPC        *warden.ClientConfig
	LocationGRPC      *warden.ClientConfig
	UpArcGRPC         *warden.ClientConfig
	GarbGRPC          *warden.ClientConfig
	ChannelGRPC       *warden.ClientConfig
	OgvReviewGRPC     *warden.ClientConfig
	ESportsGRPC       *warden.ClientConfig
	SportsGRPC        *warden.ClientConfig
	GuardGRPC         *warden.ClientConfig
	DynamicTopicGRPC  *warden.ClientConfig
	NatGRPC           *warden.ClientConfig
	ActivityClient    *warden.ClientConfig
	SpaceClient       *warden.ClientConfig
	SpacePhotoClient  *warden.ClientConfig
	SpaceUpRcmdClient *warden.ClientConfig
	ResClient         *warden.ClientConfig
	MemClient         *warden.ClientConfig
	CampusSvrClient   *warden.ClientConfig
	SeriesClient      *warden.ClientConfig
	ElecClient        *warden.ClientConfig
	UpArcClient       *warden.ClientConfig
	SiriExtClieng     *warden.ClientConfig
	NoteClient        *warden.ClientConfig
	GrabClient        *warden.ClientConfig
	ResourceGRPC      *warden.ClientConfig
	FeedClient        *warden.ClientConfig
	AppDynamicGRPC    *warden.ClientConfig
	ToViewGRPC        *warden.ClientConfig
	DynGRPC           *warden.ClientConfig
	DynShareGRPC      *warden.ClientConfig
	ManagerSearchGRPC *warden.ClientConfig
	DynTopicGRPC      *warden.ClientConfig
	OTTGRPC           *warden.ClientConfig
	ContractGRPC      *warden.ClientConfig
	GalleryGRPC       *warden.ClientConfig
	StaticKVGRPC      *warden.ClientConfig
	PassportUser      *warden.ClientConfig
	Live2dGRPC        *warden.ClientConfig
	GameEntryGRPC     *warden.ClientConfig
	BgroupGRPC        *warden.ClientConfig
	CfcGRPC           *warden.ClientConfig
	DigitalGRPC       *warden.ClientConfig
	CheckinClient     *warden.ClientConfig
	NewmontClient     *warden.ClientConfig
	// custom config
	Cfg *Cfg
	// func switch
	Switch *Switch
	// player build limit
	PlayerBuildLimit map[string]int
	// DegradeConfig
	DegradeConfig  *DegradeConfig
	Degrade2Config *DegradeConfig
	// SpaceTabABTest
	SpaceTabABTest *SpaceTabABTest
	// SpaceTabABTest
	SpaceNewABTest *SpaceNewABTest
	// Fav build limit
	FavBuildLimit   *FavBuildLimit
	SearchPageTitle *SearchPageTitle
	SpaceGame       *SpaceGame
	// Custom
	Custom       *Custom
	SpaceArchive SpaceArchive
	// cron
	Cron *Cron
	// 下发资源
	Resource    *Resource
	Creative    *Creative
	Teenagers   *teenagers
	IOSPreIcons []*IOSPreIcons
	AndPreIcons []*AndPreIcons
	// feature配置
	Feature     *Feature
	HisIcon     *HisIcon
	HisLTimeMap map[string]int64
	HisRTimeMap map[string]int64
	//历史记录空页面时候的跳转链接business:link
	HisEmptyLink map[string]string
	LegoToken    *legoToken
	LivePlayback *livePlayback
	GameModuleID map[string]int64
	// 防沉迷
	AntiAddictionRule *AntiAddictionRule
	// content.flow.control.service 配置
	CfcSvrConfig *CfcSvrConfig
	// 睡眠提醒
	SleepRemind *SleepRemind
	// 显示IP地址开关
	ActiveLocationSwitch *ActiveLocationSwitch
	// 天马搜索导航位配置
	SearchRcmdTagsConfig *SearchRcmdTagsConfig
	// 显示MCN机构Tag开关
	MCNTagSwitch *MCNTagSwitch
	//硬核会员生日祝福配置
	SeniorGateBirthday SeniorGateBirthday
	//成就达成配置
	AchievementConf AchievementConf
}

type SearchRcmdTagsConfig struct {
	CloseRcmdTagsSwitch bool
	AiRcmdTimeout       string
}

type SleepRemind struct {
	Gray       int64
	Whitelist  []int64
	Switch     bool
	BgroupBiz  string
	BgroupName string
	FullSwitch bool
}

// 由服务方定义
type CfcSvrConfig struct {
	BusinessID int64
	Secret     string
	Source     string
}

type AntiAddictionRule struct {
	Switch         bool
	Id             int64
	Version        string
	Frequency      int64
	SeriesDuration xtime.Duration
	SeriesInterval xtime.Duration
	BgroupBiz      string
	BgroupName     string
	PushTitle      string
	PushSubtitle   string
}

type livePlayback struct {
	UpFrom []int32
}

type legoToken struct {
	SpaceIPLimit string
}

// HostDiscovery Http Discovery
type HostDiscovery struct {
	Data   string
	Mall   string
	Pay    string
	Search string
	Music  string
	Space  string
}

type HisIcon struct {
	Phone string
	Pad   string
	TV    string
	PC    string
	Car   string
	Iot   string
}

type LiveTipConf struct {
	//提示条图标
	Icon string
	//提示条样式
	Mod int64
	//跳链
	Url string
	//文案
	Text string
	//按钮文案
	ButtonText string
	//按钮图案
	ButtonIcon string
	//跳链文案
	UrlText string
}

type AndPreIcons struct {
	ID           int64
	Title        string
	URL          string
	Icon         string
	RedDot       int8
	GlobalRedDot int8
	NeedLogin    int8
	Display      int32
}

type IOSPreIcons struct {
	ID           int64
	Title        string
	URL          string
	Icon         string
	RedDot       int8
	GlobalRedDot int8
	NeedLogin    int8
	Display      int32
}

type teenagers struct {
	InnerInterval     int64
	OuterInterval     int64
	OuterZone         map[string][]int64
	NoneZone          []*noneZone
	ForceOpen         bool  //强制拉入开关
	ForceClose        bool  //强制退出开关
	ForceOpenInterval int64 //强拉时间间隔
	ForceOnlineTime   int64 //强拉上线时间
}

type noneZone struct {
	Index int
	Zone  map[string][]int64
}

type Creative struct {
	BeUpTitle        string
	UpTitle          string
	TipIcon          string
	TipTitle         string
	ButtonName       string
	DefaultUpTitle   string
	DefaultBeUpTitle string
}

// Custom is
type Custom struct {
	SpaceSeriesPLaylistBuvid map[string]int // space series 白名单
	PhotoMallTitle           string
	SetArchiveText           string
	PhotoArcCount            int
	CreatedActCnt            int
	UpArcHasShare            bool
	UpArcAddToViewIcon       string
	UpArcAddToViewText       string
	UpArcShareIcon           string
	UpArcShareText           string
	UpArcShareSuccToast      string
	UpArcShareFailToast      string
	HisHasShare              bool
	RecommendTimeout         xtime.Duration
	LatestHistoryPro         float64
	MallIcon                 string
	HisLTime                 int64
	HisRTime                 int64
	NewFollowersRTime        int64
	EncyclopediaBWToken      string
	AndroidPadSectionExp     int64
	IpadNewSectionMid        map[string]int64
	ChannelLink              string
	TopLevelExTime           int64
	SeniorGateExTime         int64
	BirthdaySwitchOn         bool
}

type SeniorGateBirthday struct {
	Icon       string
	BubbleText string
	Url        string
}

type AchievementConf struct {
	TopLevelIcon   string
	SeniorGateIcon string
}

// Cfg def.
type Cfg struct {
	PgcSearchCard *PgcSearchCard
	GarbCfg       *GarbCfg
	MaxDisplay    int
}

type GarbCfg struct {
	GoodsAvailable      bool
	PurchaseButtonTitle string
	PurchaseButtonURI   string
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

// LoginBuild is
type LoginBuild struct {
	Iphone   int
	Android  int
	AndroidB int
	IphoneB  int
	IpadHD   int
}

// BuildLimit is
type BuildLimit struct {
	IOSMineCreative         int
	AndMineCreative         int
	IPhoneBMineCreative     int
	IOSCheese               int
	AndroidCheese           int
	IPadCheese              int
	IPadHDCheese            int
	AndroidHDCheese         int
	AndroidBCheese          int
	IPhoneBCheese           int
	IOSMineLive             int
	AndMineLive             int
	NewMineIOSBuild         int
	NewMineAndBuild         int
	NewMineIPhoneBBuild     int
	NewMineAndBBuild        int
	NewMineAndIBuild        int
	NewMineIPad             int
	NewMineIPadHD           int
	NewMineAndroidHD        int
	NewPlayurlAndBuild      int
	NewPlayurlIOSBuild      int
	NewCreativeIOSBuild     int
	NewCreativeAndBuild     int
	NewCreativeIOSBBuild    int
	NewCreativeAndBBuild    int
	SkinOpenIOSBuild        int
	SkinOpenAndroidBuild    int
	NewMixCreativeIOSBuild  int
	NewMixCreativeAndBuild  int
	NewMixCreativeIOSBBuild int
	NewMixCreativeAndBBuild int
	NewFansIOSBuild         int
	NewFansIOSBBuild        int
	OGVChanIOSBuild         int64
	OGVChanAndroidBuild     int64
}

// HTTPServers Http Servers
type HTTPServers struct {
	Outer *bm.ServerConfig
}

// Host struct
type Host struct {
	Account   string
	Bangumi   string
	APICo     string
	Im9       string
	Search    string
	Game      string
	GameCo    string
	Space     string
	Elec      string
	VC        string
	APILiveCo string
	WWW       string
	Show      string
	Pay       string
	Member    string
	Mall      string
	Manga     string
	Ad        string
	AdTopic   string
	Data      string
	Manager   string
	FeedAdmin string
	Workshop  string
	//游戏中心人群包
	GameDmp string
}

// MySQL struct
type MySQL struct {
	Show      *sql.Config
	Teenagers *sql.Config
}

// Redis struct
type Redis struct {
	Contribute *struct {
		*redis.Config
	}
	Interface *struct {
		*redis.Config
		EmptyCacheExpire    xtime.Duration
		EmptyCacheRand      xtime.Duration
		AntiAddictionExpire xtime.Duration
		FamilyRelsExpire    xtime.Duration
		SleepRemindExpire   xtime.Duration
		UnlockErrorExpire   xtime.Duration
	}
	Attention *struct {
		*redis.Config
	}
	Family *struct {
		*redis.Config
		QrcodeExpire       xtime.Duration
		QrcodeStatusExpire xtime.Duration
		LockExpire         xtime.Duration
		TimelockPwdExpire  xtime.Duration
	}
}

// Search struct
type Search struct {
	SeasonNum             int
	MovieNum              int
	SeasonMore            int
	MovieMore             int
	UpUserNum             int
	UVLimit               int
	UserNum               int
	UserVideoLimit        int
	UserVideoLimitMix     int
	BiliUserNum           int
	BiliUserVideoLimit    int
	BiliUserVideoLimitMix int
	OperationNum          int
	IPadSearchBangumi     int
	IPadSearchFt          int
	TrendingLimit         int
	EggCloseCount         int
	BackgroundSwitch      bool
	LiveFaceSwitch        bool
	SearchRankingSwitch   bool
	SpaceEntrance         *SpaceEntrance
}

type SquareSortOptAbTest struct {
	DisableSquareSortOptAbTest bool           // 搜索发现页排布优化开关
	SquareSortOptAbTestMid     map[string]int // 搜索发现页排布优化白名单
	SquareSortOptAbTestLimit   int            // 搜索发现页排布优化分流
}

type SpaceEntrance struct {
	TextMore        string
	TextMoreWithNum string
	TextColor       string
	TextColorNight  string
}

// Space struct
type Space struct {
	ForbidMid []int64
}

// SearchBuildLimit struct
type SearchBuildLimit struct {
	PGCHighLightIOS          int
	PGCHighLightAndroid      int
	PGCALLIOS                int
	PGCALLAndroid            int
	SpecialerGuideIOS        int
	SpecialerGuideAndroid    int
	SearchArticleIOS         int
	SearchArticleAndroid     int
	ComicIOS                 int
	ComicAndroid             int
	ChannelIOS               int
	ChannelAndroid           int
	CooperationIOS           int
	CooperationAndroid       int
	CooperationIPadHD        int
	QueryCorIOS              int
	QueryCorAndroid          int
	SugDetailIOS             int
	SugDetailAndroid         int
	NewTwitterIOS            int
	NewTwitterAndroid        int
	NewOrderIOS              int
	NewOrderAndroid          int
	DefaultWordJumpIOS       int
	DefaultWordJumpAndroid   int
	DefaultWordJumpAndroidI  int
	NewChannelIOS            int
	NewChannelAndroid        int
	ESportsIOS               int
	ESportsAndroid           int
	VideoDurationIOS         int
	VideoDurationAndroid     int
	OGVURLAndroid            int
	OGVURLIOS                int
	SpecialCardIOS           int
	SpecialCardAndroid       int
	UpNewAndroid             int
	UpNewIOS                 int
	CardOptimizeAndroid      int
	CardOptimizeIPhone       int
	CardOptimizeIpadHD       int
	TipsCardIOS              int
	TipsCardAndroid          int
	ADCardIOS                int
	ADCardAndroid            int
	UserInlineLiveIOS        int
	UserInlineLiveAndroid    int
	FlowInlineCardIOS        int
	FlowInlineCardAndroid    int
	FlowOGVInlineCardIOS     int
	FlowOGVInlineCardAndroid int

	// type search
	TypeSearchWithPlayURLIOS     int
	TypeSearchWithPlayURLAndroid int
	TypeSearchChannelESIOS       int
	TypeSearchChannelESAndroid   int
}

// SearchDynamicSwitch .
type SearchDynamicSwitch struct {
	IsUP    bool
	IsCount bool
}

// SpaceLikeRule .
type SpaceLikeRule struct {
	SkrTip string
}

// SpaceSkipRule .
type SpaceSkipRule struct {
	AchieveURL   string
	AchieveImage string
	PendantURL   string
}

// SpaceBuildLimit struct
type SpaceBuildLimit struct {
	FavIOS              int
	FavAndroid          int
	CooperationIOS      int
	CooperationAndroid  int
	UGCSeasonIOS        int
	UGCSeasonAndroid    int
	UGCSeasonAndroidI   int
	ComicIOS            int
	ComicAndroid        int
	ComicAndroidI       int
	HideSexSwitch       bool
	HideSexAndroidStart int
	HideSexAndroidEnd   int
	SubComicIOS         int
	SubComicAndroid     int
	ContinuePlayAndroid int
	ContinuePlayIOS     int
	PlayurlIOS          int
	PlayurlAndroid      int
	IPadHDArchiveSort   int
	ButtonTextAnd       int
	ButtonTextIOS       int
	ButtonTextIpad      int
	PrInfoCardIOS       int
	PrInfoCardAndroid   int
	PrInfoCardIPadHD    int
}

// Switch func switch.
type Switch struct {
	SearchRecommend bool
	SearchSuggest   bool
	SearchMainRcmd  bool
	// 商品店铺开关
	AdOpen bool
	// 大航海开关
	GuardOpen bool
	// 皮肤装扮开关
	SkinOpen bool
	// 空间投稿全部
	SpaceContributeAll bool
	// 搜索三点
	SearchThreePoint bool
	// 秒开新参数开启
	PlayerArgs bool
}

// DegradeConfig struct.
type DegradeConfig struct {
	Expire            int32
	Memcache          *memcache.Config
	NewSearchMemcache *memcache.Config
}

// SpaceTabABTest .
type SpaceNewABTest struct {
	// 关注abtest
	AtTestNum int
}

// SpaceTabABTest struct.
type SpaceTabABTest struct {
	Exp1     int
	Exp2     int
	IsParams bool
}

// FavBuildLimit struct.
type FavBuildLimit struct {
	ComicIOS     int
	ComicAndroid int
	NoteIOS      int
	NoteAndroid  int
	//频道
	ChannelTabIOSBuild     int
	ChannelTabAndroidBuild int
}

type SearchPageTitle struct {
	HistoryTitle string
	FindTitle    string
}

type SpaceGame struct {
	JumpUri string
	Image   string
}

type SpaceArchive struct {
	EpisodicOpen  bool
	EpisodicMid   []int64
	EpisodicText  string
	EpisodicText1 string
	EpisodicDesc  string
}

type Cron struct {
	LoadSidebar         string
	LoadBlacklist       string
	LoadHotCache        string
	LoadSearchTipsCache string
	LoadSpecialCache    string
	LoadUpRcmdBlockList string
	LoadSystemNotice    string
}

type Resource struct {
	SearchThreePoint *SearchThreePoint
}

type SearchThreePoint struct {
	WaitIcon   string
	WaitTitle  string
	ShareIcon  string
	ShareTitle string
}

type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}

type FeatureBuildLimit struct {
	Switch           bool
	ShowLive         string // done
	SearchPlayer     string // done
	SpaceFansEffect  string // done
	SpaceTabActivity string // done
	SpaceShop        string // done
	SpaceShopElse    string // done
	SearchParamOGV   string // done
	HistoryPlayurl   string // done
	MineSkin         string // done
}

type ActiveLocationSwitch struct {
	OwnerSwitch bool
	GuestSwitch bool
}

type MCNTagSwitch struct {
	Switch bool
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
