package fawkes

import (
	"context"
	"fmt"
	"strings"

	xsql "go-common/library/database/sql"
	"go-common/library/xstr"

	confmdl "go-gateway/app/app-svr/fawkes/service/model/config"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_existConfigVersion = `SELECT count(*) FROM config_version WHERE app_key=? AND env=? AND version=? AND version_code=? AND state=1`
	_setConfigVersion   = `INSERT INTO config_version (app_key,env,version,version_code,state,operator) VALUES(?,?,?,?,1,?)`
	_configVersionCount = `SELECT count(*) FROM config_version WHERE app_key=? AND env =? AND state=1`
	_configVersionList  = `SELECT id,app_key,env,version,version_code FROM config_version WHERE app_key=? AND env=? AND 
state=1 ORDER BY version_code DESC LIMIT ?,?`
	_configDefaultVersion  = `SELECT id,app_key,env,version,version_code FROM config_version WHERE app_key=? AND env=? AND state=1 AND version = 'default'`
	_delConfigVersionState = `DELETE FROM config_version WHERE id=?`
	_configVersionID       = `SELECT id FROM config_version WHERE app_key=? AND env=? %v`
	_configVersionByID     = `SELECT id,app_key,env,version,version_code,modify_desc FROM config_version WHERE id=?`
	_upConfigVersionDesc   = `UPDATE config_version SET modify_desc=? WHERE id=?`
	_configVersionByIDs    = `SELECT id,app_key,env,version,version_code FROM config_version WHERE id IN(%v)`

	_configModifyCountAll = `SELECT count(*) FROM config WHERE app_key=? AND env=? AND state <>3`
	_configModifyCount    = `SELECT cvid,count(*) FROM config WHERE app_key=? AND env=? AND state<>3 AND cvid IN(%v) GROUP BY cvid`
	_addConfig            = `INSERT INTO config (app_key,env,cvid,cgroup,ckey,cvalue,state,operator,description) 
VALUES(?,?,?,?,?,?,1,?,?) ON DUPLICATE KEY UPDATE cvalue=?,operator=?,description=?`
	_upConfig  = `UPDATE config SET cvalue=?,state=2,operator=?,description=? WHERE app_key=? AND env=? AND cvid=? AND cgroup=? AND ckey=?`
	_delConfig = `UPDATE config SET state=-1,operator=?,description=? WHERE app_key=? AND env=? AND cvid=? AND cgroup=? AND ckey=?`
	_config    = `SELECT app_key,env,cvid,cgroup,ckey,cvalue,state,operator,description,unix_timestamp(mtime) 
FROM config WHERE app_key=? AND env=? AND cvid=? ORDER BY id`
	_configGroup           = `SELECT app_key,env,cvid,cgroup,ckey,cvalue,state,operator,description,unix_timestamp(mtime) FROM config WHERE app_key=? AND env=? AND cvid=? AND cgroup=? AND state != -1 ORDER BY id`
	_configItem            = `SELECT app_key,env,cvid,cgroup,ckey,cvalue,state,operator,description,unix_timestamp(mtime) FROM config WHERE app_key=? AND env=? AND cvid=? AND cgroup=? AND ckey=? AND state != -1 ORDER BY id`
	_upConfigState         = `UPDATE config SET state=? WHERE app_key=? AND env=? AND cvid=?`
	_upConfigStateMultiple = `UPDATE config SET state=? WHERE app_key=? AND env=? AND cvid=? AND (%v)`
	_upConfigItemState     = `UPDATE config SET state=? WHERE app_key=? AND env=? AND cvid=? AND cgroup=? AND ckey=?`
	_delConfig2            = `DELETE FROM config WHERE app_key=? AND env=? AND cvid=? AND cgroup=? AND ckey=?`
	_flushConfig           = `DELETE FROM config WHERE app_key=? AND env=? AND cvid=? AND state=-1`
	_flushConfigMultiple   = `DELETE FROM config WHERE app_key=? AND env=? AND cvid=? AND state=-1 AND (%v)`
	_addConfigs            = `INSERT INTO config (app_key,env,cvid,cgroup,ckey,cvalue,state,operator,description) VALUES %v`
	_delAllConfig          = `DELETE FROM config WHERE cvid=?`
	_upConfigDesc          = `UPDATE config SET description=? WHERE app_key=? AND env=? AND cvid=? AND cgroup=? AND ckey=?`

	_configPublish = `SELECT id,app_key,env,cvid,cv,state,md5,cdn_url,local_path,diffs,total_url,total_path,operator,
description,unix_timestamp(ctime),unix_timestamp(mtime) FROM config_publish WHERE app_key=? AND env=? AND cvid=? ORDER BY cv DESC %v`
	_configPublishs = `SELECT id,app_key,env,cvid,cv,state,md5,cdn_url,local_path,diffs,total_url,total_path,operator,
description,unix_timestamp(ctime),unix_timestamp(mtime) FROM config_publish WHERE app_key=? AND env=? AND cvid IN(%v) ORDER BY cv DESC`
	_configPublishByID = `SELECT id,app_key,env,cvid,cv,state,md5,cdn_url,local_path,diffs,total_url,total_path,operator,
description,unix_timestamp(ctime),unix_timestamp(mtime) FROM config_publish WHERE app_key=? AND id=?`
	_configPublishAll = `SELECT cp.app_key,cp.env,cp.cvid,cv.version,cv.version_code,cp.cv,cp.state,cp.md5,cp.cdn_url,
cp.local_path,cp.diffs,cp.total_url,cp.total_path,cp.operator,cp.description,unix_timestamp(cp.ctime),unix_timestamp(cp.mtime) 
FROM config_publish AS cp LEFT JOIN config_version AS cv ON cp.cvid=cv.id WHERE cp.app_key=? AND cp.env=? ORDER BY cp.cv DESC limit ?,?`
	_configLastCV         = `SELECT cv FROM config_publish WHERE app_key=? AND env=? AND cvid=? AND cv<? ORDER BY cv DESC LIMIT 1`
	_addConfigPublish     = `INSERT INTO config_publish (app_key,env,cvid,cv,description,operator) VALUES(?,?,?,?,?,?)`
	_upConfigPublishFiles = `UPDATE config_publish SET md5=?,cdn_url=?,local_path=?,diffs=? WHERE app_key=? AND env=? AND id=?`
	_upConfigPublishState = `UPDATE config_publish SET state=? WHERE app_key=? AND env=? AND cvid=? AND cv=?`
	_delAllConfigPublish  = `DELETE FROM config_publish WHERE cvid=?`
	_allNewConfigPublish  = `SELECT app_key,env,cvid,cv,md5,cdn_url,local_path,diffs,total_url,total_path,operator,
description,unix_timestamp(mtime) FROM config_publish WHERE app_key=? AND env=? AND state=1 ORDER BY cvid ASC,cv DESC`
	_upConfigPublishTotal     = `UPDATE config_publish SET total_url=?,total_path=? WHERE app_key=? AND env=? AND cvid=? AND cv=?`
	_configPublishCount       = `SELECT count(*) FROM config_publish WHERE app_key=? AND env=?`
	_configPublishCountByCvid = `SELECT count(*) FROM config_publish WHERE app_key=? AND env=? AND cvid=?`

	_configFile = `SELECT app_key,env,cv,cvid,cgroup,ckey,cvalue,state,operator,description,unix_timestamp(mtime) 
FROM config_file WHERE app_key=? AND env=? AND cv=?`
	_addConfigFile    = `INSERT INTO config_file (app_key,env,cvid,cv,cgroup,ckey,cvalue,state,operator,description) VALUES %v`
	_delAllConfigFile = `DELETE FROM config_file WHERE cvid=?`

	// Config 单个配置的发布历史
	_configKeyPublishHistory = `SELECT cf.app_key,cf.env,cf.cvid,cf.cv,cf.cgroup,cf.ckey,cf.cvalue,cf.state,cf.operator,cf.description,unix_timestamp(cf.mtime) FROM config_file AS cf, config_publish AS cp WHERE cf.app_key=cp.app_key AND cf.env=cp.env AND cf.cv=cp.cv AND cp.app_key=? AND cp.env=? AND cp.cvid=? AND cf.ckey=? AND cf.cgroup=? ORDER BY cp.id DESC`
)

// ExistConfigVersion if exist app config version.
func (d *Dao) ExistConfigVersion(c context.Context, appKey, env, version string, versionCode int64) (count int, err error) {
	row := d.db.QueryRow(c, _existConfigVersion, appKey, env, version, versionCode)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("ExistConfigVersion %v", err)
		}
	}
	return
}

// ConfigVersionIDs get app config cvid.
func (d *Dao) ConfigVersionIDs(c context.Context, appKey, env, version string, versionCode int64) (cvids []int64, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey, env)
	if version != "*" {
		args = append(args, version, versionCode)
		sqlAdd = "AND version=? AND version_code=?"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_configVersionID, sqlAdd), args...)
	if err != nil {
		log.Error("ConfigVersionIDs %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cvid int64
		if err = rows.Scan(&cvid); err != nil {
			log.Error("ConfigVersionIDs %v", err)
			return
		}
		cvids = append(cvids, cvid)
	}
	err = rows.Err()
	return
}

// ConfigVersionByID get config version by id.
func (d *Dao) ConfigVersionByID(c context.Context, cvid int64) (res *confmdl.Version, err error) {
	row := d.db.QueryRow(c, _configVersionByID, cvid)
	res = &confmdl.Version{}
	if err = row.Scan(&res.ID, &res.AppKey, &res.Env, &res.Version, &res.VersionCode, &res.Desc); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("ConfigVersionByID %v", err)
		}
	}
	return
}

// ConfigVersionByIDs get config version by ids.
func (d *Dao) ConfigVersionByIDs(c context.Context, cvids []int64) (res map[int64]*confmdl.Version, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, cvid := range cvids {
		sqls = append(sqls, "?")
		args = append(args, cvid)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_configVersionByIDs, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("ConfigVersionByIDs %v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*confmdl.Version)
	for rows.Next() {
		re := &confmdl.Version{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Env, &re.Version, &re.VersionCode); err != nil {
			log.Error("ConfigVersionByIDs %v", err)
			return
		}
		res[re.ID] = re
	}
	err = rows.Err()
	return
}

// TxUpConfigVersionDesc update config version desc where config save.
func (d *Dao) TxUpConfigVersionDesc(tx *xsql.Tx, cvid int64, desc string) (r int64, err error) {
	res, err := tx.Exec(_upConfigVersionDesc, desc, cvid)
	if err != nil {
		log.Error("TxUpConfigVersionDesc %v", err)
		return
	}
	return res.RowsAffected()
}

// TxSetConfigVersion set new config.
func (d *Dao) TxSetConfigVersion(tx *xsql.Tx, appKey, env, version string, versionCode int64, userName string) (r int64, err error) {
	res, err := tx.Exec(_setConfigVersion, appKey, env, version, versionCode, userName)
	if err != nil {
		log.Error("TxSetConfigVersion %v", err)
		return
	}
	return res.LastInsertId()
}

// ConfigVersionCount get app config count.
func (d *Dao) ConfigVersionCount(c context.Context, appKey, env string) (count int, err error) {
	row := d.db.QueryRow(c, _configVersionCount, appKey, env)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("ConfigVersionCount %v", err)
		}
	}
	return
}

// ConfigVersionList get config version list.
func (d *Dao) ConfigVersionList(c context.Context, appKey, env string, pn, ps int) (res []*confmdl.Version, err error) {
	rows, err := d.db.Query(c, _configVersionList, appKey, env, (pn-1)*ps, ps)
	if err != nil {
		log.Error("ConfigVersionList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &confmdl.Version{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Env, &re.Version, &re.VersionCode); err != nil {
			log.Error("ConfigVersionList %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// GetDefaultConfigVersion get default version.
func (d *Dao) GetDefaultConfigVersion(c context.Context, appKey, env string) (res []*confmdl.Version, err error) {
	rows, err := d.db.Query(c, _configDefaultVersion, appKey, env)
	if err != nil {
		log.Error("GetDefaultConfigVersion %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &confmdl.Version{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Env, &re.Version, &re.VersionCode); err != nil {
			log.Error("GetDefaultConfigVersion %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// ConfigModifyCountsAll get app config modify count.
func (d *Dao) ConfigModifyCountsAll(c context.Context, appKey, env string) (count int, err error) {
	row := d.db.QueryRow(c, _configModifyCountAll, appKey, env)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("_configModifyCountAll %v", err)
		}
	}
	return
}

// ConfigModifyCounts get app config modify counts.
func (d *Dao) ConfigModifyCounts(c context.Context, appKey, env string, cvids []int64) (counts map[int64]int, err error) {
	var (
		sqls []string
		args []interface{}
	)
	args = append(args, appKey, env)
	for _, cvid := range cvids {
		sqls = append(sqls, "?")
		args = append(args, cvid)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_configModifyCount, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("ConfigModifyCounts %v", err)
		return
	}
	defer rows.Close()
	counts = make(map[int64]int)
	for rows.Next() {
		var (
			count int
			cvID  int64
		)
		if err = rows.Scan(&cvID, &count); err != nil {
			log.Error("ConfigModifyCounts %v", err)
			return
		}
		counts[cvID] = count
	}
	err = rows.Err()
	return
}

// ConfigPublishs get app config publish .
func (d *Dao) ConfigPublishs(c context.Context, appKey, env string, cvids []int64) (res map[int64][]*confmdl.Publish, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_configPublishs, xstr.JoinInts(cvids)), appKey, env)
	if err != nil {
		log.Error("ConfigPublishs %v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64][]*confmdl.Publish)
	for rows.Next() {
		re := &confmdl.Publish{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Env, &re.CVID, &re.CV, &re.State, &re.MD5, &re.URL, &re.LocalPath, &re.Diffs,
			&re.TotalURL, &re.TotalLocalPath, &re.Operator, &re.Desc, &re.CTime, &re.PTime); err != nil {
			log.Error("ConfigPublishs %v", err)
			return
		}
		res[re.CVID] = append(res[re.CVID], re)
	}
	err = rows.Err()
	return
}

// TxDelConfigVersion del app config.
func (d *Dao) TxDelConfigVersion(tx *xsql.Tx, id int64) (r int64, err error) {
	res, err := tx.Exec(_delConfigVersionState, id)
	if err != nil {
		log.Error("TxDelConfigVersion %v", err)
		return
	}
	return res.RowsAffected()
}

// TxAddConfig add app config.
func (d *Dao) TxAddConfig(tx *xsql.Tx, appKey, env string, cvid int64, group, key, value, userName, description string) (r int64, err error) {
	res, err := tx.Exec(_addConfig, appKey, env, cvid, group, key, value, userName, description, value, userName, description)
	if err != nil {
		log.Error("TxSetConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpConfig update app config.
func (d *Dao) TxUpConfig(tx *xsql.Tx, appKey, env string, cvid int64, group, key, value, userName, description string) (r int64, err error) {
	res, err := tx.Exec(_upConfig, value, userName, description, appKey, env, cvid, group, key)
	if err != nil {
		log.Error("TxUpConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// TxDelConfig del app config.
func (d *Dao) TxDelConfig(tx *xsql.Tx, appKey, env string, cvid int64, group, key, userName, description string) (r int64, err error) {
	res, err := tx.Exec(_delConfig, userName, description, appKey, env, cvid, group, key)
	if err != nil {
		log.Error("TxDelConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// ConfigPublish get app config publish.
func (d *Dao) ConfigPublish(c context.Context, appKey, env string, cvid int64, pn, ps int) (res []*confmdl.Publish, err error) {
	var (
		sqlLimit string
		args     []interface{}
	)
	args = append(args, appKey, env, cvid)
	if pn != -1 && ps != -1 {
		args = append(args, (pn-1)*ps, ps)
		sqlLimit += "LIMIT ?,?"
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_configPublish, sqlLimit), args...)
	if err != nil {
		log.Error("ConfigPublish %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &confmdl.Publish{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Env, &re.CVID, &re.CV, &re.State, &re.MD5, &re.URL, &re.LocalPath, &re.Diffs,
			&re.TotalURL, &re.TotalLocalPath, &re.Operator, &re.Desc, &re.CTime, &re.PTime); err != nil {
			log.Error("ConfigPublish %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// AppConfigVersionHistoryByID get config history by id.
func (d *Dao) AppConfigVersionHistoryByID(c context.Context, appKey string, cid int64) (re *confmdl.Publish, err error) {
	row := d.db.QueryRow(c, _configPublishByID, appKey, cid)
	re = &confmdl.Publish{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.Env, &re.CVID, &re.CV, &re.State, &re.MD5, &re.URL, &re.LocalPath, &re.Diffs,
		&re.TotalURL, &re.TotalLocalPath, &re.Operator, &re.Desc, &re.CTime, &re.PTime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("AppConfigVersionHistoryByID %v", err)
		}
	}
	return
}

// ConfigPublishAll get all config publish.
func (d *Dao) ConfigPublishAll(c context.Context, appKey, env string, pn, ps int) (res []*confmdl.Publish, err error) {
	rows, err := d.db.Query(c, _configPublishAll, appKey, env, (pn-1)*ps, ps)
	if err != nil {
		log.Error("ConfigPublishAll %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &confmdl.Publish{}
		if err = rows.Scan(&re.AppKey, &re.Env, &re.CVID, &re.Version, &re.VersionCode, &re.CV, &re.State, &re.MD5,
			&re.URL, &re.LocalPath, &re.Diffs, &re.TotalURL, &re.TotalLocalPath, &re.Operator, &re.Desc, &re.CTime, &re.PTime); err != nil {
			log.Error("ConfigPublishAll %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// Config get app config.
func (d *Dao) Config(c context.Context, appKey, env string, vid int64) (res []*confmdl.Config, err error) {
	rows, err := d.db.Query(c, _config, appKey, env, vid)
	if err != nil {
		log.Error("Config %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &confmdl.Config{}
		if err = rows.Scan(&re.AppKey, &re.Env, &re.CVID, &re.Group, &re.Key, &re.Value, &re.State, &re.Operator, &re.Desc, &re.MTime); err != nil {
			log.Error("Config %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// ConfigFile get app config filt.
func (d *Dao) ConfigFile(c context.Context, appKey, env string, cv int64) (res []*confmdl.Config, err error) {
	rows, err := d.db.Query(c, _configFile, appKey, env, cv)
	if err != nil {
		log.Error("ConfigFile %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &confmdl.Config{}
		if err = rows.Scan(&re.AppKey, &re.Env, &re.CV, &re.CVID, &re.Group, &re.Key, &re.Value, &re.State, &re.Operator, &re.Desc, &re.MTime); err != nil {
			log.Error("ConfigFile %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// ConfigLastCV get last publish cv.
func (d *Dao) ConfigLastCV(c context.Context, appKey, env string, cvid, cv int64) (lastCV int64, err error) {
	row := d.db.QueryRow(c, _configLastCV, appKey, env, cvid, cv)
	if err = row.Scan(&lastCV); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("ConfigLastCV %v", err)
		}
	}
	return
}

// TxAddConfigPublish add app config publish.
func (d *Dao) TxAddConfigPublish(tx *xsql.Tx, appKey, env string, cvid, cv int64, desc, userName string) (id int64, err error) {
	res, err := tx.Exec(_addConfigPublish, appKey, env, cvid, cv, desc, userName)
	if err != nil {
		log.Error("TxAddConfigPublish %v", err)
		return
	}
	return res.LastInsertId()
}

// TxAddConfigFile add app config file.
func (d *Dao) TxAddConfigFile(tx *xsql.Tx, sqls []string, args []interface{}) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_addConfigFile, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TxAddConfigFile %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpConfigPublishFiles update config publish fils.
func (d *Dao) TxUpConfigPublishFiles(tx *xsql.Tx, appKey, env string, id int64, md5, url, localPath, diffs string) (r int64, err error) {
	res, err := tx.Exec(_upConfigPublishFiles, md5, url, localPath, diffs, appKey, env, id)
	if err != nil {
		log.Error("TxUpConfigPublishFiles %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpConfigPublishState update config publish state.
func (d *Dao) TxUpConfigPublishState(tx *xsql.Tx, appKey, env string, cvid, cv int64, state int) (r int64, err error) {
	res, err := tx.Exec(_upConfigPublishState, state, appKey, env, cvid, cv)
	if err != nil {
		log.Error("TxUpConfigPublishState %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpConfigState update config state.
func (d *Dao) TxUpConfigState(tx *xsql.Tx, appKey, env string, cvid int64, state int) (r int64, err error) {
	res, err := tx.Exec(_upConfigState, state, appKey, env, cvid)
	if err != nil {
		log.Error("TxUpConfigState %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpConfigStateMultiple update config state.
func (d *Dao) TxUpConfigStateMultiple(tx *xsql.Tx, filterSql []string, filterArg []interface{}, state int) (r int64, err error) {
	var args []interface{}
	args = append(args, state)
	args = append(args, filterArg...)
	res, err := tx.Exec(fmt.Sprintf(_upConfigStateMultiple, strings.Join(filterSql, " OR ")), args...)
	if err != nil {
		log.Error("TxUpConfigState %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpConfigItemState update config items status.
func (d *Dao) TxUpConfigItemState(tx *xsql.Tx, appKey, env string, cvid int64, group, key string, state int) (r int64, err error) {
	res, err := tx.Exec(_upConfigItemState, state, appKey, env, cvid, group, key)
	if err != nil {
		log.Error("TxUpConfigState %v", err)
		return
	}
	return res.RowsAffected()
}

// TxDelConfig2 delete config.
func (d *Dao) TxDelConfig2(tx *xsql.Tx, appKey, env string, cvid int64, group, key string) (r int64, err error) {
	res, err := tx.Exec(_delConfig2, appKey, env, cvid, group, key)
	if err != nil {
		log.Error("TxDelConfig2 %v", err)
		return
	}
	return res.RowsAffected()
}

// TxFlushConfig delete all state -1.
func (d *Dao) TxFlushConfig(tx *xsql.Tx, appKey, env string, cvid int64) (r int64, err error) {
	res, err := tx.Exec(_flushConfig, appKey, env, cvid)
	if err != nil {
		log.Error("TxFlushConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// TxFlushConfigMultiple delete all state -1.
func (d *Dao) TxFlushConfigMultiple(tx *xsql.Tx, filterSql []string, filterArg []interface{}) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_flushConfigMultiple, strings.Join(filterSql, " OR ")), filterArg...)
	if err != nil {
		log.Error("TxFlushConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// TxAddConfigs add app configs.
func (d *Dao) TxAddConfigs(tx *xsql.Tx, sqls []string, args []interface{}) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_addConfigs, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TxAddConfigs %v", err)
		return
	}
	return res.RowsAffected()
}

// TxDelAllConfig del all configs.
func (d *Dao) TxDelAllConfig(tx *xsql.Tx, cvid int64) (r int64, err error) {
	res, err := tx.Exec(_delAllConfig, cvid)
	if err != nil {
		log.Error("TxDelAllConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpConfigDesc update config desc.
func (d *Dao) TxUpConfigDesc(tx *xsql.Tx, appKey, env string, cvid int64, desc string, group string, key string) (r int64, err error) {
	res, err := tx.Exec(_upConfigDesc, desc, appKey, env, cvid, group, key)
	if err != nil {
		log.Error("TxUpConfigDesc %v", err)
		return
	}
	return res.RowsAffected()
}

// TxDelAllConfigPublish del all configs_publish.
func (d *Dao) TxDelAllConfigPublish(tx *xsql.Tx, cvid int64) (r int64, err error) {
	res, err := tx.Exec(_delAllConfigPublish, cvid)
	if err != nil {
		log.Error("TxDelAllConfigPublish %v", err)
		return
	}
	return res.RowsAffected()
}

// TxDelAllConfigFile del all config_file.
func (d *Dao) TxDelAllConfigFile(tx *xsql.Tx, cvid int64) (r int64, err error) {
	res, err := tx.Exec(_delAllConfigFile, cvid)
	if err != nil {
		log.Error("TxDelAllConfigFile %v", err)
		return
	}
	return res.RowsAffected()
}

// AllNewConfigPublish get app config publish all newest.
func (d *Dao) AllNewConfigPublish(c context.Context, appKey, env string) (res map[int64]*confmdl.Publish, err error) {
	rows, err := d.db.Query(c, _allNewConfigPublish, appKey, env)
	if err != nil {
		log.Error("AllNewConfigPublish %v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*confmdl.Publish)
	idm := make(map[int64]int64)
	for rows.Next() {
		re := &confmdl.Publish{}
		if err = rows.Scan(&re.AppKey, &re.Env, &re.CVID, &re.CV, &re.MD5, &re.URL, &re.LocalPath, &re.Diffs, &re.TotalURL,
			&re.TotalLocalPath, &re.Operator, &re.Desc, &re.PTime); err != nil {
			log.Error("AllNewConfigPublish %v", err)
			return
		}
		if _, ok := idm[re.CVID]; ok {
			continue
		}
		res[re.CVID] = re
		idm[re.CVID] = re.CVID
	}
	err = rows.Err()
	return
}

// TxUpConfigPublishTotal up config publish total.
func (d *Dao) TxUpConfigPublishTotal(tx *xsql.Tx, appKey, env string, cvid, cv int64, filePath, url string) (r int64, err error) {
	res, err := tx.Exec(_upConfigPublishTotal, url, filePath, appKey, env, cvid, cv)
	if err != nil {
		log.Error("TxUpConfigPublishTotal %v", err)
		return
	}
	return res.RowsAffected()
}

// ConfigPublishCount get app config publish count.
func (d *Dao) ConfigPublishCount(c context.Context, appKey, env string) (count int, err error) {
	row := d.db.QueryRow(c, _configPublishCount, appKey, env)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("ConfigPublishCount %v", err)
		}
	}
	return
}

// ConfigPublishCount get app config publish count by cvid.
func (d *Dao) ConfigPublishCountByCvid(c context.Context, appKey, env string, cvid int64) (count int, err error) {
	row := d.db.QueryRow(c, _configPublishCountByCvid, appKey, env, cvid)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("ConfigPublishCountByCvid %v", err)
		}
	}
	return
}

// ConfigGroup get app config group.
func (d *Dao) ConfigGroup(c context.Context, appKey, env, business string, vid int64) (res []*confmdl.Config, err error) {
	rows, err := d.db.Query(c, _configGroup, appKey, env, vid, business)
	if err != nil {
		log.Error("ConfigGroup %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &confmdl.Config{}
		if err = rows.Scan(&re.AppKey, &re.Env, &re.CVID, &re.Group, &re.Key, &re.Value, &re.State, &re.Operator, &re.Desc, &re.MTime); err != nil {
			log.Error("ConfigGroup %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// ConfigItem get app config row by cgroup ckey cvid.
func (d *Dao) ConfigItem(c context.Context, appKey, env, cgroup, ckey string, cvid int64) (re *confmdl.Config, err error) {
	row := d.db.QueryRow(c, _configItem, appKey, env, cvid, cgroup, ckey)
	re = &confmdl.Config{}
	if err = row.Scan(&re.AppKey, &re.Env, &re.CVID, &re.Group, &re.Key, &re.Value, &re.State, &re.Operator, &re.Desc, &re.MTime); err != nil {
		if err == xsql.ErrNoRows {
			re = nil
			err = nil
		} else {
			log.Error("ConfigItem %v", err)
		}
	}
	return
}

// ConfigKeyPublishHistory config key publish history
func (d *Dao) ConfigKeyPublishHistory(c context.Context, appKey, env, ckey, cgroup string, cvid int64) (res []*confmdl.Config, err error) {
	rows, err := d.db.Query(c, _configKeyPublishHistory, appKey, env, cvid, ckey, cgroup)
	if err != nil {
		log.Error("ConfigKeyPublishHistory %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &confmdl.Config{}
		if err = rows.Scan(&re.AppKey, &re.Env, &re.CVID, &re.CV, &re.Group, &re.Key, &re.Value, &re.State, &re.Operator, &re.Desc, &re.MTime); err != nil {
			log.Error("ConfigKeyPublishHistory %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}
