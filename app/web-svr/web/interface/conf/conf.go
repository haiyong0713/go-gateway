package conf

import (
	xtime "time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"

	"go-common/library/conf/paladin.v2"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"
	"go-common/library/net/trace"
	"go-common/library/time"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"

	infocV2 "go-common/library/log/infoc.v2"

	"github.com/BurntSushi/toml"
)

// global var
var (
	Conf = &Config{}
)

// Config config set
type Config struct {
	// log
	Log *log.Config
	// ecode
	Ecode *ecode.Config
	// http client
	HTTPClient *httpClient
	// HTTPServer
	HTTPServer *blademaster.ServerConfig
	// GRPCServer
	GRPCCfg *GRPCCfg
	// tracer
	Tracer *trace.Config
	// auth
	Auth *auth.Config
	// verify
	Verify *verify.Config
	// DynamicRPC
	DynamicRPC *rpc.ClientConfig
	// TagRPC
	TagRPC *rpc.ClientConfig
	//  ActivityRPC
	ActivityRPC *rpc.ClientConfig
	// Host
	Host host
	// redis
	Redis *redisConf
	// degrade
	DegradeConfig      *degradeConfig
	SearchDegradeCache *SearchDegradeCache
	// WEB
	WEB *web
	// Tag
	Tag *tag
	// DefaultTop
	DefaultTop *defaultTop
	// Bfs
	Bfs *bfs
	// Infoc2
	InfocV2      *infocV2.Config
	InfocV2LogID *InfocV2LogID
	// Rule
	Rule *rule
	// Warden Client
	BroadcastClient *warden.ClientConfig
	CoinClient      *warden.ClientConfig
	ArcClient       *warden.ClientConfig
	AccClient       *warden.ClientConfig
	AccmClient      *warden.ClientConfig
	PanguGSClient   *warden.ClientConfig
	ShareClient     *warden.ClientConfig
	UGCClient       *warden.ClientConfig
	ResClient       *warden.ClientConfig
	Res2Client      *warden.ClientConfig
	CreativeClient  *warden.ClientConfig
	ThumbupClient   *warden.ClientConfig
	UpClient        *warden.ClientConfig
	RelationClient  *warden.ClientConfig
	DMClient        *warden.ClientConfig
	SeasonClient    *warden.ClientConfig
	HisClient       *warden.ClientConfig
	EpClient        *warden.ClientConfig
	DynamicClient   *warden.ClientConfig
	AnswerClient    *warden.ClientConfig
	TagClient       *warden.ClientConfig
	// TagService 目前只给vlog用，其他不要调用
	TagSvrClient     *warden.ClientConfig
	ArtClient        *warden.ClientConfig
	CouponClient     *warden.ClientConfig
	SteinsClient     *warden.ClientConfig
	UGCSeasonClient  *warden.ClientConfig
	LocClient        *warden.ClientConfig
	FavClient        *warden.ClientConfig
	PayRankClient    *warden.ClientConfig
	VideoUpClient    *warden.ClientConfig
	RoomGRPC         *warden.ClientConfig
	GaiaGRPC         *warden.ClientConfig
	ShareAdminGRPC   *warden.ClientConfig
	GarbGRPC         *warden.ClientConfig
	UpArcGRPC        *warden.ClientConfig
	CommonActiveGRPC *warden.ClientConfig
	ActGRPC          *warden.ClientConfig
	ActPlatGRPC      *warden.ClientConfig
	// bnj
	Bnj2019 *bnj2019
	Bnj2020 *bnj2020
	// switch
	Switch *Switch
	// grpc
	ChannelGRPC       *warden.ClientConfig
	PGCRPC            *warden.ClientConfig
	ArchiveGRPC       *warden.ClientConfig
	CoinGRPC          *warden.ClientConfig
	FavoriteGRPC      *warden.ClientConfig
	TagGRPC           *warden.ClientConfig
	CheeseDynamicGRPC *warden.ClientConfig
	CheeseSeasonGRPC  *warden.ClientConfig
	DynamicFeedGRPC   *warden.ClientConfig
	PGCShareGRPC      *warden.ClientConfig
	VoteGRPC          *warden.ClientConfig
	DMGRPC            *warden.ClientConfig
	AccountGRPC       *warden.ClientConfig
	PopularGRPC       *warden.ClientConfig
	SeasonGRPC        *warden.ClientConfig
	EmoteGRPC         *warden.ClientConfig
	ESportsGRPC       *warden.ClientConfig
	ESportsConfGRPC   *warden.ClientConfig
	HonorGRPC         *warden.ClientConfig
	SmsGRPC           *warden.ClientConfig
	FeedAdminGRPC     *warden.ClientConfig
	TopicGRPC         *warden.ClientConfig
	DynTopicGRPC      *warden.ClientConfig
	WatchedGRPC       *warden.ClientConfig
	CfcGRPC           *warden.ClientConfig
	OnlineGRPC        *warden.ClientConfig
	LiveGRPC          *warden.ClientConfig
	TradeGRPC         *warden.ClientConfig
	CampusGRPC        *warden.ClientConfig
	BgroupGRPC        *warden.ClientConfig
	PcdnAccGRPC       *warden.ClientConfig
	PcdnRewardGRPC    *warden.ClientConfig
	PcdnVerifyGRPC    *warden.ClientConfig
	// Cron
	Cron *Cron
	// resource grpc
	ResourceClient *warden.ClientConfig
	// grpc article
	ArticleGRPC *warden.ClientConfig
	// PR limit
	PRLimit *PRLimit
	// popular
	PopularPrecious *PopularPrecious
	PopularSeries   *PopularSeries
	Custom          *Custom
	// ActivitySeason
	ActivitySeason *ActivitySeason
	// Rcmd
	Rcmd        *rcmd
	LandingPage map[string]*landingPage
	// ArcNoSearch
	ArcNoSearch *arcNoSearch
	// anticrawler
	Anticrawler     *anticrawler.Config
	ShowDB          *sql.Config
	PopularActivity *PopularActivity
	// pwd_appeal
	PwdAppeal *PwdAppeal
	// basis season gray
	BasisSeasonABTest *BasisSeasonABTest
	BanArcGRPCToken   string
	// content.flow.control.service gRPC config
	CfcSvrConfig *CfcSvrConfig
	// switch
	SeniorMemberSwitch *SeniorMemberSwitch
	RiskManagement     *RiskManagement
}

type PwdAppeal struct {
	EncryptKey string
	Boss       *Boss
	Captcha    *Captcha
}

type Boss struct {
	Bucket          string
	EndPoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

type Captcha struct {
	Biz          string //业务类型
	IpDailyLimit int64  //单IP每天发送验证码限制
	ValidTime    int64  //验证码有效时间
	IntervalTime int64  //验证码发送间隔时长
	VerifyLimit  int64  //验证码最大验证次数
	Digit        int    //验证码位数
	Tcode        string //验证码模板编码
	TField       string //验证码模板字段名
}

type landingPage struct {
	Newlist map[string]int64
}

type arcNoSearch struct {
	Referers []string
}

type GRPCCfg struct {
	GRPC     *warden.ServerConfig
	QuotaCfg *quota.Config
}

type ActivitySeason struct {
	RightRelatedTitle string
	BreakCycle        int64
}

type Custom struct {
	RecommendTimeout time.Duration
}

type PopularSeries struct {
	Reminder string
	MediaMid int64
}

type PopularPrecious struct {
	Title      string
	ExplainFmt string
}

type InfocV2LogID struct {
	UserActLogID string
	PopularLogID string
	RcmdLogID    string
}

type Cron struct {
	Type                  string
	NewCount              string
	OnlineTotal           string
	OnlineList            string
	IndexIcon             string
	IndexIconRand         string
	SearchEgg             string
	Manager               string
	SpecialRcmd           string
	WxHot                 string
	SteinsGuide           string
	BvIDJs                string
	Bnj2019               string
	Bnj2020               string
	WebTop                string
	RankIndex             string
	RankRecommend         string
	RankRegion            string
	RankV2                string
	ParamConfig           string
	InformationRegionCard string
	Popular               string
	ActivitySeason        string
	RegionList            string
	SearchTipDetail       string
	Fawkes                string
}

type rule struct {
	// min cache rank count
	MinRankCount int
	// min cache rank index count
	MinRankIndexCount int
	// min cache rank region count
	MinRankRegionCount int
	MaxRankRegionCount int
	// min cache rank recommend count
	MinRankRecCount int
	// min cache rank tag count
	MinRankTagCount int
	// min cache dynamic count
	MinDyCount int
	// min newlist tid arc count
	MinNewListCnt int
	// Elec
	ElecShowTypeIDs []int32
	// AuthorRecCnt author recommend count
	AuthorRecCnt int
	// UpGroupCnt up test group count
	UpGroupCnt int32
	// RelatedArcCnt related archive limit count
	RelatedArcCnt int
	// MaxHelpPageSize help detail search max page  count
	MaxHelpPageSize int
	// newlist
	MaxArcsPageSize int64
	// max size of second region newlist.
	MaxSecondCacheSize int
	// max size of first region newlist.
	MaxFirstCacheSize int
	// default num of dynamic archives
	DynamicNumArcs int64
	// regions count
	RegionsCount int
	// bangumi count
	BangumiCount int64
	// MaxArtPageSize max article page size
	MaxArtPageSize int
	// article up list get count
	ArtUpListGetCnt int32
	// article up list count
	ArtUpListCnt int
	// min wechat hot count
	MinWxHotCount int
	// Rids first region ids
	Rids []int32
	// web top limit
	WebTop         int
	Grayscale      []int32
	GrayscaleAll   int
	SteinsGuideAid int64
	// new main rids newlist and top count rid
	MainRids             []int32
	RecommendRids        []int64
	RankFirstRegion      []int64
	RankOfflineRegion    []int64
	RankNoOriginalRegion []int64
	DayAllRegion         []int64
	// online access token
	OnlineToken string
	// rank v2 rids
	RankV2Rids []int64
	// hot share
	HotShare     int
	HotShareOut  int
	HotShareOpen bool
	// forbid jump to act page
	ForbidBnjJump bool
	Recommend     *recommend
	// tiny package rids
	TinyPackageRegion []int64
	// teenage mode rids
	TeenageModeRegion []int64
	// custom tags for h5 test
	H5SubdivisionTags []string
	// pcdn bgroup conf
	PcdnGroup *pcdnGroup
}

type recommend struct {
	AidCount  int
	ItemCount int
}

type pcdnGroup struct {
	Top      string
	Pop      string
	Business string
}

type httpClient struct {
	Read    *blademaster.ClientConfig
	Write   *blademaster.ClientConfig
	BigData *blademaster.ClientConfig
	Help    *blademaster.ClientConfig
	Search  *blademaster.ClientConfig
	Pay     *blademaster.ClientConfig
	Game    *blademaster.ClientConfig
	Rcmd    *blademaster.ClientConfig
}

type host struct {
	Rank                 string
	API                  string
	Data                 string
	Space                string
	Elec                 string
	ArcAPI               string
	LiveAPI              string
	HelpAPI              string
	HelpAPINew           string
	Mall                 string
	Search               string
	Manager              string
	Pay                  string
	AbServer             string
	Game                 string
	VcAPI                string
	RcmdDiscovery        string
	ReplyDiscovery       string
	SearchDiscovery      string
	SearchMainDiscovery  string
	MallDiscovery        string
	PayDiscovery         string
	TeenageRcmdDiscovery string
	MusicAPI             string
	AdDiscovery          string
	CampusRcmdDiscovery  string
	FawkesAPI            string
}

type tag struct {
	PageSize int
	MaxSize  int
}

type redisConf struct {
	LocalRedis *localRedis
	BakRedis   *bakRedis
	IndexRedis *indexSet
	Popular    *localRedis
}

type localRedis struct {
	*redis.Config
	RankingExpire time.Duration
}

type indexSet struct {
	*redis.Config
}

type bakRedis struct {
	*redis.Config
	RankingExpire     time.Duration
	NewlistExpire     time.Duration
	RegionExpire      time.Duration
	ArchiveExpire     time.Duration
	TagExpire         time.Duration
	CardExpire        time.Duration
	RcExpire          time.Duration
	ArtUpExpire       time.Duration
	IndexIconExpire   time.Duration
	HelpExpire        time.Duration
	OlListExpire      time.Duration
	AppealLimitExpire time.Duration
}

type web struct {
	PullRegionInterval    time.Duration
	PullOnlineInterval    time.Duration
	PullIndexIconInterval time.Duration
	SearchEggInterval     time.Duration
	OnlineCount           int
	SpecailInterval       time.Duration
	SpecRecmInterval      time.Duration
	WxHotInterval         time.Duration
}

type defaultTop struct {
	SImg string
	LImg string
}

type bfs struct {
	Addr        string
	Bucket      string
	Key         string
	Secret      string
	MaxFileSize int
	Timeout     time.Duration
}

type degradeConfig struct {
	Expire   int32
	Memcache *memcache.Config
}

type SearchDegradeCache struct {
	Expire   int32
	Memcache *memcache.Config
}

type bnj2019 struct {
	Open        bool
	LiveAid     int64
	BnjMainAid  int64
	FakeElec    int64
	BnjListAids []int64
	BnjTick     time.Duration
	Timeline    []*struct {
		Name    string
		Start   xtime.Time
		End     xtime.Time
		Cover   string
		H5Cover string
	}
}

type bnj2020 struct {
	Open           bool
	MainAid        int64
	SpAid          int64
	LiveAid        int64
	ListAids       []int64
	RelateAidToRid map[string]int64
	Tick           time.Duration
	Timeline       []*struct {
		Type     int
		Name     string
		Start    xtime.Time
		End      xtime.Time
		Cover    string
		H5Cover  string
		Subtitle string
		Tag      []string
	}
}

type Switch struct {
	DetailVerify bool
	ListOGVMore  bool
	ListOGVFold  bool
}

type PRLimit struct {
	ChannelList []int64
}

type rcmd struct {
	Timeout   time.Duration
	Whitelist []int64
	Bucket    uint64
	GroupTest *struct {
		Group0 bool
		Group1 bool
		Group2 bool
		Group3 bool
		Group4 bool
		Group5 bool
		Group6 bool
		Group7 bool
		Group8 bool
		Group9 bool
		// 空buvid实验开启状态
		EmptyBuvid bool
	}
	AdResource map[string]int
}

type PopularActivity struct {
	Sid                  int64
	ArchiveCounter       string
	ShortTermSkinAwardId int64
	LongTermSkinAwardId  int64
	BadgeAwardId         int64
	MessageIDsGroup1     []int64
	MessageIDsGroup2     []int64
	MessageIDsGroup3     []int64
	MessageIDsGroup4     []int64
	HasBadgeStock        bool
	SingleEmoteId        []int64
	AllEmoteId           []int64
}

type BasisSeasonABTest struct {
	Group uint32
	Gray  uint32
}

type CfcSvrConfig struct {
	BusinessID int64
	Secret     string // 由服务方下发
	Source     string
}

type SeniorMemberSwitch struct {
	ShowSeniorMember bool
}

type RiskManagement map[string][]string

func Init() (err error) {
	err = paladin.Init()
	if err != nil {
		return
	}
	return remote()
}

func remote() (err error) {
	if err = load(); err != nil {
		return err
	}
	err = paladin.Watch("web-interface.toml", Conf)
	if err != nil {
		return err
	}

	return
}

func load() (err error) {
	err = paladin.Get("web-interface.toml").UnmarshalTOML(Conf)
	if err != nil {
		return
	}
	return
}

func Close() {
	paladin.Close()
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	log.Info("web-interface-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}
