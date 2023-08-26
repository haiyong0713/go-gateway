package conf

import (
	"go-common/library/conf/paladin.v2"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/time"
	"go-gateway/app/web-svr/player/interface/model"

	"github.com/BurntSushi/toml"
)

// global var
var (
	Conf = &Config{}
)

// Config is service conf.
type Config struct {
	// 广播
	Broadcast Broadcast
	// policy
	Policy *model.Policy
	// policy items
	Pitem []*model.Pitem
	// 拜年祭
	Matsuri Matsuri
	// XLog
	XLog *log.Config
	// ecode
	Ecode *ecode.Config
	// host
	Host *Host
	// tracer
	Tracer *trace.Config
	// auth
	Auth *auth.Config
	// verify
	Verify *verify.Config
	// bm
	BM *HTTPServers
	// rpc
	ResourceRPC *rpc.ClientConfig
	TagRPC      *rpc.ClientConfig
	// HTTPClient
	HTTPClient *bm.ClientConfig
	// Rule
	Rule *Rule
	Tick *Tick
	// Infoc2
	Infoc2 *infoc.Config
	// PlayURLToken
	PlayURLToken *PlayURLToken
	// grpc client
	AccClient       *warden.ClientConfig
	ArcClient       *warden.ClientConfig
	UGCPayClient    *warden.ClientConfig
	PlayURLClient   *warden.ClientConfig
	SteinsClient    *warden.ClientConfig
	DMClient        *warden.ClientConfig
	AnswerClient    *warden.ClientConfig
	AssistClient    *warden.ClientConfig
	HistoryClient   *warden.ClientConfig
	LocationClient  *warden.ClientConfig
	PugvClient      *warden.ClientConfig
	MemberClient    *warden.ClientConfig
	ResourceClient  *warden.ClientConfig
	AppConfigClient *warden.ClientConfig
	VideoUpGRPC     *warden.ClientConfig
	PlayURLGRPC     *warden.ClientConfig
	CfcGRPC         *warden.ClientConfig
	// bnj
	Bnj *Bnj
	//Cron
	Cron          *cronConf
	InfocLog      *infocLog
	LongProgress  *longProgress
	OnlineGray    *onlineGray
	GrayVideoShot *GrayVideoShot
	// content.flow.control.service gRPC config
	CfcSvrConfig *CfcSvrConfig
}

type CfcSvrConfig struct {
	BusinessID int64
	Secret     string // 由服务方下发
	Source     string
}

type onlineGray struct {
	Open          bool
	Whitelist     []int64
	Bucket        uint64
	RealBucket    uint64
	RealWhitelist []int64
}

type longProgress struct {
	UGC time.Duration
}

type cronConf struct {
	Resource  string
	Param     string
	Mat       string
	GuideCid  string
	BnjView   string
	Fawkes    string
	LimitFree string
}

type Bnj struct {
	Tick     time.Duration
	MainAid  int64
	SpAid    int64
	ListAids []int64
}

// Tick tick time.
type Tick struct {
	// tick time
	CarouselTick time.Duration
	ParamTick    time.Duration
}

// Rule rules
type Rule struct {
	// timeout
	VsTimeout      time.Duration
	NoAssistMid    int64
	AutoQn         int64
	SteinsGuideAid int64
	H5GRPCGray     int64
	GrayMids       []int64
	WithOutVipAids []int64
}

// Host is host info
type Host struct {
	APICo        string
	AccCo        string
	PlayurlCo    string
	H5Playurl    string
	HighPlayurl  string
	Fawkes       string
	MusicAPI     string
	LimitFreeUrl string
}

// Broadcast breadcast.
type Broadcast struct {
	TCPAddr string
	WsAddr  string
	WssAddr string
	Begin   string
	End     string
}

// Matsuri matsuri.
type Matsuri struct {
	PastID  int64
	MatID   int64
	MatTime string
	Tick    time.Duration
}

// PlayURLToken playurl auth token.
type PlayURLToken struct {
	Secret      string
	PlayerToken string
}

// HTTPServers bm servers config.
type HTTPServers struct {
	Outer *bm.ServerConfig
}

type infocLog struct {
	ShowLogID string
}

type GrayVideoShot struct {
	Group uint32
	Gray  uint32
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
	err = paladin.Watch("player-interface.toml", Conf)
	if err != nil {
		return err
	}

	return
}

func load() (err error) {
	err = paladin.Get("player-interface.toml").UnmarshalTOML(Conf)
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
	log.Info("player-interface-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}
