package conf

import (
	"sync/atomic"
	innerTime "time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/http/blademaster/middleware/rate/quota"
	"go-common/library/net/http/blademaster/middleware/verify"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	rpcquota "go-common/library/net/rpc/warden/ratelimiter/quota"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/time"
	newConf "go-gateway/app/web-svr/activity/tools/lib/conf"

	"go-gateway/app/web-svr/esports/interface/tool"
)

// global var
var (
	// Conf config
	Conf = &Config{}
)

// Config config set
type Config struct {
	// base
	// elk
	Log *log.Config
	// rpc server2
	RPCServer2 *rpc.ServerConfig
	// tracer
	Tracer *trace.Config
	// bm
	BM *bm.ServerConfig
	// Ecode
	Ecode *ecode.Config
	// grpc
	ArcClient            *warden.ClientConfig
	FavClient            *warden.ClientConfig
	ActClient            *warden.ClientConfig
	AccClient            *warden.ClientConfig
	CoinClient           *warden.ClientConfig
	LiveClient           *warden.ClientConfig
	EsportsServiceClient *warden.ClientConfig
	// Mysql
	Mysql *sql.Config
	// MysqlMaster
	MysqlMaster *sql.Config
	// Redis
	Redis *Redis
	// Auto subscribe redis
	AutoSubCache *redis.Config
	// HTTP client
	HTTPClient       *bm.ClientConfig
	Elastic          *elastic.Config
	TunnelDatabusPub *databus.Config
	// Host
	Host *Host
	// Auth
	Auth *auth.Config
	// verify
	Verify *verify.Config
	// reload
	Rule *Rule
	// leidata
	Leidata *Leidata
	// GameTypes game types.
	GameTypes []*types
	// Interval
	Interval *interval

	CustomLimiters map[string]int64

	Memcached4UserGuess *memcache.Config

	Memcached          *memcache.Config
	SeasonContestWatch *SeasonContestWatch
	Limiter            *quota.Config
	RpcLimiter         *rpcquota.Config

	// contest component.
	SeasonContestComponentWatch *SeasonContestComponentWatch

	RankingDataWatch *RankingDataWatch
	// score
	Score *Score

	CircuitBreaker   map[string]tool.BreakerSetting
	HotContestIDList []int64
	// 进行中赛事.
	GoingMatchs *GoingMatch
	// databus v2.
	Databus *DatabusV2
	// tunnelBGroup
	TunnelBGroup            *TunnelBGroup
	SeriesIgnoreTeamsIDList []int64
}

type GoingMatch struct {
	MatchIDs     []int64
	GoingSeasons []int64
	ReserveMap   map[string]int64
}

type DatabusV2 struct {
	Target string `toml:"target"`
	AppID  string `toml:"app_id"`
	Token  string `toml:"token"`
	Topic  struct {
		BGroupMessage string
	}
}

type TunnelBGroup struct {
	Source      string
	SendNew     int64
	NewContests []int64
}

type SeasonContestComponentWatch struct {
	CanWatch bool
}

type SeasonContestWatch struct {
	SeasonID                  int64
	ShowLPL                   bool
	MatchType                 int // 1: LOL
	UniqKey                   string
	StartTime                 string
	EndTime                   string
	ContestAvCIDListCacheKey  string
	ContestListCacheKey       string
	ContestIDListCacheKey     string
	ContestMatchIDMapCacheKey string
	ContestSeriesMapCacheKey  string
	CacheKey4TeamAnalysis     string
	CacheKey4PlayerAnalysis   string
	CacheKey4HeroAnalysis     string
	CacheKey4PosterList       string
	TeamScoreMapCacheKey      string
	ExpiredDuration           int32
	TabCovers                 map[string]string
	SeasonConfiguration       *SeasonConfiguration
}

type SeasonConfiguration struct {
	Champion  ContestChampion
	Desc      string
	OffSeason bool
}

type ContestChampion struct {
	TeamName string `json:"team_name"`
}

type RankingDataWatch struct {
	CurrentRoundIDCacheKey string
	RoundIDListCacheKey    string
	RoundDataCacheKeyPre   string
	InterventionCacheKey   string
	WatchDuration          time.Duration
	Description            struct {
		Finalist          string
		Final             string
		FinalistPoint     string
		FinalistEliminate string
		FinalPoint        string
		FinalEliminate    string
	}
}

// Host hosts.
type Host struct {
	Search string
	Es     string
}

// Redis redis struct
type Redis struct {
	*redis.Config
	FilterExpire time.Duration
	ListExpire   time.Duration
	GuessExpire  time.Duration
	TreeExpire   time.Duration
}

// Rule rule .
type Rule struct {
	S9GuessMax      int
	S9SwitchSID     int64
	S9Sleep         time.Duration
	JumpURL         string
	TunnelPushBizID int64
}

// Score score.
type Score struct {
	LiveInterval int64
}

// Leidata lei da data .
type Leidata struct {
	Timeout         time.Duration
	GameEnd         time.Duration
	LolPlayersCron  string
	DotaPlayersCron string
	OwPlayersCron   string
	InfoCron        string
	BigDataCron     string
}

// interval .
type interval struct {
	KnockTreeCron string
	S9ContestCron string
}

type types struct {
	ID       int64
	Name     string
	DbGameID int64
}

func (s *SeasonContestWatch) CanWatch() bool {
	if s.StartTime == "" || s.EndTime == "" || s.SeasonID <= 0 || s.ContestListCacheKey == "" ||
		s.ExpiredDuration <= 0 || s.ContestIDListCacheKey == "" {
		return false
	}

	startTime, err := innerTime.ParseInLocation("2006-01-02 15:04:05", s.StartTime, innerTime.Local)
	if err != nil {
		return false
	}
	endTime, err := innerTime.ParseInLocation("2006-01-02 15:04:05", s.EndTime, innerTime.Local)
	if err != nil {
		return false
	}

	now := innerTime.Now()
	if now.Before(startTime) || now.After(endTime) {
		return false
	}

	return true
}

var (
	S10SeasonContestWatch atomic.Value
	S10TabCovers          atomic.Value
	S10ContestCfg         atomic.Value
	S10ShowLPL            atomic.Value

	SeasonContestComponentCfgWatch atomic.Value

	GlobalTabCovers map[string]string
)

func init() {

	GlobalTabCovers = make(map[string]string, 0)
}

// Init init conf
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
	storeSeasonContestWatch(tmpConf.SeasonContestWatch)
	storeS10TabCovers(tmpConf.SeasonContestWatch.TabCovers)
	storeS10SeasonCfg(tmpConf.SeasonContestWatch.SeasonConfiguration)
	tool.RestoreLimiters(tmpConf.CustomLimiters)
	tool.RestoreCbsByCfg(tmpConf.CircuitBreaker)

	// contest component
	storeSeasonContestComponentWatch(tmpConf.SeasonContestComponentWatch)

	return
}

func LoadSeasonContestWatch() *SeasonContestWatch {
	return S10SeasonContestWatch.Load().(*SeasonContestWatch)
}

func LoadShowLPL() bool {
	return S10ShowLPL.Load().(bool)
}

func storeSeasonContestWatch(d *SeasonContestWatch) {
	S10ShowLPL.Store(d.ShowLPL)
	S10SeasonContestWatch.Store(d)
}

func LoadS10TabCovers() map[string]string {
	return S10TabCovers.Load().(map[string]string)
}

func storeS10TabCovers(d map[string]string) {
	S10TabCovers.Store(d)

	GlobalTabCovers = d
}

func LoadS10SeasonCfg() *SeasonConfiguration {
	cfg := new(SeasonConfiguration)
	if d := S10ContestCfg.Load().(*SeasonConfiguration); d != nil {
		cfg = d
	}

	return cfg
}

func storeS10SeasonCfg(cfg *SeasonConfiguration) {
	S10ContestCfg.Store(cfg)
}

func LoadSeasonContestComponentWatch() *SeasonContestComponentWatch {
	return SeasonContestComponentCfgWatch.Load().(*SeasonContestComponentWatch)
}

func storeSeasonContestComponentWatch(d *SeasonContestComponentWatch) {
	SeasonContestComponentCfgWatch.Store(d)
}
