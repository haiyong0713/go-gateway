package article

import (
	"go-common/library/database/taishan"
	"go-gateway/app/app-svr/hkt-note/common"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/hkt-note/job/conf"

	artgrpc "git.bilibili.co/bapis/bapis-go/article/service"
	frontgrpc "git.bilibili.co/bapis/bapis-go/frontend/bilinote/v1"

	replygrpc "git.bilibili.co/bapis/bapis-go/community/interface/reply"
)

// Dao is archive dao.
type Dao struct {
	c              *conf.Config
	db             *xsql.DB
	redis          *redis.Pool
	ArtExpire      int
	ArtTmpExpire   int // 发布失败时，短期缓存
	cache          *fanout.Fanout
	client         *bm.Client
	articleClient  artgrpc.ArticleGRPCClient
	frontendClient frontgrpc.BiliNoteServerClient
	//arcClient      archivegrpc.ArchiveClient
	replyClient replygrpc.ReplyInterfaceClient
	TaishanCli  taishan.TaishanProxyClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:            c,
		client:       bm.NewClient(c.HTTPClient),
		db:           xsql.NewMySQL(c.DB.Note),
		redis:        redis.NewPool(c.Redis.Config),
		ArtExpire:    int(time.Duration(c.Redis.ArtExpire) / time.Second),
		ArtTmpExpire: int(time.Duration(c.Redis.ArtTmpExpire) / time.Second),
		cache:        fanout.New("cache", fanout.Worker(10), fanout.Buffer(10240)),
	}
	var err error

	common.TaishanConfig.TaishanRpc = c.TaishanRpc
	common.TaishanConfig.NoteReply = c.TaishanNoteReply
	if d.articleClient, err = artgrpc.NewClient(c.ArticleClient); err != nil {
		panic(err)
	}
	if d.frontendClient, err = frontgrpc.NewClient(c.FrontendClient); err != nil {
		panic(err)
	}
	//if d.arcClient, err = archivegrpc.NewClient(c.ArcClient); err != nil {
	//	panic(err)
	//}
	if d.replyClient, err = replygrpc.NewClient(c.ReplyClient); err != nil {
		panic(err)
	}
	if d.TaishanCli, err = taishan.NewClient(common.TaishanConfig.TaishanRpc); err != nil {
		panic(err)
	}
	return
}
