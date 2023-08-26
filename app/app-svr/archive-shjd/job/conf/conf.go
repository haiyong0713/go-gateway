package conf

import (
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"
	"go-common/library/railgun"

	"github.com/BurntSushi/toml"
)

type Config struct {
	// Env
	Env string
	// interface XLog
	Log *log.Config
	// databus
	Databus        *databus.Config
	ViewSubRedis   *databus.Config
	DmSubRedis     *databus.Config
	ReplySubRedis  *databus.Config
	FavSubRedis    *databus.Config
	CoinSubRedis   *databus.Config
	ShareSubRedis  *databus.Config
	RankSubRedis   *databus.Config
	LikeSubRedis   *databus.Config
	FollowSubRedis *databus.Config
	NotifyPub      *databus.Config
	CacheSub       *databus.Config
	// http
	BM *bm.ServerConfig
	// redis
	Redis *redis.Config
	// stat-jd-job 自用的redis
	StatRedis *redis.Config
	// stat-jd-job 需更新的arc service的redis集群
	ArcRedises []*redis.Config
	// 简易版稿件缓存
	SimpleArcRedis []*redis.Config
	// DB
	DB      *DB
	Custom  *Custom
	Taishan *Taishan
	// 新增点赞云立方、嘉定消息
	LikeYLFRailgun *SingleRailgun
	LikeJDRailgun  *SingleRailgun
	// 切流点赞新消息灰度控制
	LikeRailgunWhitelist map[string]int64
	LikeRailgunGray      int64
	//社区IP地址转化服务
	LocationClient *warden.ClientConfig
}

type SingleRailgun struct {
	Cfg     *railgun.Config
	Databus *railgun.DatabusV1Config
	Single  *railgun.SingleConfig
}

type Taishan struct {
	Table string
	Token string
}

// Custom is
type Custom struct {
	RedisAvExpireTime int64 // stat-redis 中的稿件缓存过期时间
	ProcCount         int
	CanalAlertTime    int64 // canal延迟时间报警
}

// DB is db config.
type DB struct {
	Result *sql.Config
	Stat   *sql.Config
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
