package note

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/bfs"
	xsql "go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/pipeline/fanout"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/hkt-note/interface/conf"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	cssngrpc "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"
	seqgrpc "git.bilibili.co/bapis/bapis-go/infra/service/sequence"
	bcgrpc "git.bilibili.co/bapis/bapis-go/push/service/broadcast/v2"
)

type Dao struct {
	c            *conf.Config
	db           *xsql.DB
	redis        *redis.Pool
	noteExpire   int64
	cache        *fanout.Fanout
	seqClient    seqgrpc.SeqClient
	noteClient   notegrpc.HktNoteClient
	bfsClientSdk *bfs.BFS
	bfsClient    *http.Client
	arcClient    arcapi.ArchiveClient
	syncClient   bcgrpc.BroadcastAPIClient
	chSsnClient  cssngrpc.SeasonClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:            c,
		db:           xsql.NewMySQL(c.DB.Note),
		redis:        redis.NewPool(c.Redis.Config),
		bfsClientSdk: bfs.New(c.BfsClient),
		bfsClient:    NewClient(c.HTTPClients.Inner),
		noteExpire:   int64(time.Duration(c.Redis.NoteExpire) / time.Second),
		cache:        fanout.New("cache", fanout.Worker(10), fanout.Buffer(10240)),
	}
	var err error
	if d.noteClient, err = notegrpc.NewClient(c.NoteClient); err != nil {
		panic(err)
	}
	if d.seqClient, err = seqgrpc.NewClient(c.SeqClient); err != nil {
		panic(err)
	}
	if d.arcClient, err = arcapi.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if d.syncClient, err = bcgrpc.NewClient(c.BroadCastClient); err != nil {
		panic(err)
	}
	if d.chSsnClient, err = cssngrpc.NewClient(c.SeasonClient); err != nil {
		panic(err)
	}
	return
}

// NewClient new a http client.
//
//nolint:gosec
func NewClient(c *bm.ClientConfig) (client *http.Client) {
	var (
		transport *http.Transport
		dialer    *net.Dialer
	)
	dialer = &net.Dialer{
		Timeout:   time.Duration(c.Dial),
		KeepAlive: time.Duration(c.KeepAlive),
	}
	transport = &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{
		Transport: transport,
	}
	return
}

func (d *Dao) Close() {
	d.db.Close()
	d.redis.Close()
}
