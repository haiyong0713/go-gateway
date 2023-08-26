package conf

import (
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/rpc/warden/ratelimiter/quota"
	"go-common/library/net/trace"
	"go-common/library/time"

	"github.com/BurntSushi/toml"
)

// Config struct
type Config struct {
	// base
	Tick      time.Duration
	Videoshot *Videoshot
	// xlog
	Xlog *log.Config
	// tracer
	Tracer *trace.Config
	// http
	BM *BM
	// http client
	PlayerClient *bm.ClientConfig
	// switch get player
	PlayerSwitch bool
	PlayerNum    int64
	// PlayerAPI path
	PlayerAPI       string
	PlayerDiscovery string
	PGCPlayerAPI    string
	PGCPlayerV2API  string
	// db
	DB *DB

	// grpc client
	AccClient          *warden.ClientConfig
	PlayurlClient      *warden.ClientConfig
	VipClient          *warden.ClientConfig
	HisClient          *warden.ClientConfig
	MngClient          *warden.ClientConfig
	StClient           *warden.ClientConfig
	VolumeClient       *warden.ClientConfig
	LocationGRPC       *warden.ClientConfig
	VasGRPC            *warden.ClientConfig
	PassportUserClient *warden.ClientConfig
	SteampunkClient    *warden.ClientConfig
	IPDisplayClient    *warden.ClientConfig

	// app/app-svr/archive/service/service/shot.go
	ArcRedis  *redis.Config
	Redis     *Redis
	Custom    *Custom
	Switch    *Switch
	Taishan   *Taishan
	Cron      *Cron
	QuotaConf *quota.Config
	// ipad清晰度灰度控制
	IpadClarityGrayControl *IpadClarityGrayControl
}

// IpadClarityGrayControl ipad清晰度灰度控制
type IpadClarityGrayControl struct {
	Mid  map[string]int64
	Gray int64
}

type Switch struct {
	VipControl     bool
	HistorySeek    bool
	NoMultiPlayer  bool
	VoiceBalance   bool
	DegreePayCheck bool
}

type Taishan struct {
	Table string
	Token string
}

// Custom is
type Custom struct {
	PlayerQn                   int64
	SteinsGuideAid             int64
	SteinsCallers              []string
	DurationLimit              int64
	FlvProjectGray             uint32
	FourkAndBuild              int64
	FourkIOSBuild              int64
	FourkIPadHDBuild           int64
	BackupNum                  int
	SimplePlayurlIOS           int64
	SimplePlayurlAnd           int64
	SimplePlayurlIpad          int64
	HdrIOS                     int64
	HdrAnd                     int64
	VipFreeAids                []int64
	UserQnGray                 uint32
	WifiUserQnGray             uint32
	UserQnGrayMids             []int64
	WifiUserQnGrayMids         []int64
	ShortLinkHost              string
	HistoryPlayUrlBuildIphone  int64
	HistoryPlayUrlBuildAndroid int64
	StoryQnGroup1              uint32
	StoryQnGroup2              uint32
	StoryQnGroup1Mids          []int64
	StoryQnGroup2Mids          []int64
	CdnScoreGray               uint32
	ScoreInternal              int64
	ScoreRank                  float64
	CdnMids                    []int64
	NologinQnBuvids            []string
	NologinQnGray              uint32
	VideoShotNew               bool
	QnChangeGrey               uint32
	PCDNGrey                   int64
}

// BM http
type BM struct {
	Inner *bm.ServerConfig
}

// Videoshot videoshot uri and key
type Videoshot struct {
	NewURI  string
	BossURI string
}

// DB db config
type DB struct {
	ArcResult *sql.Config
	Stat      *sql.Config
}

// Redis redis config
type Redis struct {
	Archive *struct {
		*redis.Config
		UpRdsExpire int32
	}
	SimpleArc *redis.Config
	CdnScore  *redis.Config
}

type Cron struct {
	LoadShortHost          string
	LoadTypes              string
	LoadRecentPremiereArc  string
	LoadFixedLocation      string
	LoadBuvidFixedLocation string
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
