package conf

import (
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	ecode "go-common/library/ecode/tip"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	// Env
	Env string
	// db
	MySQL *MySQL
	// show  XLog
	Log *log.Config
	// tracer
	Tracer *trace.Config
	// cache tick time
	CacheTick xtime.Duration
	// httpClinet
	HTTPClient *bm.ClientConfig
	// HTTPTelecom
	HTTPTelecom *bm.ClientConfig
	// HTTPBroadband
	HTTPBroadband *bm.ClientConfig
	// HTTPUnicom
	HTTPUnicom *bm.ClientConfig
	// HTTPUnicom
	HTTPActive *bm.ClientConfig
	// HTTPUnicom
	HTTPActivate *bm.ClientConfig
	// bm http
	BM *HTTPServers
	// rpc location
	LocationRPC *rpc.ClientConfig
	// seq
	SeqGRPC *warden.ClientConfig
	// live
	LiveGRPC *warden.ClientConfig
	// host
	Host *Host
	// ecode
	Ecode *ecode.Config
	// Report
	Report *databus.Config
	// iplimit
	IPLimit *IPLimit
	// Seq
	Seq *Seq
	// Telecom
	Telecom *Telecom
	// Redis
	Redis *Redis
	// mc
	Memcache *Memcache
	// reddot
	Reddot *Reddot
	// unicom
	Unicom *Unicom
	ES     *elastic.Config
	// databus
	UnicomDatabus *databus.Config
	PackPub       *databus.Config
	// grpc location
	LocationGRPC *warden.ClientConfig
	AccountGRPC  *warden.ClientConfig
	VIPGRPC      *warden.ClientConfig
	// logID
	IPLogID string
	// mobile
	Mobile *Mobile
	// Rule
	Rule       map[string][]*Rule
	AccountVIP map[string]*AccountVIP
	// CouponV2
	CouponV2 *CouponV2
}

type Rule struct {
	M  string
	Tf bool
	P  string
	A  string
	// 表示备用操作参数，用于客户端多域名重试
	ABackup []string
}

type Host struct {
	APICo               string
	Dotin               string
	Live                string
	APILive             string
	Telecom             string
	TelecomCard         string
	Unicom              string
	UnicomFlow          string
	Broadband           string
	Sms                 string
	Mall                string
	TelecomReturnURL    string
	TelecomCancelPayURL string
	TelecomActive       string
	Comic               string
	Gdt                 string
	UnicomActivate      string
	UnicomVerify        string
	UnicomUsermob       string
	MallDiscovery       string
	UnicomFlowTryout    string
}

type HTTPServers struct {
	Outer *bm.ServerConfig
}

type Seq struct {
	BusinessID int64
	Token      string
}

// App bilibili intranet authorization.
type App struct {
	Key    string
	Secret string
}

type MySQL struct {
	Show    *sql.Config
	ShowLog *sql.Config
}

type IPLimit struct {
	Addrs map[string][]string
}

type Reddot struct {
	StartTime string
	EndTime   string
}

type Unicom struct {
	KeyExpired                xtime.Duration
	FlowWait                  xtime.Duration
	UnicomAppKey              string
	UnicomSecurity            string
	UnicomAppMethodFlow       string
	UnicomMethodNumber        string
	UnicomMethodFlowPre       string
	UnicomMethodQryFlowChange string
	Cpid                      string
	Password                  string
	Activate                  *Activate
	FlowProduct               []*Product
	CardProduct               []*Product
	ExchangeLimit             *exchangeLimit
	UnicomUsermob             *UnicomUsermob
	Verify                    *Verify
	FlowTryout                *FlowTryout
}

type exchangeLimit struct {
	PhoneWhitelist []string
}

type Activate struct {
	User     string
	Password string
}

type AccountVIP struct {
	BatchID int64
	AppKey  string
	Days    int64
}

type Telecom struct {
	KeyExpired         xtime.Duration
	PayKeyExpired      xtime.Duration
	SMSTemplate        string
	SMSMsgTemplate     string
	SMSFlowTemplate    string
	SMSOrderTemplateOK string
	FlowPercentage     int
	Area               map[string][]string
	CardSpid           string
	CardPass           string
	// flow
	AppSecret   string
	AppID       string
	PackageID   int
	CardProduct []*Product
}

type Redis struct {
	Recommend *struct {
		*redis.Config
		Expire xtime.Duration
	}
	Wall *struct {
		*redis.Config
	}
}

type Memcache struct {
	Operator *struct {
		*memcache.Config
		Expire        xtime.Duration
		EmptyExpire   xtime.Duration
		UsermobExpire xtime.Duration
	}
}

type Mobile struct {
	FlowProduct []*Product
	CardProduct []*Product
}

type Product struct {
	ID   string
	Spid string
	Desc string
	Type int
	Way  string
	Tag  string
}

// 联通取伪码所需账号信息
type UnicomUsermob struct {
	User  string
	Pass  string
	AppID string
}

type Verify struct {
	User     string
	Password string
}

type CouponV2 struct {
	SourceID string
}

type FlowTryout struct {
	Channel  string
	Password string
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
