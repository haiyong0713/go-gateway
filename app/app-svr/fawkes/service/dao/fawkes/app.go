package fawkes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/database/xsql"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_app = `SELECT a.id,aa.id,aa.datacenter_app_id,a.app_id,a.app_key,a.mobi_app,a.platform,a.name,a.icon,a.tree_path,a.description,a.git_path,
a.git_prj_id,a.robot_name,a.robot_webhook_url,aa.owners,aa.operator,aa.state,aa.manager_plat,aa.debug_url,aa.app_dsym_name,aa.app_symbolso_name,aa.server_zone,aa.laser_webhook,aa.is_host,aa.is_highest_peak,
unix_timestamp(a.ctime),unix_timestamp(aa.mtime) FROM app AS a,app_attribute AS aa WHERE a.id=? AND a.app_key=aa.app_key AND a.id=aa.app_table_id AND a.app_key=?`
	// nolint:gosec
	_appPass = `SELECT a.id,aa.id,aa.datacenter_app_id,a.app_id,a.app_key,a.mobi_app,a.platform,a.name,a.icon,a.tree_path,a.description,a.git_path,
a.git_prj_id,a.robot_name,a.robot_webhook_url,aa.owners,aa.operator,aa.state,aa.manager_plat,aa.debug_url,aa.app_dsym_name,aa.app_symbolso_name,aa.server_zone,aa.laser_webhook,aa.is_host,aa.is_highest_peak,
unix_timestamp(a.ctime),unix_timestamp(aa.mtime) FROM app AS a,app_attribute AS aa WHERE a.app_key=aa.app_key AND a.id=aa.app_table_id AND aa.state=1 AND a.app_key=?`
	_appByID = `SELECT a.id,aa.id,aa.datacenter_app_id,a.app_id,a.app_key,a.mobi_app,a.platform,a.name,a.icon,a.tree_path,a.description,a.git_path,
a.git_prj_id,a.robot_name,a.robot_webhook_url,aa.owners,aa.operator,aa.state,aa.manager_plat,aa.debug_url,aa.app_dsym_name,aa.app_symbolso_name,aa.server_zone,aa.laser_webhook,aa.is_host,
unix_timestamp(a.ctime),unix_timestamp(aa.mtime) FROM app AS a,app_attribute AS aa WHERE a.app_key=aa.app_key AND a.id=aa.app_table_id AND a.id=?`
	// nolint:gosec
	_appsPass = `SELECT a.id,aa.id,aa.datacenter_app_id,aa.server_zone,aa.laser_webhook,a.app_id,a.app_key,a.mobi_app,a.platform,a.name,a.icon,a.tree_path,a.description,a.git_path,
a.git_prj_id,a.robot_name,a.robot_webhook_url,aa.owners,aa.operator,aa.state,aa.manager_plat,aa.debug_url,aa.app_dsym_name,aa.app_symbolso_name,IFNULL(au.role,5),aa.is_host,aa.is_highest_peak,
unix_timestamp(a.ctime),unix_timestamp(aa.mtime) FROM app AS a LEFT JOIN auth_user as au ON au.app_key=a.app_key AND au.name=?,app_attribute AS aa WHERE a.app_key=aa.app_key AND a.id=aa.app_table_id AND aa.state=1 %v`
	_inApp          = `INSERT INTO app (app_id,app_key,mobi_app,platform,name,icon,tree_path,description,git_path,git_prj_id) VALUES(?,?,?,?,?,?,?,?,?,?)`
	_appCount       = `SELECT count(*) FROM app AS a,app_attribute AS aa WHERE a.app_key=aa.app_key AND a.id=aa.app_table_id AND a.app_key=? AND aa.state!=-2`
	_activeAppCount = `SELECT count(*) FROM app AS a,app_attribute AS aa WHERE a.app_key=aa.app_key AND a.id=aa.app_table_id AND a.app_key=? AND aa.state>=0`
	_upApp          = `UPDATE app SET app_id=?,mobi_app=?,platform=?,git_path=?,name=?,icon=?,description=?,tree_path=?,git_prj_id=? WHERE id=?`
	_appAll         = `SELECT id,app_id,app_key,mobi_app,platform,name,icon,tree_path,description,git_path,git_prj_id FROM app`

	_auditApp = `SELECT a.id,aa.id,aa.datacenter_app_id,aa.workflow_id,aa.server_zone,a.app_id,a.app_key,a.mobi_app,a.platform,a.name,a.icon,a.tree_path,a.description,a.git_path,
a.git_prj_id,aa.owners,aa.operator,aa.state,aa.refusal_reason,aa.is_host,unix_timestamp(a.ctime),unix_timestamp(aa.mtime) FROM app AS a,app_attribute AS aa WHERE a.app_key=aa.app_key AND a.id=aa.app_table_id ORDER BY aa.ctime DESC`
	_inAppAttribute              = `INSERT INTO app_attribute (app_table_id,datacenter_app_id,workflow_id,app_key,is_host,owners,operator) VALUES(?,?,?,?,?,?,?)`
	_upAppAttribute              = `UPDATE app_attribute SET owners=?,app_dsym_name=?,app_symbolso_name=?,operator=?,datacenter_app_id=?,server_zone=?,laser_webhook=?,is_host=? WHERE app_table_id=?`
	_upAppAudit                  = `UPDATE app_attribute SET state=?,refusal_reason=? WHERE app_key=? AND app_table_id=?`
	_upAppAttributeIsHighestPeak = `UPDATE app_attribute SET is_highest_peak=? WHERE app_key=?`

	_appFollow    = `SELECT app_key FROM app_follow WHERE uname=?`
	_inAppFollow  = `INSERT INTO app_follow (app_key,uname) VALUES(?,?)`
	_delAppFollow = `DELETE FROM app_follow WHERE app_key=? AND uname=?`

	_appBaseInfo         = `SELECT a.app_id,a.git_path,a.git_prj_id FROM app AS a,app_attribute AS aa WHERE a.id=aa.app_table_id AND aa.state=1 AND a.app_key=?`
	_robotWebhook        = `SELECT webhook FROM bot_manage WHERE FIND_IN_SET(?, app_keys) AND func_module=?`
	_appRobotCount       = `SELECT count(*) FROM bot_manage %s`
	_appRobotList        = `SELECT id,bot_name,webhook,app_keys,func_module,users,state,is_global,description,operator,unix_timestamp(mtime),unix_timestamp(ctime),is_default FROM bot_manage %s ORDER BY id DESC`
	_appRobotById        = `SELECT id,bot_name,webhook,app_keys,func_module,users,state,is_global,description,operator,unix_timestamp(mtime),unix_timestamp(ctime),is_default FROM bot_manage WHERE id=?`
	_appRobotByWebhook   = `SELECT id,bot_name,webhook,app_keys,func_module,users,state,is_global,description,operator,unix_timestamp(mtime),unix_timestamp(ctime),is_default FROM bot_manage WHERE webhook=?`
	_addAppRobot         = `INSERT INTO bot_manage (bot_name,webhook,app_keys,func_module,users,state,is_global,description,operator,is_default) VALUES (?,?,?,?,?,?,?,?,?,?)`
	_upAppRobot          = `UPDATE bot_manage SET bot_name=?,webhook=?,operator=?,app_keys=? %v WHERE id=?`
	_upAppRobotAppKey    = `UPDATE bot_manage SET app_keys=? WHERE id=?`
	_delAppRobot         = `DELETE FROM bot_manage WHERE id=?`
	_appNotificationList = `SELECT id,app_keys,platform,route_path,title,content,url,closeable,state,is_global,type,operator,
unix_timestamp(effect_time),unix_timestamp(expire_time),unix_timestamp(mtime),unix_timestamp(ctime) FROM notification_config %v ORDER BY id DESC`
	_updateNotification = `UPDATE notification_config SET app_keys=?,platform=?,route_path=?,title=?,content=?,url=?,closeable=?,state=?,is_global=?,type=?,effect_time=?,expire_time=?,operator=? WHERE id=?`
	_addNotification    = `INSERT INTO notification_config (app_keys,platform,route_path,title,content,url,closeable,state,is_global,type,effect_time,expire_time,operator) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`

	_appListByMobiApp = `SELECT a.id,aa.id as attr_id,aa.datacenter_app_id,a.app_id,a.app_key,a.mobi_app,a.platform,a.name,a.icon,a.tree_path,a.description,a.git_path,
a.git_prj_id,a.robot_name,a.robot_webhook_url,aa.owners,aa.operator,aa.state,aa.manager_plat,aa.debug_url,aa.app_dsym_name,aa.app_symbolso_name,aa.server_zone,unix_timestamp(a.ctime) as ctime,
unix_timestamp(aa.mtime) as ptime FROM (app AS a INNER JOIN app_attribute AS aa ON a.id=aa.app_table_id) WHERE a.mobi_app=? AND aa.state=1`

	_appListByDatacenterAppId = `SELECT a.id,aa.id as attr_id,aa.datacenter_app_id,a.app_id,a.app_key,a.mobi_app,a.platform,a.name,a.icon,a.tree_path,a.description,a.git_path,
a.git_prj_id,a.robot_name,a.robot_webhook_url,aa.owners,aa.operator,aa.state,aa.manager_plat,aa.debug_url,aa.app_dsym_name,aa.app_symbolso_name,aa.server_zone,unix_timestamp(a.ctime) as ctime,
unix_timestamp(aa.mtime) as ptime FROM (app AS a INNER JOIN app_attribute AS aa ON a.id=aa.app_table_id) WHERE aa.datacenter_app_id=? AND aa.state=1`

	_appHost = `SELECT aa.app_key FROM app_attribute as aa INNER JOIN app as a ON aa.app_table_id=a.id WHERE aa.is_host=1 AND aa.datacenter_app_id=? AND a.platform=?`
)

// AppInfo get app info.
func (d *Dao) AppInfo(c context.Context, appKey string, ID int64) (re *appmdl.APP, err error) {
	// 历史原因：由于appInfo sql 会出现重复app_key的查出多条的情况，故增加一个id字段查询，缺省则查appPass
	var (
		_sql string
		args []interface{}
	)
	if ID == -1 {
		_sql = _appPass
		args = append(args, appKey)
	} else {
		_sql = _app
		args = append(args, ID, appKey)
	}
	row := d.db.QueryRow(c, _sql, args...)
	re = &appmdl.APP{}
	if err = row.Scan(&re.ID, &re.AttrID, &re.DataCenterAppID, &re.AppID, &re.AppKey, &re.MobiApp, &re.Platform, &re.Name, &re.Icon, &re.TreePath,
		&re.Desc, &re.GitPath, &re.GitPrjID, &re.RobotName, &re.RobotWebhookUrl, &re.Owners,
		&re.Operator, &re.State, &re.ManagerPlat, &re.DebugUrl, &re.AppDsymName, &re.AppSymbolsoName, &re.ServerZone, &re.LaserWebhook, &re.IsHost, &re.IsHighestPeak, &re.CTime, &re.PTime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("AppInfo %v", err)
		}
	}
	return
}

// AppPass get app info pass.
func (d *Dao) AppPass(c context.Context, appKey string) (re *appmdl.APP, err error) {
	row := d.db.QueryRow(c, _appPass, appKey)
	re = &appmdl.APP{}
	if err = row.Scan(&re.ID, &re.AttrID, &re.DataCenterAppID, &re.AppID, &re.AppKey, &re.MobiApp, &re.Platform, &re.Name, &re.Icon, &re.TreePath,
		&re.Desc, &re.GitPath, &re.GitPrjID, &re.RobotName, &re.RobotWebhookUrl, &re.Owners,
		&re.Operator, &re.State, &re.ManagerPlat, &re.DebugUrl, &re.AppDsymName, &re.AppSymbolsoName, &re.ServerZone, &re.LaserWebhook, &re.IsHost, &re.IsHighestPeak, &re.CTime, &re.PTime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("AppInfo %v", err)
		}
	}
	return
}

// AppByID get app info by id.
func (d *Dao) AppByID(c context.Context, id int64) (re *appmdl.APP, err error) {
	row := d.db.QueryRow(c, _appByID, id)
	re = &appmdl.APP{}
	if err = row.Scan(&re.ID, &re.AttrID, &re.DataCenterAppID, &re.AppID, &re.AppKey, &re.MobiApp, &re.Platform, &re.Name, &re.Icon, &re.TreePath,
		&re.Desc, &re.GitPath, &re.GitPrjID, &re.RobotName, &re.RobotWebhookUrl, &re.Owners,
		&re.Operator, &re.State, &re.ManagerPlat, &re.DebugUrl, &re.AppDsymName, &re.AppSymbolsoName, &re.ServerZone, &re.LaserWebhook, &re.IsHost, &re.CTime, &re.PTime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("AppInfo %v", err)
		}
	}
	return
}

// TxUpApp update app name desc tree.
func (d *Dao) TxUpApp(tx *sql.Tx, id int64, appID, mobiApp, platform, gitPath, name, icon, desc, treePath, projectID string) (r int64, err error) {
	res, err := tx.Exec(_upApp, appID, mobiApp, platform, gitPath, name, icon, desc, treePath, projectID, id)
	if err != nil {
		log.Error("TxUpApp %v", err)
		return
	}
	return res.RowsAffected()
}

// AppFollow get follow appkeys list.
func (d *Dao) AppFollow(c context.Context, username string) (appKeys []string, err error) {
	rows, err := d.db.Query(c, _appFollow, username)
	if err != nil {
		log.Error("AppFollow %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var appKey string
		if err = rows.Scan(&appKey); err != nil {
			log.Error("AppFollow %v", err)
			return
		}
		appKeys = append(appKeys, appKey)
	}
	err = rows.Err()
	return
}

// AppsPass get app list passed.
func (d *Dao) AppsPass(c context.Context, appKeys []string, username string, datacenterAppId int64) (apps []*appmdl.APP, err error) {
	var (
		sqlAdd string
		sqls   []string
		args   []interface{}
	)
	args = append(args, username)
	if len(appKeys) > 0 {
		for _, appKey := range appKeys {
			sqls = append(sqls, "?")
			args = append(args, appKey)
		}
		sqlAdd = fmt.Sprintf(" AND a.app_key IN(%v)", strings.Join(sqls, ","))
	}
	if datacenterAppId != 0 {
		sqlAdd += " AND aa.datacenter_app_id=? "
		args = append(args, datacenterAppId)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_appsPass, sqlAdd), args...)
	if err != nil {
		log.Error("AppsPass %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		app := &appmdl.APP{}
		if err = rows.Scan(&app.ID, &app.AttrID, &app.DataCenterAppID, &app.ServerZone, &app.LaserWebhook, &app.AppID, &app.AppKey, &app.MobiApp, &app.Platform, &app.Name, &app.Icon,
			&app.TreePath, &app.Desc, &app.GitPath, &app.GitPrjID, &app.RobotName, &app.RobotWebhookUrl, &app.Owners,
			&app.Operator, &app.State, &app.ManagerPlat, &app.DebugUrl, &app.AppDsymName, &app.AppSymbolsoName, &app.Role, &app.IsHost, &app.IsHighestPeak, &app.CTime, &app.PTime); err != nil {
			log.Error("AppsPass %v", err)
			return
		}
		apps = append(apps, app)
	}
	err = rows.Err()
	return
}

// TxInAppFollow add app follow.
func (d *Dao) TxInAppFollow(tx *sql.Tx, appKey, username string) (r int64, err error) {
	res, err := tx.Exec(_inAppFollow, appKey, username)
	if err != nil {
		log.Error("TxInAppFollow %v", err)
		return
	}
	return res.RowsAffected()
}

// TxDelAppFollow del app follow.
func (d *Dao) TxDelAppFollow(tx *sql.Tx, appKey, username string) (r int64, err error) {
	res, err := tx.Exec(_delAppFollow, appKey, username)
	if err != nil {
		log.Error("TxDelAppFollow %v", err)
		return
	}
	return res.RowsAffected()
}

// TxAppAdd add app.
func (d *Dao) TxAppAdd(tx *sql.Tx, appID, appKey, mobiApp, platform, name, treePath, icon, desc, gitPath, gitPrjID string) (id int64, err error) {
	res, err := tx.Exec(_inApp, appID, appKey, mobiApp, platform, name, icon, treePath, desc, gitPath, gitPrjID)
	if err != nil {
		log.Error("TxAppAdd %v", err)
		return
	}
	return res.LastInsertId()
}

// TxAppAttributeAdd add app_attribute.
func (d *Dao) TxAppAttributeAdd(tx *sql.Tx, id, datacenterAppId, isHost int64, workflowId, appKey, owners, userName string) (r int64, err error) {
	res, err := tx.Exec(_inAppAttribute, id, datacenterAppId, workflowId, appKey, isHost, owners, userName)
	if err != nil {
		log.Error("TxAppAttributeAdd %v", err)
		return
	}
	return res.RowsAffected()
}

// TxAppAttributeUpdate owners.
func (d *Dao) TxAppAttributeUpdate(tx *sql.Tx, id, datacenterAppID, serverZone, isHost int64, owners, dsymName, symbolsoName, userName, laserWebhook string) (r int64, err error) {
	res, err := tx.Exec(_upAppAttribute, owners, dsymName, symbolsoName, userName, datacenterAppID, serverZone, laserWebhook, isHost, id)
	if err != nil {
		log.Error("TxAppAttributeUpdate %v", err)
		return
	}
	return res.RowsAffected()
}

// TxAppUpdateIsHighestPeak
func (d *Dao) TxAppUpdateIsHighestPeak(tx *sql.Tx, appKey string, isHighestPeak int64) (r int64, err error) {
	res, err := tx.Exec(_upAppAttributeIsHighestPeak, isHighestPeak, appKey)
	if err != nil {
		log.Error("TxAppUpdateIsHighestPeak %v", err)
		return
	}
	return res.RowsAffected()
}

// AppCount
func (d *Dao) AppCount(c context.Context, appKey string) (count int64, err error) {
	row := d.db.QueryRow(c, _appCount, appKey)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("CheckAppKey %v", err)
		}
	}
	return
}

// AppAll get all apps
func (d *Dao) AppAll(c context.Context) (apps []*appmdl.APP, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, _appAll); err != nil {
		log.Errorc(c, "select error: %v", err)
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &apps); err != nil {
		log.Errorc(c, "scan error: %v", err)
		return
	}
	return
}

// AppsAudit get app app audit.
func (d *Dao) AppsAudit(c context.Context) (auditApps []*appmdl.APP, err error) {
	rows, err := d.db.Query(c, _auditApp)
	if err != nil {
		log.Error("AppsAudit %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		app := &appmdl.APP{}
		if err = rows.Scan(&app.ID, &app.AttrID, &app.DataCenterAppID, &app.WorkflowId, &app.ServerZone, &app.AppID, &app.AppKey, &app.MobiApp, &app.Platform, &app.Name, &app.Icon,
			&app.TreePath, &app.Desc, &app.GitPath, &app.GitPrjID, &app.Owners,
			&app.Operator, &app.State, &app.Reason, &app.IsHost, &app.CTime, &app.PTime); err != nil {
			log.Error("AppsAudit %v", err)
			return
		}
		auditApps = append(auditApps, app)
	}
	err = rows.Err()
	return
}

// TxUpAppAudit audit app.
func (d *Dao) TxUpAppAudit(tx *sql.Tx, appKey, reason string, status int, id int64) (r int64, err error) {
	res, err := tx.Exec(_upAppAudit, status, reason, appKey, id)
	if err != nil {
		log.Error("TxUpAppAudit %v", err)
		return
	}
	return res.RowsAffected()
}

// ActiveAppCount
func (d *Dao) ActiveAppCount(c context.Context, appKey string) (count int, err error) {
	row := d.db.QueryRow(c, _activeAppCount, appKey)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("AppInfo %v", err)
		}
	}
	return
}

// AppBasicInfo get app basic Info
func (d *Dao) AppBasicInfo(c context.Context, appKey string) (appID, gitPath, gitlabPrjID string, err error) {
	row := d.db.QueryRow(c, _appBaseInfo, appKey)
	if err = row.Scan(&appID, &gitPath, &gitlabPrjID); err != nil {
		log.Error("d.AppBasicInfo d.db.QueryRow(%v) error(%v)", appKey, err)
		return
	}
	return
}

// AppRobotWebhook get app robot webhook url
func (d *Dao) AppRobotWebhook(c context.Context, appKey string) (robotWebhookURL string, err error) {
	row := d.db.QueryRow(c, _robotWebhook, appKey, appmdl.MessageBot)
	if err = row.Scan(&robotWebhookURL); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("AppRobotWebhook %v", err)
		}
	}
	return
}

func (d *Dao) AppRobotCount(c context.Context, appKey, funcModule, botName string, state int) (count int64, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if appKey != "" {
		sqlAdd += "AND (FIND_IN_SET(?, app_keys) or is_global=1)"
		args = append(args, appKey)
	}
	if funcModule != "" {
		sqlAdd += " AND (func_module=?)"
		args = append(args, funcModule)
	}
	if state != -1 {
		sqlAdd += "AND (state=?)"
		args = append(args, state)
	}
	if botName != "" {
		sqlAdd += "AND (bot_name LIKE ?)"
		args = append(args, "%"+botName+"%")
	}
	if len(sqlAdd) > 0 {
		sqlAdd = strings.Replace(sqlAdd, "AND", "WHERE", 1)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_appRobotCount, sqlAdd), args...)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return

}

// AppRobotList get app robots
func (d *Dao) AppRobotList(c context.Context, appKey, funcModule, botName string, state int) (res []*appmdl.Robot, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if appKey != "" {
		sqlAdd += "AND (FIND_IN_SET(?, app_keys) or is_global=1)"
		args = append(args, appKey)
	}
	if funcModule != "" {
		sqlAdd += " AND (func_module=?)"
		args = append(args, funcModule)
	}
	if state != -1 {
		sqlAdd += "AND (state=?)"
		args = append(args, state)
	}
	if botName != "" {
		sqlAdd += "AND (bot_name LIKE ?)"
		args = append(args, "%"+botName+"%")
	}
	if len(sqlAdd) > 0 {
		sqlAdd = strings.Replace(sqlAdd, "AND", "WHERE", 1)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_appRobotList, sqlAdd), args...)
	if err != nil {
		log.Error("Dao AppRobotList: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &appmdl.Robot{}
		if err = rows.Scan(&re.ID, &re.BotName, &re.WebHook, &re.AppKeys, &re.FuncModule, &re.Users, &re.State, &re.IsGlobal, &re.Description,
			&re.Operator, &re.Mtime, &re.Ctime, &re.IsDefault); err != nil {
			log.Error("AppRobotList Scan: %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) AppRobotInfoById(c context.Context, id int64) (res *appmdl.Robot, err error) {
	row := d.db.QueryRow(c, _appRobotById, id)
	res = &appmdl.Robot{}
	if err = row.Scan(&res.ID, &res.BotName, &res.WebHook, &res.AppKeys, &res.FuncModule, &res.Users, &res.State, &res.IsGlobal, &res.Description,
		&res.Operator, &res.Mtime, &res.Ctime, &res.IsDefault); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		}
	}
	return
}

func (d *Dao) AppRobotInfoByWebhook(c context.Context, webhook string) (res *appmdl.Robot, err error) {
	row := d.db.QueryRow(c, _appRobotByWebhook, webhook)
	res = &appmdl.Robot{}
	if err = row.Scan(&res.ID, &res.BotName, &res.WebHook, &res.AppKeys, &res.FuncModule, &res.Users, &res.State, &res.IsGlobal, &res.Description,
		&res.Operator, &res.Mtime, &res.Ctime, &res.IsDefault); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		}
	}
	return
}

// TxAppRobotAdd add robot
func (d *Dao) TxAppRobotAdd(tx *sql.Tx, botName, webhook, appKeys, funcModule, users, description, userName string, state, isGlobal, isDefaultDisplay int) (err error) {
	res, err := tx.Exec(_addAppRobot, botName, webhook, appKeys, funcModule, users, state, isGlobal, description, userName, isDefaultDisplay)
	if err != nil {
		return
	}
	_, err = res.RowsAffected()
	return err
}

// TxAppRobotUpdate update robot
func (d *Dao) TxAppRobotUpdate(tx *sql.Tx, botName, webhook, appKeys, funcModule, users, description, userName string, state, isGlobal, isDefault int, id int64) (err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, botName, webhook, userName, appKeys)
	if funcModule != appmdl.MessageBot {
		sqlAdd += ",func_module=?,users=?,state=?,is_global=?,description=?,is_default=?"
		args = append(args, funcModule, users, state, isGlobal, description, isDefault)
	}
	args = append(args, id)
	res, err := tx.Exec(fmt.Sprintf(_upAppRobot, sqlAdd), args...)
	if err != nil {
		return
	}
	_, err = res.RowsAffected()
	return err
}

// TxAppRobotUpdateAppKey update appKey by id
func (d *Dao) TxAppRobotUpdateAppKey(tx *sql.Tx, appKeys string, id int64) (err error) {
	_, err = tx.Exec(_upAppRobotAppKey, appKeys, id)
	return
}

// TxAppRobotDelete update robot
func (d *Dao) TxAppRobotDel(tx *sql.Tx, ID int64) (err error) {
	res, err := tx.Exec(_delAppRobot, ID)
	if err != nil {
		log.Error("TxAppRobotDelete %v", err)
		return
	}
	_, err = res.RowsAffected()
	return err
}

// RobotNotify robot notify
func (d *Dao) RobotNotify(webhookURL string, notify interface{}) (err error) {
	var (
		req    *http.Request
		reqMdl *appmdl.RobotReq
		res    *appmdl.RobotRes
	)
	switch robotMessage := notify.(type) {
	case *appmdl.Text:
		reqMdl = &appmdl.RobotReq{MsgType: "text", Text: robotMessage}
	case *appmdl.Markdown:
		reqMdl = &appmdl.RobotReq{MsgType: "markdown", Markdown: robotMessage}
	case *appmdl.Image:
		reqMdl = &appmdl.RobotReq{MsgType: "image", Image: robotMessage}
	case *appmdl.News:
		reqMdl = &appmdl.RobotReq{MsgType: "news", News: robotMessage}
	default:
		log.Error("RobotNotify: unknown type")
		return
	}
	// Push Message
	byteBuf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuf)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(reqMdl); err != nil {
		log.Error("d.RobotNotify json encode error(%v)", err)
		return
	}
	if req, err = http.NewRequest(http.MethodPost, webhookURL, strings.NewReader(byteBuf.String())); err != nil {
		log.Error("s.RobotNotify call http.NewRequest error(%v)", err)
		return
	}
	req.Header.Add("content-type", "application/json")
	if err = d.httpClient.Do(context.Background(), req, &res); err != nil {
		log.Error("RobotNotify error(%v)", err)
		return
	}
	return
}

// GetWXUsersByDepartmentId Get Users
func (d *Dao) GetWXUsersByDepartmentId(ctx context.Context, departmentID, token string) (users []*appmdl.User, err error) {
	var (
		req         *http.Request
		userlistRes *appmdl.WXNotifyUserListRes
	)
	if req, err = http.NewRequest(http.MethodGet, fmt.Sprintf(d.c.WXNotify.UserListURL, token, departmentID), nil); err != nil {
		log.Error("d.FawkesNotify call http.NewRequest error(%v)", err)
		return
	}
	req.Header.Add("content-type", "application/json")
	if err = d.httpClient.Do(ctx, req, &userlistRes); err != nil {
		log.Error("FawkesNotify error(%v)", userlistRes)
		return
	}
	users = userlistRes.UserList
	return
}

// AppNotificationList get notif list
func (d *Dao) AppNotificationList(c context.Context, appKey, platform string, state int64) (res []*appmdl.Notif, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if state != -1 {
		args = append(args, state)
		sqlAdd += "AND (state=?)"
	}
	if appKey != "" {
		args = append(args, appKey)
		sqlAdd += "AND (FIND_IN_SET(?, app_keys) OR is_global=1 )"
	}
	if platform != "" {
		args = append(args, platform)
		sqlAdd += "AND FIND_IN_SET(?, platform)"
	}
	if len(sqlAdd) > 0 {
		sqlAdd = strings.Replace(sqlAdd, "AND", "WHERE", 1)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_appNotificationList, sqlAdd), args...)
	if err != nil {
		log.Error("AppNotificationList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &appmdl.Notif{}
		if err = rows.Scan(&re.ID, &re.AppKeys, &re.Platform, &re.RoutePath, &re.Title, &re.Content, &re.URL, &re.Closeable, &re.State, &re.IsGlobal, &re.Type, &re.Operator, &re.EffectTime, &re.ExpireTime, &re.Mtime, &re.Ctime); err != nil {
			log.Error("AppNotificationList %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxAppNotificationUpdate update notif
func (d *Dao) TxAppNotificationUpdate(tx *sql.Tx, id int64, appKeys, platform, routePath, title, content, url string, state, isGlobal, showType, closeable int64, effectTime, expireTime, username string) (r int64, err error) {
	if expireTime == "" {
		expireTime = "0000-00-00 00:00:00"
	}
	res, err := tx.Exec(_updateNotification, appKeys, platform, routePath, title, content, url, closeable, state, isGlobal, showType, effectTime, expireTime, username, id)
	if err != nil {
		log.Error("TxUpNotification %v", err)
		return
	}
	return res.RowsAffected()
}

// TxAppNotificationAdd add notif
func (d *Dao) TxAppNotificationAdd(tx *sql.Tx, appKeys, platform, routePath, title, content, url string, state, isGlobal, showType, closeable int64, effectTime, expireTime, username string) (r int64, err error) {
	if expireTime == "" {
		expireTime = "0000-00-00 00:00:00"
	}
	res, err := tx.Exec(_addNotification, appKeys, platform, routePath, title, content, url, closeable, state, isGlobal, showType, effectTime, expireTime, username)
	if err != nil {
		log.Error("TxUpNotification %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) AppListByMobiApp(c context.Context, mobiApp string) (res []*appmdl.APP, err error) {
	rows, err := d.db.Query(c, _appListByMobiApp, mobiApp)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) AppListByDatacenterAppId(c context.Context, datacenterAppId int64) (res []*appmdl.APP, err error) {
	rows, err := d.db.Query(c, _appListByDatacenterAppId, datacenterAppId)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) AppHost(c context.Context, datacenterAppId int64, platform string) (appKey string, err error) {
	row := d.db.QueryRow(c, _appHost, datacenterAppId, platform)
	if err = row.Scan(&appKey); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}
	return
}
