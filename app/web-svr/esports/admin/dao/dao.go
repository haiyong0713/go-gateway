package dao

import (
	"context"
	"go-common/library/database/elastic"
	"go-common/library/database/orm"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/admin/conf"

	"github.com/jinzhu/gorm"
)

const (
	_esports   = "esports"
	_replyReg  = "/x/internal/v2/reply/subject/regist"
	_jobURL    = "/x/internal/esports/job/big/info"
	_jobBigURL = "/x/internal/esports/job/big/init"
)

// Dao .
type Dao struct {
	c       *conf.Config
	DB      *gorm.DB
	Elastic *elastic.Elastic
	// client
	replyClient *bm.Client
	jobClient   *bm.Client
	replyURL    string
	jobURL      string
	jobBigURL   string
	genPostURL  string
	savePostURL string
}

// New .
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// conf
		c: c,
		// db
		DB: orm.NewMySQL(c.ORM),
		// elastic
		Elastic:     elastic.NewElastic(nil),
		replyClient: bm.NewClient(c.HTTPReply),
		jobClient:   bm.NewClient(c.HTTPJob),
		replyURL:    c.Host.APICo + _replyReg,
		jobURL:      c.Host.APICo + _jobURL,
		jobBigURL:   c.Host.APICo + _jobBigURL,
		genPostURL:  c.Host.GenPost,
		savePostURL: c.Host.SavePost,
	}
	return
}

// Ping .
func (d *Dao) Ping(c context.Context) error {
	return d.DB.DB().PingContext(c)
}
