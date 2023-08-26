package image

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"go-common/library/database/bfs"
	xsql "go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/hkt-note/interface/conf"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"
)

type Dao struct {
	c          *conf.Config
	db         *xsql.DB
	bfsClient  *bfs.BFS
	httpClient *http.Client
	noteClient notegrpc.HktNoteClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		db:         xsql.NewMySQL(c.DB.Note),
		bfsClient:  bfs.New(c.BfsClient),
		httpClient: NewClient(c.HTTPClients.Inner),
	}
	var err error
	if d.noteClient, err = notegrpc.NewClient(c.NoteClient); err != nil {
		panic(err)
	}
	return
}

// NewClient new a http client.
// nolint:gosec
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
