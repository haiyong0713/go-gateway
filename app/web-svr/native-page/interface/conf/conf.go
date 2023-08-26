package conf

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc/warden"
	rpcquota "go-common/library/net/rpc/warden/ratelimiter/quota"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/time"

	"github.com/BurntSushi/toml"
)

// Config service config
type Config struct {
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
	// grpc
	TagClient        *warden.ClientConfig
	RelClient        *warden.ClientConfig
	AccClient        *warden.ClientConfig
	FliClient        *warden.ClientConfig
	UpClient         *warden.ClientConfig
	ArtClient        *warden.ClientConfig
	FavoriteClient   *warden.ClientConfig
	CharGRPC         *warden.ClientConfig
	TopicClient      *warden.ClientConfig
	ArcClient        *warden.ClientConfig
	LiveClient       *warden.ClientConfig
	RoomGateClient   *warden.ClientConfig
	ReplyClient      *warden.ClientConfig
	PopularClient    *warden.ClientConfig
	PgcClient        *warden.ClientConfig
	ChannelClient    *warden.ClientConfig
	ActClient        *warden.ClientConfig
	PlatClient       *warden.ClientConfig
	SpaceClient      *warden.ClientConfig
	UpRatingClient   *warden.ClientConfig
	AegisClient      *warden.ClientConfig
	HmtChannelClient *warden.ClientConfig
	DynvoteClient    *warden.ClientConfig
	ScoreClient      *warden.ClientConfig
	EsportsGRPC      *warden.ClientConfig
	// httpClient
	HTTPClient   *blademaster.ClientConfig
	HTTPDynamic  *blademaster.ClientConfig
	HTTPBusiness *blademaster.ClientConfig
	HTTPGameCo   *blademaster.ClientConfig
	HTTPMangaCo  *blademaster.ClientConfig
	HTTPShowCo   *blademaster.ClientConfig
	HTTPActAdmin *blademaster.ClientConfig
	// Rule
	Rule *Rule
	// Host
	Host Host
	// Log
	Log *log.Config
	// ecode
	Ecode *ecode.Config
	// redis
	Redis *Redis
	// DataBus databus
	DataBus *DataBus
	Limiter *quota.Config
	// native-page
	NativePage         *NativePage
	QuotaConf          *rpcquota.Config
	WhitelistCondition *WhitelistCondition //白名单条件
	FilterClass        *FilterClass        //敏感词等级
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

type WhitelistCondition struct {
	MinRatingLevel int64 //最小电磁力等级
	MustNoCompany  bool  //是否必须非企业号认证
	LockExpire     int64 //并发锁过期时间
}

type FilterClass struct {
	MinDynamic  int32 //动态库
	MinDynTopic int32 //动态话题库
}

type NativePage struct {
	WhiteListByMidExpire     int64
	WhiteListByMidNullExpire int64
}

// DataBus multi databus collection.
type DataBus struct {
	NativePub *databus.Config
}

// Host remote host.
type Host struct {
	APICo    string
	Dynamic  string
	Business string
	GameCo   string
	ShowCo   string
	MangaCo  string
	ActAdmin string
}

// Rule   rule config.
type Rule struct {
	OpenDynamic bool
	//up 主发起活动活动id
	UpSenderUid       uint64
	RegularExpire     time.Duration
	NotifyCodeOffline string
	//up发起人添加up主活动开启
	UpActOpen   bool
	UpFansLimit int64
}

// MySQL define MySQL config
type MySQL struct {
	Like *sql.Config
}

// Redis struct
type Redis struct {
	*redis.Config
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
