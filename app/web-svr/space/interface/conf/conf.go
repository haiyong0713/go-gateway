package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"

	"go-common/library/database/hbase.v2"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/antispam"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/supervisor"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/time"
	"go-gateway/app/app-svr/app-card/middleware/anticrawler"

	"github.com/BurntSushi/toml"
)

const (
	_space_conf = "space-interface.toml"
)

// global var
var (
	// Conf config
	Conf = &Config{}
)

// Config config set
type Config struct {
	// elk
	Log *log.Config
	// App
	App *blademaster.App
	// tracer
	Tracer *trace.Config
	// Auth
	Auth *auth.Config
	// Verify
	Verify *verify.Config
	// Supervisor
	Supervisor *supervisor.Config
	// BM
	BM *httpServers
	// HTTPServer
	HTTPServer *blademaster.ServerConfig
	// GRPCServer
	GRPCServer *warden.ServerConfig
	// Ecode
	Ecode *ecode.Config
	// grpc
	AccClient         *warden.ClientConfig
	ArcClient         *warden.ClientConfig
	AccmClient        *warden.ClientConfig
	PanguGSClient     *warden.ClientConfig
	CoinClient        *warden.ClientConfig
	ThumbupClient     *warden.ClientConfig
	PGCFollowClient   *warden.ClientConfig
	PGCSeasonClient   *warden.ClientConfig
	PGCProgressClient *warden.ClientConfig
	ArticleClient     *warden.ClientConfig
	FilterClient      *warden.ClientConfig
	AssistClient      *warden.ClientConfig
	LiveGRPC          *warden.ClientConfig
	LiveXRoomGRPC     *warden.ClientConfig
	FavGRPC           *warden.ClientConfig
	PugvGRPC          *warden.ClientConfig
	AccCPClient       *warden.ClientConfig
	RelationGRPC      *warden.ClientConfig
	MemberClient      *warden.ClientConfig
	NoteClient        *warden.ClientConfig
	UpArcClient       *warden.ClientConfig
	PayRankClient     *warden.ClientConfig
	ActivityClient    *warden.ClientConfig
	SeriesGRPC        *warden.ClientConfig
	ResourceGRPC      *warden.ClientConfig
	LiveUserGRPC      *warden.ClientConfig
	TagGRPC           *warden.ClientConfig
	DynamicSearchGRPC *warden.ClientConfig
	PgcCardClient     *warden.ClientConfig
	PGCCardClient     *warden.ClientConfig
	SpaceGRPC         *warden.ClientConfig
	UGCSeasonClient   *warden.ClientConfig
	GalleryGRPC       *warden.ClientConfig
	CfcGRPC           *warden.ClientConfig
	DynamicFeedGRPC   *warden.ClientConfig
	GaiaGRPC          *warden.ClientConfig
	// Mysql
	Mysql *sql.Config
	// Redis
	Redis *redisConf
	// Rule
	Rule *rule
	// HTTP client
	HTTPClient *httpClient
	// Host
	Host *host
	// HBase hbase config
	HBase *Hbase
	// Antispam
	Antispam *antispam.Config
	// fake
	Fake *fake
	// Spec
	Spec *Spec
	// play button
	PlayButton *playButton
	// databus
	Databus *Databus
	// bfs
	Bfs *Bfs
	// MidRPC
	MidGRPC *warden.ClientConfig
	// NaPageRPC
	NaPageRPC     *warden.ClientConfig
	UGCSeasonGRPC *warden.ClientConfig
	AccountGRPC   *warden.ClientConfig
	// anticrawler
	Anticrawler *anticrawler.Config
	// live_playback
	LivePlayback *livePlayback
	// degrade
	DegradeConfig *degradeConfig
	LegoToken     *legoToken
	Series        *series
	// content.flow.control.service gRPC config
	CfcSvrConfig *CfcSvrConfig
	// switch
	SeniorMemberSwitch *SeniorMemberSwitch
	// risk management
	RiskManagement *RiskManagement
}

type series struct {
	Open      bool
	Bucket    uint64
	Whitelist []int64
}

type legoToken struct {
	SpaceIPLimit string
}

type degradeConfig struct {
	Expire   int32
	Memcache *memcache.Config
}

type livePlayback struct {
	UpFrom []int32
}

type Bfs struct {
	ReadTimeout time.Duration
	Bucket      string
	Dir         string
}

type playButton struct {
	Open       bool
	Text       string
	BaseURI    string
	ForbidMids []int64
}

type Databus struct {
	VisitPub *databus.Config
}

type Spec struct {
	PhotoMall    string
	BlackList    string
	SysNotice    string
	LivePlayback string
}

type redisConf struct {
	*redis.Config
	ClExpire          time.Duration
	OfficialExpire    time.Duration
	UserTabExpire     time.Duration
	WhitelistExpire   time.Duration
	MinExpire         time.Duration
	MaxExpire         time.Duration
	TopPhotoArcExpire time.Duration
	SettingExpire     time.Duration
	NoticeExpire      time.Duration
	TopArcExpire      time.Duration
	MpExpire          time.Duration
	ThemeExpire       time.Duration
	TopDyExpire       time.Duration
}

type rule struct {
	MaxChNameLen     int
	MaxChIntroLen    int
	MaxChLimit       int
	MaxChArcLimit    int
	MaxChArcAddLimit int
	MaxChArcsPs      int
	MaxRiderPs       int
	MaxArticlePs     int
	ChIndexCnt       int
	MaxNoticeLen     int
	MaxTopReasonLen  int
	MaxMpReasonLen   int
	MaxMpLimit       int
	// RealNameOn
	RealNameOn bool
	// No limit notice mids
	NoNoticeMids []int64
	// default top photo
	TopPhoto string
	// dynamic list switch
	Merge   bool
	ActFold bool
	// block mids
	BlockMids []int64
	// default photo
	DftPhotoID    int64
	DftPhotoLowID int64
	DftPhotoBuild struct {
		Android  int32
		Iphone   int32
		AndroidI int32
	}
	// top photo arc build
	TopPhotoArcBuild struct {
		Android int32
		Iphone  int32
	}
	// mcn info switch
	McnOn bool
}

type host struct {
	Bangumi  string
	API      string
	Mall     string
	APIVc    string
	APILive  string
	Acc      string
	Game     string
	AppGame  string
	Search   string
	Space    string
	Dynamic  string
	LinkDraw string
}

type fake struct {
	Home  string
	Guest string
	Url   string
}

type httpClient struct {
	Read    *blademaster.ClientConfig
	Write   *blademaster.ClientConfig
	Game    *blademaster.ClientConfig
	Dynamic *blademaster.ClientConfig
}

type httpServers struct {
	Outer *blademaster.ServerConfig
}

// Hbase .
type Hbase struct {
	*hbase.Config
	ReadTimeout time.Duration
}

type CfcSvrConfig struct {
	BusinessID int64
	Secret     string // 由服务方下发
	Source     string
}

type SeniorMemberSwitch struct {
	ShowSeniorMember bool
}
type RiskManagement struct {
	RiskDecisions []string
}

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
	err = paladin.Watch(_space_conf, Conf)
	if err != nil {
		return err
	}

	return
}

func load() (err error) {
	err = paladin.Get(_space_conf).UnmarshalTOML(Conf)
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
	log.Info("space-interface-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}
