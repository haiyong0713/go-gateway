package conf

import (
	"encoding/json"
	"errors"

	"go-common/library/cache/redis"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	wardensdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden"

	"github.com/BurntSushi/toml"
)

var (
	WardenSDKBuilder *wardensdk.InterceptorBuilder
)

type WardenSDKConfig struct {
	SDKBuilderConfig *wardensdk.SDKBuilderConfig
}

// Config .
type Config struct {
	Log                 *log.Config
	BM                  *bm.ServerConfig
	Tracer              *trace.Config
	Ecode               *ecode.Config
	GRPC                *warden.ServerConfig
	HTTPClient          *bm.ClientConfig
	HTTPCopyRightClient *bm.ClientConfig
	ResourceRPC         *rpc.ClientConfig
	Host                *Host
	// http discovery
	HostDiscovery *HostDiscovery
	// Warden Client
	AccountClient    *warden.ClientConfig
	ArchiveClient    *warden.ClientConfig
	FavClient        *warden.ClientConfig
	UGCpayClient     *warden.ClientConfig
	UGCpayRankClient *warden.ClientConfig
	PGCPlayerClient  *warden.ClientConfig
	PlayurlClient    *warden.ClientConfig
	SteinsClient     *warden.ClientConfig
	OTTClient        *warden.ClientConfig
	VipClient        *warden.ClientConfig
	TaiShanClient    *warden.ClientConfig
	DmClient         *warden.ClientConfig
	AppConfClient    *warden.ClientConfig
	UGCSeasonClient  *warden.ClientConfig
	Res2GRPC         *warden.ClientConfig
	Broadcast        *warden.ClientConfig
	VipProfileClient *warden.ClientConfig
	H5PlayurlClient  *warden.ClientConfig
	HqPlayurlClient  *warden.ClientConfig
	HlsPlayurlClient *warden.ClientConfig
	VolumeClient     *warden.ClientConfig
	//distribution
	DistributionClient *warden.ClientConfig
	//playurl灾备
	PlayurlDisasterClient *warden.ClientConfig
	SteampunkClient       *warden.ClientConfig
	// Custom
	Custom      *Custom
	Redis       *Redis
	TaiShanConf *TaiShanConf
	// cron
	Cron       *Cron
	InfocConf  *InfocConf
	BuildLimit *BuildLimit
	// feature平台
	Feature    *Feature
	LegoToken  *LegoToken
	GlanceConf *GlanceConf
	//无法投屏的原因
	CastDisabledMsg       map[string]string
	BackgroundDisabledMsg map[string]string
	// ipad清晰度灰度控制
	IpadClarityGrayControl *IpadClarityGrayControl
	NewDeviceWhiteList     map[string]string
}

type GlanceConf struct {
	Times    int64
	Duration int64
	Ratio    int64
}

type LegoToken struct {
	PlayOnlineToken string
}

type BuildLimit struct {
	NewDeviceAndBuild int32
	NewDeviceIOSBuild int32
}

type InfocConf struct {
	CloudInfoc *infoc.Config
	CloudLogID string
	DolbyLogID string
	LiteLogID  string
}

// HostDiscovery Http Discovery
type HostDiscovery struct {
	CopyRight string
}

type Cron struct {
	LoadChronos       string
	LoadPasterCID     string
	LoadSteinsWhite   string
	LoadCustomConfig  string
	LoadManagerConfig string
	LoadVipConfig     string
}

type TaiShanConf struct {
	PlayConfTable string
	PlayConfToken string
}

// Host struct
type Host struct {
	Playurl     string
	PlayurlBk   string
	ManagerHost string
	APICo       string
}

// Custom struct
type Custom struct {
	SteinsWhiteAid  int64
	SteinsCallers   []string
	OTTPlayVerify   int
	FlvProjectGray  uint32
	ElecShowTypeIDs []int32
	//tf gray
	TFGray int64
	//CloudGray grdy
	CloudGray       uint32
	AndChronosBuild int32
	IOSChronosBuild int32
	//vip free aids
	StoryQnGroup1     uint32
	StoryQnGroup2     uint32
	StoryQnGroup1Mids []int64
	StoryQnGroup2Mids []int64
	VipFreeAids       []int64
	ArchiveGray       int64
	// new device ab test
	NewDeviceTime       int64
	NewDeviceSwitchon   bool
	MusicMids           []int64
	MusicAids           []int64
	PlayurlVolumeSwitch bool
	PayArcDegreeAid     int64
	PayArcDegreeCid     int64
	PCDNGrey            int64
}

type Redis struct {
	Vip      *redis.Config
	ArcRedis *redis.Config
	MixRedis *redis.Config //jd,ylf共用一份配置
}

type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}

type FeatureBuildLimit struct {
	Switch      bool
	Chronos     string
	NeedAbility string
}

// IpadClarityGrayControl ipad清晰度灰度控制
type IpadClarityGrayControl struct {
	Mid  map[string]int64
	Gray int64
}

func (c *Config) Set(s string) error {
	var tmp Config
	if _, err := toml.Decode(s, &tmp); err != nil {
		return err
	}
	old, _ := json.Marshal(c)
	nw, _ := json.Marshal(tmp)
	log.Info("service config changed, old=%+v new=%+v", string(old), string(nw))
	*c = tmp
	return nil
}

func (c *WardenSDKConfig) Set(s string) error {
	var tmp WardenSDKConfig
	if _, err := toml.Decode(s, &tmp); err != nil {
		return err
	}
	if tmp.SDKBuilderConfig == nil {
		return errors.New("invalid sdk builder config, maybe empty config")
	}
	old, _ := json.Marshal(c)
	nw, _ := json.Marshal(tmp)
	log.Info("WardenSDKConfig changed, old=%+v new=%+v", string(old), string(nw))
	*c = tmp
	if WardenSDKBuilder == nil {
		WardenSDKBuilder = wardensdk.NewBuilder(*tmp.SDKBuilderConfig)
	} else {
		if err := WardenSDKBuilder.Reload(*tmp.SDKBuilderConfig); err != nil {
			log.Error("WardenSDKConfig changed error, old=%+v new=%+v, err(%+v)", string(old), string(nw), err)
			return err
		}
	}
	return nil
}
