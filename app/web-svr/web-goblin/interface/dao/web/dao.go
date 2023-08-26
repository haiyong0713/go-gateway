package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/stat/prom"
	"go-gateway/app/web-svr/web-goblin/interface/conf"

	"github.com/pkg/errors"
)

const (
	_pgcFullURL  = "/ext/internal/archive/channel/content"
	_pgcIncreURL = "/ext/internal/archive/channel/content/change"
	_rankURL     = "/data/rank/%s.json"
)

// Dao dao .
type Dao struct {
	c      *conf.Config
	db     *sql.DB
	showDB *sql.DB
	// redis
	redis                   *redis.Pool
	httpR                   *bm.Client
	httpJob                 *bm.Client
	pgcFullURL, pgcIncreURL string
	ela                     *elastic.Elastic
	rankURL                 string
	cusExpire               int32
}

// New init mysql db .
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		c:           c,
		db:          sql.NewMySQL(c.DB.Goblin),
		redis:       redis.NewPool(c.Redis.Config),
		showDB:      sql.NewMySQL(c.DB.Show),
		httpR:       bm.NewClient(c.SearchClient),
		httpJob:     bm.NewClient(c.JobClient),
		pgcFullURL:  c.Host.PgcURI + _pgcFullURL,
		pgcIncreURL: c.Host.PgcURI + _pgcIncreURL,
		ela:         elastic.NewElastic(c.Es),
		rankURL:     c.Host.Rank + _rankURL,
		cusExpire:   int32(time.Duration(c.Redis.CustomerExpire) / time.Second),
	}
	return
}

// Close close the resource .
func (d *Dao) Close() {
}

// Ping dao ping .
func (d *Dao) Ping(c context.Context) error {
	return nil
}

// PromError stat and log .
func PromError(name string, format string, args ...interface{}) {
	prom.BusinessErrCount.Incr(name)
	log.Error(format, args...)
}

func (d *Dao) ReadURLContent(c context.Context, outURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", outURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadURLContent http.NewRequest(%s)", outURL)
	}
	ctx, cancel := context.WithTimeout(c, time.Duration(d.c.Rule.ReadTimeout))
	defer cancel()
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadURLContent httpClient.Do(%s)", outURL)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.New(fmt.Sprintf("ReadURLContent url(%s) resp.StatusCode(%v)", outURL, resp.StatusCode))
	}
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ReadURLContent ioutil.ReadAll error:%v")
	}
	defer resp.Body.Close()
	return res, nil
}
