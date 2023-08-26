package conf

import (
	"errors"
	"flag"

	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/database/sql"
	xlog "go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	Conf     = &Config{}
	client   *conf.Client
)

type Config struct {
	// Log
	Log *xlog.Config
	// tracer
	Tracer *trace.Config
	// bm http
	BM *HTTPServers
	// httpClinet
	HTTPClient *bm.ClientConfig
	// httpSearch
	HTTPSearch *bm.ClientConfig
	// httpPGC
	HTTPPGC *bm.ClientConfig
	// httpData
	HTTPData *bm.ClientConfig
	// host
	Host *Host
	// HostDiscovery
	HostDiscovery *HostDiscovery
	// redis
	Redis *Redis
	// taishan
	Taishan *Taishan
	// Custom
	Custom *Custom

	// ArchiveGRPC grpc
	ArchiveGRPC *warden.ClientConfig
	// FlowControlGRPC grpc client config
	FlowControlGRPC *warden.ClientConfig
	// AccountGRPC grpc
	AccountGRPC *warden.ClientConfig
	// RelationGRPC grpc
	RelationGRPC *warden.ClientConfig
	// HistoryGRPC grpc
	HistoryGRPC *warden.ClientConfig
	// PGCRPC grpc
	PGCRPC *warden.ClientConfig
	// VIPGRPC grpc
	VIPGRPC *warden.ClientConfig
	// UpGRPC grpc
	UpGRPC *warden.ClientConfig
	// dynamicRPC client
	DynamicRPC *rpc.ClientConfig
	// playurl client
	PlayURLClient *warden.ClientConfig
	// thumbup client
	ThumbupClient *warden.ClientConfig
	// archiveHonor client
	GaiaClient *warden.ClientConfig
	// SilverClient
	SilverClient *warden.ClientConfig
	// ResourceClient
	ResourceClient *warden.ClientConfig
	// ChannelClient
	ChannelClient *warden.ClientConfig
	// SerialClient
	SerialClient *warden.ClientConfig
	// grpc location
	LocationGRPC *warden.ClientConfig
	// automotive.channel.recommend
	FmRecommendGRPC *warden.ClientConfig
	// ott-ab实验
	ABClientCfg *warden.ClientConfig

	// feature配置
	Feature *Feature
	// MySQL
	MySQL *MySQL
	// VipConfig
	VipConfig *VipConfig
	// VideoTabsV2Conf
	VideoTabsV2Conf *VideoTabsV2Conf
	// FlowControl 内容分发管控
	FlowControl    *FlowControl
	FlowControlAll *FlowControl
	// FM上报算法
	FmReportMq     *DataBusV2Conf
	BannerPlaylist []*BannerPlaylist // banner位播放列表，指定哪个tab配置哪个播放列表。key：tab id，value：播放列表
	CustomModule   *CustomModule
	// v1版本中，发现tab页列表
	CustomModule51          *CustomModule // 51-特辑
	CustomModule61Childhood *CustomModule // 61-童年回来了
	CustomModule61Eden      *CustomModule // 61-小朋友乐园
	CustomModuleDW          *CustomModule // 端午-粽有陪伴
	// v2版本中，视频tab的二级tab
	CustomTab61Childhood   *CustomModule // 61-童年回来了
	CustomTab61Eden        *CustomModule // 61-小朋友乐园
	CustomTabDWRicePudding *CustomModule // 端午-粽有陪伴
	CustomTabDWEden        *CustomModule // 端午-小朋友乐园
	CustomTabLYJ           *CustomModule // 小鹏727-老友记
	CustomTabZhouJieLun    *CustomModule // 小鹏727-周杰伦

	TabExchange   *HomeTabExchange // 首页视频/FM tab互换配置
	PinPageCfgAll *PinPageCfgAll   // 金刚位配置
	// v2.3没有视频合集和频道数据，但产品要预埋这个功能，因此写了一些代码mock数据；
	// 但测试验收时没有数据，导致release包验不了功能，必须带上mock代码上线；
	// 此配置为debug开关，验收通过后关闭
	V23Debug            *V23Debug
	MediaParseAccessIps []string // 给车企解析短链接口的ip白名单
	ExpIds              *ExpCfg  // ab实验id配置
	EnableXP727Tabs     bool     // 是否启用小鹏727特别视频tab
}

type V23Debug struct {
	Switch bool
	Mids   []int64
}

type PinPageCfgAll struct {
	TopText string      // 金刚位更多页顶部标题
	HasPin  *PinPageCfg // 是否展示金刚位
	PinMore *PinPageCfg // 金刚位是否展示更多
}

type PinPageCfg struct {
	BlackChannel []string // 黑名单
}

type HomeTabExchange struct {
	Channels []string // 需要互换的渠道
}

type BannerPlaylist struct {
	Id           int64   // 唯一标识
	PlayList     []int64 // 播放列表
	ShowId       int64   // 入口稿件
	Title        string  // 入口标题
	Cover        string  // 入口封面
	MaterialType string  // 稿件类型：ugc、ogv_season
	StyleType    int     // 样式类型：1banner多个稿件、2合集
	TabId        int64   // 该配置作用于哪个tab banner上
}

type CustomModule struct {
	ChannelAids        map[string][]int64
	EnableCustomModule bool
	MinNumbers         int
	//时尚分区 - 美妆护肤二级分区 - 化妆教程频道 - 综合 - 近期热门的渠道ID
	//uat 600, prod:261355
	ChannelMakeups        int64
	XiaoPengKeywordRegion string
	XiaoPengKeywordTab    string
	//加载小鹏干预卡片任务
	XiaoPengInterveneCron string
	//熔断干预开关 true 打开 不走干预; false 关闭(默认)
	CircuitIntervene bool
}

type DataBusV2Conf struct {
	AppId string
	Token string
	Topic string
}

type FlowControl struct {
	BusinessID int
	Source     string
	Secret     string
}

type VideoTabsV2Conf struct {
	VideoTabs []*VideoTabsV2
	DefaultPs int
}

type VideoTabsV2 struct {
	Type      int64
	Id        int64
	Name      string
	IsDefault bool
}

// Host is
type Host struct {
	Search  string
	APICo   string
	APICom  string
	Bangumi string
	VcCo    string
}

type HostDiscovery struct {
	Data      string
	PGCPlayer string
}

// HTTPServers is
type HTTPServers struct {
	Outer *bm.ServerConfig
}

// Redis is
type Redis struct {
	Entrance *redis.Config
}

type MySQL struct {
	Show *sql.Config
	Car  *sql.Config
}

type Taishan struct {
	ChannelTable  *TaishanTable // 频道信息表
	ChannelClient *warden.ClientConfig
}

type TaishanTable struct {
	Table string
	Token string
}

// Custom config
type Custom struct {
	// TabConfigs
	TabConfigs []*TabConfig
	// v1.1增加新的tab，老版本不下发
	TabConfigs2 []*TabConfig
	// TabConfigs web使用
	TabConfigsWeb []*TabConfig
	// banner轮播图
	Banners map[string][]*BannerConfig
	// 详情页二维码
	ViewQRCode map[string]string
	// BackupNum
	BackupNum uint32
	// default qn
	DefaultQn int64
	// 渠道类型
	ChannelType *ChannelType
	// web我的页配置
	MineWebTab map[string][]string
	// FM首页卡片配置
	FmTabConfigs []*FmTabConfig
}

type ChannelType struct {
	Sound map[string]string
}

type TabConfig struct {
	ID           int64
	Name         string
	URI          string
	TabID        string
	IsDefault    bool
	Icon         string
	IconSelected string
	HideChannel  map[string]int
	Goto         string
}

type BannerConfig struct {
	ID    int64
	Image string
	URL   string
}

type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}

type FeatureBuildLimit struct {
	Switch      bool
	TabConfig2  string
	Media       string
	MineTag     string
	HotBanner   string
	Feed        string
	Region      string
	ShowListTab string
	ShowFromTag string
}

type VipConfig struct {
	AppKey     string
	BatchToken string
}

type FmTabConfig struct {
	FmType   string
	FmId     int64
	Title    string
	SubTitle string
	Cover    string
	Style    int
}

type ExpCfg struct {
	Season SeasonExpCfg
}

type SeasonExpCfg struct {
	ExpId      int64 // 实验ID
	ExpGroupId int64 // 实验组ID
}

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
}

// Init init config.
func Init() (err error) {
	if confPath != "" {
		_, err = toml.DecodeFile(confPath, &Conf)
		return
	}
	err = remote()
	return
}

func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	client.Watch("app-car.toml")
	// nolint:biligowordcheck
	go func() {
		for range client.Event() {
			xlog.Info("config reload")
			if load() != nil {
				xlog.Error("config reload error (%v)", err)
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

// GetDB 获取数据库链接
func GetDB(c *Config) (db *sql.DB) {
	return sql.NewMySQL(c.MySQL.Car)

}
