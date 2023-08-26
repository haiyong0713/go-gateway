package conf

import (
	xtime "time"

	"go-common/library/cache/redis"
	"go-common/library/database/bfs"
	"go-common/library/database/boss"
	"go-common/library/database/orm"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/time"

	"go-common/library/conf/paladin.v2"
	"go-common/library/log/infoc.v2"

	clickSQL "go-gateway/app/app-svr/fawkes/service/dao/database"

	"github.com/BurntSushi/toml"
)

var (
	Conf = &Config{}
)

func Init() (err error) {
	if err = paladin.Init(); err != nil {
		panic(err)
	}
	if err := paladin.Get("fawkes-admin.toml").UnmarshalTOML(Conf); err != nil {
		panic(err)
	}
	if err := paladin.Watch("fawkes-admin.toml", Conf); err != nil {
		panic(err)
	}
	return nil
}

// Config struct.
type Config struct {
	// reload time
	Reload time.Duration
	// patch limit
	PatchLimit       int
	PatchSteadyLimit int
	// interface XLog
	XLog *log.Config
	// http
	HTTPServers *HTTPServers
	// http client
	HTTPClient *bm.ClientConfig
	// oss
	Oss *OssWrapper
	// app store connect
	AppstoreConnect *AppstoreConnect
	// db
	MySQL *MySQL
	// db
	ORM *orm.Config
	// bfs config
	BFS *bfs.Config
	// tracer
	Tracer *trace.Config
	// System Version
	System *System
	// LocalPath
	LocalPath *LocalPath
	// Keys
	Keys *Keys
	// gitlab
	Gitlab *Gitlab
	// hosts
	Host *Host
	// easyst
	Easyst *Easyst
	// white list
	Whitelist []string
	// mail
	Mail *Mail
	// wxnotify
	WXNotify *WXNotify
	// CDN
	CDN *CDN
	// bfs CDN
	BFSCDN *BFSCDN
	// ClickHouse
	ClickHouse *ClickHouse
	// Cron
	Cron *Cron
	// comet
	Comet *Comet
	// EP
	Ep *Ep
	// boss cfg
	BossConfig *boss.Config
	// zip
	ZipContentType []string
	// broadcasePush
	BroadcastPush *BroadcastPush
	// redis
	Redis *Redis
	// mod
	Mod *Mod
	// broadcast grpc
	BroadcastGrpc *BroadcastGrpc
	// 告警接收人
	AlarmReceiver *AlarmReceiver
	// selector databases
	Prometheus *Prometheus
	// 流处理器任务和事件关联
	FlinkJob *FlinkJob
	// 监控配置
	Moni *Moni
	// 开关控制
	Switch *Switch
	// 数据中心外部接口
	Datacenter *Datacenter
	// 日志系统外部接口
	Billions *Billions
	// databus collect
	Databus *Databus
	// ipdb
	IPDB *IPDB
	// 定时任务配置
	Task *Task
	// Elasticsearch-Proxy
	ElasticsearchProxy *ElasticsearchProxy
	// tapd配置
	TAPD *TAPD
	// 可执行文件地址
	ExePath *ExePath
	// billions-alert
	BillionsAlert *BillionsAlert
	FissionGRPC   *warden.ClientConfig
	// mobiApp白名单
	MobiAppWhiteList []string
	// prometheus 模板
	PrometheusTemplate *PrometheusTemplate
	// infoc
	Infoc infoc.Infoc
}

func (c *Config) Set(text string) error {
	var tmp *Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	log.Info("config changed, old=%+v new=%+v", Conf, tmp)
	Conf = tmp
	return nil
}

// HTTPServers Http Servers.
type HTTPServers struct {
	Inner *bm.ServerConfig
	Outer *bm.ServerConfig
}

// WXNotify struct
type WXNotify struct {
	AccessTokenURL   string
	UserListURL      string
	MessageSendURL   string
	UploadTmpFileURL string
	AgentID          string
	CorpID           string
	CorpSecret       string
	DepartmentIDs    string
}

// Gitlab struct.
type Gitlab struct {
	Host           string
	API            string
	Token          string
	CronExpression string
}

// OssWrapper struct.
type OssWrapper struct {
	Inland *Oss // 国内 - 对象存储
	Abroad *Oss // 海外 - 对象存储
}

// Oss struct.
type Oss struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	Bucket          string
	OriginDir       string
	PublishDir      string
	CDNDomain       string
}

// MySQL struct.
type MySQL struct {
	Fawkes  *sql.Config
	Macross *sql.Config
	Show    *sql.Config
	Veda    *sql.Config
}

// LocalPath struct.
type LocalPath struct {
	LocalDir    string
	LocalDomain string
	PatcherPath string
}

// Mail struct.
type Mail struct {
	AppBuilder *MailConfig
	BanBenJi   *MailConfig
}

type MailConfig struct {
	Host    string
	Port    int
	Address string
	Pwd     string
	Name    string
}

// CDN struct.
type CDN struct {
	SecretID          string
	Signature         string
	RefreshURL        string
	RefreshAction     string
	RefreshAccountIDs string
}

// BFSCDN struct.
type BFSCDN struct {
	RefreshURL string
}

// Keys struct.
type Keys struct {
	AesKey string
}

// System version.
type System struct {
	IOS     []string
	Android []string
}

// Host hosts
type Host struct {
	Easyst string
	Saga   string
	Sven   string
	Bap    string
	Fawkes string
	Bender string
}

// Easyst struct
type Easyst struct {
	User     string
	Platform string
}

// ClickHouse struct.
type ClickHouse struct {
	Monitor  *clickSQL.Config
	Monitor2 *clickSQL.Config
}

// AppstoreConnect struct.
type AppstoreConnect struct {
	Expire           int64
	Audience         string
	BaseURL          string
	ITMSTransporter  string
	KeyPath          string
	TestersThreshold int
	BuglyUploader    string
	DisPermilLimit   int
}

// Cron struct.
type Cron struct {
	CronSwitch                string
	LoadUsers                 string
	LoadApmParams             string
	LoadVersion               string
	LoadModuleListAll         string
	LoadFawkesMoni            string
	LoadFawkesMoniMergeNotice string
	LoadPackAll               string
	LoadBizApkListAll         string
	LoadTribeListAll          string
	LoadUpgradConfigAll       string
	LoadVersionAll            string
	LoadHotfixAll             string
	LoadFlowConfigAll         string
}

type AlarmReceiver struct {
	// testflight 包上传app store 监控
	UploadMonitorReceiver []string
	// 技术埋点 监控
	EventMonitorReceiver []string
	// 渠道包自动构建
	ChannelPackAutoBuildReceiver []string
	// MOD网络带宽告警
	ModTrafficReceiver []string
}

// Comet struct
type Comet struct {
	FawkesAppID  string
	MonitorAppID string
	SecretID     string
	Signature    string
	CometUrl     string
	WorkflowUrl  string
	ProcessUrl   string
}

// EP struct
type Ep struct {
	MonkeyUrl  string
	MonkeyAuth string
}

type BroadcastPush struct {
	Operation int    // operation number
	QPS       int    // qps limit
	URL       string // push url
	Expire    time.Duration
}

type BroadcastGrpc struct {
	Laser        *BroadcastChannel
	LaserCommand *BroadcastChannel
	Module       *BroadcastChannel
	SGPProxy     *GrpcProxy // 新加坡proxy
}
type BroadcastChannel struct {
	TargetPath string
	Token      string
	Ratelimit  int32
}

type GrpcProxy struct {
	Host        string
	DiscoveryId string
}

type Redis struct {
	Fawkes *redis.Config
}

type Databus struct {
	Discovery string
	AppID     string
	Token     string
	Topics    *DatabusTopics
}

type DatabusTopics struct {
	PackGreyDataPub *PackGreyDataPub
}

type PackGreyDataPub struct {
	Group string
	Name  string
}

type Mod struct {
	ModCDN        map[string]string
	DisableModule map[string][]string
	PoolKey       map[string][]string
	PriorityMod   map[string][]string
	CDN           map[string]*modCDN
	PCDN          *PCDN
	TrafficMoni   *TrafficMoni
	Switch        *ModSwitch
	Peak          []*Peak
}

type TrafficMoni struct {
	NotifyReceiver []string
	Threshold      int64
	TimeOffSet     string
	TimeSlice      string
	ModUrl         string
	SamplingRate   float64
	DownloadRate   float64
	CDNRatio       float64 // 带宽与5min下载量之间的换算系数 下载量乘以该系数得到带宽
	PatchRate      float64
	Boundary       map[string]float64
	Advice         map[string]string
	DocURL         string
}

type PCDN struct {
	AppKey []string
}

type Peak struct {
	Start string
	End   string
}

type modCDN struct {
	NewDomain string
	OldDomain string
	Bucket    uint64
}

type Prometheus struct {
	LocalPath *LocalPath
	Database  *Database
}

type Database struct {
	Name     string
	Host     string
	Port     int64
	User     string
	Password string
}

// FlinkJob struct
type FlinkJob struct {
	LocalPath *LocalPath
}

type LongMerge struct {
	Duration            string
	StatisticalDuration string
}

// Moni struct
type Moni struct {
	LongMerge *LongMerge
}

// Switch struct
type Switch struct {
	PackAutoUploadCDN *PackAutoUploadCDN
	TribeDir          *TribeDir
}

type PackAutoUploadCDN struct {
	WhiteList []string
}

type TribeDir struct {
	WhiteList []string
}

// Datacenter struct
type Datacenter struct {
	Host    string
	Dir     string
	Add     string
	Update  string
	Del     string
	OpenAPI *DatacenterOpenAPI
}

type DatacenterOpenAPI struct {
	Account   string
	Dir       string
	SecretKey string
}

type Billions struct {
	Host               string
	Dir                string
	AutoAdd            string
	MappingUpdate      string
	Lifecycle          string
	TreeID             string
	DeployLocations    string
	AuthorizationToken string
	Cluster            string
}

type ElasticsearchProxy struct {
	Host    string
	Dir     string
	Search  string
	Cluster string
	Token   string
}

type IPDB struct {
	Ipv4 string
}

type Task struct {
	NasClean   *NasClean
	MoveTribe  *MoveTribe
	VedaUpdate *VedaUpdate
}

type NasClean struct {
	CIDelete      *CIDelete
	PatchDelete   *PatchDelete
	ChannelDelete *ChannelDelete
}

type CIDelete struct {
	PackType    []int64
	Start       xtime.Time
	End         xtime.Time
	AppKey      string
	Persistence int // 保留时长 单位月
}

type PatchDelete struct {
	AppKey        []string
	ExcludeAppKey []string
	Persistence   int // 保留时长 单位月
}

type ChannelDelete struct {
	AppKey        []string
	ExcludeAppKey []string
	Persistence   int // 保留时长 单位月
}

type MoveTribe struct {
	Apps      []string
	OldDir    string
	NewDir    string
	OldUrl    string
	NewUrl    string
	BatchSize int // 一次执行的条数
	Batch     int // 执行次数
}

type VedaUpdate struct {
	Apps        []string
	Persistence int // 更新多久之前 单位月
	Count       int // 每次更新数量
}

type ExePath struct {
	Tribe *Tribe
}

type Tribe struct {
	TribeAPI string
}

type TAPD struct {
	Token string
}

type BillionsAlert struct {
	Host    string
	Dir     string
	Alert   string
	RuleOpt string
	Token   string
}

type PrometheusTemplate struct {
	Key   string
	Value string
}

type ModSwitch struct {
	Patch *PatchSwitch
}

type PatchSwitch struct {
	FileUrl []string
}
