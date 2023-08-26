package conf

import (
	"encoding/json"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/database/tidb"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-common/library/rate/limit/quota"

	"github.com/BurntSushi/toml"
)

// Config is
type Config struct {
	// interface XLog
	XLog *log.Config
	// host
	Host *Host
	// httpClinet
	HTTPClient *bm.ClientConfig
	//grpc
	FlowClient *warden.ClientConfig
	// databus
	VideoupSub         *databus.Config
	ArchiveResultPub   *databus.Config
	CacheSub           *databus.Config
	SeasonNotifyArcSub *databus.Config
	SteinsGateSub      *databus.Config
	InternalSub        *databus.Config
	Taishan            *Taishan
	//rail_gun config
	VideoUpSubV2Config         *RailGunConfig //稿件aid分发rail_gun配置
	SeasonNotifyArcSubV2Config *RailGunConfig //合集稿件rail_gun配置
	SteinsGateSubV2Config      *RailGunConfig //互动视频rail_gun配置
	CacheSubV2Config           *RailGunConfig //缓存rail_gun配置
	InternalSubConfig          *RailGunConfig //稿件部分禁止项rail_gun配置
	LoadTypesCronConfig        *RailGunConfig //更新archive_type定时任务配置
	SyncCreativeTypeCronConfig *RailGunConfig //更新"creative_type"定时任务配置
	CheckConsumeCronConfig     *RailGunConfig //检查consumer数据 定时任务配置
	// DB
	DB *DB
	// BM
	BM *bm.ServerConfig
	// Redis
	Redis          *redis.Config
	UpperRedis     *redis.Config
	ArcRedises     []*redis.Config
	SimpleArcRedis []*redis.Config
	Custom         *Custom
	Limiter        *Limiter
	//b端稿件服务
	VideoupClient *warden.ClientConfig
	// PGCRPC grpc
	PGCRPC *warden.ClientConfig
	//社区IP地址转化服务
	LocationClient *warden.ClientConfig
}

type RailGunConfig struct {
	Cfg             *railgun.Config
	Databus         *railgun.DatabusV1Config
	SingleConfig    *railgun.SingleConfig
	CronInputConfig *railgun.CronInputerConfig
	CronProcConfig  *railgun.CronProcessorConfig
}

type Limiter struct {
	UGC   *quota.WaiterConfig
	OGV   *quota.WaiterConfig
	Retry *quota.WaiterConfig
	Other *quota.WaiterConfig
}

type Taishan struct {
	Table string
	Token string
}

type Custom struct {
	ChanSize                   int
	MonitorSize                int
	DBAlertSec                 int64
	VsMonitorSize              int
	FlowSecret                 string
	PremiereCloseAfter         int64
	PremiereCloseSystemMsg     string
	PremiereEndTipSystemMsg    string
	PremiereRiskCloseSystemMsg string
	//灰度
	GetArchiveAdditGrey         int  //GetArchiveAddit
	GetArchiveGrey              int  //GetArchive
	GetArchiveBizGrey           int  //GetArchiveBiz
	GetArchiveFirstPassGrey     int  //GetArchiveFirstPass
	GetArchiveStaffGrey         int  //GetArchiveStaff
	GetArchiveTypeSwitch        bool //GetArchiveType
	GetArchiveVideoRelationGrey int  //GetArchiveVideoRelation
}

// Host is
type Host struct {
	APICo string
}

// DB is db config.
type DB struct {
	Result      *sql.Config
	Stat        *sql.Config
	ArchiveTiDB *tidb.Config
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
