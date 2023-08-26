package fawkes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/database/xsql"

	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_buildPacks = `SELECT id,app_id,git_path,app_key,gl_prj_id,gl_job_id,git_type,git_name,commit,pkg_type,
version,version_code,internal_version_code,operator,size,md5,pkg_path,pkg_url,mapping_url,bbr_url,r_url,r_mapping_url,status,did_push,change_log,notify_group,unix_timestamp(ctime),unix_timestamp(mtime),test_status,task_ids,ci_env_vars,description,build_start_time,build_end_time,dep_gl_job_id,is_compatible,webhook_url,features,send FROM build_pack WHERE app_key=? AND state=0 %s`
	_buildPackById = `SELECT id,app_id,git_path,app_key,gl_prj_id,gl_job_id,git_type,git_name,commit,pkg_type,
version,version_code,internal_version_code,operator,size,md5,pkg_path,pkg_url,mapping_url,bbr_url,r_url,r_mapping_url,status,did_push,change_log,notify_group,unix_timestamp(ctime),unix_timestamp(mtime),test_status,task_ids,ci_env_vars,description,build_start_time,build_end_time,dep_gl_job_id,is_compatible,webhook_url,features,send FROM build_pack WHERE id=? AND state=0`
	_buildPacksCount         = `SELECT count(*) FROM build_pack WHERE state=0 AND app_key=? %s`
	_buildPacksShouldRefresh = `SELECT bp.id, bp.app_id, bp.git_path, bp.app_key, bp.gl_prj_id, bp.gl_job_id, bp.git_type, bp.git_name, bp.commit, bp.pkg_type, bp.version, bp.version_code, bp.internal_version_code, bp.operator, bp.size, bp.md5, bp.pkg_path, bp.pkg_url, bp.mapping_url, bp.r_url, bp.r_mapping_url, bp.status, bp.did_push, bp.change_log, bp.notify_group, unix_timestamp(bp.ctime), unix_timestamp(bp.mtime) FROM ( SELECT id FROM build_pack ORDER BY id DESC LIMIT 5000 ) res, build_pack bp WHERE res.id = bp.id AND bp.state = 0 AND bp.status IN (1, 2) AND bp.gl_job_id != 0`
	_buildPacksByJobIds      = "SELECT id,app_key,gl_job_id,ci_env_vars FROM build_pack WHERE app_key=? AND gl_job_id IN (%s)"
	_insertBuildPack         = `INSERT INTO build_pack (app_id,git_path,app_key,gl_prj_id,gl_job_id,git_type,git_name,
commit,pkg_type,version,version_code,internal_version_code,operator,status) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,2)`
	_insertBuildPackCreate = `INSERT INTO build_pack (app_id,git_path,app_key,gl_prj_id,git_type,git_name,pkg_type,
operator,ci_env_vars,description,webhook_url,notify_group,dep_gl_job_id,send) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_updateBuildPackBaseInfo = `UPDATE build_pack SET gl_job_id=?,commit=?,version=?,version_code=?,internal_version_code=?,
status=2,build_start_time=? WHERE id=? AND state=0`
	_updateBuildPackInfo = `UPDATE build_pack SET size=?,md5=?,pkg_path=?,pkg_url=?,mapping_url=?,r_url=?,r_mapping_url=?,bbr_url=?,
change_log=?,features=?,status=3,build_end_time=?,is_compatible=? WHERE id=? AND state=0`
	_updateBuildPackStatus      = `UPDATE build_pack SET status=? WHERE id=?`
	_delBuild                   = `UPDATE build_pack SET state=1 WHERE id=?`
	_buildPackDidPush           = `UPDATE build_pack SET did_push=1 WHERE id=?`
	_updateTestStatus           = `UPDATE build_pack SET test_status=?,task_ids=? WHERE gl_job_id=? AND state=0`
	_buildPackSubRepoCommits    = `SELECT id,app_key,build_id,gl_job_id,repo_name,commit,unix_timestamp(ctime),unix_timestamp(mtime) FROM build_pack_subrepo WHERE app_key=? AND gl_job_id=?`
	_insertSubrepoCommits       = `INSERT INTO build_pack_subrepo (app_key,build_id,gl_job_id,repo_name,commit) VALUES %s`
	_selectBuildPackVersionInfo = "SELECT gl_job_id,version_code,version,did_push FROM build_pack WHERE app_key=? AND state=? AND gl_job_id IN (%s)"

	// build_pack_subrepo
	_buildPackSubRepoList         = `SELECT id,app_key,build_id,gl_job_id,repo_name,commit,ctime,mtime FROM build_pack_subrepo WHERE app_key=? AND build_id=?`
	_buildPackSubRepoListByCommit = `SELECT id,app_key,build_id,gl_job_id,repo_name,commit,ctime,mtime FROM build_pack_subrepo WHERE app_key=? AND commit=?`

	// crontab
	_ciCrontabCount     = `SELECT count(1) FROM crontab_ci WHERE app_key=?`
	_ciCrontabAll       = `SELECT id,app_key,unix_timestamp(stime),tick,git_type,git_name,pkg_type,build_id,ci_env_vars,send,state,operator FROM crontab_ci WHERE state IN(-1,0,1) AND stime<=?`
	_ciCrontab          = `SELECT id,app_key,unix_timestamp(stime),tick,git_type,git_name,pkg_type,build_id,ci_env_vars,send,state,operator FROM crontab_ci WHERE app_key=? AND build_id=?`
	_ciCrontabList      = `SELECT id,app_key,unix_timestamp(stime),tick,git_type,git_name,pkg_type,build_id,ci_env_vars,send,state,operator FROM crontab_ci WHERE app_key=? LIMIT ?,?`
	_addCiCrontab       = `INSERT INTO crontab_ci(app_key,stime,tick,git_type,git_name,pkg_type,send,ci_env_vars,operator) VALUES(?,?,?,?,?,?,?,?,?)`
	_upCiCrontabBuildID = `UPDATE crontab_ci SET build_id=? WHERE id=?`
	_upCiCrontabStatus  = `UPDATE crontab_ci SET state=? WHERE id=?`
	_delCiCrontabStatus = "DELETE FROM crontab_ci WHERE id=?"
	// for ep
	_packByID = `SELECT id,app_id,git_path,app_key,gl_prj_id,gl_job_id,git_type,git_name,commit,pkg_type,
version,version_code,internal_version_code,operator,size,md5,pkg_path,pkg_url,mapping_url,r_url,r_mapping_url,bbr_url,status,did_push,change_log,notify_group,unix_timestamp(ctime),unix_timestamp(mtime),test_status,task_ids,ci_env_vars,description,build_start_time,build_end_time FROM build_pack WHERE app_key=? %s`

	_ciEnvList        = `SELECT id,env_key,env_val,env_type,description,is_default,is_global,push_cd_able,platform,app_keys,operator,unix_timestamp(mtime),unix_timestamp(ctime) from ci_env WHERE is_global=1 %s`
	_addCiEnv         = `INSERT INTO ci_env(env_key,env_val,env_type,description,is_default,is_global,push_cd_able,platform,app_keys, operator) VALUES (?,?,?,?,?,?,?,?,?,?)`
	_upCiEnv          = `UPDATE ci_env SET env_key=?,env_val=?,env_type=?,description=?,is_default=?,is_global=?,push_cd_able=?,platform=?,app_keys=?,operator=? WHERE id=?`
	_delCiEnv         = `DELETE FROM ci_env WHERE %s`
	_delCiEnvByAppKey = `UPDATE ci_env SET app_keys=?, operator=? WHERE env_key=?`

	_monkeyTestList     = `SELECT id,app_key,build_id,osver,status,exec_duration,scheme_url,log_url,play_url,message_to,operator,unix_timestamp(ctime),unix_timestamp(mtime) FROM monkey_test WHERE app_key=? AND build_id=? LIMIT ?,?`
	_inMonkeyTest       = `INSERT INTO monkey_test(app_key,build_id,osver,status,exec_duration,scheme_url,message_to,operator) VALUES(?,?,?,?,?,?,?,?)`
	_updateMonkeyStatus = `UPDATE monkey_test SET status=?,log_url=?,play_url=? WHERE app_key=? AND id=?`
	_addCIJobRecord     = `INSERT INTO ci_job_time(app_key,build_id,pipeline_id,pkg_version,job_id,job_name,job_status,stage,job_start_time,job_end_time,job_url,tag_list,runner_info) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_ciJobInfo          = `SELECT DISTINCT %s FROM ci_job_time WHERE %s!='' AND DATE_SUB(CURDATE(), INTERVAL 7 DAY) <= ctime;`
	_addCICompileRecord = `INSERT INTO ci_compile_time(app_key,pkg_type,build_env,build_log_url,job_id,status,steps_count,uptodate_count,cache_count,executed_count,operator,start_time,end_time,fast_total,fast_remote,fast_local,after_sync_task,build_source_local,build_source_remote,optimize_level) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	// query specify time
	_ciSpecifyTimeList            = `SELECT id,app_key,pkg_type,unix_timestamp(ctime) FROM build_pack WHERE is_expired=0 %s`
	_updateBuildPackExpiredStatus = `UPDATE build_pack SET is_expired=1 WHERE id IN (%s)`
)

// BuildPacks get the ci list.
func (d *Dao) BuildPacks(c context.Context, appKey string, pn, ps, pkgType, status, gitType int, gitKeyword,
	operator, order, sort string, glJobID, ID int64, didPushCD string, hasBbrUrl bool) (buildPacks []*cimdl.BuildPack, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if pkgType != 0 {
		args = append(args, pkgType)
		sqlAdd += " AND pkg_type=?"
	}
	if status != 0 {
		args = append(args, status)
		sqlAdd += " AND status=?"
	}
	if didPushCD != "" {
		parseInt, _ := strconv.ParseInt(didPushCD, 10, 64)
		args = append(args, parseInt)
		sqlAdd += " AND did_push=?"
	}
	if hasBbrUrl {
		sqlAdd += " AND LENGTH(trim(bbr_url))>0"
	}
	if gitKeyword != "" {
		if gitType != cimdl.GitTypeCommit {
			args = append(args, gitType)
			sqlAdd += " AND git_type=?"
		}
		var (
			tmpSqls string
			tmpArgs []interface{}
			params  = []string{"git_name"}
		)
		if gitType == cimdl.GitTypeCommit {
			params = append(params, "commit")
		}
		tmpSqls, tmpArgs = d.FormLike(gitKeyword, params, "OR")
		args = append(args, tmpArgs...)
		sqlAdd += " AND " + tmpSqls
	}
	if operator != "" {
		args = append(args, operator)
		sqlAdd += " AND operator=?"
	}
	if glJobID != 0 {
		args = append(args, glJobID)
		sqlAdd += " AND gl_job_id=?"
	}
	if ID != 0 {
		args = append(args, ID)
		sqlAdd += " AND id=?"
	}
	if strings.ToLower(order) == "mtime" {
		sqlAdd += " ORDER BY mtime "
	} else {
		sqlAdd += " ORDER BY id "
	}
	if strings.ToLower(sort) == "asc" {
		sqlAdd += "ASC"
	} else {
		sqlAdd += "DESC"
	}

	args = append(args, (pn-1)*ps, ps)
	sqlAdd += " LIMIT ?,?"
	rows, err := d.db.Query(c, fmt.Sprintf(_buildPacks, sqlAdd), args...)
	if err != nil {
		log.Error("d.BuildPacks d.dbQuery(%v,%v,%v) error(%v)", appKey, pn, ps, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var bp = &cimdl.BuildPack{}
		if err = rows.Scan(&bp.BuildID, &bp.AppID, &bp.GitPath, &bp.AppKey, &bp.GitlabProjectID, &bp.GitlabJobID,
			&bp.GitType, &bp.GitName, &bp.Commit, &bp.PkgType, &bp.Version,
			&bp.VersionCode, &bp.InternalVersionCode, &bp.Operator, &bp.Size, &bp.Md5, &bp.PkgPath, &bp.PkgURL,
			&bp.MappingURL, &bp.BbrURL, &bp.RURL, &bp.RMappingURL, &bp.Status, &bp.DidPush, &bp.ChangeLog, &bp.NotifyGroup, &bp.CTime,
			&bp.MTime, &bp.TestStatus, &bp.TaskIds, &bp.EnvVars, &bp.Description, &bp.BuildStartTime, &bp.BuildEndTime, &bp.DepGitlabJobID, &bp.IsCompatible, &bp.WebhookURL, &bp.Features, &bp.Send); err != nil {
			log.Error("d.BuildPacks rows.Scan error(%v)", err)
			return
		}
		buildPacks = append(buildPacks, bp)
	}
	err = rows.Err()
	return
}

// BuildPack get single ci info.
func (d *Dao) BuildPack(c context.Context, appKey string, buildID int64) (bp *cimdl.BuildPack, err error) {
	var sqlAdd string
	sqlAdd += fmt.Sprintf(" AND id=%v", buildID)
	bp = &cimdl.BuildPack{}
	row := d.db.QueryRow(c, fmt.Sprintf(_buildPacks, sqlAdd), appKey)
	if err = row.Scan(&bp.BuildID, &bp.AppID, &bp.GitPath, &bp.AppKey, &bp.GitlabProjectID, &bp.GitlabJobID, &bp.GitType,
		&bp.GitName, &bp.Commit, &bp.PkgType, &bp.Version,
		&bp.VersionCode, &bp.InternalVersionCode, &bp.Operator, &bp.Size, &bp.Md5, &bp.PkgPath, &bp.PkgURL, &bp.MappingURL, &bp.BbrURL,
		&bp.RURL, &bp.RMappingURL, &bp.Status, &bp.DidPush, &bp.ChangeLog, &bp.NotifyGroup, &bp.CTime, &bp.MTime, &bp.TestStatus, &bp.TaskIds,
		&bp.EnvVars, &bp.Description, &bp.BuildStartTime, &bp.BuildEndTime, &bp.DepGitlabJobID, &bp.IsCompatible, &bp.WebhookURL, &bp.Features, &bp.Send); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("d.BuildPackAppKey row.Scan error(%v)", err)
		}
	}
	return
}

// BuildPackByJobId get single ci info.
func (d *Dao) BuildPackByJobId(c context.Context, appKey string, jobId int64) (bp *cimdl.BuildPack, err error) {
	var sqlAdd string
	sqlAdd += fmt.Sprintf(" AND gl_job_id=%v", jobId)
	bp = &cimdl.BuildPack{}
	row := d.db.QueryRow(c, fmt.Sprintf(_buildPacks, sqlAdd), appKey)
	if err = row.Scan(&bp.BuildID, &bp.AppID, &bp.GitPath, &bp.AppKey, &bp.GitlabProjectID, &bp.GitlabJobID, &bp.GitType,
		&bp.GitName, &bp.Commit, &bp.PkgType, &bp.Version,
		&bp.VersionCode, &bp.InternalVersionCode, &bp.Operator, &bp.Size, &bp.Md5, &bp.PkgPath, &bp.PkgURL, &bp.MappingURL, &bp.BbrURL,
		&bp.RURL, &bp.RMappingURL, &bp.Status, &bp.DidPush, &bp.ChangeLog, &bp.NotifyGroup, &bp.CTime, &bp.MTime, &bp.TestStatus, &bp.TaskIds,
		&bp.EnvVars, &bp.Description, &bp.BuildStartTime, &bp.BuildEndTime, &bp.DepGitlabJobID, &bp.IsCompatible, &bp.WebhookURL, &bp.Features, &bp.Send); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("d.BuildPackByJobId row.Scan error(%v)", err)
		}
	}
	return
}

// BuildPackById get build info from build id
func (d *Dao) BuildPackById(c context.Context, id int64) (res *cimdl.BuildPack, err error) {
	res = &cimdl.BuildPack{}
	row := d.db.QueryRow(c, _buildPackById, id)
	if err = row.Scan(&res.BuildID, &res.AppID, &res.GitPath, &res.AppKey, &res.GitlabProjectID, &res.GitlabJobID, &res.GitType,
		&res.GitName, &res.Commit, &res.PkgType, &res.Version,
		&res.VersionCode, &res.InternalVersionCode, &res.Operator, &res.Size, &res.Md5, &res.PkgPath, &res.PkgURL, &res.MappingURL, &res.BbrURL,
		&res.RURL, &res.RMappingURL, &res.Status, &res.DidPush, &res.ChangeLog, &res.NotifyGroup, &res.CTime, &res.MTime, &res.TestStatus, &res.TaskIds,
		&res.EnvVars, &res.Description, &res.BuildStartTime, &res.BuildEndTime, &res.DepGitlabJobID, &res.IsCompatible, &res.WebhookURL, &res.Features, &res.Send); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Errorc(c, "row.Scan error(%v)", err)
		}
	}
	return
}

// BuildPacksCount get the total counts
func (d *Dao) BuildPacksCount(c context.Context, appKey string, pkgType, status, gitType int, gitKeyword, operator string, glJobID, ID int64, didPushCD string, hasBbrUrl bool) (r int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if pkgType != 0 {
		args = append(args, pkgType)
		sqlAdd += " AND pkg_type=?"
	}
	if status != 0 {
		args = append(args, status)
		sqlAdd += " AND status=?"
	}
	if didPushCD != "" {
		parseInt, _ := strconv.ParseInt(didPushCD, 10, 64)
		args = append(args, parseInt)
		sqlAdd += " AND did_push=?"
	}
	if hasBbrUrl {
		sqlAdd += " AND LENGTH(trim(bbr_url))>0"
	}
	if gitKeyword != "" {
		if gitType != cimdl.GitTypeCommit {
			args = append(args, gitType)
			sqlAdd += " AND git_type=?"
		}
		var (
			tmpSqls string
			tmpArgs []interface{}
			params  = []string{"git_name"}
		)
		if gitType == cimdl.GitTypeCommit {
			params = append(params, "commit")
		}
		tmpSqls, tmpArgs = d.FormLike(gitKeyword, params, "OR")
		args = append(args, tmpArgs...)
		sqlAdd += " AND " + tmpSqls
	}
	if operator != "" {
		args = append(args, operator)
		sqlAdd += " AND operator=?"
	}
	if glJobID != 0 {
		args = append(args, glJobID)
		sqlAdd += " AND gl_job_id=?"
	}
	if ID != 0 {
		args = append(args, ID)
		sqlAdd += " AND id=?"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_buildPacksCount, sqlAdd), args...)
	if err = row.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("d.BuildPacksCount row.Scan error(%v)", err)
		}
	}
	return
}

// BuildPacksShouldRefresh get build packs which should be refresh
func (d *Dao) BuildPacksShouldRefresh(c context.Context) (buildPacks []*cimdl.BuildPack, err error) {
	rows, err := d.db.Query(c, _buildPacksShouldRefresh)
	if err != nil {
		log.Error("d.BuildPacks d.dbQuery() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var bp = &cimdl.BuildPack{}
		if err = rows.Scan(&bp.BuildID, &bp.AppID, &bp.GitPath, &bp.AppKey, &bp.GitlabProjectID, &bp.GitlabJobID, &bp.GitType,
			&bp.GitName, &bp.Commit, &bp.PkgType, &bp.Version,
			&bp.VersionCode, &bp.InternalVersionCode, &bp.Operator, &bp.Size, &bp.Md5, &bp.PkgPath, &bp.PkgURL, &bp.MappingURL,
			&bp.RURL, &bp.RMappingURL, &bp.Status, &bp.DidPush, &bp.ChangeLog, &bp.NotifyGroup, &bp.CTime, &bp.MTime); err != nil {
			log.Error("d.BuildPacks rows.Scan error(%v)", err)
			return
		}
		buildPacks = append(buildPacks, bp)
	}
	err = rows.Err()
	return
}

// TxInsertBuildPack record a ci build.
func (d *Dao) TxInsertBuildPack(tx *sql.Tx, appKey, appID, gitPath, gitlabPrjID string, gitlabJobID int64, pkgType,
	gitType int, gitName, commit, version string, versionCode, internalVersionCode int64, operator string) (r int64, err error) {
	res, err := tx.Exec(_insertBuildPack, appID, gitPath, appKey, gitlabPrjID, gitlabJobID, gitType, gitName, commit,
		pkgType, version, versionCode, internalVersionCode, operator)
	if err != nil {
		log.Error("d.TxInsertBuildPack tx.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

// TxInsertBuildPackCreate create a ci build.
func (d *Dao) TxInsertBuildPackCreate(tx *sql.Tx, appKey, appID, gitPath, gitlabPrjID, send string, pkgType, gitType int,
	gitName, operator, envVars, description, webhookURL string, notifyGroup int8, depGitlabJobId int64) (r int64, err error) {
	res, err := tx.Exec(_insertBuildPackCreate, appID, gitPath, appKey, gitlabPrjID, gitType, gitName, pkgType, operator, envVars, description, webhookURL, notifyGroup, depGitlabJobId, send)
	if err != nil {
		log.Error("d.TxInsertBuildPackCreate tx.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

// TxUpdateBuildPackBaseInfo update build pack basic info
func (d *Dao) TxUpdateBuildPackBaseInfo(tx *sql.Tx, buildID int64, gitlabJobID int64, commit, version string, versionCode,
	internalVersionCode int64) (r int64, err error) {
	res, err := tx.Exec(_updateBuildPackBaseInfo, gitlabJobID, commit, version, versionCode, internalVersionCode, time.Now(), buildID)
	if err != nil {
		log.Error("d.TxUpdateBuildPackBaseInfo tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TxUpdateBuildPackInfo update ci build info.
func (d *Dao) TxUpdateBuildPackInfo(tx *sql.Tx, buildID int64, size int64, md5, pkgPath, pkgURL, mappingURL, rURL,
	rMappingURL, bbrUrl, changeLog, features string, isCompatible bool) (r int64, err error) {
	var compatible int8
	if isCompatible {
		compatible = 1
	}
	res, err := tx.Exec(_updateBuildPackInfo, size, md5, pkgPath, pkgURL, mappingURL, rURL, rMappingURL, bbrUrl, changeLog, features, time.Now(), compatible, buildID)
	if err != nil {
		log.Error("d.TxUpdateBuildPackInfo tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TxUpdateBuildPackStatus delete a ci build.
func (d *Dao) TxUpdateBuildPackStatus(tx *sql.Tx, buildID int64, status int) (r int64, err error) {
	res, err := tx.Exec(_updateBuildPackStatus, status, buildID)
	if err != nil {
		log.Error("d.TxUpdateBuildPackStatus tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TxDelBuild delete a ci build.
func (d *Dao) TxDelBuild(tx *sql.Tx, buildID int64) (r int64, err error) {
	res, err := tx.Exec(_delBuild, buildID)
	if err != nil {
		log.Error("d.TxDelBuild tx.Exec error(%v)", err)
		return
	}
	return res.RowsAffected()
}

// TxBuildPackDidPush update did push flag.
func (d *Dao) TxBuildPackDidPush(tx *sql.Tx, buildID int64) (r int64, err error) {
	res, err := tx.Exec(_buildPackDidPush, buildID)
	if err != nil {
		log.Error("d.TxBuildPackDidPush tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TxUpdateTestStatus update build pack basic info
func (d *Dao) TxUpdateTestStatus(tx *sql.Tx, pack *cimdl.BuildPack) (r int64, err error) {
	res, err := tx.Exec(_updateTestStatus, pack.TestStatus, pack.TaskIds, pack.GitlabJobID)
	if err != nil {
		log.Error("d.TxUpdateTestStatus tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// CiCrontabCount count of ci crontab
func (d *Dao) CiCrontabCount(c context.Context, appKey string) (count int, err error) {
	row := d.db.QueryRow(c, _ciCrontabCount, appKey)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		} else {
			log.Error("d.CiCrontabCount row.Scan error(%v)", err)
		}
	}
	return
}

// CiCrontabAll all of ci crontab
func (d *Dao) CiCrontabAll(c context.Context, now time.Time) (res []*cimdl.Contab, err error) {
	rows, err := d.db.Query(c, _ciCrontabAll, now)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &cimdl.Contab{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.STime, &re.Tick, &re.GitType, &re.GitName, &re.PkgType, &re.BuildID, &re.CIEnvVars, &re.Send, &re.State, &re.Operator); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// CiCrontabList list ci crontab
func (d *Dao) CiCrontabList(c context.Context, appKey string, pn, ps int) (res []*cimdl.Contab, err error) {
	rows, err := d.db.Query(c, _ciCrontabList, appKey, ps*(pn-1), ps)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &cimdl.Contab{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.STime, &re.Tick, &re.GitType, &re.GitName, &re.PkgType, &re.BuildID, &re.CIEnvVars, &re.Send, &re.State, &re.Operator); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxAddCiCrontab add a crontab job
func (d *Dao) TxAddCiCrontab(tx *sql.Tx, appKey, stime, tick string, gitType int, gitName string, pkgType int, send, envVars, userName string) (r int64, err error) {
	res, err := tx.Exec(_addCiCrontab, appKey, stime, tick, gitType, gitName, pkgType, send, envVars, userName)
	if err != nil {
		log.Error("d.TxAddCiCrontab tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TxUpStatusCiCrontab update crontab status
func (d *Dao) TxUpStatusCiCrontab(tx *sql.Tx, id int64, status int) (r int64, err error) {
	res, err := tx.Exec(_upCiCrontabStatus, status, id)
	if err != nil {
		log.Error("d.TxUpStatusCiCrontab tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TxDelCiCrontab delete a crontab
func (d *Dao) TxDelCiCrontab(tx *sql.Tx, id int64) (r int64, err error) {
	res, err := tx.Exec(_delCiCrontabStatus, id)
	if err != nil {
		log.Error("d.TxDelCiCrontab tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// CIByBuild get ci info by id
func (d *Dao) CIByBuild(c context.Context, appKey string, buildID, glJobID int64) (res *cimdl.BuildPack, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if buildID != 0 {
		sqlAdd += " AND id=?"
		args = append(args, buildID)
	}
	if glJobID != 0 {
		sqlAdd += " AND gl_job_id=?"
		args = append(args, glJobID)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_packByID, sqlAdd), args...)
	res = &cimdl.BuildPack{}
	if err = row.Scan(&res.BuildID, &res.AppID, &res.GitPath, &res.AppKey, &res.GitlabProjectID, &res.GitlabJobID,
		&res.GitType, &res.GitName, &res.Commit, &res.PkgType, &res.Version,
		&res.VersionCode, &res.InternalVersionCode, &res.Operator, &res.Size, &res.Md5, &res.PkgPath, &res.PkgURL,
		&res.MappingURL, &res.RURL, &res.RMappingURL, &res.BbrURL, &res.Status, &res.DidPush, &res.ChangeLog, &res.NotifyGroup, &res.CTime,
		&res.MTime, &res.TestStatus, &res.TaskIds, &res.EnvVars, &res.Description, &res.BuildStartTime, &res.BuildEndTime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("CIByBuild %v", err)
		}
	}
	return
}

// TxUpBuildIDCiCrontab update build id of a crontab
func (d *Dao) TxUpBuildIDCiCrontab(tx *sql.Tx, id, buildID int64) (r int64, err error) {
	res, err := tx.Exec(_upCiCrontabBuildID, buildID, id)
	if err != nil {
		log.Error("d.TxUpBuildIDCiCrontab tx.Exec error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// CronInfo get info of a crontab by build id
func (d *Dao) CronInfo(c context.Context, appKey string, buildID int64) (re *cimdl.Contab, err error) {
	row := d.db.QueryRow(c, _ciCrontab, appKey, buildID)
	re = &cimdl.Contab{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.STime, &re.Tick, &re.GitType, &re.GitName, &re.PkgType, &re.BuildID, &re.CIEnvVars, &re.Send, &re.State, &re.Operator); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("CronInfo %v", err)
		}
	}
	return
}

func (d *Dao) JenkinsJobMonkey(c context.Context, appKey, pkgUrl, mappingUrl, osver, bundleId, schemeUrl, userName string, execDuration int, monkeyTestId int64) (err error) {
	var (
		req  *http.Request
		resp *http.Response
	)
	body := &cimdl.EPMonkeyRequstBody{}
	body.Ref = strconv.Itoa(int(time.Now().UnixNano()))
	body.MobEnv = "prd"
	body.MobAndroidOS = osver
	body.BundleID = bundleId
	body.AppKey = appKey
	body.ApkURL = pkgUrl
	body.MappingURL = mappingUrl
	body.ExecDuration = strconv.Itoa(execDuration)
	body.MonoDuration = strconv.Itoa(execDuration)
	body.Schemes = schemeUrl
	body.CC = userName
	body.CallbackURL = "http://fawkes.bilibili.co/x/admin/fawkes/app/ci/monkey/update/status"
	body.HookID = strconv.Itoa(int(monkeyTestId))
	bodyString, _ := json.Marshal(body)
	bodyBuffer := bytes.NewBuffer([]byte(bodyString))
	req, err = http.NewRequest("POST", d.c.Ep.MonkeyUrl, bodyBuffer)
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	if err != nil {
		fmt.Println(err)
		return
	}
	basicAuth := strings.Split(d.c.Ep.MonkeyAuth, ":")
	req.SetBasicAuth(basicAuth[0], basicAuth[1])
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Error("JenkinsJobMonkey %v", err)
		return
	}
	defer resp.Body.Close()
	respBody, error := ioutil.ReadAll(resp.Body)
	if error != nil {
		fmt.Println(error)
		return
	}
	log.Error("JenkinsJobMonkey %v", string(respBody))
	return
}

// CiEnvList get ci env list.
func (d *Dao) CiEnvList(c context.Context, envKey, appKey, platform string) (res []*cimdl.BuildEnvs, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if envKey != "" {
		sqlAdd += "AND env_key=? "
		args = append(args, envKey)
	}
	if platform != "" {
		sqlAdd += "OR platform=? "
		args = append(args, platform)
	}
	if appKey != "" {
		sqlAdd += "OR FIND_IN_SET(?, app_keys) "
		args = append(args, appKey)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_ciEnvList, sqlAdd), args...)
	if err != nil {
		log.Error("CiEnvList Dao %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var env = &cimdl.BuildEnvs{}
		if err = rows.Scan(&env.ID, &env.EnvKey, &env.EnvVal, &env.EnvType, &env.Descrition, &env.IsDefault, &env.IsGlobal, &env.PushCDAble, &env.Platform, &env.AppKeys, &env.Operator, &env.Mtime, &env.Ctime); err != nil {
			log.Error("CiEnvList rows.Scan error(%v)", err)
			return
		}
		res = append(res, env)
	}
	err = rows.Err()
	return
}

// TxAddCiEnv add ci env
func (d *Dao) TxAddCiEnv(tx *sql.Tx, envKey, envValue, description, platform, appKeys, username string, envType, isDefault, isGlobal, pushCDAble int) (err error) {
	_, err = tx.Exec(_addCiEnv, envKey, envValue, envType, description, isDefault, isGlobal, pushCDAble, platform, appKeys, username)
	if err != nil {
		log.Error("d.TxAddCiEnv tx.Exec error(%v)", err)
	}
	return
}

// TxUpdateCiEnv update ci env
func (d *Dao) TxUpdateCiEnv(tx *sql.Tx, envKey, envValue, description, platform, appKeys, username string, envType, isDefault, isGlobal, pushCDAble int, id int64) (err error) {
	_, err = tx.Exec(_upCiEnv, envKey, envValue, envType, description, isDefault, isGlobal, pushCDAble, platform, appKeys, username, id)
	if err != nil {
		log.Error("d.TxUpdateCiEnv tx.Exec error(%v)", err)
	}
	return
}

// TxDeleteCiEnv delete ci env
func (d *Dao) TxDeleteCiEnv(tx *sql.Tx, id int64, envKey string) (err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	if envKey != "" {
		sqlAdd = "env_key=?"
		args = append(args, envKey)
	} else {
		sqlAdd = "id=?"
		args = append(args, id)
	}
	_, err = tx.Exec(fmt.Sprintf(_delCiEnv, sqlAdd), args...)
	if err != nil {
		log.Error("d.TxDeleteCiEnv tx.Exec error(%v)", err)
	}
	return
}
func (d *Dao) TxDeleteCiEnvByAppKey(tx *sql.Tx, envKey, appKeys, userName string) (err error) {
	_, err = tx.Exec(_delCiEnvByAppKey, appKeys, userName, envKey)
	if err != nil {
		log.Error("d.TxDeleteCiEnvByAppKey tx.Exec error(%v)", err)
	}
	return
}
func (d *Dao) GetMonkeyList(c context.Context, appKey string, buildId int64, pn, ps int) (res []*cimdl.EPMonkey, err error) {
	rows, err := d.db.Query(c, _monkeyTestList, appKey, buildId, pn-1, ps)
	if err != nil {
		log.Error("d.GetMonkeyList d.dbQuery(%v,%v,%v,%v) error(%v)", appKey, buildId, ps*(pn-1), ps, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var row = &cimdl.EPMonkey{}
		if err = rows.Scan(&row.ID, &row.AppKey, &row.BuildID, &row.OSVer, &row.Status, &row.ExecDuration, &row.SchemeUrl, &row.LogUrl, &row.PlayUrl, &row.MessageTo, &row.Operator, &row.Ctime, &row.Mtime); err != nil {
			log.Error("d.GetMonkeyList rows.Scan error(%v)", err)
			return
		}
		res = append(res, row)
	}
	err = rows.Err()
	return
}

func (d *Dao) TxAddMonkey(tx *sql.Tx, appKey, osver, schemeUrl, messageTo, userName string, buildId int64, status, execDuration int) (r int64, err error) {
	res, err := tx.Exec(_inMonkeyTest, appKey, buildId, osver, status, execDuration, schemeUrl, messageTo, userName)
	if err != nil {
		log.Error("d.TxAddMonkey tx.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

func (d *Dao) UpdateMonkeyStatus(tx *sql.Tx, appKey, logUrl, playUrl string, id int64, status int) (err error) {
	_, err = tx.Exec(_updateMonkeyStatus, status, logUrl, playUrl, appKey, id)
	if err != nil {
		log.Error("d.UpdateMonkeyStatus tx.Exec error(%v)", err)
	}
	return
}

func (d *Dao) TxRecordCIJob(tx *sql.Tx, params *cimdl.JobRecordParam, jobURL string) (err error) {
	_, err = tx.Exec(_addCIJobRecord, params.AppKey, params.BuildID, params.PipelineID, params.PkgVersion,
		params.JobID, params.JobName, params.JobStatus, params.Stage, params.JobStartTime, params.JobEndTime, jobURL, params.TagList, params.RunnerInfo)
	if err != nil {
		log.Error("d.TxRecordCIJob tx.Exec error(%v)", err)
	}
	return
}

func (d *Dao) CIJobInfo(c context.Context, typeName string) (res []string, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_ciJobInfo, typeName, typeName))
	if err != nil {
		log.Error("d.CIJobInfo %s", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cols string
		if err = rows.Scan(&cols); err != nil {
			log.Error("d.CIJobInfo rows.Scan error(%v)", err)
			return
		}
		res = append(res, cols)
	}
	err = rows.Err()
	return
}

func (d *Dao) TxRecordCICompile(tx *sql.Tx, params *cimdl.CICompileRecordParam) (err error) {
	_, err = tx.Exec(_addCICompileRecord, params.AppKey, params.PkgType, params.BuildEnv, params.BuildLogURL,
		params.JobID, params.Status, params.StepsCount, params.UptodateCount, params.CacheCount, params.ExecutedCount, params.Operator,
		params.StartTime, params.EndTime, params.FastTotal, params.FastRemote, params.FastLocal, params.AfterSyncTask, params.BuildSourceLocal, params.BuildSourceRemote, params.OptimizeLevel)
	if err != nil {
		log.Error("d.TxRecordCICompile tx.Exec error(%v)", err)
	}
	return
}

func (d *Dao) CISubRepoCommits(c context.Context, appKey string, piplineId string) (res []*cimdl.BuildPackSubRepo, err error) {
	rows, err := d.db.Query(c, _buildPackSubRepoCommits, appKey, piplineId)
	if err != nil {
		log.Error("d.CISubRepoCommits d.dbQuery error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var row = &cimdl.BuildPackSubRepo{}
		if err = rows.Scan(&row.SubRepoID, &row.AppKey, &row.BuildID, &row.PipelineID, &row.RepoName, &row.Commit, &row.CTime, &row.MTime); err != nil {
			log.Error("d.CISubRepoCommits rows.Scan error(%v)", err)
			return
		}
		res = append(res, row)
	}
	err = rows.Err()
	return
}

func (d *Dao) TXAddCISubRepoCommits(tx *sql.Tx, buildId, jobId int64, appKey string, subrepoCommits []*cimdl.BuildPackSubRepo) (r int64, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, subrepoCommit := range subrepoCommits {
		sqls = append(sqls, "(?,?,?,?,?)")
		args = append(args, appKey, buildId, jobId, subrepoCommit.RepoName, subrepoCommit.Commit)
	}
	res, err := tx.Exec(fmt.Sprintf(_insertSubrepoCommits, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TXAddCISubRepoCommits %v", err)
		return
	}
	return res.RowsAffected()
}

// BuildPackSubRepos get the ci subrepo list.
func (d *Dao) BuildPackSubRepos(c context.Context, appKey string, buildID int64) (res []*cimdl.BuildPackSubRepo, err error) {
	rows, err := d.db.Query(c, _buildPackSubRepoList, appKey, buildID)
	if err != nil {
		log.Error("d.BuildPackSubRepos error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var bp = &cimdl.BuildPackSubRepo{}
		if err = rows.Scan(&bp.SubRepoID, &bp.AppKey, &bp.BuildID, &bp.PipelineID, &bp.RepoName, &bp.Commit, &bp.CTime, &bp.MTime); err != nil {
			log.Error("d.BuildPackSubRepos rows.Scan error(%v)", err)
			return
		}
		res = append(res, bp)
	}
	err = rows.Err()
	return
}

// BuildPackSubReposByCommit get the ci subrepo list.
func (d *Dao) BuildPackSubReposByCommit(c context.Context, appKey, commit string) (res []*cimdl.BuildPackSubRepo, err error) {
	rows, err := d.db.Query(c, _buildPackSubRepoListByCommit, appKey, commit)
	if err != nil {
		log.Error("d.BuildPackSubReposByCommit error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var bp = &cimdl.BuildPackSubRepo{}
		if err = rows.Scan(&bp.SubRepoID, &bp.AppKey, &bp.BuildID, &bp.PipelineID, &bp.RepoName, &bp.Commit, &bp.CTime, &bp.MTime); err != nil {
			log.Error("d.BuildPackSubReposByCommit rows.Scan error(%v)", err)
			return
		}
		res = append(res, bp)
	}
	err = rows.Err()
	return
}

// CINasList 获取指定时间的ci包
func (d *Dao) CINasList(c context.Context, appKey string, pkgTypes []int64, startTime, endTime time.Time) (res []*cimdl.BuildPack, err error) {
	var (
		sqlAdd string
		pkgStr string
		args   []interface{}
	)
	if appKey != "" {
		args = append(args, appKey)
		sqlAdd += " AND app_key = ?"
	}
	for index, pkgType := range pkgTypes {
		if index != 0 {
			pkgStr += ","
		}
		pkgStr += strconv.Itoa(int(pkgType))
	}
	sqlAdd += " AND pkg_type IN (" + pkgStr + ")"
	args = append(args, endTime)
	sqlAdd += " AND ctime <= ?"
	args = append(args, startTime)
	sqlAdd += " AND ctime >= ?"
	rows, err := d.db.Query(c, fmt.Sprintf(_ciSpecifyTimeList, sqlAdd), args...)
	if err != nil {
		log.Error("d.CISpecifyTimeListInfo d.dbQuery(%v) error(%v)", appKey, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var bp = &cimdl.BuildPack{}
		if err = rows.Scan(&bp.BuildID, &bp.AppKey, &bp.PkgType, &bp.CTime); err != nil {
			log.Error("d.CISpecifyTimeListInfo rows.Scan error(%v)", err)
			return
		}
		res = append(res, bp)
	}
	err = rows.Err()
	return
}

// UpdateCIExpiredStatus 更新过期ci包字段
func (d *Dao) UpdateCIExpiredStatus(c context.Context, buildIds []int64) (r int64, err error) {
	var (
		sqlAdd []string
		args   []interface{}
	)
	for _, buildId := range buildIds {
		sqlAdd = append(sqlAdd, "?")
		args = append(args, buildId)
	}
	res, err := d.db.Exec(c, fmt.Sprintf(_updateBuildPackExpiredStatus, strings.Join(sqlAdd, ",")), args...)
	if err != nil {
		log.Error("d.UpdateCIExpiredStatus error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

func (d *Dao) SelectBuildPackVersionInfo(c context.Context, appKey string, state int64, id []int64) (res []*cimdl.BuildPack, err error) {
	if len(id) == 0 {
		return
	}

	var (
		sqls = make([]string, 0, len(id))
		args = make([]interface{}, 0)
	)
	args = append(args, appKey, state)
	for _, v := range id {
		sqls = append(sqls, "?")
		args = append(args, v)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_selectBuildPackVersionInfo, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("SelectBuildPackVersionInfo error: %#v", err)
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Error("ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) SelectBuildPackByJobIds(c context.Context, appKey string, id []int64) (res []*cimdl.BuildPack, err error) {
	if len(id) == 0 {
		return
	}
	var (
		sqls = make([]string, 0, len(id))
		args = make([]interface{}, 0)
	)
	args = append(args, appKey)
	for _, v := range id {
		sqls = append(sqls, "?")
		args = append(args, v)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_buildPacksByJobIds, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("SelectBuildPackVersionInfo error: %#v", err)
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Error("ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	return
}
