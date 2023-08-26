package article

import (
	"go-common/library/database/taishan"
	"go-gateway/app/app-svr/hkt-note/common"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/hkt-note/service/conf"

	artgrpc "git.bilibili.co/bapis/bapis-go/article/service"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

type Dao struct {
	c             *conf.Config
	dbr           *xsql.DB
	redis         *redis.Pool
	artExpire     int
	cache         *fanout.Fanout
	artClient     artgrpc.ArticleGRPCClient
	thumbupClient thumbupgrpc.ThumbupClient
	TaishanCli    taishan.TaishanProxyClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:         c,
		dbr:       xsql.NewMySQL(c.DB.NoteRead),
		redis:     redis.NewPool(c.Redis.Config),
		artExpire: int(time.Duration(c.Redis.ArticleExpire) / time.Second),
		cache:     fanout.New("cache", fanout.Worker(10), fanout.Buffer(10240)),
	}
	var err error
	common.TaishanConfig.TaishanRpc = c.TaishanRpc
	common.TaishanConfig.NoteReply = c.TaishanNoteReply
	if d.artClient, err = artgrpc.NewClient(c.ArtClient); err != nil {
		panic(err)
	}
	if d.thumbupClient, err = thumbupgrpc.NewClient(c.ThumbupClient); err != nil {
		panic(err)
	}
	if d.TaishanCli, err = taishan.NewClient(common.TaishanConfig.TaishanRpc); err != nil {
		panic(err)
	}
	return
}

// Close close resource.
func (d *Dao) Close() {
	d.dbr.Close()
	d.redis.Close()
}
