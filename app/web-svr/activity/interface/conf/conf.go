package conf

import (
	"sync/atomic"
	xtime "time"

	"go-gateway/app/web-svr/activity/tools/lib/conf"

	rpcquota "go-common/library/net/rpc/warden/ratelimiter/quota"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/bfs"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/antispam"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/stat/prom"
	"go-common/library/time"
	gtime "go-common/library/time"

	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-gateway/app/web-svr/activity/interface/model/bnj"
	"go-gateway/app/web-svr/activity/interface/model/like"
	modell "go-gateway/app/web-svr/activity/interface/model/lottery_v2"
	"go-gateway/app/web-svr/activity/interface/model/newyear2021"
	"go-gateway/app/web-svr/activity/interface/model/s10"
	model "go-gateway/app/web-svr/activity/interface/model/wishes_2021_spring"
	"go-gateway/app/web-svr/activity/interface/tool"

	"go-common/library/database/hbase.v2"
)

var (
	// Conf is global config
	Conf = &Config{}
)

type DatabusConfigV2 struct {
	Target string `toml:"target"`
	AppID  string `toml:"app_id"`
	Token  string `toml:"token"`
	Topic  string `toml:"topic"`
}

// Config service config
type Config struct {
	ManuScriptConfMap map[string]*model.CommonActivityConfig `toml:"manu_script_conf"`

	BackupMQ                 *redis.Config `toml:"backup_mq"`
	VoteRedis                *redis.Config
	ExamProducer             *newyear2021.ExamProducer `toml:"exam_producer"`
	Bnj2021ARPub             *DatabusConfigV2          `toml:"bnj_2021_ar_pub"`
	Bnj2021LiveLotteryRec    *DatabusConfigV2          `toml:"bnj_2021_live_lottery_rec"`
	Bnj2021LiveDrawARCoupon  *DatabusConfigV2          `toml:"bnj_2021_live_draw_ar_coupon"`
	MainWebSvrActivity       *DatabusConfigV2
	MainWebSvrJob            *DatabusConfigV2
	AsyncReserveConfig       *AsyncReserveConfig
	ActPlatConfig            *ActPlatConfig
	UpActReserveConfig       *UpActReserveConfig
	CardsComposeConfig       *CardsComposeConfig
	ARDeviceProducer         *newyear2021.ExamProducer `toml:"AR_device_producer"`
	RewardsAwardSendingPubV2 *DatabusConfigV2          `toml:"rewards_award_sending_pub_v2"`

	AsyncReservePub          *databus.Config
	AsyncReserveSub          *databus.Config
	GaiaRiskPub              *databus.Config
	ManuScriptAuditPub       *databus.Config
	ActGuessPub              *databus.Config
	StockServerSyncPubConfig *databus.Config
	NoCancelReserve          map[string]bool
	CustomLimiters           map[string]int64
	CircuitBreaker           map[string]tool.BreakerSetting

	Static string
	// reload
	Reload ReloadInterval
	// auth
	Auth *auth.Config
	// verify
	Verify *verify.Config
	// HTTPServer
	HTTPServer *blademaster.ServerConfig
	// tracer
	Tracer *trace.Config
	// db
	MySQL *MySQL

	RewardsMySQL *sql.Config
	// rpc
	RPCClient2 *RPCClient2
	// grpc
	TagClient *warden.ClientConfig
	// acc client
	SilverGaiaClient   *warden.ClientConfig
	AccClient          *warden.ClientConfig
	ActPlatClient      *warden.ClientConfig
	SilverBullet       *warden.ClientConfig
	ArchiveClient      *warden.ClientConfig
	RelClient          *warden.ClientConfig
	FliClient          *warden.ClientConfig
	VideoClient        *warden.ClientConfig
	CoinClient         *warden.ClientConfig
	ThumbupClient      *warden.ClientConfig
	SuitClient         *warden.ClientConfig
	OgvClient          *warden.ClientConfig
	UpClient           *warden.ClientConfig
	ArtClient          *warden.ClientConfig
	SpyClient          *warden.ClientConfig
	VipClient          *warden.ClientConfig
	LocationRPC        *warden.ClientConfig
	Figure             *warden.ClientConfig
	PassClient         *warden.ClientConfig
	SteinClient        *warden.ClientConfig
	CheeseClient       *warden.ClientConfig
	CheesePayClient    *warden.ClientConfig
	SeasonClient       *warden.ClientConfig
	VipActClient       *warden.ClientConfig
	SilverClient       *warden.ClientConfig
	ResourceClient     *warden.ClientConfig
	FavoriteClient     *warden.ClientConfig
	CouponClient       *warden.ClientConfig
	VipInfoClient      *warden.ClientConfig
	FeatureClient      *warden.ClientConfig
	GarbClient         *warden.ClientConfig
	PassportClient     *warden.ClientConfig
	UpClientNew        *warden.ClientConfig
	MemberClient       *warden.ClientConfig
	RelationClient     *warden.ClientConfig
	TagNewClient       *warden.ClientConfig
	AcpClient          *warden.ClientConfig
	BbqTaskClient      *warden.ClientConfig
	NaPageClient       *warden.ClientConfig
	FlowControlClient  *warden.ClientConfig
	GaiaClient         *warden.ClientConfig
	TunnelClient       *warden.ClientConfig
	AuditClient        *warden.ClientConfig
	LiveClient         *warden.ClientConfig
	LiveActivityClient *warden.ClientConfig
	EsportSercieClient *warden.ClientConfig
	ActivityClient     *warden.ClientConfig
	LiveDataClient     *warden.ClientConfig
	PublishClient      *warden.ClientConfig

	DataMartClient *warden.ClientConfig
	// httpClient
	HTTPClient *blademaster.ClientConfig
	// HTTPClientRewards
	HTTPClientRewards *blademaster.ClientConfig
	// HTTPClientSports
	HTTPClientSports *blademaster.ClientConfig
	// HTTPClientBnj
	HTTPClientBnj   *blademaster.ClientConfig
	HTTPClientComic *blademaster.ClientConfig
	// HTTPClientKfc
	HTTPClientKfc *blademaster.ClientConfig
	// HTTPClientSingle
	HTTPClientSingle *blademaster.ClientConfig
	HTTPDynamic      *blademaster.ClientConfig
	// HTTPClientAddFavInner
	HttpClientAddFavInner *blademaster.ClientConfig
	// Rule
	Rule *Rule
	// bws
	Bws *Bws
	// Host
	Host Host
	// Log
	Log *log.Config
	// ecode
	Ecode *ecode.Config
	// ip
	IPFile string
	// mc
	Memcache *Memcache
	// redis
	Redis *Redis
	// hbase
	Hbase     *hbase.Config
	RPCServer *rpc.ServerConfig
	// interval
	Interval *Interval
	// Elastic
	Elastic *elastic.Config
	// ArcClient
	ArcClient       *warden.ClientConfig
	LiveXRoomClient *warden.ClientConfig
	TvBwClient      *warden.ClientConfig
	PopularClient   *warden.ClientConfig
	PgcClient       *warden.ClientConfig
	PgcActClient    *warden.ClientConfig
	// Bnj
	Bnj2019 *bnj2019
	Lottery *Lottery
	BFS     *bfs.Config
	// Scholarship act
	Scholarship *Scholarship
	// image act
	Image *Image
	// DataBus databus
	DataBus *DataBus
	// star project
	Star  *Star
	Stars []*Star
	// stein act
	Stein *Stein
	// Live props
	Live *Live
	// staff
	Staff *Staff
	// Double Eleven act
	Eleven *Eleven
	// ArticleList
	ArticleList  *ArticleList
	Taaf         *Taaf
	SteinV2      *SteinV2
	Box          *Box
	Ent          *Ent
	Bnj2020      *Bnj2020
	AnnualVoting *AnnualVoting
	Vogue        *Vogue
	// timemachine
	Timemachine *Timemachine
	// UpAct
	UpAct         *UpAct
	SpringCardAct *SpringCardAct
	// cron
	Cron *Cron
	Shad *Shad
	// bdf
	Bdf *Bdf
	// wx lottery
	WxLottery *WxLottery
	// fission lottery
	Fission *Fission
	// special
	Special        *Special
	Restart2020    *Restart2020
	YellowAndGreen *YellowAndGreen
	ReadDay        *ReadDay
	ImageV2        *ImageV2
	MobileGame     *MobileGame
	Faction        *Faction
	StarSpring     *StarSpring
	GiantV4        *GiantV4
	VipSpecial     *VipSpecial
	Stupid         *Stupid
	BdfOnline      *BdfOnline
	// bml
	Bml20         []*Bml20
	StrategyLevel int32
	Antispam      *antispam.Config
	ModelNameOpen bool
	// bws online
	BwsOnline *BwsOnline
	// SongFestival 毕业歌会
	SongFestival *SongFestival
	// Brand
	Brand *Brand
	// HandWrite
	HandWrite *HandWrite
	// HandWrite2021
	HandWrite2021 *HandWrite2021
	// Remix
	Remix *Remix
	// ArticleDay
	ArticleDay *ArticleDay
	// GameHoliday ...
	GameHoliday *GameHoliday
	// Funny
	Funny *Funny
	// S10 Answer
	S10Answer *S10Answer
	// Wechat
	Wechat *Wechat

	S10Redis       *redis.Config
	S10MySQL       *sql.Config
	S10MC          *memcache.Config
	S10Tasks       *s10.ActTask
	S10CacheExpire *s10.CacheExpire
	Limiter        *quota.Config
	RpcLimiter     *rpcquota.Config
	S10Client      *S10Client
	S10CoinCfg     *S10CoinConfig
	S10Cache       *S10Cache
	S10WhiteList   *S10White

	S10PointCostMC    *memcache.Config
	S10PointShopRedis *redis.Config
	S10Matches        *s10.MatchesConf
	S10TimePeriod     *s10.S10TimePeriod
	S10General        *S10General

	// College
	College *College
	// Invite
	Invite *Invite
	// s10 contribution
	S10Contribution *S10Contribution
	// double eleven
	DoubleEleven *DoubleEleven
	// Dubbing
	Dubbing *Dubbing
	Acg2020 *Acg2020
	// selection
	Selection     *Selection
	UpWholeActive *UpWholeActive
	Ticket        *Ticket
	// knowledge
	Knowledge                             *Knowledge
	TiDB                                  *sql.Config
	System                                *System
	Party2021                             *Party2021
	OperationSource                       *OperationSource
	UpActReserveWhiteList                 *UpActReserveWhiteList
	UpActReserveCreateMoreLiveReserveMIDs map[string]int64
	Amusement                             *Amusement

	SpringFestival2021 *SpringFestival2021
	Cards              *Cards
	// amusement

	WinterStudy              *WinterStudy
	ReserveDoveAct           *ReserveDoveAct
	UpActReserveAudit        *UpActReserveAudit
	UpActReserveCreateConfig *UpActReserveCreateConfig
	// 2021愚人节日报活动
	AprilFoolsAct                 *AprilFoolsAct
	EsportsArena                  *EsportsArena
	UpActReserveRelationInfo4Live *UpActReserveRelationInfo4Live
	ActDomainConf                 *ActDomainConf
	UpActReserveAuthConf          *UpActReserveAuthConf
	Vote                          *VoteConfig
	Rewards                       *RewardsConfig
	GaoKaoActConf                 *GaoKaoActConf
	StockRedis                    *StockRedis
	Cpc100                        *Cpc100Config
	BindConf                      *BindConfig
	PassportSNSClient             *warden.ClientConfig
	BiliOAuth2Client              *warden.ClientConfig
	// fit健身打卡活动热门视频配置
	FitHotVideoConf map[string]int
	FitHotVideoList map[string][]int64
	FitCounter      *FitCounter
	// 暑期夏令营配置
	SummerCampConf     *SummerCampConf
	SCCounterCourseMap map[string]string

	BMLGuessAct               *BMLGuessAct
	AppSecrets                map[string]*AppSecret
	SelfClient                *warden.ClientConfig
	MissionActivityConf       *MissionActivityConf
	ReserveTotalShowLimitsMap map[string]int64
	OlympicConf               *OlympicConf
}

type OlympicConf struct {
	ContestSourceId     int64
	QuerySourceId       int64
	ValidContestSize    int
	RefreshCacheSeconds int
}

type SummerCampConf struct {
	LotteryActionType int64
	ActivityIdRW      string
	ReserveId         int64
	LotteryPoolId     string
	LotteryCid        int64
	StartPlanPoint    int64
	ViewPoint         int64
	SharePoint        int64
	ArchivePoint      int64
	AwardPoint        int64
	FitHotVideoConf   map[string]int
	FitHotVideoList   map[string][]int64
	FitCounter        *FitCounter
}

type StockRedis struct {
	*redis.Config
	ConfExpire time.Duration
	DataExpire time.Duration
}

type MissionActivityConf struct {
	CacheRule         *MissionCacheRule
	SpecialActivityId int64
}

type MissionCacheRule struct {
	ValidBeforeTime                  time.Duration
	ValidAfterTime                   time.Duration
	ValidActivitySize                int
	RefreshActivityCacheSeconds      int
	RefreshActivityTasksCacheSeconds int
}

type AppSecret struct {
	Key    string
	Secret string
}

type BMLGuessAct struct {
	ActName      string
	ActBeginTime int64
	ActEndTime   int64

	Common30dayRewardId     int64
	Common30dayStock        int
	Common30dayStockVersion int64

	CommonForeverRewardId     int64
	CommonForeverStock        int
	CommonForeverStockVersion int64

	CommonKey string

	JokerRewardId           int64
	JokerRewardStock        int
	JokerRewardStockVersion int64
	JokerKey                []string

	FollowMids           []int64
	FrequencyControlTime int64
	JokerKeyGuessMax     int64
}

type FitCounter struct {
	HasLimit string
	NoLimit  string
}

type Cpc100Config struct {
	Vid                                        int64
	QuestionBusiness, UnlockBusiness, Activity string
	QuestionSid                                []int64
}

type BindConfig struct {
	ExternalConfig      []*BindExternalInfo
	Games               []*BindGame
	CacheSize           int
	CacheExpireSeconds  int
	RefreshTickerSecond int
}

type BindExternalInfo struct {
	BindExternal int64  `json:"bind_external"`
	ExternalName string `json:"external_name"`
}

type BindGame struct {
	GameId       int64  `json:"game_id"`
	GameName     string `json:"game_name"`
	GameTitle    string `json:"game_title"`
	ExternalName string `json:"external_name"`
	ExternalId   int64  `json:"external_id"`
	Business     string `json:"business"`
	ClientId     string `json:"client_id"`
	OriginId     string `json:"origin_id"`
	AppId        string `json:"app_key"`
	SignKey      string `json:"sign_key"`
	Version      string `json:"version"`
	BasePath     string `json:"base_path"`
}

type GaoKaoActConf struct {
	ActStartTime int64
	ActEndTime   int64
	MaxExamTime  int
	MaxScore     int
	SpitTag      string
	EnableCache  bool
	QtypeMap     map[string]int64
}

type RewardsConfig struct {
	VipMallSourceId       string
	VipMallSourcePlatform string
}

type VoteConfig struct {
	DataSourceItemsInfoCacheExpire         time.Duration
	OutdatedDataSourceItemsInfoCacheExpire time.Duration
	VoteRankZsetExpire                     time.Duration
	RealTimeVoteRankWithInfoExpire         time.Duration
	ManualVoteRankWithInfoExpire           time.Duration
	OnTimeVoteRankWithInfoExpire           time.Duration
	AdminVoteRankWithInfoExpire            time.Duration
	OutdatedVoteRankWithInfoExpire         time.Duration
	ActivityCacheExpire                    time.Duration
	BlackListCacheExpire                   time.Duration
	UserVoteCountExpire                    time.Duration
}

type EsportsArena struct {
	Stime      int64
	Etime      int64
	Sid        string
	LTimeLimit int64
	ZDCid      int64
	UpMids     []int64
}

type ReserveDoveAct struct {
	Stime         int64
	Etime         int64
	Svga          string
	LastImg       string
	PlayTimes     []int64
	ActUrl        string
	ShareUrl      string
	AwardActId    string
	DefaultUpName string
	DefaultUpFace string
	BlessingMsg   []string
	BlackList     []int64
}

type AprilFoolsAct struct {
	Sid        string
	ActivityId int64
	JsonData   string
	Clues      []*struct {
		TimePoint xtime.Time
		Process   float64
		KeyEvent  string
	}
	CluesSrcs []*modell.Item
}

type UpActReserveWhiteList struct {
	Dynamic map[string][]int64
}

type Wechat struct {
	PublicKey string
}

type Ticket struct {
	Mid []int64
}

type OperationSource struct {
	InfoSids      []int64
	OperationSids []int64
	//UpdateTicker  xtime.Duration
}

type Acg2020Task struct {
	StartTime   xtime.Time `json:"start_time"`
	EndTime     xtime.Time `json:"end_time"`
	StatTime    xtime.Time `json:"stat_time"`
	AwardTime   xtime.Time `json:"award_time"`
	FinishCount int        `json:"finish_count"`
	FinishScore int64      `json:"finish_score"`
}

type Acg2020 struct {
	Task []*Acg2020Task
}

// SpringFestival2021 ...
type SpringFestival2021 struct {
	NeedRetry int64
	// LotterySid 奖池id
	LotterySid string
	// Card1GiftID 卡1映射的giftid
	Card1GiftID []int64
	// Card2GiftID 卡2映射的giftid
	Card2GiftID []int64
	// Card3GiftID 卡3映射的giftid
	Card3GiftID []int64
	// Card4GiftID 卡4映射的giftid
	Card4GiftID []int64
	// Card5GiftID 卡5映射的giftid
	Card5GiftID      []int64
	Activity         string
	ActivityID       int64
	Counter          string
	Filter           string
	JoinReserveSid   int64
	FollowMidUri     string
	OgvLinkUri       string
	AllowClickFinish []string
	MaxCard          int64
}

// Cards ...
type Cards struct {
	NeedRetry int64
	// LotterySid 奖池id
	LotterySid string
	// Card1GiftID 卡1映射的giftid
	Card1GiftID []int64
	// Card2GiftID 卡2映射的giftid
	Card2GiftID []int64
	// Card3GiftID 卡3映射的giftid
	Card3GiftID []int64
	// Card4GiftID 卡4映射的giftid
	Card4GiftID []int64
	// Card5GiftID 卡6映射的giftid
	Card5GiftID []int64
	// Card6GiftID 卡6映射的giftid
	Card6GiftID []int64
	// Card7GiftID 卡7映射的giftid
	Card7GiftID []int64
	// Card8GiftID 卡8映射的giftid
	Card8GiftID []int64
	// Card9GiftID 卡9映射的giftid
	Card9GiftID      []int64
	Activity         string
	ActivityID       int64
	Counter          string
	Filter           string
	JoinReserveSid   int64
	FollowMidUri     string
	AllowClickFinish []string
	MaxCard          int64
	ActivityUID      string
}

// Invite ...
type Invite struct {
	TokenExpire     string
	FaceTokenExpire string
	TelSalt         string
	BindExpire      time.Duration
}

type S10General struct {
	SplitTable      bool
	Points          int32
	UnicomSecretkey string
	MobileSecretkey string
	FlowRecvLimit   bool
	FlowSwitch      bool
	Mobile          *S10FlowControl
	Unicom          *S10FlowControl
}

type S10FlowControl struct {
	Switch  bool
	StartAt xtime.Time
	Stock   []*struct {
		StartAt   xtime.Time
		EndAt     xtime.Time
		StartHour int
		Endhour   int
	}
}

type S10White struct {
	List   []int64
	Switch bool
}

type S10Cache struct {
	S10Key *S10Key
	Redis  *redis.Config
	Mc     *memcache.Config
}

type S10Key struct {
	UserExpire          time.Duration
	ListExpire          time.Duration
	ESportsKey          string
	ContestDetailKey    string
	GuessMainDetailsKey string
}

type S10CoinConfig struct {
	McSwitch     bool
	ActivityName string
	SeasonID     int64
}

type S10Client struct {
	Coin    *warden.ClientConfig
	Formula *warden.ClientConfig
	// College
	College *College
	// Invite
	Invite *Invite
}

// College ...
type College struct {
	// FreshMidPeriod 新用户的周期
	FreshMidPeriod int64
	// MemberJoinActivity 用户加入的活动id
	MemberJoinActivity string
	MemberJoinFilter   string
	MemberJoinCounter  string
	VersionChangeCron  string
	CollegeChangeCron  string
	// FollowPoint 关注加积分
	FollowPoint int64
	// InviterPoint 邀请加积分
	InviterPoint int64
	// ViewTimes 观看次数
	ViewTimes int
	// LikeTimes 点赞次数
	LikeTimes int
	// ShareTimes 分享次数
	ShareTimes int
	// SearchShowLength 搜索展示的个数
	SearchShowLength int
	// MidActivity 活动配置
	MidActivity string
	// MidFormula 积分
	MidFormula string
}

type ArticleDay struct {
	ApplyTime  int64
	BeginTime  int64
	EndTime    int64
	ResultTime int64
}

type BwsOnline struct {
	PiecePR            []*BwsOnlinePR
	PrintPieceMinCount int64
	PrintPieceMaxCount int64
	MaxEnergy          int64
	FreeEnergy         int64
	FreeCoin           int64
	ReserveSid         int64
	BuyTicketSid       int64
	FirstEnergy        int64
	DefaultBid         int64
	PrevBid            int64
	ReserveAward       []int64
	// bwPark 哔哩乐园
	BwPark TicketParam
}

type BwsOnlinePR struct {
	ID    int64
	PR    int64
	Level int64
	Num   int64
}

type Bml20 struct {
	Sid    int64
	Mid    int64
	SnsMid int64
	ShopID int64
	DyID   int64
}

// SongFestival 毕业歌会
type SongFestival struct {
	Sid int64
}

type BdfOnline struct {
	Sids       []int64
	LotterySid string
	LotteryCid int64
}

type Stupid struct {
	Sid        int64
	LotterySid string
	Pid        int64
	PidExpire  int64
	LockExpire int32
	Num        int64
	Target1    int64
	Target2    int64
	Target3    int64
	Cid        int64
	Vid        int64
}

type VipSpecial struct {
	Sid string
}

type GiantV4 struct {
	Sid        int64
	LotterySid string
	Cid        int64
}

type Fission struct {
	Caller   string
	UpCaller string
	Sids     map[string]int
}

type Faction struct {
	Sids map[string]string
}

type MobileGame struct {
	Sid        int64
	LotterySid string
	Cid        int64
	Vid        int64
}

type ReadDay struct {
	Sid     int64
	Max     int64
	EndTime xtime.Time
	Multi   float64
}

type WxLottery struct {
	Stime              xtime.Time
	Etime              xtime.Time
	RewardSid          string
	CommonSid          string
	NewCashSid         string
	VipSid             string
	NormalSid          string
	DoJumpURL          string
	AlertURL           string
	BuvidPeriod        int64
	Vid                int64
	SenderUID          uint64
	GiftCron           string
	PlayWindowDuration int64
	JumpURLMap         map[string]string
	MoneyMap           map[string]int64
	MessageMap         map[string]string
}

type ImageV2 struct {
	Sid      int64
	DayLimit int64
	AllLimit int64
	Etime    xtime.Time
}

type YellowAndGreen struct {
	Period []*like.YellowGreenPeriod
}

type Bdf struct {
	Sid         int64
	DataSid     int64
	ImageSid    int64
	SchoolCount int
}

type Restart2020 struct {
	Sid        int64
	LotterySid string
	Cid        int64
	Vid        int64
}

type Special struct {
	SidOne      int64
	ExpireOne   int64
	SidTwo      int64
	ExpireTwo   int64
	SidThree    int64
	ExpireThree int64
}

type Shad struct {
	Sid        int64
	SuitExpire int64
	LotterySid string
	Cid        int64
	Vid        int64
}

type UpAct struct {
	WhiteMid []int64
	Image    string
}

type SpringCardAct struct {
	Sid           int64
	InviteSid     int64
	LotterySid    string
	ArcVid        int64
	CardA         int64
	CardB         int64
	CardC         int64
	CardD         int64
	CardE         int64
	CardF         int64
	InviteTimesID int64
	TimesID       int64
	NumLimit      int
	Cid           int64
}

type Vogue struct {
	Sid       string
	PrizeCost int64
}

type AnnualVoting struct {
	Max int64
	Num int64
}

type Cron struct {
	ResAudit        string
	SelectArc       string
	EntV2           string
	Bdf             string
	SpecialArc      string
	SubRule         string
	GiantV4         string
	AllTask         string
	AllAwards       string
	BwsBluetoothUps string
}

type Timemachine struct {
	Redis          *redis.Config
	TypesTick      time.Duration
	DataTick       time.Duration
	FlagTick       time.Duration
	EventSid       int64
	TagSid         int64
	RegionSid      int64
	FlagSid        int64
	Table          string
	LogDate        string
	Vid            int64
	EndLottery     xtime.Time
	LotteryNew     string
	LotteryOld     string
	PublishTimeout int64
	Gift           struct {
		GiftName string
		GiftType int
		ImgURL   string
	}
}

type Ent struct {
	Tick       time.Duration
	Vid        int64
	VidV2      int64
	UpSids     []int64
	SecondSids []int64
}

type Bnj2020 struct {
	Sid              int64
	MaxValue         int64
	Stime            xtime.Time
	BlockGame        int
	BlockGameAction  int
	BlockMaterial    int
	FinalTaskID      int64
	TimelinePic      string
	H5TimelinePic    string
	ShareTimelinePic string
	TaskTick         time.Duration
	DecreaseCD       int32
	AwardEndTime     xtime.Time
	NormalList       []string
	SpecialList      *struct {
		Good     []string
		GoodDesc []string
		Bad      []string
		BadDesc  []string
	}
	RareList []*Bnj20Material
	Info     []*struct {
		Name         string
		Pic          []string
		H5Pic        []string
		Detail       string
		H5Detail     string
		SharePic     string
		H5SharePic   string
		DynamicPic   string
		H5DynamicPic string
		Aids         []*Bnj20Aid
		Publish      xtime.Time
		Increase     []*IncreaseBnj20
		Decrease     []*DecreaseBnj20
	}
	Award []*Bnj20Award
	Live  *struct {
		Source     int64
		ExpireTime int64
		Items      []*struct {
			RewardID  int64
			StartTime int64
			Type      int
			Num       int
			ExtraData *struct {
				Type   string
				Value  int64
				Roomid int64
			}
		}
	}
}

type DecreaseBnj20 struct {
	Left  int
	Right int
	Value int64
	Type  int
}

type IncreaseBnj20 struct {
	P   float64
	IDs []int64
}

type Bnj20Aid struct {
	Aid        int64
	RcmdReason string
}

type Bnj20Award struct {
	ID           int64
	Name         string
	Pic          string
	CardPic      string
	Type         int
	LinkURL      string
	IsHide       int
	Count        int64
	SourceID     string
	SourceExpire int64
	TaskID       int64
}

type Bnj20Material struct {
	ID       int64
	Pic      string
	H5Pic    string
	Name     string
	Desc     string
	SharePic string
	CardPic  string
	Publish  xtime.Time
	TaskID   int64
}

type Box struct {
	Sid int64
}

type SteinV2 struct {
	Vid  int64
	Tick time.Duration
}

type Taaf struct {
	Sid      int64
	Vid      int64
	SidV2    int64
	Tick     time.Duration
	LikeTick time.Duration
}

type Staff struct {
	PicSid     int64
	Coupon     string
	SuitExpire int64
}

type Scholarship struct {
	Sid                  int64
	StepCurr             []int64
	StudyLikeTask        int64
	OtherLikeTask        int64
	LotterySid           int64
	GrantPid             int64
	GrantExpire          int64
	SignupTask           int64
	Stime                int64
	JoinTask             int64
	SignupLimit          int64
	AmountLimit          int64
	LikeNumLimit1        int64
	LikeNumLimit2        int64
	MallCouponId         string
	OtherSid             []int64
	FirstPrize           int64
	SecondPrize          int64
	ThirdPrize           int64
	CertificateVID       int64
	CertificateLimitNum1 int
	CertificateLimitNum2 int
	CertificateLimitNum3 int
	AllLikeNum           int64
	ArcVid               int64
	LikeSid              int64
	CountLikeID          int64
	CountStudyLikeID     int64
}

type AwardRule struct {
	Sid         int64
	NoRule      bool
	Type        string
	AwardID     int64
	AwardExpire int64
	Stime       int64
	Etime       int64
	LimitExpire int32
	ExtraSid    int64
}

// Brand 品牌
type Brand struct {
	CouponBatchToken       string
	ResourceBatchToken     string
	CouponBatchToken2      string
	ResourceAppkey         string
	CouponExperienceRemark string
	QPSLimitResourceCoupon int64
	QPSLimitExpire         time.Duration
}

// HandWrite 手书活动
type HandWrite struct {
	GodAllMoney     int
	TiredAllMoney   int
	NewAllMoney     int
	Sid             string
	ActStart        int64
	ActEnd          int64
	ActPlatCounter  string
	ActPlatActivity string
	AwardCoinLimit  int64
}

// HandWrite2021 手书2021
type HandWrite2021 struct {
	GodAllMoney int64
	Tired1Money int64
	Tired2Money int64
	Tired3Money int64
	Prefix      string
}

// GameHoliday 游戏假期
type GameHoliday struct {
	ActPlatCounter  string
	ActPlatActivity string
	AwardLikeLimit  int64
	Sid             string
	Vid             int64
}

// S10Answer
type S10Answer struct {
	BeginTime        int64
	EndTime          int64
	QuestionInterval int32
	CanPendantCount  int64
	DetailCron       string
	PendantID        int64
	PendantExpire    int64
	AnswerActExpire  time.Duration
	AnswerRound      []*AnswerRound
	AnswerPercent    []*AnswerPercent
}

// S10Answer AnswerRound
type AnswerRound struct {
	RoundDate int64
	BaseID    int64
}

// S10Answer AnswerPercent
type AnswerPercent struct {
	RightBegin   int64
	RightEnd     int64
	PercentBegin float64
	PercentEnd   float64
}

// S10Contribution s10 征稿活动
type S10Contribution struct {
	ActPlatCounter  string
	ActPlatActivity string
	AwardLikeLimit  int64
	LotteryID       string
	TotalVid        int64
	DayVid          int64
	Sid             int64
	IsWinSN         int
}

// Selection 年度动画评选
type Selection struct {
	BeginTime          int64
	EndTime            int64
	JoinTime           int32
	FilterLevel        int32
	ProductRole        []int64
	LimitWords         int
	VoteCategoryExpire time.Duration
	VoteCategoryCron   string
	VoteBegin          int64
	VoteEnd            int64
	VoteJoinTime       int32
	VoteSwitch         int
	NewKeyTime         int64
	ShowVotes          int
	VoteStage          []*like.SelCategory
}

// Amusement 娱乐之王
type Amusement struct {
	ListPs    int
	ShowVotes int
}

// UpWholeActive
type UpWholeActive struct {
	ParentSid int64
	ExtraSids map[string]int64
	ExtraNum  int64
}

// DoubleEleven
type DoubleEleven struct {
	ChannelListID int64
	VideoListVid  int64
}

// WinterStudy
type WinterStudy struct {
	PushSid         int64
	BeginTime       int64
	EndTime         int64
	ProgressUseDB   int
	ActPlatActivity string
	HistoryCounter  string
	ShareCounter    string
	UploadCounter   string
	ClockInCount    int64
}

// Remix 鬼畜活动
type Remix struct {
	Sid      int64
	DaySid   int64
	TaskID   int64
	AllMoney int64
}

// Dubbing 配音活动
type Dubbing struct {
	Sid     int64
	DaySid  int64
	TaskID  int64
	SidList []int64
}
type Stein struct {
	Sid        int64
	Stime      int64
	LikeCount  int64
	SuitPid    int64
	SuitExpire int64
}

// Star .
type Star struct {
	Sid      int64
	JoinSid  int64
	JoinLid  int64
	DownLid  int64
	FanLimit int64
}

type StarSpring struct {
	Sid           int64
	ReserveSid    int64
	CurrID        int64
	FollowerLimit int64
}

// DataBus multi databus collection.
type DataBus struct {
	LiveItemPub            *databus.Config
	BnjPub                 *databus.Config
	BnjAwardPub            *databus.Config
	SignedInPub            *databus.Config
	ActPlatPub             *databus.Config
	FreeFlowPub            *databus.Config
	NewYear2021Pub         *databus.Config
	RewardsAwardSendingPub *databus.Config
}

// Host remote host.
type Host struct {
	Sports   string
	QqNews   string
	Activity string
	Message  string
	APICo    string
	Mall     string
	LiveCo   string
	Dynamic  string
	ShowCo   string
	Comic    string
	VipBuy   string
	Pay      string
	MerakCo  string
}

type Live struct {
	Source        int64
	HeadRewardID  int64
	PropsRewardID int64
	HeadType      int64
	PropsType     int64
	ExpireTime    gtime.Time
	Num           int64
}

// Rule   rule config.
type Rule struct {
	GuessCount       int
	MaxGuessCoin     int64
	SuitPids         []int64
	SuitExpire       int64
	TickQq           time.Duration
	QqTryCount       int
	DTimeout         time.Duration
	QqStartTime      string
	QqEndTime        string
	QqYear           string
	PlayerYear       string
	DialectTags      []int64
	DialectRegions   []int32
	DialectSid       int64
	SpecialSids      []int64
	Spylike          int64
	LotteryActID     int64
	MatchLotteryID   int64
	S8Sid            int64
	S8ArcSid         int64
	S8ArtSid         int64
	KingStorySid     int64
	TmMids           []int64
	Bucket           string
	GrantPid         int64
	GrantExpire      int64
	QuestionSid      int64
	QuestionCD       int64
	QuestionLimit    int64
	OpenDynamic      bool
	TenDailys        map[string]int64
	TenImage         map[string]string
	TenH5Image       map[string]string
	TenCoupon        string
	TenCouponExpire  int32
	YellowSid        int64
	GreenSid         int64
	YeGrExpire       int32
	GuessPercent     float32
	GuessMaxOdds     float32
	MatchSids        []int64
	PaySuitExpire    int64
	AwardRule        []*AwardRule
	PollManageMid    []int64
	InterLottSids    []string
	InterQuestionIds []int64
	LimitArcSid      int64
	S9QuesSid        int64
	S9Right          int
	S9SuitID         int64
	S9SuitExpire     int64
	S9CacheExpire    int32
	S9Guess          *S9Guess
	AutoSuitIDs      map[string]int64
	SpecReserveSids  map[string]int64
	// 新星计划活动天数
	NewstarDays      time.Duration
	NewstarStop      int64
	NewstarInviteMax int64
	NewstarRole      int32
	TokenSalt        string
	ActWhiteList     []int64
}

// S9Guess s9 suits.
type S9Guess struct {
	SeasonID    int64
	SuitIDs     []int64
	SuitsExpire int64
	Stime       int64
	Etime       int64
	LimitExpire int32
	LimitCount  int
}

// Image act conf
type Image struct {
	Sid                  int64
	TenSid               int64
	TenTaskID            int64
	GrantPid             int64
	AwardEndTime         int64
	StudyTaskSid         int64
	AppointSid           int64
	AppointLid           int64
	SpecTaskID           int64
	LotteryID            int64
	LotteryName          string
	LotteryAwardAmount   int64
	LotteryAmount        int64
	LotteryCoinOne       string
	LotteryCoinOneAmount float64
	LotteryCoinTwo       string
	LotteryCoinTwoAmount float64
	LotteryCurrOne       string
	LotteryCurrOneAmount int64
	MikuGrantEndTime     int64
	MikuHeadEndTime      int64
	MikuLiveEndTime      int64
	MikuSid              int64
	MikuGrantPid         int64
	MikuGrantExpire      int64
	MikuStepCurr         []int64
	YearSid              int64
	YearLevel            int32
	YearAward            map[string]int64
	NewTask              map[string]int64
	SpecialJsonList      map[string]string
}

type Eleven struct {
	ElevenSid   int64
	LotteryID   int64
	AmountLimit int64
	LotteryName string
	ArcVid      int64
	CoinNum     float64
}

// Bws .
type Bws struct {
	VipMid          int64
	VipDate         string
	NormalMid       int64
	NormalDate      string
	IsTest          int64
	IsVip           int64
	AdminMids       []int64
	AwardMids       []int64
	LotteryMids     []int64
	LotteryAids     []int64
	SuitExpire      int64
	InitAchieveBids []int64
	SpecialBid      int64
	Bws2018Bid      int64
	NewBid          int64
	Bws2019         []int64
	SpecAcheives    map[string]int64
	VipCardAid      int64
	VipCardIDs      []int64
	VipCardStime    int64
	VipCardEtime    int64
	Bws2019Guang    int64
	Bws2019Shang    int64
	Bws2020Bid      int64
	// Bws202012Bid 2020 广州场
	Bws202012Bid  int64
	Bws2020VoteTs int64
	// StockAwardID 实物奖励id
	StockAwardID int64
	// StockAwardID2 卡券id
	StockAwardID2 int64
	// StockAwardCouponID 优惠券奖励id
	StockAwardCouponID int64
	// VipBuyToken 会员购优惠券
	VipBuyToken    string
	Bws2020GamePid int64
	// 抽中实物奖的概率
	Bws20Rand int
	// 抽中实物奖1的概率
	Bws20Rand1 int
	// GameNeedHeart 游戏需要的体力
	GameNeedHeart map[string]int64
	// BwsUpsCatchNeedStar 捕获up主可获星
	BwsUpsCatchNeedStar map[string]int64
	// RankStopHour 停止进入排行榜时间
	RankStopHour int
	// LotteryUsed 一次抽奖消耗奖券数
	LotteryUsed int64
	// RankTop50 排名50
	RankTop50 int64
	// RankTop1000 排行1000
	RankTop1000 int64
	// DefaultHeart 默认体力
	DefaultHeart int64
	// VipHeart vip体力
	VipHeart int64
	// CouponWinOnce 奖券只能抽到1次
	CouponWinOnce int64
	// WhiteMid 白名单mid
	WhiteMid []int64
}
type TicketParam struct {
	AppSecret               string
	AppKey                  string
	ProjectId               int64
	Year                    int
	SearchByIdUrl           string
	VipTag                  string
	ScreenNameMap           map[string]int64
	ShortExpireTime         int32
	OpenCache               int
	DefaultTicketExpireTime int64
	BackUpStckIds           []int64
}

type ArticleList struct {
	JoinSid int64
	ListSid int64
}

// Interval .
type Interval struct {
	NewestSubTsInterval time.Duration
	PullArcTypeInterval time.Duration
	ActSourceInterval   time.Duration
	TmInternal          time.Duration
	QuestionInterval    time.Duration
}

// Prom prom .
type Prom struct {
	LIBClient      *prom.Prom
	LIBClientState *prom.Prom
	APIClient      *prom.Prom
	HTTPServer     *prom.Prom
}

// MySQL define MySQL config
type MySQL struct {
	Like *sql.Config
	Bnj  *sql.Config
}

// ReloadInterval define reolad config
type ReloadInterval struct {
	Jobs   time.Duration
	Notice time.Duration
	Ad     time.Duration
}

// RPCClient2 define RPC client config
type RPCClient2 struct {
	Tag     *rpc.ClientConfig
	Article *rpc.ClientConfig
}

// Redis struct
type Redis struct {
	*redis.Config
	Store                 *redis.Config
	Cache                 *redis.Config
	Expire                time.Duration
	MatchExpire           time.Duration
	FollowExpire          time.Duration
	UserAchExpire         time.Duration
	UserPointExpire       time.Duration
	AchCntExpire          time.Duration
	HotDotExpire          time.Duration
	RandomExpire          time.Duration
	StochasticExpire      time.Duration
	ResetExpire           time.Duration
	RewardExpire          time.Duration
	UserTaskExpire        time.Duration
	TaskStateExpire       time.Duration
	UserCurrencyExpire    time.Duration
	SingleCurrencyExpire  time.Duration
	AnswerExpire          time.Duration
	GuessExpire           time.Duration
	EsLikesExpire         time.Duration
	LikeTotalExpire       time.Duration
	LikeMidTotalExpire    time.Duration
	OnGoingActivityExpire time.Duration
	LikeTokenExpire       time.Duration
	LotteryIPExpire       time.Duration
	LotteryExpire         time.Duration
	LotteryWinListExpire  time.Duration
	QPSLimitExpire        time.Duration
	LotteryTimesExpire    time.Duration
	AwardSubjectExpire    time.Duration
	SubRuleExpire         time.Duration
	WxLotteryLogExpire    time.Duration
	WxRedDotExpire        time.Duration
	AppstoreExpire        time.Duration
	BwsOnlineExpire       time.Duration
	BwsOnlineUserExpire   time.Duration
	NewstarExpire         time.Duration
	BwsbluetoothExpire    time.Duration
	SignedExpire          time.Duration
	ActArcsExpire         time.Duration
	BwsOfflineUserExpire  time.Duration
	BwsRankUserExpire     time.Duration

	CollegeMidCollegeExpire time.Duration
	ActExpire               time.Duration
	FirstShareExpire        time.Duration
	InviteTokenExpire       time.Duration
	LightVideoCountExpire   time.Duration

	RankConfigExpire         time.Duration
	PageExpire               time.Duration
	UserYearReport2020Expire time.Duration

	MidCardExpire                    time.Duration
	SevenDayExpire                   time.Duration
	OneDayExpire                     time.Duration
	FiveMinutesExpire                time.Duration
	SpringFestivialActivityEndExpire time.Duration

	FitPlanListExpire       time.Duration
	FitPlanDetailByIdExpire time.Duration

	SummerCampCourseListExpire       time.Duration
	SummerCampUserCostExpire         time.Duration
	SummerCampUserExchangeFlagExpire time.Duration
	RewardConfListExpire             time.Duration
}

// Memcache struct
type Memcache struct {
	Like             *memcache.Config
	LikeExpire       time.Duration
	LikeIPExpire     time.Duration
	PerpetualExpire  time.Duration
	ItemExpire       time.Duration
	SubStatExpire    time.Duration
	ViewRankExpire   time.Duration
	SourceItemExpire time.Duration
	QqExpire         time.Duration
	BwsExpire        time.Duration
	ProtocolExpire   time.Duration
	KfcExpire        time.Duration
	KfcCodeExpire    time.Duration
	GrantPidExpire   time.Duration
	TaskExpire       time.Duration
	CurrencyExpire   time.Duration
	QuestionExpire   time.Duration
	LastLogExpire    time.Duration
	RegularExpire    time.Duration
	LotteryExpire    time.Duration
	ActMissionExpire time.Duration
	LikeActExpire    time.Duration
	UserCheckExpire  time.Duration
	VogueConfExpire  time.Duration
	VogueGoodsExpire time.Duration
	VogueTaskExpire  time.Duration

	McReserveOnlyExpire time.Duration
}

type bnj2019 struct {
	ActID         int64
	SubID         int64
	GameCancel    int64
	TimelinePic   string
	H5TimelinePic string
	Start         xtime.Time
	Reward        []*bnj.Reward
	Info          []*struct {
		Nav      string
		Pic      string
		H5Pic    string
		Aid      int64
		Detail   string
		H5Detail string
		Nickname string
		Publish  xtime.Time
	}
}

// Lottery ...
type Lottery struct {
	NewLotteryTime      int64
	NewLotteryTime100   int64
	NewLotteryTime50    int64
	NewLotteryTime75    int64
	SpyScore            int32
	FigerScore          int32
	EntityContent       string
	CoinContent         string
	MemberContent       string
	MemberCouponContent string
	CouponContent       string
	GrantContent        string
	AppKey              string
	AppToken            string
	Pay                 *Pay
	VipBuy              *VipBuy
	NotifyCode          string
	MessageTitle        string
	ActivityText        string
	AddressText         string
	CouponText          string
	AddressLink         string
	PayOneYear          string
	CouponLink          string
	NewLotterSid        []string
}

// Pay ...
type Pay struct {
	Token      string
	CustomerID string
	PayHost    string
}

// VipBuy ...
type VipBuy struct {
	SourceID         int64
	SourceActivityID string
}

// Funny 搞笑迎新大会
type Funny struct {
	ActPlatCounter  string
	ActPlatActivity string
	AwardLikeLimit  int64
	Sid             string
	Vid             int64
}

type Knowledge struct {
	Mid           int64
	Sid           string
	ActiveSid     []int64
	LiveStartTime map[string]int64
	LiveEndTime   map[string]int64
	ActivityEnd   map[string]int64
}

type System struct {
	CORPID                 string
	WXCreateTokenUrl       string
	WXGetUserDetailUrl     string
	WXGetUserUserIDUrl     string
	WXCreateJSAPITicketUrl string
	OACreateTokenUrl       string
	OAClient               string
	OASecret               string
	OAGetAllUsersInfoUrl   string
	CORPSecret             map[string]string
	NotificationUrl        string
}

type Party2021 struct {
	UserExtra []string
}

type AsyncReserveConfig struct {
	Topic       string
	Switch      int
	Concurrency int
	Color       string
}

type ActPlatConfig struct {
	Topic string
}

type UpActReserveConfig struct {
	Topic string
}

type CardsComposeConfig struct {
	Topic string
}

type UpActReserveAudit struct {
	BizID1 int64
	BizID2 int64
	NetID  int64
}

type UpActReserveCreateConfig struct {
	PlayRange       int64           // 前后半小时内开播上线
	PLayContinue    int64           // 直播稳定播放5分钟推流
	LiveMaxNumLimit int64           // 直播创建预约数量限制
	ArcMaxNumLimit  int64           // 稿件创建预约数量限制
	ForceAuditFrom  map[string]bool // 开启强制敏感词审核渠道
	ForceAuditStime int64           // 开启强制敏感词审核开始时间
	ForceAuditEtime int64           // 开启强制敏感词审核结束时间
}

type UpActReserveRelationInfo4Live struct {
	GCacheLen int
}

type ActDomainConf struct {
	ExpireSecond int64
}

type UpActReserveAuthConf struct {
	EMFLevelThreshold int64
	UpGroup           int64
	ArcUpGroup        int64
	LiveUpGroup       int64
}

var (
	noCancelReserveMap *atomic.Value
	s10CoinCfg         *atomic.Value
	s10Matches         *atomic.Value
	s10TimePeriod      *atomic.Value
)

func init() {
	noCancelReserveMap = new(atomic.Value)

	m := make(map[string]bool, 0)
	storeNoCancelReserveMap(m)

	s10CoinCfg = new(atomic.Value)
	coinCfg := new(S10CoinConfig)
	storeS10CoinCfg(coinCfg)
	s10Matches = new(atomic.Value)
	s10MatchesCfg := new(s10.MatchesConf)
	storeS10MatchesCfg(s10MatchesCfg)
	s10TimePeriod = new(atomic.Value)
	s10TimePeriodCfg := new(s10.S10TimePeriod)
	storeS10TimePeriodCfg(s10TimePeriodCfg)
}

func storeS10CoinCfg(cfg *S10CoinConfig) {
	s10CoinCfg.Store(cfg)
}

func LoadS10CoinCfg() *S10CoinConfig {
	return s10CoinCfg.Load().(*S10CoinConfig)
}

func storeS10MatchesCfg(cfg *s10.MatchesConf) {
	s10Matches.Store(cfg)
}

func LoadS10MatchesCfg() *s10.MatchesConf {
	return s10Matches.Load().(*s10.MatchesConf)
}

func storeS10TimePeriodCfg(cfg *s10.S10TimePeriod) {
	s10TimePeriod.Store(cfg)
}

func LoadS10TimePeriodCfg() *s10.S10TimePeriod {
	return s10TimePeriod.Load().(*s10.S10TimePeriod)
}

func storeNoCancelReserveMap(m map[string]bool) {
	noCancelReserveMap.Store(m)
}

func LoadNoCancelReserveMap() map[string]bool {
	return noCancelReserveMap.Load().(map[string]bool)
}

// Init conf.
func Init() (err error) {
	return conf.Init(load)
}

func load() (err error) {
	var (
		tmpConf *Config
	)
	if err := conf.LoadInto(&tmpConf); err != nil {
		return err
	}
	*Conf = *tmpConf
	storeNoCancelReserveMap(tmpConf.NoCancelReserve)
	storeS10CoinCfg(tmpConf.S10CoinCfg)
	storeS10MatchesCfg(Conf.S10Matches)
	storeS10TimePeriodCfg(Conf.S10TimePeriod)
	tool.RestoreLimiters(tmpConf.CustomLimiters)
	tool.RestoreCbsByCfg(tmpConf.CircuitBreaker)
	return
}
