package conf

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
	"go-common/library/conf/paladin.v2"
	infocV2 "go-common/library/log/infoc.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	Conf = &Config{}
)

type Config struct {
	// 主动控制某些功能
	FeatureGate FeatureGate
	// show  XLog
	Log *log.Config
	// tracer
	Tracer *trace.Config
	// bm http
	BM *HTTPServers
	// grpc Archive
	ArchiveGRPC           *warden.ClientConfig
	AccountGRPC           *warden.ClientConfig
	ArcGRPC               *warden.ClientConfig
	RelaGRPC              *warden.ClientConfig
	ThumGRPC              *warden.ClientConfig
	PGCAppGRPC            *warden.ClientConfig
	ArticleGRPC           *warden.ClientConfig
	DynamicGRPC           *warden.ClientConfig
	DynaGRPC              *warden.ClientConfig
	PopularGRPC           *warden.ClientConfig
	PGCFollowGRPC         *warden.ClientConfig
	PGCDynGRPC            *warden.ClientConfig
	PGCSeasonGRPC         *warden.ClientConfig
	PGCEpisodeGRPC        *warden.ClientConfig
	PGCInlineGRPC         *warden.ClientConfig
	DynCheeseGRPC         *warden.ClientConfig
	DynDrawGRPC           *warden.ClientConfig
	CommunityGRPC         *warden.ClientConfig
	EsportsGRPC           *warden.ClientConfig
	VideoupGRPC           *warden.ClientConfig
	DynamicActivityGRPC   *warden.ClientConfig
	ActivityClient        *warden.ClientConfig
	DynamicTopicExtGRPC   *warden.ClientConfig
	ChannelClient         *warden.ClientConfig
	LivexRoomGRPC         *warden.ClientConfig
	LivexRoomFeedGRPC     *warden.ClientConfig
	LivexRoomGateGRPC     *warden.ClientConfig
	BcgGRPC               *warden.ClientConfig
	RelationGRPC          *warden.ClientConfig
	DynVoteGRPC           *warden.ClientConfig
	UGCSeasonGRPC         *warden.ClientConfig
	GarbGRPC              *warden.ClientConfig
	FavGRPC               *warden.ClientConfig
	TunnelGRPC            *warden.ClientConfig
	NatPageGRPC           *warden.ClientConfig
	PGCShareGRPC          *warden.ClientConfig
	ShortURLGRPC          *warden.ClientConfig
	ActivityServiceClient *warden.ClientConfig
	ShareGRPC             *warden.ClientConfig
	TopicGRPC             *warden.ClientConfig
	NativePageGRPC        *warden.ClientConfig
	GeoGRPC               *warden.ClientConfig
	DynamicCampusGRPC     *warden.ClientConfig
	UpGRPC                *warden.ClientConfig
	DramaseasonGRPC       *warden.ClientConfig
	PlayurlGRPC           *warden.ClientConfig
	HomePageGRPC          *warden.ClientConfig
	MemberGRPC            *warden.ClientConfig
	PassportGRPC          *warden.ClientConfig
	LocationGRPC          *warden.ClientConfig
	IpDisplayGRPC         *warden.ClientConfig
	PanguGRPC             *warden.ClientConfig
	ResourceGRPC          *warden.ClientConfig
	// httpClient
	HTTPClient *bm.ClientConfig
	// httpClient with longer timeout
	HTTPClientLongTimeOut *bm.ClientConfig
	// httpClient game
	HTTPClientGame *bm.ClientConfig
	// HTTPDataClient
	HTTPData *bm.ClientConfig
	// hosts
	Hosts *Hosts
	// tick
	Tick *Tick
	// resource
	Resource        *Resource
	BottomConfig    *BottomConfig
	FoldPublishList *FoldPublishList
	// infoc
	Infoc *Infoc
	// Grayscale
	Grayscale *Gray
	// dynamic location grpc
	DynamicLocGRPC *warden.ClientConfig
	// ctrl
	Ctrl             *Ctrl
	RecommendTimeout xtime.Duration
	BuildLimit       *BuildLimit
	Melloi           *Melloi
	// Mogul
	Mogul *mogul
	// feature配置
	Feature *Feature
	// 用于非公开API的鉴权 通常用于其他内部服务调用
	AppAuth *AppAuth
	// rpc
	MossGRPC *warden.ServerConfig
	// redis
	Redis *Redis
}

// 简单的基于app key鉴权
// 用于提供给内部服务的API作为中间件使用
type AppAuth struct {
	AuthInfo map[string]*AppAuthInfo
}

type AppAuthInfo struct {
	AppName string
	AppKey  string
}

var DefaultAppAuthMetaKeys = []string{
	"x-bili-internal-gw-auth",
}

// 根据 grpc metadata中字段进行验证
// 默认取 "x-bili-internal-gw-auth" 的值
// 例如 x-bili-internal-gw-auth: "appName appKey"
func (ai *AppAuth) authGW(md metadata.MD, info *grpc.UnaryServerInfo, metaKey ...string) (bool, []string) {
	if ai == nil || ai.AuthInfo == nil {
		log.Warn("AppAuthGW: empty config. default allow passing for %q", info.FullMethod)
		// 无配置 默认允许
		return true, nil
	}
	if md == nil || len(md) == 0 {
		return false, nil
	}
	keys := DefaultAppAuthMetaKeys
	if len(metaKey) > 0 {
		keys = append(keys, metaKey...)
	}
	var appName, appKey string
	ret := make([]string, 0, len(keys))
	for _, k := range keys {
		if v := md.Get(k); len(v) > 0 {
			ret = append(ret, fmt.Sprintf("%s:(%s)", k, v[0]))
			idx := strings.Index(v[0], " ")
			if idx == -1 || idx+1 >= len(v[0]) {
				continue
			}
			appName = v[0][0:idx]
			appKey = v[0][idx+1:]
			if au, ok := ai.AuthInfo[appName]; ok && au != nil {
				if au.AppName == appName && au.AppKey == appKey {
					return true, ret
				}
			}
		}
	}
	return false, ret
}

// 简单的基于app key鉴权
// 用于提供给内部服务的API作为中间件使用
func (ai *AppAuth) UnaryServerInterceptor(metaKey ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, _ := metadata.FromIncomingContext(ctx)
		if ok, kvs := ai.authGW(md, info, metaKey...); !ok {
			if len(kvs) > 0 {
				log.Warnc(ctx, "AppAuthGW: rejected req with wrong keys: %v", kvs)
			} else {
				log.Warnc(ctx, "AppAuthGW: rejected req without required keys: %+v", md)
			}
			return nil, ecode.Unauthorized
		}
		return handler(ctx, req)
	}
}

type mogul struct {
	// mogul config
	Databus *databus.Config
}

type Infoc struct {
	SvideoInfoc     *infocV2.Config
	SvideoLogID     string
	TabAbLogID      string
	DynAbLogID      string
	AiSchoolInfocID string
}

type Hosts struct {
	ApiCo    string
	VcCo     string
	MallCo   string
	SearchCo string
	Data     string
	Comic    string
	Game     string
	LiveCo   string
	CmCom    string
	Shopping string
}

type Tick struct {
	RcmdCron     string
	LoadHotVideo string
}

type HTTPServers struct {
	Outer *bm.ServerConfig
}

type Redis struct {
	DynamicSchool    *redis.Config
	DynamicExclusive *redis.Config
}

type Resource struct {
	LbsIcon               string
	TopicIcon             string
	HotIcon               string
	HotURI                string
	LikeDisplay           string
	GameIcon              string
	FoldPublishForward    xtime.Duration
	FoldPublishOther      xtime.Duration
	UplistMore            string
	TopicSquareMoreIcon   string
	ThreePointDislikeIcon string
	PlayIcon              string
	AdditionGoodIcon      string
	Text                  *ResourceText
	Icon                  *ResourceIcon
	Others                *ResourceOthers
	WeightIcon            string
	ReserveShare          *ReserveShare
	ReserveHasFold        int32
	ReserveShow           bool
	TabAbtestUserTime     int64
	CreativeIDmToAvid     bool
}

type ResourceText struct {
	DynVoteFaild string
	// 动态综合
	DynMixTopicSquareMore        string
	DynMixUnfollowTitle          string
	DynMixUnfollowButtonUncheck  string
	DynMixUnfollowButtonCheck    string
	DynMixUnfollowDislike        string
	DynMixLowfollowTitle         string
	DynMixLowfollowButtonUncheck string
	DynMixLowfollowButtonCheck   string
	// 三点
	ThreePointWaitAddition        string
	ThreePointWaitNotAddition     string
	ThreePointAutoPlayOpenV1      string
	ThreePointAutoPlayCloseV1     string
	ThreePointAutoPlayOpenIPADV1  string
	ThreePointAutoPlayCloseIPADV1 string
	ThreePointAutoPlayOpenV2      string
	ThreePointAutoPlayCloseV2     string
	ThreePointAutoPlayOnly        string
	ThreePointDislike             string
	ThreePointBackground          string
	ThreePointShare               string
	ThreePointFollow              string
	ThreePointFollowCancel        string
	ThreePointReport              string
	ThreePointDeleted             string
	ThreePointFav                 string
	ThreePointCancelFav           string
	ThreePointTopText             string
	ThreePointTopCannlText        string
	ThreePointHideText            string
	ThreePointCampusDelText       string
	ThreePointCampusDelToast      string
	ThreePointCampusFeedback      string
	ThreePointBatchCancel         string
	ThreePointBatchCancelDesc     string
	// 动态删除文案
	ThreePointDelReserveTitle string
	ThreePointDelReserveDesc  string
	// 用户模块
	ModuleAuthorPublishLabelDefault                string
	ModuleAuthorPublishLabelArchive                string
	ModuleAuthorPublishLabelSubscriptionNewLive    string
	ModuleAuthorPublishLabelSubscriptionNewSuffix  string
	ModuleAuthorPublishLabelSubscriptionNewESports string
	// 核心模块
	ModuleDynamicForwardDefaultTips string
	ModuleDynamicLiveBadgeFinish    string
	ModuleDynamicLiveBadgeLiving    string
	// 附加大卡模块
	ModuleAdditionalOGVHeadText                string
	ModuleAdditionalAttachedPromoHeadText      string
	ModuleAdditionalNatPageHeadText            string
	ModuleAdditionalNatPageHeadTextV2          string
	ModuleAdditionalNatPageButtonDefault       string
	ModuleAdditionalNatPageUncheck             string
	ModuleAdditionalNatPageCheck               string
	ModuleAdditionalNatPageNotStart            string
	ModuleAdditionalNatPageOver                string
	ModuleAdditionalAttachedPromoButton        string
	ModuleAdditionalMatchHeadText              string
	ModuleAdditionalMatchButtonUncheck         string
	ModuleAdditionalMatchButtonCheck           string
	ModuleAdditionalMatchStartedButtonLiveing  string
	ModuleAdditionalMatchStartedButtonPlayback string
	ModuleAdditionalGameHeadText               string
	ModuleAdditionalMangaHeadText              string
	ModuleAdditionalMangaButtonUncheck         string
	ModuleAdditionalMangaButtonCheck           string
	ModuleAdditionalDecorateHeadText           string
	ModuleAdditionalDecorateDefault            string
	ModuleAdditionalDecorateSell               string
	ModuleAdditionalDecorateButtonUncheck      string
	ModuleAdditionalDecorateButtonCheck        string
	ModuleAdditionalPUGVHeadText               string
	ModuleAdditionalVoteTips                   string
	ModuleAdditionalVoteOpen                   string
	ModuleAdditionalVoteClose                  string
	ModuleAdditionalVoteVoted                  string
	ModuleAdditionalUgcHeadText                string
	ModuleAdditionalUgcHeadTextV2              string
	ModuleAdditionalTopicHeadText              string
	ModuleAdditionalTopicHeadTextV2            string
	ModuleAdditionalTopicButtonText            string
	AdditionalButtonShareText                  string
	AdditionUpActivityOnline                   string
	AdditionUpActivityOffline                  string
	AdditionalFeedCardDramaHeadText            string
	// 附加小卡
	ModuleExtendBBQTitle            string
	ModuleExtendHotTitle            string
	ModuleExtendBiliCutDefaultTitle string
	ModuleExtendDuversionTitle      string
	ModuleExtendDuversionText       string
	// 技术模块
	ModuleStatNoComment             string
	ModuleStatNoForward             string
	ModuleStatNoForwardForwardFaild string
	// 校园
	SearchToast          string
	SearchToastDesc      string
	CampusRcmdTopicTitle string // 校园 推荐话题动态卡默认标题
}

type ResourceIcon struct {
	// 动态综合
	DynMixUplistMore      string
	DynMixTopicSquareMore string
	DynMixUnfollowDislike string
	// 三点
	ThreePointDislike        string
	ThreePointWait           string
	ThreePointWaitView       string
	ThreePointAutoPlayClose  string
	ThreePointAutoPlayOpen   string
	ThreePointBackground     string
	ThreePointShare          string
	ThreePointFollow         string
	ThreePointFollowCancel   string
	ThreePointReport         string
	ThreePointDeleted        string
	ThreePointDeletedView    string
	ThreePointFav            string
	ThreePointFavCancel      string
	ThreePointTopIcon        string
	ThreePointTopCannlIcon   string
	ThreePointHideIcon       string
	ThreePointCampusDel      string
	ThreePointCampusFeedback string
	ThreePointBatchCancel    string
	// 核心模块
	ModuleAuthorDefaultFace string
	ModuleDynamicPlayIcon   string
	ModuleDynamicItemNull   string
	// 附加大卡
	ModuleAdditionalGoods     string
	ModuleAdditionalManga     string
	AdditionalButtonShareIcon string
	AdditionalAdditionalCron  string
	AdditionalAdMarkIcon      string
	// 附加小卡
	ModuleExtendBiliCut   string
	ModuleExtendLBS       string
	ModuleExtendBBQ       string
	ModuleExtendGameTopic string
	ModuleExtendTopic     string
	ModuleExtendHot       string
	ModuleExtendAutoOGV   string
	// 新话题
	ModuleExtendNewTopic string
	// 话题发布的+号icon
	ModuleButtonPlusMark string
	// 神评
	GodReply string
}

type ResourceOthers struct {
	// 核心模块
	ModuleDynamicMedialistBadge    *ResourceBadge
	ModuleDynamicCommonBadge       *ResourceBadge
	ModuleDynamicSubscriptionBadge *ResourceBadge
	// 附加大卡模块
	ModuleAdditionalMatchTeam   *ResourceLabel
	ModuleAdditionalMatchState  *ResourceLabel
	ModuleAdditionalMatchVS     *ResourceLabel
	ModuleAdditionalMatching    *ResourceLabel
	ModuleAdditionalMatchOver   *ResourceLabel
	ModuleAdditionalMatchDard   *ResourceLabel
	ModuleAdditionalMatchLight  *ResourceLabel
	ModuleAdditionalMatchMiddle *ResourceLabel
	// 附加小卡
	ModuleExtendBBQURI            string
	ModuleExtendHotURI            string
	ModuleExtendBiliCutDefaultURI string
	// story发布卡
	StoryURI string
	// 校园
	SchoolInviteURI         string
	SchoolNoticeURI         string
	SchoolBillboardShareURI string // 校园十大排行榜分享链接
}

type ResourceBadge struct {
	Text             string
	TextColor        string
	TextColorNight   string
	BgColor          string
	BgColorNight     string
	BorderColor      string
	BorderColorNight string
	BgStyle          int32
}

type ResourceLabel struct {
	Text           string
	TextColor      string
	TextColorNight string
}

// 详情页预约分享组件配置
type ReserveShare struct {
	Name       string
	Image      string
	Channel    string
	QrCodeIcon string
	QrCodeText string
	QrCodeUrl  string
	DescAv     string
	DescLive   string
}

type Ctrl struct {
	UpListMoreLimit int
	RedClose        bool
	DescTitleLimit  int
	// 秒开新参数开启
	PlayerArgs bool
	// 在设置的时间点后显示发布IP
	IPDisplayAfter time.Time
	// 是否展示 已编辑 提示
	ShowDynEdit bool
	// 校园双列瀑布流强制视频横屏
	CampusWaterFlowForceVideoHorizontal bool
}

type BuildLimit struct {
	DescIdToTitleHightLightIOS        int64
	DescIdToTitleHightLightAndroid    int64
	DescIdToTitleHightLightPad        int64
	UplistMoreSortTypeIOS             int64
	UplistMoreSortTypeAndroid         int64
	UplistMoreSortTypePad             int64
	FakeExtendIOS                     int64
	FakeExtendAndroid                 int64
	FakeExtendPad                     int64
	NewPlayerIOS                      int64
	NewPlayerAndroid                  int64
	DynUpRcmdOldUiIOS                 int64
	DynUpRcmdOldUiIOSPad              int64
	DynUpRcmdOldUiIOSHD               int64
	DynUpRcmdOldUiAndroid             int64
	DynUnLoginIOS                     int64
	DynUnLoginIOSHD                   int64
	DynUnLoginAndroid                 int64
	DynCommonLabelIOS                 int64
	DynCommonLabelAndroid             int64
	LotteryTypeCronIOS                int64
	LotteryTypeCronAndroid            int64
	DynRedLiveAndroid                 int64
	DynSchoolIOS                      int64
	DynSchoolAndroid                  int64
	DynSchoolThreePointIOS            int64
	DynSchoolThreePointAndroid        int64
	DynSchoolBillboardIOS             int64
	DynSchoolBillboardAndroid         int64
	DynSchoolBillboardAutoOpenIOS     int64
	DynSchoolBillboardAutoOpenAndroid int64
	DynSchoolTopicDiscussAndroid      int64
	DynSchoolTopicDiscussIOS          int64
	DynSchoolTopicPublishBtnAndroid   int64
	DynSchoolTopicPublishBtnIOS       int64
	DynReservePersonalShareIOS        int64
	DynReservePersonalShareAndroid    int64
	DynStoryIOS                       int64
	DynStoryAndroid                   int64
	DynStoryIOSV2                     int64
	DynStoryAndroidV2                 int64
	DynReserveDescIOS                 int64
	DynReserveDescAndroid             int64
	DynArticleIOSPad                  int64
	DynArticleIOSHD                   int64
	DynAdFlyIOS                       int64
	DynAdFlyAndroid                   int64
	DynReplyIOS                       int64
	DynReplyAndroid                   int64
	DynSpaceLiveIOS                   int64
	DynSpaceLiveAndroid               int64
	DynThreeHideIOS                   int64
	DynThreeHideAndroid               int64
	DynStoryCardIOS                   int64
	DynStoryCardAndroid               int64
	DynReservePadIOSPad               int64
	DynReservePadHD                   int64
	DynReservePadAndroid              int64
	DynAdFlyReplyIOS                  int64
	DynAdFlyReplyAndroid              int64
	DynCampusBannerIOS                int64
	DynCampusBannerAndroid            int64
	CampusDynInteractionAndroid       int64 // 校园动态同学点赞外露
	CampusDynInteractionIOS           int64
	DynNewTopicIOS                    int64
	DynNewTopicAndroid                int64
	DynNewTopicIOSHD                  int64
	DynMidInt32IOS                    int64
	DynMidInt32IOSHD                  int64
	DynMidInt32Android                int64
	DynMidInt32AndroidHD              int64
	DynMatchIOS                       int64
	DynMatchAndroid                   int64
	DynFuFeiIOS                       int64
	DynFuFeiIOSHD                     int64
	DynFuFeiAndroidHD                 int64
	// 首映
	DynPropertyAndroid      int64
	DynPropertyIOS          int64
	DynSchoolShowTabIOS     int64
	DynSchoolShowTabAndroid int64
	DynCourUpIOS            int64
	DynCourUpIOSPAD         int64
	DynCourUpAndroid        int64
	DynCourUpIOSHD          int64
	DynCourUpAndroidHD      int64
	DynNewTopicSetIOS       int64
	DynNewTopicSetAndroid   int64
	DynViewEditIOS          int64 // 动态详情页修改按钮
	DynViewEditAndroid      int64
}

type Melloi struct {
	From string
}

type FoldPublishList struct {
	White []int64
}

type BottomConfig struct {
	TopicJumpLinks []BottomItem
}

type BottomItem struct {
	RelatedTopic []string
	Display      string
	URL          string
}

// Gray 通用实验模块配置，详细说明请参照
// https://info.bilibili.co/pages/viewpage.action?pageId=97935024
type Gray struct {
	Tab            *Scale
	StatShow       *Scale
	Relation       *Scale
	UplistMore     *Scale
	ShowInPersonal *Scale
	ShowPlayIcon   *Scale
}

type Scale struct {
	Key       string
	Switch    bool
	FlowType  string
	Bucket    int
	Flow      []*Flow
	Salt      string
	MidList   []int64
	BuvidList []string
}

type Flow struct {
	Low       int
	High      int
	MidList   []int64
	BuvidList []string
}

const (
	// 灰度分流类型
	GrayScaleFlowTypeMid   = "mid"
	GrayScaleFlowTypeBuvid = "buvid"
)

func (s *Scale) GrayCheck(mid int64, buvid string) int {
	if s == nil {
		return 0
	}
	if !s.Switch {
		return 0
	}
	// 黑名单逻辑
	for _, blackMid := range s.MidList {
		if mid == blackMid {
			return 0
		}
	}
	for _, blackBuvid := range s.BuvidList {
		if buvid == blackBuvid {
			return 0
		}
	}
	// flow 校验
	for _, flow := range s.Flow {
		if flow.Low < 0 || flow.High > s.Bucket {
			return 0
		}
	}
	// 分流逻辑
	bucket := 0
	switch s.FlowType {
	case GrayScaleFlowTypeMid:
		if s.Salt != "" {
			bucket = int(crc32.ChecksumIEEE([]byte(strconv.FormatInt(mid, 10)+s.Salt))) % s.Bucket
		} else {
			bucket = int(mid % int64(s.Bucket))
		}
	case GrayScaleFlowTypeBuvid:
		bucket = int(crc32.ChecksumIEEE([]byte(buvid+s.Salt))) % s.Bucket
	}
	for i, flow := range s.Flow {
		var value = i + 1
		// 白名单
		for _, whithMid := range flow.MidList {
			if mid == whithMid {
				return value
			}
		}
		for _, whithBuvid := range flow.BuvidList {
			if buvid == whithBuvid {
				return value
			}
		}
		if bucket >= flow.Low && bucket <= flow.High {
			return value
		}
	}
	return 0
}

// 功能开关控制 用于控制某些功能
type FeatureGate struct {
	// 不出游戏附加卡
	NoGameAttach FeatureGateItem
}

type FeatureGateItem struct {
	Enable    bool
	Platforms []TargetPlatform
}

type TargetPlatform struct {
	// mobiapp 和 channel 是且的关系 目标平台必须具有这两个值才能匹配
	MobiApp string
	Channel string
	// device 和 build是可选 不指定的话就匹配所有 device和build的值
	Device string
	Build  int64
}

type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}

type FeatureBuildLimit struct {
	Switch                     bool
	DynUpRcmdOldUi             string
	DynUnLogin                 string
	DynCommonLabel             string
	LotteryTypeCron            string
	DynRedLive                 string
	DynStory                   string
	DynReserveDesc             string
	DynReservePersonalShare    string
	DynSchool                  string
	DynSchoolThreePoint        string // 校园五期 动态三点下发
	DynSchoolBillboard         string // 校园六期 校园十大榜单
	DynSchoolBillboardAutoOpen string // 校园榜单自动开放
	DynSchoolTopicDiscuss      string // 校园六期 校园话题讨论
	DynSchoolTopicPublishBtn   string // 校园话题讨论界面发布优化
	DynArticle                 string
	DynAdFly                   string
	DynReply                   string
	DynSpaceLive               string
	DynThreeHide               string
	DynStoryCard               string
	DynReservePad              string
	DynAdFlyReply              string
	DynCampusBanner            string
	CampusDynInteraction       string // 校园动态同学点赞外露
	DynNewTopic                string // 动态新话题 包括动态顶部卡下发，动态第一页的频道与话题推荐，搜索新话题等
	DynNewTopicSet             string // 新话题 话题集订阅卡
	DynMidInt32                string
}

// Init init config.
func Init() (err error) {
	err = paladin.Init()
	if err != nil {
		return
	}
	err = remote()
	return
}

func remote() (err error) {
	err = paladin.Get("app-dynamic.toml").UnmarshalTOML(Conf)
	if err != nil {
		return
	}
	err = paladin.Watch("app-dynamic.toml", Conf)
	if err != nil {
		return err
	}
	return
}

func (c *Config) Set(str string) error {
	tmp := Config{}
	if err := toml.Unmarshal([]byte(str), &tmp); err != nil {
		return err
	}
	*c = tmp
	return nil
}

func GetInfoc(c *Config) (infoc infocV2.Infoc) {
	client, err := infocV2.New(c.Infoc.SvideoInfoc)
	if err != nil {
		log.Error("init service infoc err:%+v", err)
		panic(err)
	}
	return client
}
