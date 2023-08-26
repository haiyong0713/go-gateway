package conf

import (
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/log/infoc"
	infocv2 "go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
)

// Conf is
var (
	Conf = &Config{}
)

// Config struct
type Config struct {
	// Env
	Env      string
	DMRegion []int16
	// interface XLog
	XLog *log.Config
	// infoc
	InfocCoin     *infoc.Config
	InfocViewV2   *InfocV2Conf
	InfocRelateV2 *InfocV2Conf
	UseractPub    *databus.Config
	DislikePub    *databus.Config
	// tick time
	Tick xtime.Duration
	// tracer
	Tracer *trace.Config
	// httpClinet
	HTTPClient *bm.ClientConfig
	// httpWrite
	HTTPWrite *bm.ClientConfig
	// httpbangumi
	HTTPBangumi *bm.ClientConfig
	// httpaudio
	HTTPAudio *bm.ClientConfig
	// http_ai_client
	HTTPAiClient *bm.ClientConfig
	// http
	BM        *HTTPServers
	RpcServer *warden.ServerConfig
	// httpAd
	HTTPAD *bm.ClientConfig
	// httpGame
	HTTPGame *bm.ClientConfig
	// HTTPAsync
	HTTPAsync *bm.ClientConfig
	// HTTPGameAsync
	HTTPGameAsync *bm.ClientConfig
	// httpClinet
	HTTPSearch *bm.ClientConfig
	// host
	Host *Host
	// rpc client
	TagRPC           *rpc.ClientConfig
	ResourceRPC      *rpc.ClientConfig
	LocationRPC      *rpc.ClientConfig
	LocationGRPC     *warden.ClientConfig
	PlayerOnlineGRPC *warden.ClientConfig
	// grpc client
	CreativeClient         *warden.ClientConfig
	CoinClient             *warden.ClientConfig
	AccClient              *warden.ClientConfig
	ThumbupClient          *warden.ClientConfig
	VideoupClient          *warden.ClientConfig
	UpClient               *warden.ClientConfig
	SteinClient            *warden.ClientConfig
	AssistClient           *warden.ClientConfig
	ShareClient            *warden.ClientConfig
	ArcGRPC                *warden.ClientConfig
	RelationGRPC           *warden.ClientConfig
	ChannelClient          *warden.ClientConfig
	ArchiveHonorClient     *warden.ClientConfig
	ArchiveExtraClient     *warden.ClientConfig
	MusicClient            *warden.ClientConfig
	UGCSeasonClient        *warden.ClientConfig
	HisClient              *warden.ClientConfig
	UGCPayRankClient       *warden.ClientConfig
	ActivityClient         *warden.ClientConfig
	NatClient              *warden.ClientConfig
	LiveClient             *warden.ClientConfig
	FavClient              *warden.ClientConfig
	DMClient               *warden.ClientConfig
	ADClient               *warden.ClientConfig
	ResClient              *warden.ClientConfig
	AppConfClient          *warden.ClientConfig
	GaiaClient             *warden.ClientConfig
	VCloudClient           *warden.ClientConfig
	GarbClient             *warden.ClientConfig
	UpArcGRPC              *warden.ClientConfig
	MngClient              *warden.ClientConfig
	ContractClient         *warden.ClientConfig
	VoteClient             *warden.ClientConfig
	ReplyClient            *warden.ClientConfig
	ESportsClient          *warden.ClientConfig
	FlowControllerClient   *warden.ClientConfig
	BGroupClient           *warden.ClientConfig
	ArchiveMaterialClient  *warden.ClientConfig
	CreativeMaterialClient *warden.ClientConfig
	CreativeSparkClient    *warden.ClientConfig
	TopicClient            *warden.ClientConfig
	Buzzword               *warden.ClientConfig
	ListenerClient         *warden.ClientConfig
	NotesClient            *warden.ClientConfig
	CopyrightClient        *warden.ClientConfig
	FreyaClient            *warden.ClientConfig
	DeliveryClient         *warden.ClientConfig
	CheckinClient          *warden.ClientConfig
	TradeClient            *warden.ClientConfig
	// db
	MySQL *MySQL
	// ecode
	Ecode *ecode.Config
	// PlayURL
	PlayURL *PlayURL
	// 相关推荐秒开个数
	RelateCnt int
	// buildLimit
	BuildLimit *BuildLimit
	// play icon
	PlayIcon *PlayIcon
	// Custom
	Custom *Custom
	// BfsArc
	BfsArc *BfsArc
	// fawkes tick
	FawkesTick xtime.Duration
	// view config
	ViewConfig *ViewConfig
	// tag config
	TagConfig    *TagConfig
	InfocV2      *infocv2.Config
	InfocV2LogID *InfocV2LogID
	Redis        *Redis
	// http discovery
	HostDiscovery *HostDiscovery
	// cron
	Cron *Cron
	//ng switch
	NgSwitch *NgSwitch
	// resource
	Resource *Resource
	// ActivitySeason
	ActivitySeason *ActivitySeason
	LiveOrderMid   map[string]int64
	LiveOrderGray  map[string]int64
	PopupWhiteList map[string]int64
	//新版小窗白名单
	NewSwindowWhiteList map[string]int64
	//相关推荐双列白名单
	RelatesBiserialWhiteList map[string]int64
	// feature平台
	Feature   *Feature
	LabelIcon *LabelIcon
	Mng       *Mng
	Online    *Online
	//在看人数-简介内-左下角-特殊弹幕弹幕-三点里的开关面板4个地方是否展示
	OnlineCtrl         *OnlineCtrl
	LegoToken          *LegoToken
	LikeNumGrayControl *LikeNumGrayControl
}

// LikeNumGrayControl 点赞数灰度控制
type LikeNumGrayControl struct {
	Aid  map[string]int64
	Gray int64
}

type LegoToken struct {
	PlayOnlineToken string
}

type OnlineCtrl struct {
	Logo     string
	Mid      map[string]int64
	SwitchOn bool
	Gray     int64
}

type InfocV2Conf struct {
	LogID string
	Conf  *infocv2.Config
}

type Online struct {
	Text string
	//ugc横全屏开关
	SwitchOn bool
	//ugc竖全屏开关
	SwitchOnUS bool
	//story开关
	SwitchOnStory bool
}

type LabelIcon struct {
	Hot          *LabelIconInfo
	Act          *LabelIconInfo
	Steins       *LabelIconInfo
	Encyclopedia *LabelIconInfo
	Premiere     *LabelIconInfo
}

type LabelIconInfo struct {
	Icon        string
	IconNight   string
	IconWidth   int64
	IconHeight  int64
	Lottie      string
	LottieNight string
}

type Mng struct {
	EncyclopediaToken string
}

type ActivitySeason struct {
	//内存白名单aids
	Aids []int64
	//内存白名单season id
	Sid              int64
	RelateTitle      string
	AndroidBuild     int
	AndroidBlueBuild int
	IphoneBuild      int
	IphoneBlueBuild  int
	IpadHDBuild      int
	IpadBuild        int
	IPhoneIBuild     int
	AndroidIBuild    int
}

type InfocV2LogID struct {
	UserActLogID    string
	TagABLogID      string
	ContinuousLogID string
}

type NgSwitch struct {
	DisableAll bool // 总开关
	Feature    map[string][]string
}

type Redis struct {
	PlayerRedis *redis.Config
	OnlineRedis *redis.Config
	// 竖屏切全屏进story黑名单,实验结束需删除
	PlayStoryRedis *redis.Config
}

type TagConfig struct {
	OpenIcon bool
	ActIcon  string
	NewIcon  string
}

type ViewConfig struct {
	RelatesTitle      string
	AutoplayDesc      string
	AutoplayCountdown int
}

// Custom is
type Custom struct {
	// hot aids tick
	StoryIcon         string
	StoryDays         []int64
	HotAidsTick       xtime.Duration
	ElecShowTypeIDs   []int16
	SteinsBuild       *SteinsBuild
	DisplayHonorMids  []int64
	DisplayHonorGray  int64
	HonorRank         int32
	HonorRankMax      int32
	EndPageMids       []int64
	EndPageHalfGroup  uint32
	EndPageFullGroup  uint32
	VideoShotGray     int64
	VideoShotGrayIOS  int64
	VideoShotAndBuild int32
	VideoShotIOSBuild int32
	LikeNeedLogin     int
	// vie tag active
	ViewTag          bool
	SWindowSwitch    bool
	RecommendTimeout xtime.Duration
	AdFirstRatio     int64
	UpArcText        string
	LiveSwitchOn     bool
	PlayerArgs       bool
	LiveOrderText    string
	VideoViewNumber  int32
	//video_download version
	VideoDownloadBuildIphone  int64
	VideoDownloadBuildAndroid int64
	//pop time
	PopupExTime             int64
	PipSwitchOn             bool
	AiRecommendGray         int64
	AiRecommendMidWhiteList []int64
	//点赞场景化开关
	LikeCustomSwitch bool
	//点赞场景化-视频播放量后台配置
	LikeCustomVideoView int64
	//点赞场景化-更新次数
	LikeCustomUpdateCount int64
	//点赞场景化-全屏切非全屏进度
	FullToHalfProgress int64
	//非全屏状态
	NonFullProgress int64
	//点赞场景化-在线人数后台配置
	LikeCustomOnlineCount int64
	DisableAdTab          bool
	//新话题-tag收进简介白名单
	NewTopicDescTagWhite []int64
	NewTopicDescGrey     int64
	//相关推荐新样式
	RecStyleGrey       int
	RecStyleGreySwitch bool
	RecStyleAiGroup    uint64
	RecStyleWhite      []int64
	GroupsName         string
	GroupsBusiness     string
	// 分端陆续去除chronosV2使用限制
	//key:mobi_app
	ChronosV2SwitchOnMap        map[string]int64
	SeasonContinualButtonSwitch bool
	//禁止项secret
	FlowControllerSecret         string
	MiaoKaiWithRelateCntSwitchOn bool
	//img-host
	VideoShotHost      string
	CloseMusicIcon     bool
	PlayBackgroundGrey int64
	//首映风控原因
	PremiereRiskReason string
	//contract player switch
	ContractPlayDisplaySwitch bool
	//竖屏视频切全屏进story白名单
	PlayStoryMids []int64
	//竖屏视频切全屏黑名单时间限制
	StoryBlackTime int64
	//高清图机型黑名单
	ShotBlackModel []string
	//游戏卡强化角标开关
	PowerBadgeSwitch bool
	//展示新增稿件投稿IP的时间线
	ShowArcPubIpAfterTime string
	//稿件投稿IP白名单fromSpmid
	ShowArcPubIpFromSpmid map[string]int64
	//灵感话题灰度
	InspirationTagGrey   int64
	InspirationTagSwitch bool
	//s12赛事开关
	ESportS12Switch   bool
	ESportS12MidWhite []int64
	//听视频按钮灰度
	ListenButtonGrey int64
	ListenButtonType []int
}

type SteinsBuild struct {
	Android     int
	AndroidBlue int
	IosPink     int
	IosBlue     int
	IpadHD      int
	AndroidI    int
	IphoneI     int
}

// BfsArc is
type BfsArc struct {
	Key    string
	Secret string
}

// HTTPServers Http Servers
type HTTPServers struct {
	Outer *bm.ServerConfig
}

// Host struct
type Host struct {
	Bangumi      string
	APICo        string
	Activity     string
	Elec         string
	AD           string
	Data         string
	Archive      string
	APILiveCo    string
	Game         string
	AI           string
	Search       string
	Bvcvod       string
	BvcDiscovery string
	Fawkes       string
	Bfs          string
	ManagerHost  string
}

// HostDiscovery Http Discovery
type HostDiscovery struct {
	AD   string
	Data string
}

// MySQL struct
type MySQL struct {
	Show    *sql.Config
	Manager *sql.Config
}

// PlayURL playurl token's secret.
type PlayURL struct {
	Secret string
}

// BuildLimit for build limit
type BuildLimit struct {
	CooperationIPadHD     int
	CooperationIPad       int
	CooperationIOS        int
	CooperationAndroid    int
	ChannelIOS            int
	ChannelAndroid        int
	HonorRankIOS          int
	HonorRankAndroid      int
	AttentionLevelIOS     int
	AttentionLevelAndroid int
	OGVURLAndroid         int
	OGVURLIOS             int
	// view channel active
	ViewChannelActiveIOS      int
	ViewChannelActiveAndroid  int
	CardOptimizeAndroid       int
	CardOptimizeIPhone        int
	CardOptimizeIPadHD        int
	PlayIconIOSBuildLimit     int
	PlayIconAndroidBuildLimit int
	PlayIconIpadHDBuildLimit  int
	DmInfoAndBuild            int
	DmInfoIOSBuild            int
	LiveIOSBuildLimit         int
	LiveAndBuildLimit         int
	SeasonTypeIOSBuildLimit   int
	SeasonTypeAndBuildLimit   int
	UgcSeasonIPadBuild        int
	UgcSeasonAndroidIBuild    int
	UgcSeasonAndroidHDBuild   int
	UgcSeasonIphoneIBuild     int
	SpecialCellAndroidBuild   int
	SpecialCellIOSBuild       int
	SpecialCellIPadHDBuild    int
	SpecialCellIPadBuild      int
	NewTopicAndroidBuild      int
	NewTopicIOSBuild          int
	NewTopicIPadBuild         int
	NewTopicIPadHDBuild       int
	NewTopicGreySwitch        bool
	NewTopicActTagGreySwitch  bool
	SeasonBaseAndroidHdBuild  int
	StoryPlayIOS              int64
	StoryPlayAndroid          int64
	MusicAndroidBuild         int
	MusicIOSBuild             int
	NoteAndroidBuild          int
	NoteIOSBuild              int
	ShotAndroidHDBuild        int
	ShotIPadBuild             int
	ShotIPadHDBuild           int
	OGVChanIOSBuild           int
	OGVChanAndroidBuild       int
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

// Cron is
type Cron struct {
	LoadChronos             string
	LoadCommonActivities    string
	LoadOnlineManagerConfig string
	LoadArchiveTypes        string
	LoadInspirationTopic    string
}

type Resource struct {
	Coin *ResourceCoin
}

type ResourceCoin struct {
	Title string
}
type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}
type FeatureBuildLimit struct {
	Switch                 bool
	VideoShotBuild         string
	ViewElec               string // del
	ViewPlayIcon           string
	PageEmoji              string
	UgcPayPreview          string // del
	VideoShotIOS           string // del
	VideoShotAndroid       string // del
	ViewDmInfo             string // del
	NewRcmdRelateQn        string // del
	NewRcmdRelateQnBangumi string // del
	HonorRank              string // del
	NewLive                string // del
	NewSeasonType          string // del
	ViewPageSeason         string // del
	ViewPageInitPGC        string // del
	UpArcCount             string
	FromGame               string // del
	DisplaySteins          string
	DisplaySteinsLabel     string
	CardOptimize           string // del
	RelateGame             string // del
	ViewChannel            string // del
	ViewChannelActivity    string // del
	DisplayActSeason       string // del
	Dislike                string
	Cooperation            string // del
	RcmdRelateOgvUrl       string // del
	TcTranslateRequired    string // del
	DmCommandBuild         string
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	old, _ := json.Marshal(c)
	nw, _ := json.Marshal(tmp)
	log.Info("service-config changed, old=%+v new=%+v", string(old), string(nw))
	*c = tmp
	return nil
}
