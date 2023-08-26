package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/database/orm"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	rpcquota "go-common/library/net/rpc/warden/ratelimiter/quota"
	"go-common/library/net/trace"
	"go-common/library/time"
	newConf "go-gateway/app/web-svr/activity/tools/lib/conf"
)

var (
	// Conf of config
	Conf = &Config{}
)

type Config struct {
	// db
	SqlCfg *sql.Config
	OrmCfg *orm.Config
	// log
	Log *log.Config
	// tracer
	Tracer *trace.Config
	// client
	HTTPReply            *bm.ClientConfig
	HTTPJob              *bm.ClientConfig
	HTTPClient           *bm.ClientConfig
	SysInformsHTTPClient *bm.ClientConfig
	// Warden Client
	LiveRoomGrpc   *warden.ClientConfig
	BGroupClient   *warden.ClientConfig
	TunnelV2Client *warden.ClientConfig
	FavClient      *warden.ClientConfig
	EspClient      *warden.ClientConfig
	ActivityClient *warden.ClientConfig
	AccClient      *warden.ClientConfig
	// GameTypes game types.
	GameTypes []*types
	// Host
	Host       Host
	TunnelPush *TunnelPush
	// tunnelBGroup
	TunnelBGroup *TunnelBGroup
	// tunnelCardMsg
	TunnelCardMsg *TunnelCardMsg
	Memcached     *memcache.Config
	// Auto subscribe redis
	CommonRedis *redis.Config

	RpcLimiter *rpcquota.Config

	// System Informs
	SysInforms SysInforms

	// Push
	Push *Push

	// AppConfig
	AppConfig *bm.App

	// contest component
	SeasonContestComponent *SeasonContestComponent

	Rule    *Rule
	Elastic *elastic.Config
	DataBus *DataBusV2
}

type DataBusV2 struct {
	Target string `toml:"target"`
	AppID  string `toml:"app_id"`
	Token  string `toml:"token"`
	Topic  struct {
		BGroupMessage string
	}
}

type SeasonContestComponent struct {
	CanWatch        bool
	StartTimeBefore time.Duration
	EndTimeAfter    time.Duration
	ExpiredDuration int32
}

// SysInforms .
type SysInforms struct {
	BatchSize        int
	AlertTitle       string
	AlertBodyDefault string
	AlertBodySpecial string
	URL              string
	MC               string
}

// Push push.
type Push struct {
	BusinessID    int
	BusinessToken string
	PartSize      int
	Title         string
	BodyDefault   string
	BodySpecial   string
	OnlyMids      string
}

type RankingDataWatch struct {
	InterventionCacheKey string
}

type S10CoinConfig struct {
	SeasonID  int64
	GameState int64
}

// Rule .
type Rule struct {
	LogoPathPre            string
	MaxGuessStake          int64
	ContestJumpURL         string
	ContestGuessURL        string
	GuessOverBeforeSTime   int
	LRUCacheMaxContestSize int
	LRUCacheMaxTeamSize    int
	LRUCacheMaxMatchSize   int
	LRUCacheMaxSeriesSize  int
	EsSearchQueryDebug     bool
}

type TunnelPush struct {
	TunnelBizID int64
	TemplateID  int64
	Link        string
}

type TunnelBGroup struct {
	TunnelBizID   int64
	CardUniqueId  int64
	NewBusiness   string
	NewTemplateID int64
	NewCardText   string
	Link          string
	SendNew       int64
	NewCardLiveID int64
}

type TunnelCardMsg struct {
	TunnelBizID              int64
	CardUniqueId             int64
	NewBusiness              string
	DefaultContestNotifyCode string
	SpecialContestNotifyCode string
	Link                     string
	JumpText                 string
}

type types struct {
	ID   int64
	Name string
}

// Host remote host.
type Host struct {
	APICo    string
	GenPost  string
	SavePost string
}

func Init() (err error) {
	return newConf.Init(load)
}

func load() (err error) {
	var (
		tmpConf *Config
	)
	if err = newConf.LoadInto(&tmpConf); err != nil {
		return err
	}
	*Conf = *tmpConf
	return
}
