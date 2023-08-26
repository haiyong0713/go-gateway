package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	xtime "go-common/library/time"

	infocv2 "go-common/library/log/infoc.v2"

	"github.com/BurntSushi/toml"
)

type Config struct {
	// Env
	Env string
	// show  XLog
	Log *log.Config
	// tick time
	Tick xtime.Duration
	//quicker time
	QuickerTick xtime.Duration
	// tracer
	Tracer *trace.Config
	// httpClinet
	HTTPClient     *bm.ClientConfig
	HTTPGame       *bm.ClientConfig
	HTTPClientAsyn *bm.ClientConfig
	HTTPWechat     *bm.ClientConfig
	HTTPHuawei     *bm.ClientConfig
	HTTPTrace      *bm.ClientConfig
	// bm http
	BM *HTTPServers
	// db
	Ecode *ecode.Config
	MySQL *MySQL
	// duration
	Duration *Duration
	// Splash
	Splash *Splash
	// interestJSONFile
	InterestJSONFile string
	// StaticJsonFile
	StaticJSONFile string
	// guide rand
	GuideRandom *GuideRandom
	// domain
	Domain *Domain
	ABTest *ABTest
	// host
	Host *Host
	// sideBar limit id
	SideBarLimit []int64
	// resource
	ResourceRPC            *rpc.ClientConfig
	ResourceClient         *warden.ClientConfig
	GarbClient             *warden.ClientConfig
	AccountClient          *warden.ClientConfig
	CollectionSplashClient *warden.ClientConfig
	SplashInfoc            *infocv2.Config
	// BroadcastRPC grpc
	BroadcastRPC *warden.ClientConfig
	// White
	White *White
	// 垃圾白名单
	ShowTabMids []int64
	// location grpc
	LocationGRPC *warden.ClientConfig
	// fission grpc
	FissionGRPC *warden.ClientConfig
	ADClient    *warden.ClientConfig
	//bgroup
	BGroupClient *warden.ClientConfig
	// show hot all
	ShowHotAll bool
	// rpc server2
	RPCServer *rpc.ServerConfig
	// wechant
	WeChant *WeChant
	// mc
	Memcache *Memcache
	// mod 低优先级 pool
	ModLowPool []string
	ModMobiApp []string
	// 自定义配置
	Custom *Custom
	// tf
	TF *TF
	// cron
	Cron *Cron
	// test
	CDNTest []*CDN
	// 品牌闪屏配置
	BrandSplash *BrandSplash
	// 隐私设置配置
	Privacy *Privacy
	// dynamic location grpc
	DynamicLocGRPC *warden.ClientConfig
	// dynamic campus grpc
	DynamicCampusGRPC *warden.ClientConfig
	// WechatAuth 微信登陆
	WechatAuth *WechatAuth
	Mod        *Mod
	// Redis
	Redis      *Redis
	ModLogGray *modLogGray
	// Huawei deeplink secret key
	HuaweiSecretKey             string
	RegistrationDateEventConfig *RegistrationDateEventConfig
	// EntranceKeyExpire
	EntranceKeyExpire *EntranceKeyExpire
	// feature配置
	Feature *Feature
	// Experiment
	Experiment *Experiment
	// 杜比配置
	Dolby *Dolby
	// 泰山RPC
	TaishanRPC *warden.ClientConfig
	// 存储deeplink泰山配置
	DeeplinkTaishanCfg *DeeplinkTaishanCfg
	//版本包推送model
	PackagePushModel map[string]int64
	// topleft白名单
	TopLeftStoryMids []int64
	// topleft 老用户分桶实验
	TopLeftExpGroup map[string]int64
	// HostDiscovery
	HostDiscovery *HostDiscovery
	// HostDiscovery
	DWConfig    *DWConfig
	ABTestFlags []string
}

type DWConfig struct {
	DomainList map[string]int64
}

type HostDiscovery struct {
	CommonArch  string
	AdDiscovery string
}

type DeeplinkTaishanCfg struct {
	Open              bool
	UserTable         string
	UserTableToken    string
	ArchiveTable      string
	ArchiveTableToken string
}

type modLogGray struct {
	Open      bool
	Whitelist []int64
	Bucket    uint64
}

type ExpMidGroup struct {
	Sharding []string
}

type CDN struct {
	PoolName string
	ModName  string
	URL      string
	MD5      string
	Size     int
}

type HTTPServers struct {
	Outer *bm.ServerConfig
}

type Host struct {
	Ad      string
	Data    string
	VC      string
	DP      string
	Fawkes  string
	API     string
	Bap     string
	Manager string
	Wechat  string
	Search  string
	Huawei  string
	School  string
}

type WechatAuth struct {
	Appid            string
	Secret           string
	ClientCredential string
}

type White struct {
	List map[string][]string
}

type ABTest struct {
	Range int
}

type GuideRandom struct {
	Random map[string]int
	Buvid  map[string]int
	Feed   uint32
}

type Duration struct {
	// splash
	Splash string
}

type EntranceKeyExpire struct {
	LiveReserveBlock int
}

type Splash struct {
	Random      map[string][]string
	WhiteFile   string
	AbtestState int
}

type BrandSplash struct {
	LogoURL             string
	Duration            int64
	PullInterval        int64
	DefaultTitle        string
	DefaultType         string
	ProbabilityTitle    string
	ProbabilityType     string
	Desc                string
	ShowTitle           string
	CollectionShowTitle string
	SplitDuration       BrandSplashSplitDuration
}

type BrandSplashSplitDuration struct {
	HalfDuration int64
	FullDuration int64
}

type MySQL struct {
	Show *sql.Config
}

type Domain struct {
	Addr      []string
	ImageAddr []string
}

// WeChant is
type WeChant struct {
	Token  string
	Secret string
	Users  []string
}

// Memcache struct
type Memcache struct {
	Bubble *struct {
		*memcache.Config
	}
}

type Custom struct {
	NoLoginAvatarAll bool
	NoLoginAvatar    map[string]struct {
		URL  string
		Type int
	}
	TopActivityInterval                 int64
	TopActivityMngSwitch                bool
	TopActivityFissionSwitch            bool
	PopUpAutoClose                      bool
	PopUpAutoCloseTime                  int
	TopLeftHeadTag                      string
	TopLeftDefaultUrl                   string
	TopLeftSpecialUrl                   string
	TopLeftBlackTime                    int64
	AndroidTopLeftStoryBackgroundImage  string
	AndroidTopLeftStoryForegroundImage  string
	AndroidTopLeftListenBackgroundImage string
	AndroidTopLeftListenForegroundImage string
	IosTopLeftStoryBackgroundImage      string
	IosTopLeftStoryForegroundImage      string
	IosTopLeftListenBackgroundImage     string
	IosTopLeftListenForegroundImage     string
	TinyUpgradeInform                   string
	TinyUpgradeInformTiming             int64
	TinyUpgradeInformTitle              string
	ShowTabLogID                        string
}

type TF struct {
	Rule string
}

type Cron struct {
	LoadAbTest                string
	LoadAuditCache            string
	LoadFawkes                string
	LoadModuleCache           string
	LoadNotice                string
	LoadParam                 string
	LoadPlugin                string
	LoadShowCache             string
	LoadBubbleCache           string
	LoadSkinExtCache          string
	LoadSidebar               string
	LoadSplash                string
	LoadBirth                 string
	LoadWhiteListCache        string
	LoadStaticCache           string
	LoadVersion               string
	LoadBrandSplash           string
	LoadModCache              string
	LoadCollectionBrandSplash string
	LoadDwTime                string
}

type Privacy struct {
	City *struct {
		Title       string
		SubTitle    string
		SubTitleURL string
	}
}

type Mod struct {
	ModuleForbid map[string][]string
	GrayDuration xtime.Duration
	FileHost     FileHost
}

type FileHost struct {
	BFS  string
	BOSS string
}

type Redis struct {
	Resource *struct {
		*redis.Config
		Expire xtime.Duration
	}
	Fawkes *struct {
		*redis.Config
	}
	TopLeft *redis.Config
}

type RegistrationDateEventConfigItem struct {
	ResourceType string
	ImageURL     string
	VideoURI     string
	VideoHash    string
	AccountCard  struct {
		Enable            bool
		MaxWidthPX        int64
		PaddingTopPercent float64
	}
	Greeting struct {
		Enable            bool
		MaxWidthPX        int64
		PaddingTopPercent float64
		Fontsize          int64
		Text              string
	}
	Text struct {
		Enable            bool
		MaxWidthPX        int64
		PaddingTopPercent float64
		Fontsize          int64
		Text              string
	}
}

type RegistrationDateEventConfig struct {
	Disable    bool
	LogoURL    string // http://i0.hdslb.com/bfs/archive/1b1a8a4fc78a3b1b2992402ebdc19808b9d251ed.png
	ShowTimes  int64
	Duration   int64
	URI        string
	SkipButton bool
	Param      string
	Normal     RegistrationDateEventConfigItem
	Full       RegistrationDateEventConfigItem
	Pad        RegistrationDateEventConfigItem
	SE640      RegistrationDateEventConfigItem
	P1080      RegistrationDateEventConfigItem
}

type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}

type FeatureBuildLimit struct {
	Switch         bool
	UpdateAndroidB string
	Splash610      string
	Skin554        string // todo
	SkinDress      string
}

type Experiment struct {
	Config map[string]*ExperimentConfig
}

type ExperimentConfig struct {
	Switch  bool
	ExpName string
	Bucket  int
	ExpType string
	Groups  []*ExperimentGroup
}

type ExperimentGroup struct {
	GroupName string
	Start     int
	End       int
	WhiteList string
}

type Dolby struct {
	DolbyConfig []*DolbyConfig
}

type DolbyConfig struct {
	Brand string
	Model string
	File  string
	Hash  string
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
