package conf

import (
	"sync/atomic"
	innerTime "time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/time"
	newConf "go-gateway/app/web-svr/activity/tools/lib/conf"

	"go-gateway/app/web-svr/esports/job/tool"
)

var (
	// Conf config
	Conf = &Config{}

	seasonContestsMap        atomic.Value
	scoreAnalysisCfg         atomic.Value
	ScoreAnalysisRestartChan chan int

	SeasonNotifies atomic.Value
	// contest component.
	seasonContestComponentCfg atomic.Value
)

// FTP represents the ftp login info
type FTP struct {
	Pass    string
	User    string
	Host    string
	Timeout time.Duration // timeout in seconds
}

// Search represents the config for the search suggestion module
type Search struct {
	FTP               *FTP // the ftp info
	FtpDataCron       string
	LocalFile         string
	LocalMD5File      string
	RemotePath        string
	RemoteFileName    string
	RemoteMD5FileName string
}

type SeasonNotify struct {
	SeasonID         int64
	UniqID           string
	HttpNotifies     []string
	StartTime        string
	EndTime          string
	NotifyInterval   int64
	NotifyTimes      int
	WebhookUrl       string
	WebhookNotify    bool
	WebhookReceivers []string
	WebhookTels      []string
}

// Config .
type Config struct {
	Log      *log.Config
	BM       *bm.ServerConfig
	Tracer   *trace.Config
	Memcache *memcache.Config
	// db
	Mysql *sql.Config
	Ecode *ecode.Config
	// Host
	Host Host
	// HTTP client
	HTTPClient        *bm.ClientConfig
	MessageHTTPClient *bm.ClientConfig
	LeidaHTTPClient   *bm.ClientConfig
	// Rule
	Rule *Rule
	//Push push urls
	Push *Push
	//Message
	Message Message
	// App
	App *bm.App
	// Warden Client
	TagRPC           *warden.ClientConfig
	ArcClient        *warden.ClientConfig
	FavClient        *warden.ClientConfig
	TunnelClient     *warden.ClientConfig
	ActClient        *warden.ClientConfig
	EspClient        *warden.ClientConfig
	EspServiceClient *warden.ClientConfig
	// leidata
	Leidata *Leidata
	// GameTypes game types.
	GameTypes []*types
	// score
	Score  *Score
	Search *Search
	// databus
	ArchiveNotifySub *databus.Config
	EsportsBinlog    *databus.Config
	// Interval
	Interval *interval

	// berserker configuration
	Berserker             tool.BerserkerCfg
	CorpWeChat            tool.CorpWeChat
	SeasonContestWatchMap map[string]*SeasonContestWatch
	RankingDataWatch      *RankingDataWatch
	ScoreAnalysisConfig   *ScoreAnalysisConfig
	// contest component
	SeasonContestComponent *SeasonContestComponent
	// redis
	Redis *Redis

	SeasonStatusNotifier map[string]SeasonNotify
	TunnelDatabusPub     *databus.Config
	ContestSchedulePub   *databus.Config

	// Auto subscribe redis
	AutoSubCache *redis.Config
	// databus v2.
	Databus *DatabusV2
	// tunnelBGroup
	TunnelBGroup         *TunnelBGroup
	SeriesRefresh        *SeriesRefreshConfig
	ContestStatusRefresh *ContestStatusRefresh
	OlympicConf          *OlympicConf
	DeployOption         *DeployOption
}

type DeployOption struct {
	// 可以支持写的部署
	WriteDeploy bool
	Switch      map[string]bool
}

type OlympicConf struct {
	Open                     bool
	PreContest               bool
	OlympicLocalFile         string
	OlympicLocalMD5File      string
	OlympicRemotePath        string
	OlympicRemoteFileName    string
	OlympicRemoteMD5FileName string
}

type ContestStatusRefresh struct {
	RefreshSwitchDo bool
	RefreshDuration time.Duration
}

type TunnelBGroup struct {
	Source      string
	SendNew     int64
	NewContests []int64
}

type DatabusV2 struct {
	Target string `toml:"target"`
	AppID  string `toml:"app_id"`
	Token  string `toml:"token"`
	Topic  struct {
		BGroupMessage   string
		ContestSchedule string
	}
}

type SeriesRefreshConfig struct {
	RefreshDuration     time.Duration
	RefreshIgnoreIDList []int64
}

type ScoreAnalysisConfig struct {
	TournamentID    int64
	StartTime       int64
	EndTime         int64
	Interval        int64
	Enabled         bool
	CacheKey4Team   string
	CacheKey4Player string
	CacheKey4Hero   string
	Expiration      int32
}

type RankingDataWatch struct {
	TournamentID           string
	CurrentRoundIDCacheKey string
	RoundIDListCacheKey    string
	InterventionCacheKey   string
	RoundDataCacheKeyPre   string
	Cron                   string
}

type SeasonContestWatch struct {
	SeasonID                  int64
	MatchType                 int // 1: LOL
	UniqKey                   string
	StartTime                 string
	EndTime                   string
	FetchAll                  bool
	ContestAvCIDListCacheKey  string
	ContestListCacheKey       string
	ContestIDListCacheKey     string
	ContestMatchIDMapCacheKey string
	ContestSeriesMapCacheKey  string
	CacheKey4PosterList       string
	TeamScoreMapCacheKey      string
	ExpiredDuration           int32
}

type SeasonContestComponent struct {
	CanWatch        bool
	StartTimeBefore time.Duration
	EndTimeAfter    time.Duration
	ExpiredDuration int32
}

// Redis config
type Redis struct {
	*redis.Config
	ScoreLiveExpire time.Duration
}

// Push push.
type Push struct {
	BusinessID    int
	BusinessToken string
	PartSize      int
	RetryTimes    int
	Title         string
	BodyDefault   string
	BodySpecial   string
	OnlyMids      string
}

// Message .
type Message struct {
	URL string
	MC  string
}

// Rule .
type Rule struct {
	SleepInterval    time.Duration
	Before           time.Duration
	ScoreSleep       time.Duration
	AlertTitle       string
	AlertBodyDefault string
	AlertBodySpecial string
	CoinPercent      float64
	FavPercent       float64
	DmPercent        float64
	ReplyPercent     float64
	ViewPercent      float64
	LikePercent      float64
	SharePercent     float64
	NewDay           float64
	NewPercent       float64
	TunnelBizID      int64
	AutoFileSwitch   int64
	AutoOfficialTid  int64
	AutoPassView     int32
}

// Host remote host
type Host struct {
	API string
}

// DB define MySQL config
type DB struct {
	Esports *sql.Config
}

// Leidata lei da data .
type Leidata struct {
	Timeout     time.Duration
	RecentSleep time.Duration
	ConnTime    time.Duration
	BindTime    time.Duration
	GroupURL    string
	Socket      string
	Key         string
	Origin      string
	IP          string
	Hero        *HeroVersion
	After       *struct {
		Retry         int
		BigDataCron   string
		InfoDataCron  string
		GameSleepCron string
		URL           string
		Key           string
		LolGameID     int
		DotaGameID    int
		GameEnd       time.Duration
	}
}

// Score score.
type Score struct {
	Key                 string
	Secret              string
	URL                 string
	LiveBackupImg       string
	OfflineTournamentID string
}

// HeroVersion hero version .
type HeroVersion struct {
	Version string
	IDs     []int
}

type types struct {
	ID   int64
	Name string
}

type interval struct {
	AutoArcRuleCron  string
	OffLineImageCron string
	AutoArcPassCron  string
}

func init() {
	m := make(map[string]*SeasonContestWatch)
	storeSeasonWatchMap(m)

	ScoreAnalysisRestartChan = make(chan int, 1)

	tmpCfg := new(ScoreAnalysisConfig)
	scoreAnalysisCfg.Store(tmpCfg)

	// contest component.
	tmpConponentCfg := new(SeasonContestComponent)
	seasonContestComponentCfg.Store(tmpConponentCfg)
}

func LoadSeasonNotifies(m map[string]SeasonNotify) {
	SeasonNotifies.Store(m)
}

// Init init conf
func Init() error {
	return newConf.Init(load)
}

func load() (err error) {
	var tmpConf *Config
	if err = newConf.LoadInto(&tmpConf); err != nil {
		return err
	}
	tool.UpdateCropWeChat(tmpConf.CorpWeChat)
	storeScoreAnalysisCfg(tmpConf.ScoreAnalysisConfig)
	storeSeasonWatchMap(tmpConf.SeasonContestWatchMap)
	LoadSeasonNotifies(tmpConf.SeasonStatusNotifier)
	//contest component.
	storeSeasonContestComponentCfg(tmpConf.SeasonContestComponent)

	*Conf = *tmpConf
	return
}

func storeScoreAnalysisCfg(cfg *ScoreAnalysisConfig) {
	if isScoreAnalysisCfgUpdated(cfg) {
		defer func() {
			_ = recover()
		}()

		close(ScoreAnalysisRestartChan)
	}

	scoreAnalysisCfg.Store(cfg)
}

func LoadScoreAnalysisCfg() *ScoreAnalysisConfig {
	return scoreAnalysisCfg.Load().(*ScoreAnalysisConfig)
}

func isScoreAnalysisCfgUpdated(new *ScoreAnalysisConfig) bool {
	old := scoreAnalysisCfg.Load().(*ScoreAnalysisConfig)
	if old == nil || old.TournamentID == 0 {
		return false
	}

	return old.StartTime != new.StartTime || old.EndTime != new.EndTime ||
		old.CacheKey4Hero != new.CacheKey4Hero || old.CacheKey4Player != new.CacheKey4Player ||
		old.CacheKey4Team != new.CacheKey4Team || old.TournamentID != new.TournamentID
}

func IsScoreAnalysisEnabled() bool {
	if cfg := scoreAnalysisCfg.Load().(*ScoreAnalysisConfig); cfg != nil {
		return cfg.Enabled
	}

	return false
}

func storeSeasonWatchMap(m map[string]*SeasonContestWatch) {
	seasonContestsMap.Store(m)
}

func LoadSeasonWatchMap() map[string]*SeasonContestWatch {
	return seasonContestsMap.Load().(map[string]*SeasonContestWatch)
}

func storeSeasonContestComponentCfg(cfg *SeasonContestComponent) {
	seasonContestComponentCfg.Store(cfg)
}

func LoadSeasonContestComponentCfg() *SeasonContestComponent {
	return seasonContestComponentCfg.Load().(*SeasonContestComponent)
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
