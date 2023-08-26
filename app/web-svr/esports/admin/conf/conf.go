package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/orm"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	newConf "go-gateway/app/web-svr/activity/tools/lib/conf"
)

var (
	// Conf of config
	Conf = &Config{}
)

// Config def.
type Config struct {
	// base
	// http
	BM *bm.ServerConfig
	// db
	ORM *orm.Config
	// log
	Log *log.Config
	// tracer
	Tracer *trace.Config
	// rule
	Rule *Rule
	// client
	HTTPReply *bm.ClientConfig
	HTTPJob   *bm.ClientConfig
	// Warden Client
	ArcClient      *warden.ClientConfig
	AccClient      *warden.ClientConfig
	ActClient      *warden.ClientConfig
	ACPRPC         *warden.ClientConfig
	RoomGRPC       *warden.ClientConfig
	TunnelClient   *warden.ClientConfig
	BGroupClient   *warden.ClientConfig
	TunnelV2Client *warden.ClientConfig
	// GameTypes game types.
	GameTypes []*types
	// Host
	Host                  Host
	S10CoinCfg            *S10CoinConfig
	EspClient             *warden.ClientConfig
	EsportsServiceClient  *warden.ClientConfig
	ActivityServiceClient *warden.ClientConfig
	TunnelPush            *TunnelPush
	// tunnelBGroup
	TunnelBGroup     *TunnelBGroup
	Memcached        *memcache.Config
	RankingDataWatch *RankingDataWatch

	// Auto subscribe redis
	AutoSubCache *redis.Config
}

type RankingDataWatch struct {
	InterventionCacheKey string
}

type S10CoinConfig struct {
	SeasonID  int64
	GameState int64
}

// Rule .
type Rule struct {
	MaxCSVRows       int
	MaxAutoRows      int
	MaxBatchArcLimit int
	MaxTreeContests  int
	MaxGuessStake    int64
	MatchFixLimit    int64
	UserNameLimit    []string
}

type TunnelPush struct {
	TunnelBizID int64
	TemplateID  int64
	Link        string
}

type TunnelBGroup struct {
	TunnelBizID   int64
	NewBusiness   string
	NewTemplateID int64
	NewCardText   string
	Link          string
	SendNew       int64
	NewCardLiveID int64
}

type types struct {
	ID   int64
	Name string
}

// Host remote host.
type Host struct {
	APICo    string
	GenPost  string
	SavePost string
}

func Init() (err error) {
	return newConf.Init(load)
}

func load() (err error) {
	var (
		tmpConf *Config
	)
	if err = newConf.LoadInto(&tmpConf); err != nil {
		return err
	}
	*Conf = *tmpConf
	return
}
