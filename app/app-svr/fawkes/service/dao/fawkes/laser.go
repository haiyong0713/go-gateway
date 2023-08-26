package fawkes

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/database/sql"
	xsql "go-common/library/database/sql"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_laserCount = `SELECT count(*) FROM app_laser WHERE app_key=? %s`
	_laserList  = `SELECT id,app_key,platform,mid,buvid,email,log_date,url,status,silence_url,silence_status,parse_status,channel,description,mobi_app,recall_mobi_app,build,error_msg,operator,msg_id,md5,unix_timestamp(ctime),
unix_timestamp(mtime) FROM app_laser WHERE app_key=? %s ORDER BY id DESC LIMIT ?,?`
	_addLaser                 = `INSERT INTO app_laser (app_key,platform,mid,buvid,log_date,operator,status,silence_status,channel,description,mobi_app) VALUES(?,?,?,?,?,?,?,?,?,?,?)`
	_delLaser                 = `DELETE FROM app_laser WHERE id=?`
	_upLaserStatus            = `UPDATE app_laser SET status=?,recall_mobi_app=?,build=? WHERE id=?`
	_upLaserMsgId             = `UPDATE app_laser SET msg_id=? WHERE id=?`
	_upLaserURL               = `UPDATE app_laser SET url=?,md5=?,raw_upos_uri=? WHERE id=?`
	_upLaserSilenceStatus     = `UPDATE app_laser SET silence_status=?,recall_mobi_app=?,build=? WHERE id=?`
	_upLaserSilenceURL        = `UPDATE app_laser SET silence_url=? WHERE id=?`
	_upLaserErrorMessage      = `UPDATE app_laser SET error_msg=? WHERE id=?`
	_upLaserParseStatus       = `UPDATE app_laser SET parse_status=? WHERE id=?`
	_upActiveLaserParseStatus = `UPDATE app_laser_active SET parse_status=? WHERE id=?`
	_addLaser2                = `INSERT INTO app_laser_active (app_key,mid,buvid,url,recall_mobi_app,build,error_msg,md5,raw_upos_uri,task_status) VALUES(?,?,?,?,?,?,?,?,?,?)`
	_updateLaser2             = `UPDATE app_laser_active SET url=?,task_status=?,error_msg=?,md5=?,raw_upos_uri=? WHERE app_key=? AND id=?`
	_laserActiveCount         = `SELECT count(*) FROM app_laser_active WHERE app_key=? %s`
	_laserActiveList          = `SELECT id,app_key,platform,mid,buvid,url,recall_mobi_app,build,error_msg,task_status,parse_status,md5,unix_timestamp(ctime),unix_timestamp(mtime) FROM
app_laser_active WHERE app_key=? %s ORDER BY mtime DESC LIMIT ?,?`
	_laserActiveByID = `SELECT id,app_key,platform,mid,buvid,url,recall_mobi_app,build,error_msg,task_status,unix_timestamp(ctime),unix_timestamp(mtime) FROM
app_laser_active WHERE app_key=? AND id=?`
	_laserCmdCount = `SELECT count(*) FROM app_laser_command WHERE app_key=? %s`
	_laserCmdList  = `SELECT id,app_key,platform,mid,buvid,action,params,url,result,status,operator,description,mobi_app,recall_mobi_app,build,error_msg,unix_timestamp(ctime),unix_timestamp(mtime) 
	FROM app_laser_command WHERE app_key=? %s ORDER BY id DESC LIMIT ?,?`
	_laserCmdInfo = `SELECT id,app_key,platform,mid,buvid,action,params,url,result,status,operator,description,mobi_app,recall_mobi_app,build,error_msg,unix_timestamp(ctime),unix_timestamp(mtime) 
	FROM app_laser_command WHERE id=?`
	_addLaserCmd        = `INSERT INTO app_laser_command (app_key,platform,mid,buvid,action,params,status,operator,description,mobi_app) VALUES (?,?,?,?,?,?,?,?,?,?)`
	_upLaserCmd         = `INSERT app_laser_command (id, status,recall_mobi_app,build,url,error_msg,result,md5,raw_upos_uri) VALUES (?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE status=?,recall_mobi_app=?,build=?,url=?,error_msg=?,result=?,md5=?,raw_upos_uri=?`
	_upLaserCmdStatus   = `UPDATE app_laser_command SET status=?,recall_mobi_app=?,build=? WHERE id=?`
	_delLaserCmd        = `DELETE FROM app_laser_command WHERE id=?`
	_addLaserCmdAction  = `INSERT INTO app_laser_command_actions (action_name,platform,params,operator,description) VALUES (?,?,?,?,?)`
	_upLaserCmdAction   = `UPDATE app_laser_command_actions SET action_name=?,platform=?,params=?,operator=?,description=? WHERE id=?`
	_laserCmdActionList = `SELECT id,action_name,platform,params,operator,description,unix_timestamp(mtime),unix_timestamp(ctime) FROM app_laser_command_actions %s ORDER BY id DESC`
	_delLaserCmdAction  = `DELETE FROM app_laser_command_actions WHERE id=?`
	_laserWithCrash     = `SELECT l.id,l.app_key,l.platform,l.mid,l.buvid,l.email,l.log_date,l.url,l.status,l.silence_url,l.silence_status,l.parse_status,l.channel,l.description,l.mobi_app,l.recall_mobi_app,l.build,l.error_msg,l.operator,unix_timestamp(l.ctime),
unix_timestamp(l.mtime) FROM app_laser as l,veda_crash_laser_relation as r WHERE r.laser_id=l.id AND r.error_stack_hash_without_useless=? ORDER BY l.ctime DESC`
)

// LaserCount get laser count.
func (d *Dao) LaserCount(c context.Context, appKey, platform, buvid, operator string, id, mid, startTime, endTime int64, status int) (count int, err error) {
	var (
		sqls string
		args []interface{}
	)
	args = append(args, appKey)
	if id != 0 {
		sqls += " AND id=?"
		args = append(args, id)
	}
	if mid != 0 {
		sqls += " AND mid=?"
		args = append(args, mid)
	}
	if platform != "" {
		sqls += " AND platform=?"
		args = append(args, platform)
	}
	if buvid != "" {
		sqls += " AND buvid=?"
		args = append(args, buvid)
	}
	if status != 0 {
		sqls += " AND status=?"
		args = append(args, status)
	}
	if operator != "" {
		sqls += " AND operator=?"
		args = append(args, operator)
	}
	if startTime != 0 {
		sqls += " AND unix_timestamp(ctime)>=?"
		args = append(args, startTime)
	}
	if endTime != 0 {
		sqls += " AND unix_timestamp(ctime)<=?"
		args = append(args, endTime)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_laserCount, sqls), args...)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("LaserCount %v", err)
		}
	}
	return
}

// LaserList get laser list.
func (d *Dao) LaserList(c context.Context, appKey, platform, buvid, operator string, id, mid, startTime, endTime int64, status, pn, ps int) (res []*appmdl.Laser, err error) {
	var (
		sqls string
		args []interface{}
	)
	args = append(args, appKey)
	if id != 0 {
		sqls += " AND id=?"
		args = append(args, id)
	}
	if mid != 0 {
		sqls += " AND mid=?"
		args = append(args, mid)
	}
	if platform != "" {
		sqls += " AND platform=?"
		args = append(args, platform)
	}
	if buvid != "" {
		sqls += " AND buvid=?"
		args = append(args, buvid)
	}
	if status != 0 {
		sqls += " AND status=?"
		args = append(args, status)
	}
	if operator != "" {
		sqls += " AND operator=?"
		args = append(args, operator)
	}
	if startTime != 0 {
		sqls += " AND unix_timestamp(ctime)>=?"
		args = append(args, startTime)
	}
	if endTime != 0 {
		sqls += " AND unix_timestamp(ctime)<=?"
		args = append(args, endTime)
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_laserList, sqls), args...)
	if err != nil {
		log.Error("LaserList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &appmdl.Laser{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Platform, &re.MID, &re.Buvid, &re.Email, &re.LogDate, &re.URL,
			&re.Status, &re.SilenceURL, &re.SilenceStatus, &re.ParseStatus, &re.Channel, &re.Description, &re.MobiApp, &re.RecallMobiApp, &re.Build, &re.ErrorMessage, &re.Operator, &re.MsgId, &re.MD5, &re.CTime, &re.MTime); err != nil {
			log.Error("LaserList%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxAddLaser add laser.
func (d *Dao) TxAddLaser(tx *sql.Tx, appKey, platform, buvid, logDate, userName, description, mobiApp string, mid int64, status, silenceStatus, channel int) (r int64, err error) {
	res, err := tx.Exec(_addLaser, appKey, platform, mid, buvid, logDate, userName, status, silenceStatus, channel, description, mobiApp)
	if err != nil {
		log.Error("TxAddLaser %v", err)
		return
	}
	return res.LastInsertId()
}

// TxDelLaser del laser.
func (d *Dao) TxDelLaser(tx *sql.Tx, id int64) (r int64, err error) {
	res, err := tx.Exec(_delLaser, id)
	if err != nil {
		log.Error("TxDelLaser %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpLaserStatus update laser status.
func (d *Dao) TxUpLaserStatus(tx *sql.Tx, id int64, status int, recallMobiApp, build string) (r int64, err error) {
	res, err := tx.Exec(_upLaserStatus, status, recallMobiApp, build, id)
	if err != nil {
		log.Error("TxUpLaserStatus %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxUpMsgId(tx *sql.Tx, id int64, msgId int64) (err error) {
	_, err = tx.Exec(_upLaserMsgId, msgId, id)
	if err != nil {
		log.Error("TxUpLaserStatus %v", err)
		return
	}
	return
}

// TxUpLaserURL update laser url.
func (d *Dao) TxUpLaserURL(tx *sql.Tx, id int64, url, md5, rawUposUri string) (r int64, err error) {
	res, err := tx.Exec(_upLaserURL, url, md5, rawUposUri, id)
	if err != nil {
		log.Error("TxUpLaserURL %v", err)
		return
	}
	return res.RowsAffected()
}

// TxAddLaser2 update laser url.
func (d *Dao) TxAddLaser2(tx *sql.Tx, appKey, buvid, url, recallMobiApp, build, errorMessage, md5, rawUposUri string, status int, mid int64) (r int64, err error) {
	res, err := tx.Exec(_addLaser2, appKey, mid, buvid, url, recallMobiApp, build, errorMessage, md5, rawUposUri, status)
	if err != nil {
		log.Error("TxAddLaser2 %v", err)
		return
	}
	return res.LastInsertId()
}

// TxUpdateLaser2 update laser url.
func (d *Dao) TxUpdateLaser2(tx *sql.Tx, appKey, url, errorMessage, md5, rawUposUri string, status int, taskID int64) (r int64, err error) {
	res, err := tx.Exec(_updateLaser2, url, status, errorMessage, md5, rawUposUri, appKey, taskID)
	if err != nil {
		log.Error("TxUpdateLaser2 %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) LaserActiveCount(c context.Context, appKey, buvid string, mid, laserId, startTime, endTime int64) (count int, err error) {
	var (
		sqls string
		args []interface{}
	)
	args = append(args, appKey)
	if mid != 0 {
		sqls += " AND mid=?"
		args = append(args, mid)
	}
	if buvid != "" {
		sqls += " AND buvid=?"
		args = append(args, buvid)
	}
	if laserId != 0 {
		sqls += " AND id=?"
		args = append(args, laserId)
	}
	if startTime != 0 {
		sqls += " AND unix_timestamp(ctime)>=?"
		args = append(args, startTime)
	}
	if endTime != 0 {
		sqls += " AND unix_timestamp(ctime)<=?"
		args = append(args, endTime)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_laserActiveCount, sqls), args...)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("LaserActiveCount %v", err)
		}
	}
	return
}

// 判断主动上报日志是否存在
func (d *Dao) LaserActiveByID(c context.Context, appKey string, taskId int64) (res []*appmdl.Laser, err error) {
	rows, err := d.db.Query(c, _laserActiveByID, appKey, taskId)
	if err != nil {
		log.Error("LaserActiveByID %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &appmdl.Laser{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Platform, &re.MID, &re.Buvid, &re.URL, &re.RecallMobiApp, &re.Build, &re.ErrorMessage, &re.Status, &re.CTime, &re.MTime); err != nil {
			log.Error("LaserActiveByID %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// LaserList get laser list.
func (d *Dao) LaserActiveList(c context.Context, appKey, buvid string, mid, laserId, startTime, endTime int64, pn, ps int) (res []*appmdl.Laser, err error) {
	var (
		sqls string
		args []interface{}
	)
	args = append(args, appKey)
	if mid != 0 {
		sqls += " AND mid=?"
		args = append(args, mid)
	}
	if buvid != "" {
		sqls += " AND buvid=?"
		args = append(args, buvid)
	}
	if laserId != 0 {
		sqls += " AND id=?"
		args = append(args, laserId)
	}
	if startTime != 0 {
		sqls += " AND unix_timestamp(ctime)>=?"
		args = append(args, startTime)
	}
	if endTime != 0 {
		sqls += " AND unix_timestamp(ctime)<=?"
		args = append(args, endTime)
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_laserActiveList, sqls), args...)
	if err != nil {
		log.Error("LaserList %v", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		re := &appmdl.Laser{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Platform, &re.MID, &re.Buvid, &re.URL, &re.RecallMobiApp, &re.Build, &re.ErrorMessage, &re.Status, &re.ParseStatus, &re.MD5, &re.CTime, &re.MTime); err != nil {
			log.Error("LaserList%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxUpLaserStatus update laser status.
func (d *Dao) TxUpLaserSilenceStatus(tx *sql.Tx, id int64, status int, recallMobiApp, build string) (r int64, err error) {
	res, err := tx.Exec(_upLaserSilenceStatus, status, recallMobiApp, build, id)
	if err != nil {
		log.Error("TxUpLaserSilenceStatus %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpLaserSilenceURL update laser url.
func (d *Dao) TxUpLaserSilenceURL(tx *sql.Tx, id int64, url string) (r int64, err error) {
	res, err := tx.Exec(_upLaserSilenceURL, url, id)
	if err != nil {
		log.Error("TxUpLaserSilenceURL %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpLaserErrorMessage update laser error_msg.
func (d *Dao) TxUpLaserErrorMessage(tx *sql.Tx, id int64, errorMessage string) (r int64, err error) {
	res, err := tx.Exec(_upLaserErrorMessage, errorMessage, id)
	if err != nil {
		log.Error("TxUpLaserErrorMessage %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpLaserParseStatus update laser parse_status.
func (d *Dao) TxUpLaserParseStatus(tx *sql.Tx, status int, laserID int64) (err error) {
	_, err = tx.Exec(_upLaserParseStatus, status, laserID)
	if err != nil {
		log.Error("TxUpLaserParseStatus %v", err)
	}
	return
}

// TxUpActiveLaserParseStatus update laser parse_status.
func (d *Dao) TxUpActiveLaserParseStatus(tx *sql.Tx, status int, laserID int64) (err error) {
	_, err = tx.Exec(_upActiveLaserParseStatus, status, laserID)
	if err != nil {
		log.Error("TxUpActiveLaserParseStatus %v", err)
	}
	return
}

// LaserCmdCount get laser command count.
func (d *Dao) LaserCmdCount(c context.Context, appKey, platform, buvid, action, operator string, mid, id int64, status int) (count int, err error) {
	var (
		sqls string
		args []interface{}
	)
	args = append(args, appKey)
	if id != 0 {
		sqls += " AND id=?"
		args = append(args, id)
	}
	if mid != 0 {
		sqls += " AND mid=?"
		args = append(args, mid)
	}
	if platform != "" {
		sqls += " AND platform=?"
		args = append(args, platform)
	}
	if buvid != "" {
		sqls += " AND buvid=?"
		args = append(args, buvid)
	}
	if action != "" {
		sqls += " AND action=?"
		args = append(args, action)
	}
	if status != 0 {
		sqls += " AND status=?"
		args = append(args, status)
	}
	if operator != "" {
		sqls += " AND operator=?"
		args = append(args, operator)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_laserCmdCount, sqls), args...)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("LaserCount %v", err)
		}
	}
	return
}

// LaserCmdList get laser command list.
func (d *Dao) LaserCmdList(c context.Context, appKey, platform, buvid, action, operator string, id, mid int64, status, pn, ps int) (res []*appmdl.LaserCmd, err error) {
	var (
		sqls string
		args []interface{}
	)
	args = append(args, appKey)
	if id != 0 {
		sqls += " AND id=?"
		args = append(args, id)
	}
	if mid != 0 {
		sqls += " AND mid=?"
		args = append(args, mid)
	}
	if platform != "" {
		sqls += " AND platform=?"
		args = append(args, platform)
	}
	if buvid != "" {
		sqls += " AND buvid=?"
		args = append(args, buvid)
	}
	if action != "" {
		sqls += " AND action=?"
		args = append(args, action)
	}
	if status != 0 {
		sqls += " AND status=?"
		args = append(args, status)
	}
	if operator != "" {
		sqls += " AND operator=?"
		args = append(args, operator)
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_laserCmdList, sqls), args...)
	if err != nil {
		log.Error("LaserList error: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &appmdl.LaserCmd{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Platform, &re.MID, &re.Buvid, &re.Action, &re.Params, &re.URL, &re.Result,
			&re.Status, &re.Operator, &re.Description, &re.MobiApp, &re.RecallMobiApp, &re.Build, &re.ErrorMessage, &re.CTime, &re.MTime); err != nil {
			log.Error("LaserList error: %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// LaserCmdInfo get single cmd info
func (d *Dao) LaserCmdInfo(c context.Context, taskID int64) (re *appmdl.LaserCmd, err error) {
	row := d.db.QueryRow(c, _laserCmdInfo, taskID)
	re = &appmdl.LaserCmd{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.Platform, &re.MID, &re.Buvid, &re.Action, &re.Params, &re.URL, &re.Result,
		&re.Status, &re.Operator, &re.Description, &re.MobiApp, &re.RecallMobiApp, &re.Build, &re.ErrorMessage, &re.CTime, &re.MTime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("LaserCmdInfo %v", err)
		}
	}
	return
}

// TxAddLaserCmd add laser cmd
func (d *Dao) TxAddLaserCmd(tx *sql.Tx, appKey, mobiApp, platform, buvid, action, description, paramsStr, operator string, mid int64) (id int64, err error) {
	res, err := tx.Exec(_addLaserCmd, appKey, platform, mid, buvid, action, paramsStr, appmdl.StatusQueuing, operator, description, mobiApp)
	if err != nil {
		log.Error("TxAddLaserCmd error: %v", err)
		return
	}
	return res.LastInsertId()
}

// TxUpLaserCmd update laser cmd info
func (d *Dao) TxUpLaserCmd(c context.Context, taskID int64, status int, recallMobiApp, build, url, errorMsg, result, md5, rawUposUri string) (err error) {
	_, err = d.db.Exec(c, _upLaserCmd, taskID, status, recallMobiApp, build, url, errorMsg, result, md5, rawUposUri, status, recallMobiApp, build, url, errorMsg, result, md5, rawUposUri)
	if err != nil {
		log.Error("TxUpLaserCmd error: %v", err)
	}
	return
}

// TxUpLaserStatus update laser cmd status.
func (d *Dao) TxUpLaserCmdStatus(tx *sql.Tx, id int64, status int, recallMobiApp, build string) (r int64, err error) {
	res, err := tx.Exec(_upLaserCmdStatus, status, recallMobiApp, build, id)
	if err != nil {
		log.Error("TxUpLaserCmdStatus %v", err)
		return
	}
	return res.RowsAffected()
}

// DelLaserCmd delete lase cmd
func (d *Dao) DelLaserCmd(c context.Context, taskID int64) (err error) {
	_, err = d.db.Exec(c, _delLaserCmd, taskID)
	return
}

// AddLaserCmdAction add laser cmd action
func (d *Dao) AddLaserCmdAction(c context.Context, name, platform, paramsJSON, operator, description string) (err error) {
	_, err = d.db.Exec(c, _addLaserCmdAction, name, platform, paramsJSON, operator, description)
	return
}

// UpdateLaserCmdAction update laser cmd action
func (d *Dao) UpdateLaserCmdAction(c context.Context, id int64, name, platform, paramsJSON, operator, description string) (err error) {
	_, err = d.db.Exec(c, _upLaserCmdAction, name, platform, paramsJSON, operator, description, id)
	return
}

// DelLaserCmdAction del laser cmd action
func (d *Dao) DelLaserCmdAction(c context.Context, id int64) (err error) {
	_, err = d.db.Exec(c, _delLaserCmdAction, id)
	return
}

// LaserCmdActionList get laser cmd action list
func (d *Dao) LaserCmdActionList(c context.Context, name, platform string) (res []*appmdl.LaserCmdAction, err error) {
	var (
		sqls string
		args []interface{}
	)
	if name != "" {
		sqls += "AND (name LIKE ?)"
		args = append(args, "%"+name+"%")
	}
	if platform != "" {
		sqls += "AND (platform=? OR platform='')"
		args = append(args, platform)
	}
	if len(args) > 0 {
		sqls = strings.Replace(sqls, "AND", "WHERE", 1)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_laserCmdActionList, sqls), args...)
	if err != nil {
		log.Error("_laserCmdActionList error: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &appmdl.LaserCmdAction{}
		if err = rows.Scan(&re.ID, &re.Name, &re.Platform, &re.Params, &re.Operator, &re.Description, &re.Mtime, &re.Ctime); err != nil {
			log.Error("LaserCmdActionList rows.Scan err: %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) LaserWithCrash(c context.Context, errorStackHashWithoutUseless string) (res []*appmdl.Laser, err error) {
	rows, err := d.db.Query(c, _laserWithCrash, errorStackHashWithoutUseless)
	if err != nil {
		log.Error("LaserWithCrash d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &appmdl.Laser{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Platform, &re.MID, &re.Buvid, &re.Email, &re.LogDate, &re.URL,
			&re.Status, &re.SilenceURL, &re.SilenceStatus, &re.ParseStatus, &re.Channel, &re.Description, &re.MobiApp, &re.RecallMobiApp, &re.Build, &re.ErrorMessage, &re.Operator, &re.CTime, &re.MTime); err != nil {
			log.Error("LaserCmdActionList rows.Scan err: %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}
