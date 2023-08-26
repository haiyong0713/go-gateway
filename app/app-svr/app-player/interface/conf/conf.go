package conf

import (
	"go-common/library/cache/redis"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"

	"github.com/BurntSushi/toml"
)

var (
	// Conf all conf
	Conf = &Config{}
)

// Config is
type Config struct {
	HTTPClient  *bm.ClientConfig
	ResourceRPC *rpc.ClientConfig
	Ecode       *ecode.Config
	Log         *log.Config
	Host        *Host
	Custom      *Custom
	// Warden Client
	ArchiveClient   *warden.ClientConfig
	UGCpayClient    *warden.ClientConfig
	AccountClient   *warden.ClientConfig
	PlayURLClient   *warden.ClientConfig
	VipClient       *warden.ClientConfig
	LocationGRPC    *warden.ClientConfig
	Switch          *Switch
	HlsSign         *HlsSign
	OttClient       *warden.ClientConfig
	Redis           *Redis
	Feature         *Feature
	RpcServer       *warden.ServerConfig
	AndroidQnShield map[string]string
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

type Redis struct {
	CdnScore *redis.Config
}

type HlsSign struct {
	Key    string
	Secret string
}

type Switch struct {
	VipControl bool
}

// Custom is
type Custom struct {
	PadAid           int64
	PadCid           int64
	PhoneAid         int64
	PhoneCid         int64
	PadHDAid         int64
	PadHDCid         int64
	SteinsBuild      *SteinsBuild
	FourkAndBuild    int32
	FourkIOSBuild    int32
	FourkIPadHDBuild int32
	BackupNum        uint32
	CdnScoreGray     uint32
	ScoreRank        float64
	ScoreInternal    int64
	CdnMids          []int64
	UpgradeInfo      *UpgradeInfo
}

type UpgradeInfo struct {
	UpgradeLimitMessage    string
	UpgradeLimitButtonText string
	PlayLimitMessage       string
	PlayLimitButtonText    string
}

type SteinsBuild struct {
	Android       int32
	Iphone        int32
	IphoneB       int32
	IpadHD        int32
	AndroidI      int32
	IphoneI       int32
	Message       string
	Image         string
	ButtonText    string
	ButtonLink    string
	UseCustomLink bool
	LinkHD        string
	LinkPink      string
	LinkBlue      string
	LinkAndroid   string
}

// Host struct
type Host struct {
	Playurl   string
	PlayurlBk string
}

// feature版本控制
type Feature struct {
	FeatureBuildLimit *FeatureBuildLimit
}

type FeatureBuildLimit struct {
	Switch            bool
	CheckFourk        string
	PlayurlValidPhone string // del
	PlayurlValidPad   string // del
	PlayurlValidHD    string // del
	PlayurlSteins     string
}
