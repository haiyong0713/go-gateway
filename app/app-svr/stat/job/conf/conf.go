package conf

import (
	"errors"
	"flag"

	"go-common/library/cache/redis"
	"go-common/library/conf"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-common/library/time"

	"github.com/BurntSushi/toml"
)

var (
	confPath string
	client   *conf.Client
	// Conf config
	Conf = &Config{}
)

// Config stat-job config
type Config struct {
	// interface XLog
	XLog *log.Config
	// BM
	BM *bm.ServerConfig
	// http client
	HTTPClient *bm.ClientConfig
	// redis databus
	ViewSubRedis   *databus.Config
	DmSubRedis     *databus.Config
	ReplySubRedis  *databus.Config
	FavSubRedis    *databus.Config
	CoinSubRedis   *databus.Config
	ShareSubRedis  *databus.Config
	RankSubRedis   *databus.Config
	LikeSubRedis   *databus.Config
	FollowSubRedis *databus.Config
	// cache
	ArcRedises []*redis.Config
	StatRedis  *redis.Config
	// DB
	DB *sql.Config
	// rpc
	ArchiveGRPC *warden.ClientConfig
	Custom      *Custom
	// 新增点赞云立方、嘉定消息
	LikeYLFRailgun *SingleRailgun
	LikeJDRailgun  *SingleRailgun
	// 切流点赞新消息灰度控制
	LikeRailgunWhitelist map[string]int64
	LikeRailgunGray      int64
}

type SingleRailgun struct {
	Cfg     *railgun.Config
	Databus *railgun.DatabusV1Config
	Single  *railgun.SingleConfig
}

// Custom is
type Custom struct {
	RedisAvExpireTime int64         // stat-redis 中的稿件缓存过期时间
	BabySleepTick     time.Duration // 在处理完一批baby之后sleep多久
	BabyExpire        int64         // 存放冷门稿件的set的过期时间
	LastChangeTime    int64         // 至少距离上次落库多久可以再落
	ProcCount         int           // 处理stat的proc数目
}

func init() {
	flag.StringVar(&confPath, "conf", "", "config path")
}

// Init init conf
func Init() error {
	if confPath != "" {
		return local()
	}
	return remote()
}

func local() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}

// nolint:biligowordcheck
func remote() (err error) {
	if client, err = conf.New(); err != nil {
		return
	}
	if err = load(); err != nil {
		return
	}
	client.Watch("stat-job.toml")
	go func() {
		for range client.Event() {
			log.Info("config reload")
			if load() != nil {
				log.Error("config reload error (%v)", err)
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
