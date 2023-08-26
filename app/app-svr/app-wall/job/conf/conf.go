package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
)

// Config config set
type Config struct {
	// base
	// elk
	Log *log.Config
	// http
	BM *HTTPServers
	// tracer
	Tracer *trace.Config
	// redis
	Redis *Redis
	// memcache
	Memcache *Memcache
	// db
	MySQL *MySQL
	// databus
	PackPub       *databus.Config
	ReportDatabus *railgun.DatabusV1Config
	ReportRailgun *railgun.ReplaceConfig
	ComicDatabus  *railgun.DatabusV1Config
	ComicRailgun  *railgun.SingleConfig
	PackSub       *railgun.DatabusV1Config
	PackRailgun   *railgun.SingleConfig
	CanalSub      *railgun.DatabusV1Config
	CanalRailgun  *railgun.SingleConfig
	// ecode
	Ecode *ecode.Config
	// Report
	Report *databus.Config
	// client
	Consumer *Consumer
	// HTTPClient
	HTTPClient *bm.ClientConfig
	// HTTPUnicom
	HTTPUnicom *bm.ClientConfig
	// HTTPSMS
	HTTPSMS *bm.ClientConfig
	// host
	Host *Host
	// unicom
	Unicom *Unicom
	// monthly
	Monthly bool
	// seq
	SeqGRPC *warden.ClientConfig
	VIPGRPC *warden.ClientConfig
	// Seq
	Seq        *Seq
	AccountVIP map[string]*AccountVIP
	// CouponV2
	CouponV2 *CouponV2
}

type Seq struct {
	BusinessID int64
	Token      string
}

// HTTPServers Http Servers
type HTTPServers struct {
	Outer *bm.ServerConfig
	Inner *bm.ServerConfig
	Local *bm.ServerConfig
}

type MySQL struct {
	Show *sql.Config
}

type Unicom struct {
	PackKeyExpired            xtime.Duration
	KeyExpired                xtime.Duration
	UnicomUser                string
	UnicomPass                string
	UnicomAppKey              string
	UnicomSecurity            string
	UnicomMethodQryFlowChange string
	FlowProduct               []*Product
	CardProduct               []*Product
}

type Product struct {
	ID       string
	Integral int
}

type Memcache struct {
	Operator *struct {
		*memcache.Config
		Expire xtime.Duration
	}
}

type Redis struct {
	Wall *struct {
		*redis.Config
		LockExpire      xtime.Duration
		MonthLockExpire xtime.Duration
	}
}

type Consumer struct {
	Group   string
	Topic   string
	Offset  string
	Brokers []string
}

type Host struct {
	APP           string
	UnicomFlow    string
	Unicom        string
	APICo         string
	APILive       string
	Mall          string
	Comic         string
	MallDiscovery string
}

type AccountVIP struct {
	BatchID int64
	AppKey  string
}

type CouponV2 struct {
	SourceID string
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
