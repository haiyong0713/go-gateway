package conf

import (
	"go-common/library/cache/redis"
	"go-common/library/database/bfs"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/middleware/auth"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/hkt-note/common"
)

var (
	Conf = &Config{}
)

type Config struct {
	// show  XLog
	Log    *log.Config
	Tracer *trace.Config
	Auth   *auth.Config
	// bm http
	HTTPServer  *bm.ServerConfig
	HTTPClients *HTTPClients
	// db
	DB                       *DB
	Redis                    *Redis
	NoteCfg                  *NoteCfg
	ArcClient                *warden.ClientConfig
	SeasonClient             *warden.ClientConfig
	EpisodeClient            *warden.ClientConfig
	Bfs                      *Bfs
	BfsClient                *bfs.Config
	NoteClient               *warden.ClientConfig
	SeqClient                *warden.ClientConfig
	BroadCastClient          *warden.ClientConfig
	AccClient                *warden.ClientConfig
	ArtClient                *warden.ClientConfig
	UpClient                 *warden.ClientConfig
	AccountRelationClientCfg *warden.ClientConfig
	NotePub                  *databus.Config
	NoteAuditPub             *databus.Config
	Gray                     *Gray
	Hosts                    *Hosts
	InfocV2                  *infocV2.Config

	TaishanRpc       *warden.ClientConfig
	TaishanNoteReply *common.TaishanTableConfig
}

type Hosts struct {
	BfsHost string
}

type Gray struct {
	NoteWebGray int
	WhiteList   []int64
}

type Seq struct {
	BusinessId int64
	Token      string
}

type Bfs struct {
	MaxSize     int
	Bucket      string
	Key         string
	Secret      string
	Host        string
	PublicUrl   string // 不进行用户鉴权的url
	PublicToken string // 直传mid时需加token进行校验
}

type NoteCfg struct {
	Seq            *Seq
	MaxSize        int64
	MaxContSize    int64 // 单篇正文字数上限
	MaxSummarySize int   // summary字节数上限
	ImageHost      string
	BroadcastToken string
	CheeseQALink   string    // 课堂问卷链接
	Messages       *Messages // 展示文案
}

type Messages struct {
	UpSwitchMsg string // up主未开启笔记展示
	ListNoneMsg string // 笔记列表为空
}

// BM http
type HTTPClients struct {
	Inner *bm.ClientConfig
}

// DB define MySQL config
type DB struct {
	Note *xsql.Config
}

// Redis redis
type Redis struct {
	*redis.Config
	Expire     xtime.Duration
	NoteExpire xtime.Duration
	ArtExpire  xtime.Duration
}
