package fawkes

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/model/bugly"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_getCrashIndexMessageList = `SELECT vci.*,IFNULL(vcs.analyse_error_stack,'') AS analyse_error_stack FROM (SELECT error_stack_hash_without_useless,error_stack_before_hash,app_key,error_type,error_msg,solve_status,solve_operator,assign_operator,solve_version_code,solve_description,happen_newest_version_code,happen_oldest_version_code,unix_timestamp(happen_time) AS happen_time,unix_timestamp(ctime) AS ctime,unix_timestamp(mtime) AS mtime FROM %s WHERE (%s) AND app_key=?) AS vci LEFT JOIN %s AS vcs ON vci.error_stack_hash_without_useless=vcs.error_stack_hash_without_useless AND vci.app_key=vcs.app_key`
	_getJankIndexMessageList  = `SELECT analyse_jank_stack,analyse_jank_stack_hash,app_key,solve_status,solve_operator,solve_version_code,solve_description,happen_newest_version_code,happen_oldest_version_code,unix_timestamp(happen_time) AS happen_time,unix_timestamp(ctime) AS ctime,unix_timestamp(mtime) AS mtime FROM veda_crash_db.veda_jank_index WHERE (%s) AND app_key=?`
	_getOOMIndexMessageList   = `SELECT hash,analyse_stack,leak_reason,gc_root,app_key,solve_status,solve_version_code,solve_operator,solve_description,happen_newest_version_code,happen_oldest_version_code,unix_timestamp(happen_time) AS happen_time,unix_timestamp(ctime) AS ctime,unix_timestamp(mtime) AS mtime FROM veda_crash_db.veda_oom_index WHERE (%s) AND app_key=?`
	_getCrashLogList          = `SELECT hash,app_key,operator,log_text,unix_timestamp(ctime) AS ctime,unix_timestamp(mtime) AS mtime FROM veda_log WHERE hash=? AND app_key=?`

	_getIndexStatus              = `SELECT %s AS hash,app_key,solve_status,solve_version_code,assign_operator,solve_operator,solve_description FROM %s where %s=? AND app_key=?`
	_getHashListLimitVersionCode = `SELECT %s AS hash FROM %s WHERE app_key=? AND happen_newest_version_code<? LIMIT ?`

	_updateCrashIndex = `UPDATE %s SET solve_operator=?,solve_version_code=?,solve_status=?,solve_description=? WHERE error_stack_hash_without_useless=? AND app_key=?`
	_updateJankIndex  = `UPDATE veda_jank_index SET solve_operator=?,solve_version_code=?,solve_status=?,solve_description=? WHERE analyse_jank_stack_hash=? AND app_key=?`
	_updateOOMIndex   = `UPDATE veda_oom_index SET solve_operator=?,solve_version_code=?,solve_status=?,solve_description=? WHERE hash=? AND app_key=?`
	_updateIndex      = `UPDATE %s SET assign_operator=?,solve_operator=?,solve_version_code=?,solve_status=?,solve_description=? WHERE %s=? AND app_key=?`
	_addCrashLaserRel = `INSERT INTO veda_crash_laser_relation (error_stack_hash_without_useless,laser_id,operator) VALUES (?,?,?)`
	_addCrashLog      = `INSERT INTO veda_log (hash,app_key,operator,log_text) VALUES (?,?,?,?)`
)

func (d *Dao) HashListLimitVersionCode(c context.Context, vedaIndexTable, hashColumn, appKey, versionCode string, count int) (res []string, err error) {
	var args []interface{}
	args = append(args, appKey, versionCode, count)
	rows, err := d.veda.Query(c, fmt.Sprintf(_getHashListLimitVersionCode, hashColumn, vedaIndexTable), args...)
	if err != nil {
		log.Error("d.HashListLimitVersionCode error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var re string
		if err = rows.Scan(&re); err != nil {
			log.Error("d.CrashIndexMessageByHashList rows.Scan error(%v)", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) CrashIndexMessageByHashList(c context.Context, appKey, vedaIndexTable, vedaStackTable string, stackHashList []string) (res []*bugly.CrashIndex, err error) {
	var (
		sqlAdd []string
		args   []interface{}
	)
	for _, hash := range stackHashList {
		sqlAdd = append(sqlAdd, "error_stack_hash_without_useless=?")
		args = append(args, hash)
	}
	strings.Join(sqlAdd, " OR ")
	args = append(args, appKey)
	rows, err := d.veda.Query(c, fmt.Sprintf(_getCrashIndexMessageList, vedaIndexTable, strings.Join(sqlAdd, " OR "), vedaStackTable), args...)
	if err != nil {
		log.Error("d.CrashIndexMessageByHashList error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &bugly.CrashIndex{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("d.CrashIndexMessageByHashList rows.Scan error(%v)", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) JankIndexMessageByHashList(c context.Context, appKey string, stackHashList []string) (res []*bugly.JankIndex, err error) {
	var (
		sqlAdd []string
		args   []interface{}
	)
	for _, hash := range stackHashList {
		sqlAdd = append(sqlAdd, "analyse_jank_stack_hash=?")
		args = append(args, hash)
	}
	strings.Join(sqlAdd, " OR ")
	args = append(args, appKey)
	rows, err := d.veda.Query(c, fmt.Sprintf(_getJankIndexMessageList, strings.Join(sqlAdd, " OR ")), args...)
	if err != nil {
		log.Error("d.JankIndexMessageByHashList error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &bugly.JankIndex{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("d.JankIndexMessageByHashList rows.Scan error(%v)", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) OOMIndexMessageByHashList(c context.Context, appKey string, stackHashList []string) (res []*bugly.OOMIndex, err error) {
	var (
		sqlAdd []string
		args   []interface{}
	)
	for _, hash := range stackHashList {
		sqlAdd = append(sqlAdd, "hash=?")
		args = append(args, hash)
	}
	strings.Join(sqlAdd, " OR ")
	args = append(args, appKey)
	rows, err := d.veda.Query(c, fmt.Sprintf(_getOOMIndexMessageList, strings.Join(sqlAdd, " OR ")), args...)
	if err != nil {
		log.Error("d.OOMIndexMessageByHashList error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &bugly.OOMIndex{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("d.OOMIndexMessageByHashList rows.Scan error(%v)", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxCrashIndexUpdate crash index
func (d *Dao) TxIndexUpdate(tx *sql.Tx, errorStackHash, appKey, assignOperator, solveOperator, solveDescription, vedaIndexTable, hashColumn string, solveVersionCode, solveStatus int64) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_updateIndex, vedaIndexTable, hashColumn), assignOperator, solveOperator, solveVersionCode, solveStatus, solveDescription, errorStackHash, appKey)
	if err != nil {
		log.Error("TxCrashIndexUpdate %v", err)
		return
	}
	return res.RowsAffected()
}

// TxCrashIndexUpdate crash index
func (d *Dao) TxCrashIndexUpdate(tx *sql.Tx, errorStackHash, appKey, solveOperator, solveDescription, vedaIndexTable string, solveVersionCode, solveStatus int) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_updateCrashIndex, vedaIndexTable), solveOperator, solveVersionCode, solveStatus, solveDescription, errorStackHash, appKey)
	if err != nil {
		log.Error("TxCrashIndexUpdate %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxJankIndexUpdate(tx *sql.Tx, errorStackHash, appKey, solveOperator, solveDescription string, solveVersionCode, solveStatus int) (r int64, err error) {
	res, err := tx.Exec(_updateJankIndex, solveOperator, solveVersionCode, solveStatus, solveDescription, errorStackHash, appKey)
	if err != nil {
		log.Error("TxJankIndexUpdate %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxOOMIndexUpdate(tx *sql.Tx, hash, appKey, solveOperator, solveDescription string, solveVersionCode, solveStatus int) (r int64, err error) {
	res, err := tx.Exec(_updateOOMIndex, solveOperator, solveVersionCode, solveStatus, solveDescription, hash, appKey)
	if err != nil {
		log.Error("TxOOMIndexUpdate %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) CrashLaserRelationAdd(c context.Context, laserId int64, errorStackHashWithoutUseless, operator string) (err error) {
	_, err = d.db.Exec(c, _addCrashLaserRel, errorStackHashWithoutUseless, laserId, operator)
	if err != nil {
		log.Error("TxCrashLaserRelationAdd tx.Exec error(%v)", err)
	}
	return
}

func (d *Dao) IndexStatusByHash(c context.Context, hashColumn, vedaIndexTable, hash, appKey string) (res *bugly.IndexStatus, err error) {
	var (
		args []interface{}
	)
	args = append(args, hash)
	args = append(args, appKey)
	row := d.veda.QueryRow(c, fmt.Sprintf(_getIndexStatus, hashColumn, vedaIndexTable, hashColumn), args...)
	res = &bugly.IndexStatus{}
	if err = row.Scan(&res.Hash, &res.AppKey, &res.SolveStatus, &res.SolveVersionCode, &res.AssignOperator, &res.SolveOperator, &res.SolveDescription); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("ApmBusByID row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) CrashLogAdd(c context.Context, hash, appKey, operator, logText string) (err error) {
	_, err = d.veda.Exec(c, _addCrashLog, hash, appKey, operator, logText)
	if err != nil {
		log.Error("CrashLogAdd tx.Exec error(%v)", err)
	}
	return
}

func (d *Dao) LogList(c context.Context, hash, appKey string) (res []*bugly.LogText, err error) {
	var (
		args []interface{}
	)
	args = append(args, hash, appKey)
	rows, err := d.veda.Query(c, _getCrashLogList, args...)
	if err != nil {
		log.Error("LogList error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &bugly.LogText{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("LogList rows.Scan error(%v)", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}
