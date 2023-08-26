package plugin

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/component"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/plugin"

	"github.com/pkg/errors"
)

const (
	_getSQL = "SELECT `name`,`package`,`policy`,`ver_code`,`ver_name`,`size`,`md5`,`url`,`enable`,`force`,`clear`,`min_build`,`max_build`,`base_code`,`base_name`,`desc`,`coverage` FROM plugin WHERE `enable`=1 AND `state`=0"
	_trace  = "https://trace.bilibili.co/api/operation_dependencies"
)

type Dao struct {
	db        *sql.DB
	pluginGet *sql.Stmt
	client    *bm.Client
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db:     component.GlobalDB,
		client: bm.NewClient(c.HTTPTrace),
	}
	// prepare
	d.pluginGet = d.db.Prepared(_getSQL)
	return
}

func (d *Dao) All(c context.Context) (psm map[string][]*plugin.Plugin, err error) {
	rows, err := d.pluginGet.Query(c)
	if err != nil {
		log.Error("query error(%v)", err)
		return nil, err
	}
	defer rows.Close()
	psm = map[string][]*plugin.Plugin{}
	for rows.Next() {
		p := &plugin.Plugin{}
		if err = rows.Scan(&p.Name, &p.Package, &p.Policy, &p.VerCode, &p.VerName, &p.Size, &p.MD5, &p.URL, &p.Enable, &p.Force, &p.Clear, &p.MinBuild, &p.MaxBuild, &p.BaseCode, &p.BaseName, &p.Desc, &p.Coverage); err != nil {
			log.Error("row.Scan error(%v)", err)
			return nil, err
		}
		if p.MaxBuild != 0 && p.MaxBuild < p.MinBuild {
			continue
		}
		psm[p.Name] = append(psm[p.Name], p)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return psm, err
}

func (d *Dao) Traces(ctx context.Context, param *plugin.TraceParam) ([]*plugin.TraceEdge, error) {
	params := url.Values{}
	params.Set("endTs", strconv.FormatInt(time.Now().Unix()*1000, 10))
	params.Set("service", param.Service)
	params.Set("operation", param.Operation)
	traceURI, err := url.Parse(_trace)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	traceURI.RawQuery = params.Encode()
	req, err := http.NewRequest(http.MethodGet, traceURI.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Cookie", param.Cookie)
	var res struct {
		Data struct {
			Edges []*plugin.TraceEdge `json:"edges"`
		} `json:"data"`
	}
	if err := d.client.Do(ctx, req, &res); err != nil {
		return nil, err
	}
	return res.Data.Edges, nil
}

// Close close memcache resource.
func (dao *Dao) Close() {
	if dao.db != nil {
		dao.db.Close()
	}
}

func (dao *Dao) PingDB(c context.Context) (err error) {
	return dao.db.Ping(c)
}
