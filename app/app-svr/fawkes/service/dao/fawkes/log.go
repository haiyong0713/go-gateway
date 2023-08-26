package fawkes

import (
	"context"
	"fmt"
	"strings"

	xsql "go-common/library/database/sql"

	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_selLogCount = `SELECT count(*) FROM log WHERE %s`
	_insertLog   = `INSERT INTO log (app_key,env,model,operation,target,operator) VALUES (?,?,?,?,?,?)`
	_selLog      = `SELECT id,app_key,env,model,operation,target,operator,unix_timestamp(ctime),unix_timestamp(mtime) 
FROM log WHERE %s ORDER BY id DESC LIMIT ?,?`
)

// AddLog add log.
func (d *Dao) AddLog(c context.Context, appKey, env, model, operation, target, operator string) (r int64, err error) {
	rows, err := d.db.Exec(c, _insertLog, appKey, env, model, operation, target, operator)
	if err != nil {
		log.Error("AddLog() error(%v)", err)
		return
	}
	return rows.RowsAffected()
}

// Log get log list.
func (d *Dao) Log(c context.Context, appKey, env, model, operation, target, operator, stime, etime string,
	ps, pn int) (res []*mngmdl.Log, err error) {
	var (
		args []interface{}
		sqls []string
	)
	args = append(args, appKey)
	sqls = append(sqls, "app_key=?")
	if env != "" {
		args = append(args, env)
		sqls = append(sqls, "env=?")
	}
	if model != "" {
		args = append(args, model)
		sqls = append(sqls, "model=?")
	}
	if operation != "" {
		args = append(args, operation)
		sqls = append(sqls, "operation=?")
	}
	if operator != "" {
		args = append(args, operator)
		sqls = append(sqls, "operator=?")
	}
	if target != "" {
		args = append(args, "%"+target+"%")
		sqls = append(sqls, "target LIKE ?")
	}
	if stime != "" {
		args = append(args, stime)
		sqls = append(sqls, "ctime>?")
	}
	if etime != "" {
		args = append(args, etime)
		sqls = append(sqls, "ctime<?")
	}

	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_selLog, strings.Join(sqls, " AND ")), args...)
	if err != nil {
		log.Error("Log %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &mngmdl.Log{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Env, &re.Model, &re.Operation, &re.Target, &re.Operator, &re.CTime, &re.MTime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// LogCount get log count.
func (d *Dao) LogCount(c context.Context, appKey, env, model, operation, target, operator, stime, etime string) (count int, err error) {
	var (
		args []interface{}
		sqls []string
	)
	args = append(args, appKey)
	sqls = append(sqls, "app_key=?")
	if env != "" {
		args = append(args, env)
		sqls = append(sqls, "env=?")
	}
	if model != "" {
		args = append(args, model)
		sqls = append(sqls, "model=?")
	}
	if operation != "" {
		args = append(args, operation)
		sqls = append(sqls, "operation=?")
	}
	if operator != "" {
		args = append(args, operator)
		sqls = append(sqls, "operator=?")
	}
	if target != "" {
		args = append(args, "%"+target+"%")
		sqls = append(sqls, "target LIKE ?")
	}
	if stime != "" {
		args = append(args, stime)
		sqls = append(sqls, "ctime>?")
	}
	if etime != "" {
		args = append(args, etime)
		sqls = append(sqls, "ctime<?")
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_selLogCount, strings.Join(sqls, " AND ")), args...)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("LogCount %v", err)
		}
	}
	return
}
