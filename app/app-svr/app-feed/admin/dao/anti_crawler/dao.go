package anti_crawler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/clickhouse"
	model "go-gateway/app/app-svr/app-feed/admin/model/anti_crawler"

	"github.com/pkg/errors"
)

type Dao struct {
	c          *conf.Config
	clickhouse *clickhouse.DB
	redis      *redis.Redis
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		clickhouse: clickhouse.NewClickhouse(c.ClickHouse.AntiCrawler.Config),
		redis:      redis.NewRedis(c.SpmodeRedis.Config),
	}
	return
}

func (d *Dao) UserLog(ctx context.Context, buvid string, mid int64, host, path string, stime, etime int64, pn, ps int) ([]*model.InfocMsg, error) {
	const _getList = "SELECT mid,buvid,host,path,`method`,header,query,`body`,referer,ip,ctime,response_header,response_body FROM %s.dwd_web_anti_crawler_http_report_l_rt WHERE %s ORDER BY ctime DESC LIMIT ?,?"
	var (
		sqls []string
		args []interface{}
	)
	if buvid != "" {
		sqls = append(sqls, "buvid=?")
		args = append(args, buvid)
	}
	if mid > 0 {
		sqls = append(sqls, "mid=?")
		args = append(args, mid)
	}
	if host != "" {
		sqls = append(sqls, "host=?")
		args = append(args, host)
	}
	if path != "" {
		sqls = append(sqls, "path=?")
		args = append(args, path)
	}
	if stime > 0 {
		sqls = append(sqls, "ctime>?")
		args = append(args, stime)
	}
	if etime > 0 {
		sqls = append(sqls, "ctime<?")
		args = append(args, etime)
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.clickhouse.Query(ctx, fmt.Sprintf(_getList, d.c.ClickHouse.AntiCrawler.DatabaseName, strings.Join(sqls, " AND ")), args...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	var res []*model.InfocMsg
	for rows.Next() {
		r := &model.InfocMsg{}
		if err := rows.Scan(&r.Mid, &r.Buvid, &r.Host, &r.Path, &r.Method, &r.Header, &r.Query, &r.Body, &r.Referer, &r.IP, &r.Ctime, &r.ResponseHeader, &r.ResponseBody); err != nil {
			return nil, errors.WithStack(err)
		}
		r.CtimeHuman = time.Unix(r.Ctime, 0)
		res = append(res, r)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}
