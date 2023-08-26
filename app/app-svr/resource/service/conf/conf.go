package conf

import (
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	xlog "go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/time"

	frontpageModel "go-gateway/app/app-svr/app-feed/admin/model/frontpage"

	"github.com/BurntSushi/toml"
)

// Config service config
type Config struct {
	Version string `toml:"version"`
	// reload
	Reload *ReloadInterval
	// rpc server2
	RPCServer *rpc.ServerConfig
	// verify
	Verify *verify.Config
	// http
	BM *BM
	// tracer
	Tracer *trace.Config
	// db
	DB *DB
	// httpClient
	HTTPClient *bm.ClientConfig
	// Host
	Host *Host
	// XLog
	XLog *xlog.Config
	// rpc
	LocationGRPC  *warden.ClientConfig
	ArchiveRPC    *rpc.ClientConfig
	GarbGRPC      *warden.ClientConfig
	CrowdGRPC     *warden.ClientConfig
	ActivityGRPC  *warden.ClientConfig
	DisplayGRPC   *warden.ClientConfig
	RecommendGRPC *warden.ClientConfig
	OpIconGRPC    *warden.ClientConfig
	// redis
	Redis *Redis
	// hash number
	HashNum int64
	// databus
	ArchiveSub *databus.Config
	// qiye wechat
	WeChatToken   string
	WeChatSecret  string
	WeChantUsers  []string
	WeChantDomain string
	// kai guan off line
	MonitorArchive bool
	MonitorURL     bool
	// sp limit
	SpLimit time.Duration
	// resource label
	ResourceLabel *ResourceLabel
	// archive grpc
	ArchiveGRPC *warden.ClientConfig
	// Article grpc
	ArticleGRPC *warden.ClientConfig
	// VedioGRPC grpc
	VedioGRPC *warden.ClientConfig
	// cron
	Cron             *Cron
	PopEntranceS10Id int64
	// Taishan KV
	Taishan  *Taishan
	BannerID []int64
	// IconCacheConfig
	IconCacheConfig *IconCacheConfig
	ResourceParam   *ResourceParam
	// Frontpage 版头
	Frontpage *FrontpageConfig
	//会员购底tab id
	MallDefaultIDMap map[string]int64
	MallCustomIDMap  map[string]int64
	// hmt channel grpc
	HmtChannelGRPC *warden.ClientConfig
}

type IconCacheConfig struct {
	PreloadDuration int
}

type Taishan struct {
	Popups *struct {
		Table string
		Token string
	}
}

// BM http
type BM struct {
	Inner *bm.ServerConfig
	Local *bm.ServerConfig
}

// ReloadInterval define reolad config
type ReloadInterval struct {
	Ad time.Duration
}

// Host define host info
type Host struct {
	DataPlat string
	Ad       string
	Song     string
}

// DB define MySQL config
type DB struct {
	Res     *sql.Config
	Ads     *sql.Config
	Show    *sql.Config
	Manager *sql.Config
	Player  *sql.Config
	GWDB    *sql.Config
}

// Redis define Redis config
type Redis struct {
	Ads *struct {
		*redis.Config
		Expire time.Duration
	}
	Comm     *redis.Config
	Entrance *redis.Config
	Show     *redis.Config
	Res      *struct {
		*redis.Config
		FrontPageExpire time.Duration
	}
}

type ResourceLabel struct {
	ResourceIDs []int
	PositionIDs []int
}

type Cron struct {
	LoadIconCache        string
	LoadParamCache       string
	FeedPosRecCache      string
	LoadCardCache        string
	LoadTabExtCache      string
	LoadBWListCache      string
	LoadMaterialCache    string
	LoadSpecialCardCache string
}

type ResourceParam struct {
	AppSpecailCardTimeSize int32
}

type FrontpageConfig struct {
	BaseDefaultConfig *frontpageModel.Config
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	xlog.Info("progress-service-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}
