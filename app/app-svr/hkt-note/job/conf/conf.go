package conf

import (
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	bm "go-common/library/net/http/blademaster"
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
	XLog *log.Config
	// bm http
	HTTPServer *bm.ServerConfig
	Tracer     *trace.Config
	// db
	DB               *DB
	Redis            *Redis
	NoteBinlogSub    *databus.Config
	NoteNotifySub    *databus.Config
	NoteAuditSub     *databus.Config
	ArticleBinlogSub *databus.Config
	NoteCfg          *NoteCfg
	ArticleCfg       *ArticleCfg
	HTTPClient       *bm.ClientConfig
	InfocV2          *infocV2.Config
	ArticleClient    *warden.ClientConfig
	ArcClient        *warden.ClientConfig
	SeasonClient     *warden.ClientConfig
	FrontendClient   *warden.ClientConfig
	ReplyClient      *warden.ClientConfig
	ReplyDelSub      *databus.Config

	TaishanRpc       *warden.ClientConfig
	TaishanNoteReply *common.TaishanTableConfig
}

// DB define MySQL config
type DB struct {
	Note *xsql.Config
}

// Redis redis
type Redis struct {
	*redis.Config
	Expire       xtime.Duration
	NoteExpire   xtime.Duration
	ArtExpire    xtime.Duration
	ArtTmpExpire xtime.Duration

	BotPushExpire xtime.Duration
}

type ArticleCfg struct {
	CategoryNote int64
}

type NoteCfg struct {
	RetryFre    xtime.Duration
	DatabusFre  xtime.Duration
	Host        *Host
	FilterLimit int64
	ReplyCfg    *ReplyCfg

	//bot推送crm人群包
	BotCrmGroups []int64

	BotCrmLastestPubtime int64

	BotArcTypes []int32
	BotKey      string

	HotArcBotPushCron string
}

type ReplyCfg struct {
	WebUrl                 string
	RichTextWebUrl         string //在评论区以新样式展示时，使用该url，评论侧根据该url中标识区分新旧样式
	Template               string
	ReplyUrl               string // 通过评论区写专栏时，评论正文里的笔记链接
	ReplyOperationTitle    string //评论运营位主标题
	ReplyOperationSubTitle string //评论运营位副标题
	ReplyOperationIcon     string //评论运营位icon
	ReplyOperationEndTime  int64
	ReplyOperationWebUrl   string
}

type Host struct {
	FilterHost string
	ReplyHost  string

	HotArchiveHost string
}
