package conf

import (
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/net/trace"
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
	// bm http
	HTTPServer *bm.ServerConfig
	HTTPClient *bm.ClientConfig
	// db
	DB                     *DB
	Redis                  *Redis
	ArcClient              *warden.ClientConfig
	SsnClient              *warden.ClientConfig
	EpClient               *warden.ClientConfig
	ArtClient              *warden.ClientConfig
	ThumbupClient          *warden.ClientConfig
	NoteCfg                *NoteCfg
	ArcsForbidQuotaID      string
	TaishanRpc             *warden.ClientConfig
	TaishanNoteReply       *common.TaishanTableConfig
	GetAttachedRpidQuotaID string
	ArcTagQuotaId          string
}

type NoteCfg struct {
	WebUrlFromSpace    string // 空间页h5主人态笔记跳转链接
	WebPubUrlFromArc   string // 播放页h5笔记专栏跳转链接
	WebPubUrlFromSpace string // 空间页h5笔记专栏跳转链接
	UpPubUrl           string // up主笔记tag跳转链接
	BfsHost            string
	ForbidCfg          *ForbidCfg
	ArcTagCfg          *ArcTagCfg
	ReplyCfg           *ReplyCfg
}

type ForbidCfg struct {
	ForbidTypeIds   []int64 // 禁止记笔记分区
	FeaHost         string  // 内容特征库
	PoliticsGroupId string
	PoliticsType    string
	FeaCron         string
}

type ArcTagCfg struct {
	AllowTypeIds       []int64 //允许的稿件分区，二级分区
	TagShowText        string
	EditNoteTagLink    string
	AutoPullNoteSwitch bool
}

type ReplyCfg struct {
	ImageAllowTypeIds        []int64 //允许在评论区展示截屏图片的稿件分区,二级分区
	WebUrl                   string  //笔记在评论区的点击跳转url
	ImageShowInReplyMaxLimit int32   //笔记在评论区最多展示x张图片
	ReplySummaryTextDefault  string  //当笔记全图片，无summary的时候笔记展示的默认文案
}

// DB define MySQL config
type DB struct {
	NoteRead  *xsql.Config
	NoteWrite *xsql.Config
}

// Redis redis
type Redis struct {
	*redis.Config
	Expire        xtime.Duration
	NoteExpire    xtime.Duration
	AidNoteExpire xtime.Duration
	ImgExpire     xtime.Duration
	ArticleExpire xtime.Duration
}
