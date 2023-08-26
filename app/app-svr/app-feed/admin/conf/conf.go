package conf

import (
	"errors"
	"flag"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/database/elastic"
	"go-common/library/database/orm"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/permit"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/dao/clickhouse"
	"go-gateway/app/app-svr/app-feed/admin/dataplat"
	"go-gateway/app/app-svr/app-feed/admin/model/feature"
	frontpageModel "go-gateway/app/app-svr/app-feed/admin/model/frontpage"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
	splashModel "go-gateway/app/app-svr/app-feed/admin/model/splash_screen"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	client   *conf.Client
	// Conf of config
	Conf = &Config{}
)

// MesConfig .
type MesConfig struct {
	MC    string
	Title string
	Msg   string
}

// Message .
type Message struct {
	URL     string
	Tianma  *MesConfig
	Popular *MesConfig
}

// Memcache memcache.
type Memcache struct {
	*memcache.Config
}

type Redis struct {
	*redis.Config
}

type AggregationMemcache struct {
	*memcache.Config
}

type BubbleMemcache struct {
	*memcache.Config
}

// Bfs Bfs.
type Bfs struct {
	Timeout     xtime.Duration
	MaxFileSize int
	Bucket      string
	Addr        string
	Key         string
	Secret      string
}

// Cfg def
type Cfg struct {
	// HotCroFre hotword crontab frequency
	HotCroFre string
	// DarkCroFre darkword crontab frequency
	DarkCroFre string
	// RcmdCroFre recommend crontab frequency
	RcmdCroFre string
	// BrandCroFre brand blacklist crontab frequency
	BrandCroFre string
	// RunCront is run crontab
	RunCront bool
	SelCfg   *SelCfg
	// 让过期的模块禁用调的task
	SidebarCroFre string
	// GameCroFre game id check crontab frequency
	GameCroFre string
	// ChannelFre channel id crontab frequency
	ChannelCroFre string
}

// SelCfg is the config for the selected serie
type SelCfg struct {
	Business     string
	Index        string
	ExportTitles []string // export titles
}

// Host host
type Host struct {
	Manager   string
	Game      string
	EntryGame string
	Live      string
	API       string
	// 漫画内网
	ComicInner string
	Dynamic    string
	// 热门热点聚合
	BigData string
	// 会员购
	Vip       string
	Archive   string
	Thumbnail string
	Berserker string
	// 普通用户认证
	Easyst string
	Cmmng  string
}

// UserFeed
type UserFeed struct {
	Game      string
	Pgc       string
	Archive   string
	Account   string
	Comic     string
	Live      string
	Dynamic   string
	Article   string
	Feed      string
	MediaList string
}

// HTTPClient http client
type HTTPClient struct {
	Read      *bm.ClientConfig
	Game      *bm.ClientConfig
	ES        *bm.ClientConfig
	DataPlat  *dataplat.ClientConfig
	DataPlat2 *dataplat.ClientConfig
	Push      *bm.ClientConfig
	MediaList *bm.ClientConfig
	EntryGame *EntryGameClientConfig
}

// EntryGameClientConfig 游戏验签及DES解密配置
type EntryGameClientConfig struct {
	Dial      xtime.Duration
	Timeout   xtime.Duration
	KeepAlive xtime.Duration
	Secret    string
	DesKey    string
}

// boss平台配置
type BossCfg struct {
	Bucket     string
	EntryPoint string
	AccessKey  string
	SecretKey  string
	Region     string
	LocalDir   string
}

// 时间间隔
type TimeGapCfg struct {
	Hotword int64
}

// SplashScreenLogoCfg 闪屏LOGO配置
type SplashScreenLogoCfg struct {
	White string
	Pink  string
}

// SplashScreenImgCfg 闪屏物料配置
type SplashScreenImgCfg struct {
	KeepNewDays int32
}

// SplashScreen 闪屏配置
type SplashScreen struct {
	Img               *SplashScreenImgCfg
	Logo              *SplashScreenLogoCfg
	BaseDefaultConfig *splashModel.SplashScreenConfig
}

// PopupCfg 天马弹窗配置
type PopupCfg struct {
	BGroupBusinessName string
	AutoHideCountdown  int64
}

// 404 error 页面配置
type Error404Config struct {
	Databus  *databus.Config
	BaseConf *struct {
		ETimeOffset int64
		Operator    string
		OperatorId  int64
	}
	AuditMap []*struct {
		Codes    []int32
		Priority int32
		Reason   string
	}
}

type AllowedTabs struct {
	Tabs string
}

type FlowCtrl struct {
	Secret    string
	Source    string
	OidLength int
}
type DanmuEffectiveTime struct {
	StartTime int64
	EndTime   int64
}
type Danmu struct {
	Icon        string                //弹幕资源配置，图片路径
	CidSection  []*DanmuEffectiveTime //生效时间段
	PurifyExtra *selected.PurifyExtra
}

type WeeklySelected struct {
	SubscribedTag    int64
	AttemptCount     int
	PublishCron      string
	NewSerieCron     string
	UpdateTime       xtime.Duration
	PlaylistMid      int64
	FlowCtrl         *FlowCtrl
	HonorLink        string
	HonorLinkV2      string
	RankIndex        int
	RollBackRankCron string
	RankId           int
	RecoveryNb       int
	MaxNumber        int
	Push             struct {
		Token      string
		Title      string
		BusinessID string
		Link       string
	}
	Danmu *Danmu
}

type FeedConfig struct {
	SkipCardUrl string
	FlowCtrl    *FlowCtrl
}

// Config def.
type Config struct {
	// base
	// http
	HTTPServer *bm.ServerConfig
	// httpClinet
	HTTPClient *HTTPClient
	// host
	Host     *Host
	UserFeed *UserFeed
	// auth
	Auth *permit.Config
	// db
	ORM *orm.Config
	// db
	ORMResource *orm.Config
	ORMManager  *orm.Config
	ORMFeature  *orm.Config
	ORMTag      *orm.Config
	// log
	Log *log.Config
	// tracer
	Tracer *trace.Config
	// mc
	Memcache            *Memcache
	AggregationMemcache *AggregationMemcache
	BubbleMemcache      *BubbleMemcache
	// Bfs
	Bfs *Bfs
	// log
	ManagerReport  *databus.Config
	SelectedUpdate *databus.Config
	// BroadcastRPC grpc
	PGCRPC *warden.ClientConfig
	// rpc client
	ArchiveRPC *rpc.ClientConfig
	RPCServer  *warden.ServerConfig
	// Cfg
	Cfg *Cfg
	// grpc
	ArcClient          *warden.ClientConfig
	TagGRPClient       *warden.ClientConfig
	FlowCtrlGRPCClient *warden.ClientConfig
	GarbClient         *warden.ClientConfig
	AccountGRPC        *warden.ClientConfig
	BGroupClient       *warden.ClientConfig
	SmsClient          *warden.ClientConfig
	MemberClient       *warden.ClientConfig
	DanmuGRPC          *warden.ClientConfig
	// ecode cfg
	Ecode               *ecode.Config
	Message             *Message
	FilGRPClient        *warden.ClientConfig
	TagGRPCClient       *warden.ClientConfig
	LocationClient      *warden.ClientConfig
	LocationAdminClient *warden.ClientConfig
	ES                  *elastic.Config
	MySQL               *MySQL
	EntranceRedis       *Redis
	SelectedRedis       *Redis
	SpmodeRedis         *Redis
	Boss                *BossCfg
	TimeGap             *TimeGapCfg
	// for app entry, added by shiliang@2020.7.21
	Databus *databus.Config

	Plats        []*feature.Plat
	SplashScreen *SplashScreen
	Popup        *PopupCfg

	// for 404 error page consumer, added by shiliang@2020.10.14
	Error404Conf *Error404Config

	// 允许配置渐变色的tab
	AllowedTabs *AllowedTabs

	// 每周必看配置
	WeeklySelected *WeeklySelected
	// 荣誉稿件 databus
	ArchiveHonorDatabus *databus.Config
	// TV 稿件 databus
	OTTSeriesDatabus *databus.Config
	// Feed配置：特殊卡片跳转链接-特定字符串不下发卡片
	FeedConfig    *FeedConfig
	ShowGrpcSH004 *warden.ClientConfig

	// 版头配置
	Frontpage *FrontpageConfig
	// 忘记密码申诉
	PwdAppeal  *PwdAppeal
	ClickHouse *ClickHouse
}

type ClickHouse struct {
	AntiCrawler *struct {
		*clickhouse.Config
		DatabaseName string
	}
}

type PwdAppeal struct {
	EncryptKey  string
	ExportLimit int64                    //csv导出条数限制
	SmsCfg      map[string]*PwdAppealSms //模式对应的短信配置，key为pwd_appeal.mode
	Boss        *BossCfg
}

type PwdAppealSms struct {
	PassTcode   string //通过的短信模板编码
	RejectTcode string //驳回的短信模板编码
	AppealUrl   string //申诉页面地址
}

type MySQL struct {
	Show *sql.Config
}

type FrontpageConfig struct {
	GlobalMenu *frontpageModel.Menu
}

func local() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	client.Watch("feed-admin.toml")
	//nolint:biligowordcheck
	go func() {
		for range client.Event() {
			log.Info("config reload")
			if load() != nil {
				log.Error("config reload error (%v)", err)
			}
		}
	}()
	return
}

func load() (err error) {
	var (
		s       string
		ok      bool
		tmpConf *Config
	)
	if s, ok = client.Toml2(); !ok {
		return errors.New("load config center error")
	}
	if _, err = toml.Decode(s, &tmpConf); err != nil {
		return errors.New("could not decode config")
	}
	*Conf = *tmpConf
	return
}

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
}

// Init int config
func Init() error {
	if confPath != "" {
		return local()
	}
	return remote()
}
