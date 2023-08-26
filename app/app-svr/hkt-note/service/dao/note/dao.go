package note

import (
	"time"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/hkt-note/service/conf"

	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	cepgrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/episode"
	cssngrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
)

type Dao struct {
	c             *conf.Config
	dbr           *xsql.DB
	redis         *redis.Pool
	noteExpire    int
	aidNoteExpire int
	arcClient     arcapi.ArchiveClient
	chSsnClient   cssngrpc.SeasonClient
	chEpClient    cepgrpc.EpisodeClient
	cache         *fanout.Fanout
	client        *bm.Client
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:             c,
		dbr:           xsql.NewMySQL(c.DB.NoteRead),
		redis:         redis.NewPool(c.Redis.Config),
		noteExpire:    int(time.Duration(c.Redis.NoteExpire) / time.Second),
		aidNoteExpire: int(time.Duration(c.Redis.AidNoteExpire) / time.Second),
		cache:         fanout.New("cache", fanout.Worker(10), fanout.Buffer(10240)),
		client:        bm.NewClient(c.HTTPClient),
	}
	var err error
	if d.arcClient, err = arcapi.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if d.chSsnClient, err = cssngrpc.NewClient(c.SsnClient); err != nil {
		panic(err)
	}
	if d.chEpClient, err = cepgrpc.NewClient(c.EpClient); err != nil {
		panic(err)
	}
	return
}

// Close close resource.
func (d *Dao) Close() {
	d.dbr.Close()
	d.redis.Close()
}
