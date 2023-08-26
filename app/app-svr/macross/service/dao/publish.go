package dao

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/macross/service/model/publish"
)

const (
	_logSharding = 10
	// dashborad
	_inDashboradSQL     = `INSERT INTO dashboard (name,label,commit_info,out_url,coverage_url,text_size_arm64,res_size,extra) VALUES(?,?,?,?,?,?,?,?)`
	_inDashboradLogsSQL = `INSERT INTO dashboard_log_%02d (dashboard_id,level,msg) VALUES %s`
)

func (d *Dao) hitLogs(id int64) int64 {
	return id % _logSharding
}

// Dashborad insert dashboard.
func (d *Dao) Dashborad(c context.Context, dashboard *publish.Dashboard) (rows int64, err error) {
	res, err := d.db.Exec(c, _inDashboradSQL, dashboard.Name, dashboard.Label, dashboard.Commit, dashboard.OutURL, dashboard.CoverageURL, dashboard.TextSizeArm64, dashboard.ResSize, dashboard.Extra)
	if err != nil {
		log.Error("Dashborad() d.db.Exec() error(%v)", err)
		return
	}
	rows, err = res.LastInsertId()
	return
}

// DashboradLogs insert dashboard log.
func (d *Dao) DashboradLogs(c context.Context, id int64, logs []*publish.Log) (rows int64, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, v := range logs {
		sqls = append(sqls, "(?,?,?)")
		args = append(args, id, v.Level, v.Msg)
	}
	res, err := d.db.Exec(c, fmt.Sprintf(_inDashboradLogsSQL, d.hitLogs(id), strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("DashboradLogs d.db.Exec() error(%v)", err)
		return
	}
	rows, err = res.RowsAffected()
	return
}
