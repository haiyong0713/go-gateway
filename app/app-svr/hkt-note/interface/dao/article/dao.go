package article

import (
	"go-common/library/database/taishan"
	"go-gateway/app/app-svr/hkt-note/common"
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/hkt-note/interface/conf"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	accountRelationGrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	upgrpc "git.bilibili.co/bapis/bapis-go/archive/service/up"
	artgrpc "git.bilibili.co/bapis/bapis-go/article/service"
)

type Dao struct {
	c                     *conf.Config
	db                    *xsql.DB
	redis                 *redis.Pool
	artExpire             int64
	cache                 *fanout.Fanout
	noteClient            notegrpc.HktNoteClient
	accClient             accgrpc.AccountClient
	artClient             artgrpc.ArticleGRPCClient
	upClient              upgrpc.UpClient
	accountRelationClient accountRelationGrpc.RelationClient
	TaishanCli            taishan.TaishanProxyClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:         c,
		db:        xsql.NewMySQL(c.DB.Note),
		redis:     redis.NewPool(c.Redis.Config),
		cache:     fanout.New("cache", fanout.Worker(10), fanout.Buffer(10240)),
		artExpire: int64(time.Duration(c.Redis.ArtExpire) / time.Second),
	}
	var err error
	common.TaishanConfig.TaishanRpc = c.TaishanRpc
	common.TaishanConfig.NoteReply = c.TaishanNoteReply

	if d.noteClient, err = notegrpc.NewClient(c.NoteClient); err != nil {
		panic(err)
	}
	if d.accClient, err = accgrpc.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if d.artClient, err = artgrpc.NewClient(c.ArtClient); err != nil {
		panic(err)
	}
	if d.upClient, err = upgrpc.NewClient(c.UpClient); err != nil {
		panic(err)
	}
	if d.TaishanCli, err = taishan.NewClient(common.TaishanConfig.TaishanRpc); err != nil {
		panic(err)
	}
	if d.accountRelationClient, err = accountRelationGrpc.NewClient(c.AccountRelationClientCfg); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Close() {
	d.db.Close()
	d.redis.Close()
}
