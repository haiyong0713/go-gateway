package conf

import (
	xtime "time"

	"go-gateway/app/web-svr/activity/job/component/boss"
	"go-gateway/app/web-svr/activity/tools/lib/conf"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-common/library/time"
	"go-gateway/app/web-svr/activity/job/model/like"
	xmail "go-gateway/app/web-svr/activity/job/model/mail"
	"go-gateway/app/web-svr/activity/job/model/s10"
)

var (
	Conf = &Config{}
)

// Config so config
type Config struct {
	// interface Log
	Log *log.Config
	//HTTPClient
	HTTPClient      *bm.ClientConfig
	HTTPClientComic *bm.ClientConfig
	HTTPFate        *bm.ClientConfig
	HTTPClientBFS   *bm.ClientConfig
	// BM
	BM *bm.ServerConfig
	// rpc
	CoinRPC *rpc.ClientConfig
	ActRPC  *rpc.ClientConfig
	// grpc
	ActClient         *warden.ClientConfig
	AccClient         *warden.ClientConfig
	ArcClient         *warden.ClientConfig
	FlowControlClient *warden.ClientConfig
	ArtClient         *warden.ClientConfig
	CoinClient        *warden.ClientConfig
	SuitClient        *warden.ClientConfig
	TagClient         *warden.ClientConfig
	LiveXRoomClient   *warden.ClientConfig
	TagNewClient      *warden.ClientConfig
	CheeseClient      *warden.ClientConfig
	RelationClient    *warden.ClientConfig
	FavoriteClient    *warden.ClientConfig
	ActPlatClient     *warden.ClientConfig
	MemberClient      *warden.ClientConfig
	BbqTaskClient     *warden.ClientConfig
	VideoUpClient     *warden.ClientConfig
	GaiaLibClient     *warden.ClientConfig
	EsportsClient     *warden.ClientConfig
	SpaceClient       *warden.ClientConfig
	BGroupClient      *warden.ClientConfig
	TunnelClient      *warden.ClientConfig
	GarbClient        *warden.ClientConfig
	VideoClient       *warden.ClientConfig
	FliClient         *warden.ClientConfig
	TopicClient       *warden.ClientConfig
	Tracer            *trace.Config
	// DB
	MySQL *MySQL
	// mc
	Memcache *Memcache
	// redis
	Redis   *Redis
	GuRedis *GuRedis
	// databus
	ActSub                 *databus.Config
	BnjMainWebSvrBinlogSub *databus.Config
	ArticleLikeSub         *databus.Config
	ArticleCoinSub         *databus.Config
	ArticlePassSub         *databus.Config
	ArchiveSub             *databus.Config
	ArchiveBinLogSub       *databus.Config
	LiveFollowSub          *databus.Config
	ReplySub               *databus.Config
	LotteryAwardSub        *databus.Config
	FitTunnelPub           *databus.Config
	VipLotterySub          *databus.Config
	VipCardSub             *databus.Config
	OttVipLotterySub       *databus.Config
	BnjSub                 *databus.Config
	BnjAwardSub            *databus.Config
	CustomizeLotterySub    *databus.Config
	LiveItemPub            *databus.Config
	MessageDatabusPub      *databus.Config
	SubRuleStatSub         *databus.Config
	TunnelDatabusPub       *databus.Config
	TunnelGroupDatabusPub  *databus.Config
	ReserveNotifyPub       *databus.Config
	ReserveNotifySub       *databus.Config
	ActPlatPub             *databus.Config
	PredictTaskPub         *databus.Config
	S10GoodsCanalSub       *databus.Config
	S10PointShopRedis      *redis.Config
	S10MySQL               *sql.Config
	S10MC                  *memcache.Config
	S10CacheExpire         *S10CacheExpire
	S10CacheKey            *S10CacheKey
	S10PointCostMC         *memcache.Config
	S10General             *s10.S10General
	FreeFlowSub            *databus.Config
	GaiaRiskSub            *databus.Config
	NewYearReservePub      *databus.Config
	RewardsMySQL           *sql.Config

	BackupMQ                *redis.Config   `toml:"backup_mq"`
	Bnj2021ARSub            *databus.Config // AR游戏兑换落库
	Bnj2021LiveUser         *databus.Config // 拜年纪直播间用户观看时长达标记录落库
	Bnj2021LotteryRec       *databus.Config // 拜年纪直播间抽奖发放业务
	Bnj2021LiveDrawRec      *databus.Config // 拜年纪直播间抽奖发放业务
	Bnj2021LiveDrawARCoupon *databus.Config // 拜年纪直播间AR抽奖券业务
	RewardsAwardSendingSub  *databus.Config // 发奖平台异步奖励发放

	ActPlatHistorySub   *databus.Config // 任务平台notify流
	ActPlatHistorySpSub *databus.Config // 任务平台notify 特殊流
	LotteryAddtimesSub  *databus.Config // 抽奖增加次数流
	Databus             *DatabusV2
	LotteryDatabus      *LotteryDatabusCfg
	FitDatabusCfg       *FitDatabusCfg
	// Interval
	Interval *interval
	// Rule
	Rule *rule
	// Host
	Host *host
	// Elastic
	Elastic *elastic.Config
	// bnj
	Bnj2019 *bnj2019
	Bnj2020 *bnj2020
	// image
	Image *Image
	// fate special conf
	Fate *Fate
	// live e3 sub
	Live *Live
	//GuessMsg
	GuessImMsg ImMsg
	// Yellow green act
	YeGr *YeGr
	// Bws2019
	Bws2019 *Bws2019
	// stein
	Stein *Stein
	// lottery
	Lottery *Lottery
	// entertainment
	Ent *Ent
	// staff
	Staff          *Staff
	Taaf           *Taaf
	Restart2020    *Restart2020
	YellowAndGreen *YellowAndGrean
	MobileGame     *MobileGame
	// image v2
	ImageV2 *ImageV2
	// faction
	Faction *Faction
	Stupid  *Stupid
	// wx lottery
	WxLottery *WxLottery
	// handwrite
	HandWrite *HandWrite
	// Remix
	Remix *Remix
	// GameHoliday
	GameHoliday *GameHoliday
	// S10Contribution
	S10Contribution *Contribution
	// Selection
	Selection *Selection
	// 通用邮件配置
	Mail *RemixEmail
	// 预约通知配置
	Reserve *Reserve
	// College
	College *College
	// Share
	Share *Share
	// Dubbing
	Dubbing *Dubbing
	// Funny
	Funny   *Funny
	Column  *Column
	Acg2020 *Acg2020
	// S10 Answer
	S10Answer               *S10Answer
	Notifier                CorpWeChat
	NotifierForVote         CorpWeChat
	Knowledge               *Knowledge
	Rank                    *Rank
	NewYearReservePubConfig *ActivityReservePubConfig
	OperationSource         *OperationSource
	TunnelGroup             *TunnelGroup
	// Handwrite
	Handwrite2021            *Handwrite2021
	SpringFestival2021       *SpringFestival2021
	WechatToken              *WechatToken
	UpActReserveNotify       *UpActReserveNotify
	PushVerifyEmailReceivers *PushVerifyEmailReceivers
	UpActReserveAudit        *UpActReserveAudit
	ActDomainConfig          *ActDomainConfig
	PushVerifyUriConfig      *PushVerifyUriConfig
	PushVerifyUriResetConfig *PushVerifyUriResetConfig
	Vote                     *Vote
	FitJobConfig             *FitJobConfig
	FitTianMaCardContentConf *FitTianMaCardContentConf
	FitHotVideo              map[string][]int64
	GaoKaoAnswer             *GaoKaoAnswer
	Boss                     *boss.Config
	BwPark2021               *BwPark2021
	BindConfig               *BindConfig
	KnowledgeTask            *KnowledgeTask
	UpReservePushPub         *databus.Config
	Cpc100Config             *Cpc100Pv
	MissionTaskSub           *MissionTaskConsumer // 任务平台任务回调，用于处理任务活动用户的任务完成情况
	MissionConfig            *MissionConfig
	StockServerJobConf       *StockServerJobConf
}

type StockServerJobConf struct {
	RunCron      string
	SyncNum      int32
	SyncTimeGap  int64
	SyncPageSize int32
	SyncMaxLoop  int
	AckUrlMap    map[string]string
}

type MissionConfig struct {
	RefreshTickerSecond       int
	MakeUpReceiveRecordSecond int
}

type MissionTaskConsumer struct {
	Databus *databus.Config
	Railgun *railgun.SingleConfig
}

type Cpc100Pv struct {
	RefreshInterval      time.Duration
	PvUrl                string
	RideN                float64
	AddN                 int64
	TopicRefreshInterval time.Duration
	TopicIds             []int64
}

type KnowledgeTask struct {
	KnowTaskBatchNum    int64
	DelKnowCalcBatchNum int
	DelCron             string
	DelSleep            time.Duration
}

type BindConfig struct {
	RefreshTickerSecond int
}

type CorpWeChat struct {
	MentionUserIDs  []string
	MentionUserTels []string
	WebhookUrl      string
}

type Vote struct {
	DSItemsRefreshDurForNotEnd      time.Duration
	RealTimeRankRefreshDurForNotEnd time.Duration

	DSItemsRefreshDurForEndWithin90      time.Duration
	RealTimeRankRefreshDurForEndWithin90 time.Duration
}

type LotteryDatabusCfg struct {
	ActHistoryDatabus *databus.Config
	LotteryRailgun    *railgun.SingleConfig
}

type FitDatabusCfg struct {
	FitActivityHistorySub *databus.Config
	FitRailgun            *railgun.SingleConfig
}

// WechatToken ...
type WechatToken struct {
	LittleFlower string
}

type SpringFestival2021 struct {
	Activity string
}
type DatabusV2 struct {
	Target string `toml:"target"`
	AppID  string `toml:"app_id"`
	Token  string `toml:"token"`
	Topic  struct {
		BGroupMessage                       string
		LotteryAddTimes                     string
		UpActReserveRelation                string
		UpActReserve                        string
		UpActReserveRelationTableMonitor    string
		UpActReserveRelationChannelAudit    string
		UpActReserveLotteryUserReserveState string
	}
}

// Handwrite2021 ...
type Handwrite2021 struct {
	// EndStatisticsTime 活动统计结束时间
	EndStatisticsTime xtime.Time
	// LastPubTime 稿件最后发布时间
	LastPubTime xtime.Time
	// RankID 借用排行榜能力的id
	RankID int64
	// LastBatch 最后批次
	LastBatch int64
	// Sid 数据源id
	Sid int64
	// Coin 需要的硬币数量
	Coin int64
	// View1 需要的播放数量
	View1 int64
	// View2 需要的播放数量
	View2 int64
	// View3 需要的播放数量
	View3       int64
	GodAllMoney int64
	Tired1Money int64
	Tired2Money int64
	Tired3Money int64
	FilePath    string
	Mail        *HandWriteEmail
	Cron        string
	DataCron    string
	Prefix      string
}

// Rank ...
type Rank struct {
	TopLength     int64
	ArchiveLength int
	RankCron      string
	PublicKey     string
	NewRankCron   string
}

type Acg2020Task struct {
	Round       int
	StartTime   xtime.Time
	EndTime     xtime.Time
	StatTime    xtime.Time
	AwardTime   xtime.Time
	FinishCount int
	FinishScore int64
	Award       int
}

type Acg2020 struct {
	SID            int64
	Task           []*Acg2020Task
	DataFile       string
	CsvFile        string
	UpdateDuration time.Duration
	WxKey          string
	ExtraAward     []*struct {
		FinishTask int
		Award      int
	}
	Mail struct {
		Receiver []*xmail.Address
		Subject  string
		Content  string
	}
}

// Dubbing 配音
type Dubbing struct {
	ActivityStart      xtime.Time
	ActivityEnd        xtime.Time
	LastDay            int
	Sid                int64
	TaskID             int64
	ChildSid           []int64
	TopView            int
	TopRank            int
	ChildTopRank       int
	RankCron           string
	Mail               *DubbingEmail
	LastBatch          int64
	StartBatch         int64
	Money              int64
	FirstTime          xtime.Time
	FilePath           string
	ArchiveScoreExpire time.Duration
}

type Reserve struct {
	Notify []string
}

type S10CacheExpire struct {
	SignedExpire              time.Duration
	TaskProgressExpire        time.Duration
	RestPointExpire           time.Duration
	CoinExpire                time.Duration
	PointExpire               time.Duration
	LotteryExpire             time.Duration
	ExchangeExpire            time.Duration
	RoundExchangeExpire       time.Duration
	RestCountGoodsExpire      time.Duration
	RoundRestCountGoodsExpire time.Duration
	PointDetailExpire         time.Duration
	UserFlowExpire            time.Duration
}

type S10CacheKey struct {
	ESportsKey          string
	ContestListKey      string
	GuessMainDetailsKey string
}

// S10Answer
type S10Answer struct {
	MaxQuestion    int
	HourCron       string
	WeekCron       string
	BaseIDs        []int64
	UserInfoKey    []string
	FinishTime     map[string]string
	TimesNoRank    int64
	ScoreMaxNoRank int64
}

type GaoKaoAnswer struct {
	SpitTag    string
	ForeignId  []int64
	RunCron    string
	RandMethod []*struct {
		Code    int64
		AttrNum map[string]int
	}
}

type BwPark2021 struct {
	RunCron string
	SyncNum int32
	Bid     int32
}

type TunnelGroup struct {
	Source string
}

type OperationSource struct {
	InfoSids      []int64
	OperationSids []int64
	//UpdateTicker  xtime.Duration
}

// College 开学季
type College struct {
	// MidActivity 用户维度的活动id
	MidActivity string
	// MidFormula 用户维度的计算公式
	MidFormula string
	// MidTopLength 用户维度top榜
	MidTopLength int
	// MIDSID 用户维度活动id
	MIDSID int64
	// CollegeSID 学校维度活动id
	CollegeSID int64
	// ArchiveTopLength 稿件top榜
	ArchiveTopLength int
	// RankCron 排行榜更新时间
	RankCron string
	// VersionCron 版本更新时间更新时间
	VersionCron string
	// ArchiveCtime
	ArchiveCtime int64
	// VersionTest 是否开启测试
	VersionTest int64
	// ArchiveVideoState 稿件播放量
	ArchiveVideoState int32
	// VideoBonusPoint 超过1000额外加分
	VideoBonusPoint int64
	// BonusCron 额外加分脚本更新时间
	BonusCron string
	// ScoreCron 积分更新时间
	ScoreCron string
	// ActivityEnd 活动结束时间
	ActivityEnd xtime.Time
	// AutoAddScore 自动加积分
	AutoAddScore map[string]int64
	// ScoreAutoCron 自动加分脚本
	ScoreAutoCron string
}

// Share ...
type Share struct {
	// ShareURL
	ShareURL string
	// ShareLinkConf ...
	ShareLinkConf map[string]*SingleShareLinkConf
	// ShareCron
	ShareCron string
}

// SingleShareLinkConf .
type SingleShareLinkConf struct {
	Num        int64
	Token      string
	BaseUrl    string
	PrefixPath []string
	Hosts      []string
}

type WxLottery struct {
	TransDesc         string
	WithdrawEndTime   int64
	WithdrawStartHour int64
	PageCron          string
	PayCfg            *PayCfg
}

// HandWrite 手书
type HandWrite struct {
	Sid              int64
	CountMidArchive  int
	ArchiveCoinAward int32
	MidSQLMaxNum     int
	RankTopLength    int
	RankCron         string
	FavSyncCron      string
	FavID            int64
	FavMid           int64
	GodAllMoney      int
	TiredAllMoney    int
	NewAllMoney      int
	FilePath         string
	Mail             *HandWriteEmail
	RankLastBatch    int64
	RankStartBatch   int64
	ActivityStart    int64
	ActivityEnd      int64
	ActPlatActivity  string
	ActPlatFilter    string
	ActPlatCounter   string
}

// GameHoliday ...
type GameHoliday struct {
	Vid             int64
	ActPlatActivity string
	ActPlatCounter  string
	SyncCron        string
}

// Contribution ...
type Contribution struct {
	Sid                 int64
	DayVid              int64
	ActPlatActivity     string
	ActPlatCounter      string
	SyncDaySelectCron   string
	CalcSplitPeopleCron string
	UpDBCron            string
	TotalRankCron       string
	RankCount           int
	IsWinSN             int
	WinSnEnd            int64
	NoWinSnEnd          int64
	SnArchiveTime       int64
}

// Selection ...
type Selection struct {
	Sid                    int64
	CalcAssistanceCron     string
	VoteReportCron         string
	ResetSelectionVoteCron string
	VoteReportBegin        int64
	VoteReportEnd          int64
	VoteStage              []*like.SelCategory
	EmailReceivers         string
	TagSplit               string
	StopRiskVote           int
	ReportProduct          int
}

// Remix 鬼畜
type Remix struct {
	ActivityStart xtime.Time
	ActivityEnd   xtime.Time
	Sid           int64
	TaskID        int64
	ChildSid      []int64
	TopView       int
	TopRank       int
	ChildTopRank  int
	RankCron      string
	Mail          *RemixEmail
	LastBatch     int64
	StartBatch    int64
	Money         int64
	FilePath      string
}

// DubbingEmail 配音活动邮箱
type DubbingEmail struct {
	Host         string
	Port         int
	Address      string
	Pwd          string
	Name         string
	ToAddress    []*Address
	CcAddress    []*Address
	BccAddresses []*Address
}

// HandWriteEmail 手书活动邮箱配置
type HandWriteEmail struct {
	Host         string
	Port         int
	Address      string
	Pwd          string
	Name         string
	ToAddress    []*Address
	CcAddress    []*Address
	BccAddresses []*Address
}

// RemixEmail 鬼畜活动邮箱配置
type RemixEmail struct {
	Host         string
	Port         int
	Address      string
	Pwd          string
	Name         string
	ToAddress    []*Address
	CcAddress    []*Address
	BccAddresses []*Address
}

// Address 邮箱发送
type Address struct {
	Name    string
	Address string
}

type PayCfg struct {
	CustomerID   string
	MerchantCode string
	CoinType     string
	PayHost      string
	Token        string
	ActivityID   string
}

type Faction struct {
	Sids      []int64
	Limit     int
	Etime     xtime.Time
	BlockAids []int64
	BlockMids []int64
}

type Stupid struct {
	Sid        int64
	Sids       []int64
	Etime      int64
	Vid        int64
	LikeTaskID int64
}

type ImageV2 struct {
	Sid      int64
	DayLimit int
	Etime    xtime.Time
}

type MobileGame struct {
	LikeTaskID int64
	Vid        int64
}

type YelAndGreen struct {
	LikeTaskID        int64
	Vid               int64
	YingYuanView      int32
	YingYuanVote      int64
	YellowMid         int64
	YellowFid         int64
	GreenMid          int64
	GreenFid          int64
	GreenYingYuanSid  int64
	YellowYingYuanSid int64
	YingYuanVoteCron  string
}

type YellowAndGrean struct {
	YingYuanVoteCron string
	Period           []*like.YellowGreenPeriod
}

type Restart2020 struct {
	LikeTaskID int64
	Vid        int64
}

type Staff struct {
	Sid        int64
	PassTaskID int64
	LikeTaskID int64
}

type Ent struct {
	Vid       int64
	UpArcSids []struct {
		UpSid  int64
		ArcSid int64
	}
}

type Stein struct {
	Sid         int64
	LikeTaskID  int64
	OneViewRule int64
	OneLikeRule int64
	TwoViewRule int64
	TwoLikeRule int64
	LotteryID   int64
}

type Bws2019 struct {
	Bid            int64
	Limit          int
	VipCardCodeIDs []int64
	Stime          int64
	Etime          int64
	SpecBids       []int64
}

type YeGr struct {
	LotteryID      int64
	SuitAwardIDs   []int64
	SuitID         int64
	SuitExpire     int64
	SuitLimit      int
	CoinOneAwardID int64
	CoinTwoAwardID int64
}

type bnj2020 struct {
	PugvCount        int
	ComicCount       int
	PendantCount     int
	MallCount        int
	LiveCount        int
	Stime            xtime.Time
	GameCancel       int
	Sid              int64
	MidLimit         int64
	BlockGame        int
	BlockGameAction  int
	BlockMaterial    int
	MaxValue         int64
	TimelinePic      string
	H5TimelinePic    string
	ShareTimelinePic string
	Sender           int64
	Cron             string
	Message          []*struct {
		Start   xtime.Time
		Content string
	}
	Live *struct {
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

type bnj2019 struct {
	GameCancel    int
	LID           int64
	StartTime     xtime.Time
	TimelinePic   string
	H5TimelinePic string
	MsgSpec       string
	MidLimit      int64
	WxKey         string
	WxTitle       string
	WxUser        string
	Time          []*struct {
		Score    int64
		Second   int64
		Step     int
		WxMsg    string
		MsgTitle string
		MsgMc    string
		Msg      string
	}
	Message []*struct {
		Start   xtime.Time
		Title   string
		Content string
		Mc      string
	}
}

type interval struct {
	CoinInterval                 time.Duration
	QueryInterval                time.Duration
	ObjStatInterval              time.Duration
	ViewRankInterval             time.Duration
	KingStoryInterval            time.Duration
	QuestionInterval             time.Duration
	FateInterval                 time.Duration
	FateDataInterval             time.Duration
	GuessInterval                time.Duration
	GuessCron                    string
	UpListHisCron                string
	EntCron                      string
	SpringCron                   string
	QuestionCron                 string
	SubsRankCron                 string
	PoolCreatCron                string
	SteinListCron                string
	LoadGuessOidsCron            string
	LoadStaffCron                string
	TimingTaskCron               string
	BnjReserveCron               string
	AwardSubjectCron             string
	ImageCron                    string
	FactionCron                  string
	StupidCron                   string
	NewstarUpArcCron             string
	NewstarFinishCron            string
	NewstarIdentityCron          string
	ArticleDayCron               string
	ContestCron                  string
	ContestDetailCron            string
	ActRelationInfoCron          string
	ActSubjectInfoCron           string
	ActSubjectReserveIDsInfoCron string
	PushUpReserveVerifyCron      string
}

// MySQL is db config.
type MySQL struct {
	Like *sql.Config

	//mysql readonly slave config
	Read *sql.Config

	Bnj *sql.Config
}

// Redis config
type Redis struct {
	*redis.Config
	Cache           *redis.Config
	Store           *redis.Config
	Expire          time.Duration
	ImgUpExpire     time.Duration
	StupidExpire    time.Duration
	TaskStateExpire time.Duration
	TaskDataExpire  time.Duration
	RankDataExpire  time.Duration
}

// Redis config
type GuRedis struct {
	*redis.Config
	UserExpire time.Duration
	ListExpire time.Duration
}

// Memcache config
type Memcache struct {
	Like             *memcache.Config
	LikeExpire       time.Duration
	TimeFinishExpire time.Duration
	LessTimeExpire   time.Duration
}

type rule struct {
	BroadcastCid       int64
	BroadcastSid       int64
	ArcObjStatSid      int64
	ArtObjStatSid      int64
	KingStorySid       int64
	EleLotteryID       int64
	SpecBirthSid       int64
	SpecBirthLid       int64
	SpecBirthExpire    int32
	ThreeLotteryID     int64
	GuessSID           int64
	VipLotteryIDs      []*vipLottery
	VipLotteryExpire   int32
	LotteryAddRule     []*LotteryAddRule
	LikeAddLotterySids []int64
	GuessPercent       float64
	GuessBusiness      []int64
	GuessMaxOdds       float64
	DailyLikeSid       int64
	OttVipLottery      []*ottVipLottery
	NewstarArcTypes    []int32
	NewstarDays        time.Duration
	NewstarName        string
	ArticleDaySid      int64
	ArticleDayBegin    int64
	ArticleDayEnd      int64
	TunnelPushBizID    int64
}

type vipLottery struct {
	AppID     int64
	LotteryID int64
}

type ottVipLottery struct {
	Pid       int64
	LotteryID int64
}

// ArcLotteryRule .
type LotteryAddRule struct {
	Sid           int64
	LotteryID     int64
	NoLimit       int
	LimitDuration int // 0 all 1 hour 2 day
	LimitTimes    int
	Expire        int32
}

type Lottery struct {
	LotteryExpire time.Duration
	RedisClean    int64
	WinListCron   string
}

// Image .
type Image struct {
	Sid            int64
	AppointSid     int64
	AppointLid     int64
	TaskLikeID     int64
	TaskBeLikeID   int64
	TaskArchiveSID int64 //学习打卡活动id
	TaskArchiveID  int64 //学习打卡任务-投稿任务
	TaskMikuSid    int64 //初音未来十二周年活动id
	TaskMikuLikeID int64 //初音未来十二周年投稿任务
	TaskMikuCoinID int64 //初音未来十二周年投币任务
	TaskCoinID     int64
	TaskPassID     int64
	TenReplyID     int64
	YearTaskIDs    map[string]int64
	CommonTaskIDs  map[string]int64
	ArticleTaskIDs map[string]int64
}

type host struct {
	APICo    string
	Activity string
	ActCo    string
	MsgCo    string
	ApiVcCo  string
	MerakCo  string
	TbApi    string
	Comic    string
	Mall     string
	Manager  string
	Dynamic  string
}

// Fate .
type Fate struct {
	Stime  xtime.Time
	AppKey string
	Secret string
	Vid    int64
}

// Live .
type Live struct {
	SID        int64
	LID        int64
	Seids      []int64
	LiveExpire int32
}

type ImMsg struct {
	OfficialUID int64
	Content     string
}

type Bws struct {
	Bid           int64
	Stime         int64
	Etime         int64
	VipCardCodeID int64
}

type Taaf struct {
	Sidv2 int64
}

// Funny
type Funny struct {
	Vid                   int64
	ActPlatActivity       string
	ActPlatCounter        string
	SyncVideoCron         string
	FunnyVideoDuration    int64
	CaculatePartOne       string
	CaculatePartTwo       string
	ActSid                int64
	PartOneLikesLimit     int32
	PartTwoVideoNumLimit  int32
	PartTwoVideoViewLimit int32
	EmailReceivers        string
}

// Column
type Column struct {
	Sid            int64
	TriggerCron    string
	EmailReceivers string
}

type ActivityReservePubConfig struct {
	RuleType   int64
	ConfigName string
	RuleConfig ActivityReservePubConfigRuleTypeOne
}

type ActivityReservePubConfigRuleTypeOne struct {
	SIDs []int64
}

type Knowledge struct {
	Sid  string
	Cron string
}

type UpActReserveNotify struct {
	TmpID                string
	ArcTemplateID        string
	ResetTmpID           string
	LiveTemplateID       string
	LiveResetTempID      string
	LiveResetSenderUID   int64
	ArcResetSenderUID    int64
	LotteryReserveTmpID  string
	ResetSenderUID       int64
	PushUpVerifyID       string
	PushUpVerifySenderID int64
}

type PushVerifyEmailReceivers struct {
	EmailReceivers string
}

type UpActReserveAudit struct {
	BizID1 int64
	BizID2 int64
	NetID  int64
}

type PushVerifyUriConfig struct {
	AllUri     string
	AndroidUri string
	IphoneUri  string
	IpadUri    string
	WebUri     string
	Text       string
}

type PushVerifyUriResetConfig struct {
	AllUri     string
	AndroidUri string
	IphoneUri  string
	IpadUri    string
	WebUri     string
	Text       string
}

// Init init config.
func Init() (err error) {
	return conf.Init(load)
}

func load() (err error) {
	var (
		tmpConf = &Config{}
	)
	err = conf.LoadInto(&tmpConf)
	*Conf = *tmpConf
	return
}

type ActDomainConfig struct {
	SyncNum int
	RunCron string
}

type FitJobConfig struct {
	RunFlushCron                 string
	RunAwardSendCron             string
	RunTianMaCron                string
	RunSetMemberCron             string
	GetCounterResHasLimitCounter string
	GetCounterResNoLimitCounter  string
	PiFuDay                      int64
	AwardId                      int64
	DingYueActivityId            int64
	DaKaActivityId               int64
}

type FitTianMaCardContentConf struct {
	DefaultTemplateId int64
	ChangeTemplateId  int64
	Icon              string
	Link              string
	ButtonType        string
	ButtonText        string
	ButtonLink        string
	TianMaSTime       int
	TianMaEndSTime    int
	NeedPushWeekDay   int
}
