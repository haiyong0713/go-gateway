package conf

import (
	"time"

	"go-gateway/app/web-svr/activity/admin/component/boss"
	"go-gateway/app/web-svr/activity/tools/lib/conf"

	"go-gateway/app/web-svr/activity/admin/model/component"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/bfs"
	"go-common/library/database/elastic"
	"go-common/library/database/orm"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"
)

var (
	// Conf .
	Conf = &Config{}
)

// Config def.
type Config struct {
	Auth       *permit.Config
	HTTPServer *bm.ServerConfig
	HTTPClient *bm.ClientConfig
	EsClient   *elastic.Config
	ORM        *orm.Config
	TIDBORM    *orm.Config
	// db
	MySQL        *MySQL
	Export       *MySQL
	RewardsMySQL *sql.Config
	Log          *log.Config
	Tracer       *trace.Config
	Host         *Host
	// tag rpc client
	TagRPC             *rpc.ClientConfig
	ArcClient          *warden.ClientConfig
	AccClient          *warden.ClientConfig
	TagClient          *warden.ClientConfig
	ActPlatClient      *warden.ClientConfig
	SilverBulletClient *warden.ClientConfig
	ThumbupClient      *warden.ClientConfig
	ArtClient          *warden.ClientConfig
	ActPlatAdminClient *warden.ClientConfig
	TunnelClient       *warden.ClientConfig
	TagGRPC            *warden.ClientConfig
	VipClient          *warden.ClientConfig
	FlowControlClient  *warden.ClientConfig
	ActClient          *warden.ClientConfig
	// Elastic
	Elastic *elastic.Config
	Bnj     struct {
		Lid     int64
		Aids    []int64
		RoomID  int64
		Indexes []string
		Start   time.Time
		End     time.Time
		SidNew  int64
		Pub     *databus.Config
	}
	Bws struct {
		StartDay   int64
		EndDay     int64
		StartHour  int
		EndHour    int
		ReserveSid int64
		Bid        int64
	}
	Lottery *Lottery
	Up      *Up
	// redis
	Redis         *Redis
	Alarm         *Alarm
	Wechat        *WeChat
	VogueActivity *VogueActivity

	TunnelPush  *TunnelPush
	TunnelGroup *TunnelGroup
	BFS         *bfs.Config
	Reserve     *Reserve

	S10PointShopRedis *redis.Config
	S10MySQL          *sql.Config
	S10CacheExpire    *S10CacheExpire
	S10PointCostMC    *memcache.Config
	S10Mail           *S10Mail
	S10General        *S10General
	Rank              *Rank
	Boss              *boss.Config
	Notifier          component.CorpWeChat
	Subject           *Subject
	Cards             *Cards
	ActDomainConf     *ActDomainConf
	GaoKaoAnswer      *GaoKaoAnswer
}

type GaoKaoAnswer struct {
	SpitTag string
	BaseID  []int64
}

type FawkesConf struct {
	Host            string
	Env             string
	Operator        string
	AddUrl          string
	GetUrl          string
	ListUrl         string
	AppKey          string
	Business        string
	Description     string
	ItemGroupName   string
	ItemDescription string
	ItemKey         string
}
type ActDomainConf struct {
	DomainListUrl   string
	DefaultPageNo   int
	DefaultPageSize int
	APIHost         string
	FawkesConf      *FawkesConf
}

type Cards struct {
	Activity string
}

// Subject ...
type Subject struct {
	AuditGroupID int
}

// Rank
type Rank struct {
	ArchiveLength int
	Reviewers     []string
	Admin         []string
}
type S10General struct {
	RedeliveryHost string
	SubTabSwitch   bool
	Robins         []int64
}

type S10Mail struct {
	FilePath string
	MailInfo *component.EmailInfo
}

type S10CacheExpire struct {
	SignedExpire              xtime.Duration
	TaskProgressExpire        xtime.Duration
	RestPointExpire           xtime.Duration
	CoinExpire                xtime.Duration
	PointExpire               xtime.Duration
	LotteryExpire             xtime.Duration
	ExchangeExpire            xtime.Duration
	RoundExchangeExpire       xtime.Duration
	RestCountGoodsExpire      xtime.Duration
	RoundRestCountGoodsExpire xtime.Duration
	PointDetailExpire         xtime.Duration
}

type Reserve struct {
	Notify []string
}

type Up struct {
	SenderUid     uint64
	PassContent   string
	UnPassContent string
	ActSenderUid  uint64
}

// Redis struct
type Redis struct {
	*redis.Config
	Store *redis.Config
	Cache *redis.Config
}

// 报警
type Alarm struct {
	WeChatToken       string
	WeChatSecret      string
	WeChatShareHost   string
	WeChatMonitorTick xtime.Duration
	Username          string
	AlarmTag          string
}

type WeChat struct {
	AppId  string
	Secret string
}

// vogueActivity
type VogueActivity struct {
	ActPlatActivity   string
	ScoreInitialValue int
	Active            int
}

type Lottery struct {
	AppKey             string
	AppToken           string
	VipAppKey          string
	MoneyLimit         []string
	SenderMidLimit     []string
	NumLimit           []string
	Reviewers          []string
	MailInfo           *component.EmailInfo
	PublicKey          string
	AuditSubject       string
	AuditRejectSubject string
	AuditPassSubject   string
	EditLink           string
}

type TunnelPush struct {
	TunnelBizID    int64
	DynamicCardTag string
	Index          []string
	Letter         []string
	Dynamic        []string
}

type TunnelGroup struct {
	Source string
}

// Host remote host
type Host struct {
	API     string
	SHOW    string
	MNG     string
	Dynamic string
	MerakCo string
}

// MySQL define MySQL config
type MySQL struct {
	Lottery *sql.Config
}

func load() (err error) {
	var (
		tmpConf *Config
	)
	err = conf.LoadInto(&tmpConf)
	*Conf = *tmpConf
	return
}

// Init int config
func Init() error {
	return conf.Init(load)
}
