package fawkes

import (
	"context"
	bsql "database/sql"
	"fmt"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/database/xsql"

	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	"go-gateway/app/app-svr/fawkes/service/model/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	// tribe
	_addTribe         = "INSERT INTO tribe (app_key,name,c_name,owners,description,no_host, priority,is_build_in) VALUES (?,?,?,?,?,?,?,?)"
	_selectTribeById  = "SELECT id,app_key,name,c_name,owners,description,no_host,priority,is_build_in,mtime,ctime FROM bilibili_fawkes.tribe WHERE id=?"
	_selectTribeByIds = "SELECT id,app_key,name,c_name,owners,description,no_host,priority,mtime,ctime FROM bilibili_fawkes.tribe WHERE id IN (%s)"
	_countTribe       = "SELECT COUNT(*) FROM tribe WHERE %s"
	_selectTribe      = "SELECT id,app_key,name,c_name,owners,description,no_host,priority,is_build_in FROM tribe WHERE %s LIMIT ?,?"
	_updateTribe      = "UPDATE tribe SET %s WHERE id=?"
	_deleteTribe      = "DELETE FROM tribe WHERE id=?"

	// tribe_build_pack
	_addTribeBuildPack            = "INSERT INTO tribe_build_pack (tribe_id,dep_gl_job_id,pkg_type,git_type,app_key,git_name,operator,ci_env_vars,description,notify_group,app_id,git_path,gl_prj_id,status) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	_updateTribeBuildPackDepJobId = "UPDATE tribe_build_pack SET dep_gl_job_id=? AND status=? WHERE id=?"
	_updateTribeBuildPack         = "UPDATE tribe_build_pack SET %s WHERE id=?"
	_updateTribeBuildPackPkgInfo  = "UPDATE tribe_build_pack SET %s WHERE id=?"
	_tribeBuildPackDidPush        = `UPDATE tribe_build_pack SET did_push=? WHERE id=?`
	_updateTribeStatus            = "UPDATE tribe_build_pack SET status=?,err_msg=? WHERE id=?"
	_updateTribeMavenStatus       = "UPDATE tribe_build_pack SET push_maven=? WHERE id=?"
	_tribeBuildPacksSelect        = "SELECT id,tribe_id,gl_job_id,dep_gl_job_id,dep_feature,app_id,git_path,app_key,git_type,git_name,commit,pkg_type,operator,size,md5,pkg_path,pkg_url,mapping_url,bbr_url,state,status,did_push,change_log,notify_group,ci_env_vars,build_start_time,build_end_time,description,err_msg,mtime,ctime,version_code,version_name FROM tribe_build_pack WHERE %s %s LIMIT ?,?"
	_selectTribeBuildPackById     = "SELECT id,tribe_id,gl_job_id,dep_gl_job_id,dep_feature,app_id,git_path,app_key,git_type,git_name,commit,pkg_type,operator,size,md5,pkg_path,pkg_url,mapping_url,version_code,version_name,bbr_url,state,status,did_push,change_log,notify_group,ci_env_vars,build_start_time,build_end_time,description,err_msg,mtime,ctime FROM tribe_build_pack WHERE id=?"
	_selectTribeBuildPackByIds    = "SELECT id,tribe_id,gl_job_id,dep_gl_job_id,dep_feature,app_id,git_path,app_key,git_type,git_name,commit,pkg_type,operator,size,md5,pkg_path,pkg_url,mapping_url,bbr_url,state,status,did_push,change_log,notify_group,ci_env_vars,build_start_time,build_end_time,description,err_msg,mtime,ctime FROM tribe_build_pack WHERE id IN (%s)"
	_countTribeBuildPacks         = "SELECT count(*) FROM tribe_build_pack WHERE %s"

	//tribe_pack
	_AddTribePack               = "INSERT INTO tribe_pack (app_id,app_key,env,tribe_id,gl_job_id,dep_gl_job_id,dep_feature,version_id,git_type,git_name,commit,pack_type,change_log,operator,size,md5,pack_path,pack_url,mapping_url,bbr_url,cdn_url,description,sender) VALUE (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	_countTribePacks            = "SELECT count(*) FROM tribe_pack WHERE %s"
	_tribePacksSelectById       = "SELECT id,app_id,app_key,env,tribe_id,gl_job_id,dep_gl_job_id,dep_feature,version_id,git_type,git_name,commit,pack_type,change_log,operator,size,md5,pack_path,pack_url,mapping_url,bbr_url,cdn_url,description,sender,mtime,ctime FROM tribe_pack WHERE id=?"
	_tribePacksSelect           = "SELECT id,app_id,app_key,env,tribe_id,gl_job_id,dep_gl_job_id,dep_feature,version_id,version_code,version_name,git_type,git_name,commit,pack_type,change_log,operator,size,md5,pack_path,pack_url,mapping_url,bbr_url,cdn_url,description,sender,mtime,ctime FROM tribe_pack WHERE %s LIMIT ?,?"
	_listPackVersionByOptions   = "SELECT tpv.id,tpv.tribe_id,tpv.app_id,tpv.env,tpv.version_code,tpv.is_active,tpv.ctime,tpv.mtime,tpv.version_name FROM tribe_pack AS tp, tribe_pack_version AS tpv WHERE tpv.tribe_id=? AND tpv.env=? AND tp.version_id=tpv.id %s"
	_tribePacksSelectByVersions = "SELECT id,app_id,app_key,env,tribe_id,gl_job_id,dep_gl_job_id,dep_feature,version_id,git_type,git_name,commit,pack_type,change_log,operator,size,md5,pack_path,pack_url,mapping_url,bbr_url,cdn_url,description,sender,mtime,ctime FROM tribe_pack WHERE tribe_id=? AND env=? AND version_id IN (%s)"
	_countPackVersionByOptions  = "SELECT count(*) FROM tribe_pack AS tp, tribe_pack_version AS tpv WHERE tpv.tribe_id=? AND tpv.env=? AND tp.version_id=tpv.id"
	_updateTribePackPkgInfo     = "UPDATE tribe_pack SET %s WHERE tribe_id=? AND gl_job_id=?"

	// tribe_pack_version
	_setTribePackVersion             = "INSERT INTO tribe_pack_version (tribe_id,env,version_code,version_name,is_active) VALUE (?,?,?,?,?)"
	_selectTribePackVersionByArgs    = "SELECT id,tribe_id,app_id,env,version_code,version_name,is_active,ctime,mtime,operator FROM tribe_pack_version WHERE tribe_id=? AND env=? AND version_code=?" // unique idx
	_selectTribePackVersionById      = "SELECT id,tribe_id,app_id,env,version_code,version_name,is_active,ctime,mtime,operator FROM tribe_pack_version WHERE id=?"
	_updateTribePackVersionStatus    = "UPDATE tribe_pack_version SET is_active=?,operator=? WHERE id=?"
	_selectTribePackVersionForUpdate = "SELECT id,tribe_id,env,version_code,version_name,is_active FROM tribe_pack_version WHERE tribe_id=? AND env=? AND version_code=? FOR UPDATE"

	// tribe_pack_upgrade
	_addTribeConfigUpgrade    = "INSERT INTO tribe_pack_upgrade (tribe_id,env,tribe_pack_id,start_version_code,chosen_version_code) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE start_version_code=VALUES(start_version_code),chosen_version_code=VALUES(chosen_version_code)"
	_selectTribeConfigUpgrade = "SELECT id,tribe_id,env,tribe_pack_id,start_version_code,chosen_version_code from tribe_pack_upgrade WHERE tribe_id=? AND env=? AND tribe_pack_id=?"

	// tribe_config_filter
	_insertTribePackFilterConfig          = "INSERT INTO tribe_config_filter (tribe_id,env,tribe_pack_id,network,isp,channel,city,percent,salt,device,type,excludes_system,operator) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE network=VALUES(network),isp=VALUES(isp),channel=VALUES(channel),city=VALUES(city),percent=VALUES(percent),salt=VALUES(salt),device=VALUES(device),type=VALUES(type),excludes_system=VALUES(excludes_system),operator=VALUES(operator)"
	_selectTribePackFilterConfig          = "SELECT tribe_id,env,tribe_pack_id,network,isp,channel,city,percent,salt,device,type,excludes_system FROM tribe_config_filter WHERE %s"
	_selectTribeConfigPackFilterByPackIds = "SELECT tribe_id,env,tribe_pack_id,network,isp,channel,city,percent,salt,device,type,excludes_system,operator,ctime,mtime FROM tribe_config_filter WHERE tribe_id=? AND env=? AND tribe_pack_id IN (%s)"

	// tribe_config_flow
	_selectTribePackConfigFlow = "SELECT id,tribe_id,env,gl_job_id,flow,ctime,mtime,tribe_version_id,operator FROM tribe_config_flow WHERE tribe_id=? AND env=? AND gl_job_id IN (%s)"
	_batchInsertPackFlowConfig = "INSERT INTO tribe_config_flow (tribe_id,env,tribe_version_id,gl_job_id,flow,operator) VALUES %s ON DUPLICATE KEY UPDATE flow=VALUES(flow),operator=VALUES(operator)"

	// tribe_host_relations
	_batchInsertTribeHostRelation = "INSERT INTO tribe_host_relations (app_key,current_build_id,parent_build_id,feature) VALUES %s"
	_selectTribeHostRelation      = "SELECT id,current_build_id,parent_build_id,app_key,feature FROM tribe_host_relations WHERE %s ORDER BY id"
)

func (d *Dao) AddTribe(ctx context.Context, appKey, name, cName, owners, description string, noHost bool, priority int64, isBuildIn bool) (r int64, err error) {
	var row bsql.Result
	if row, err = d.db.Exec(ctx, _addTribe, appKey, name, cName, owners, description, noHost, priority, isBuildIn); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	return row.LastInsertId()
}

func (d *Dao) SelectTribeById(ctx context.Context, id int64) (tribeInfo *tribe.Tribe, err error) {
	row, err := d.db.Query(ctx, _selectTribeById, id)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	defer row.Close()
	err = row.Err()
	var l []*tribe.Tribe
	if err = xsql.ScanSlice(row, &l); err != nil {
		log.Errorc(ctx, "scan error: %#v", err)
		return
	}
	if len(l) == 0 {
		return
	}
	return l[0], err
}

func (d *Dao) SelectTribeByIds(ctx context.Context, ids []int64) (tribeInfos []*tribe.Tribe, err error) {
	if len(ids) == 0 {
		log.Warnc(ctx, "ids %v is empty", ids)
		return
	}
	var (
		sqls []string
		args []interface{}
	)
	for _, id := range ids {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	row, err := d.db.Query(ctx, fmt.Sprintf(_selectTribeByIds, strings.Join(sqls, ",")), args...)
	log.Warnc(ctx, "sql:%s arg: %v", fmt.Sprintf(_selectTribeByIds, strings.Join(sqls, ",")), args)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	defer row.Close()
	err = row.Err()
	if err = xsql.ScanSlice(row, &tribeInfos); err != nil {
		log.Errorc(ctx, "scan error: %#v", err)
		return
	}
	return
}

func (d *Dao) DeleteTribeById(ctx context.Context, id int64) (affect int64, err error) {
	var result bsql.Result
	if result, err = d.db.Exec(ctx, _deleteTribe, id); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	return result.RowsAffected()
}

func (d *Dao) CountTribeByArg(c context.Context, appKey, name, cName string) (total int64, err error) {
	var (
		args []interface{}
		sqls []string
	)
	if appKey != "" {
		args = append(args, appKey)
		sqls = append(sqls, "app_key=?")
	}
	if name != "" {
		args = append(args, name)
		sqls = append(sqls, "name=?")
	}
	if cName != "" {
		args = append(args, cName)
		sqls = append(sqls, "c_name=?")
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_countTribe, strings.Join(sqls, " AND ")), args...)
	if err = row.Scan(&total); err != nil && err != sql.ErrNoRows {
		log.Errorc(c, "SelectTribeByArg row.Scan error(%v)", err)
	}
	return
}

func (d *Dao) SelectTribeByArg(c context.Context, appKey, name, cName string, ps, pn int64) (tribeInfos []*tribe.Tribe, err error) {
	var (
		args []interface{}
		sqls []string
	)
	if appKey != "" {
		args = append(args, appKey)
		sqls = append(sqls, "app_key=?")
	}
	if name != "" {
		args = append(args, name)
		sqls = append(sqls, "name=?")
	}
	if cName != "" {
		args = append(args, cName)
		sqls = append(sqls, "c_name=?")
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_selectTribe, strings.Join(sqls, " AND ")), args...)
	if err != nil {
		log.Errorc(c, "sql[%s], args[%#v], SelectTribeByIds error: %#v", _selectTribe, args, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r tribe.Tribe
		if err = rows.Scan(&r.Id, &r.AppKey, &r.Name, &r.CName, &r.Owners, &r.Description, &r.NoHost, &r.Priority, &r.IsBuildIn); err != nil {
			log.Errorc(c, "SelectTribeByIds Scan error:%v r: %v", err, r)
			return
		}
		tribeInfos = append(tribeInfos, &r)
	}
	err = rows.Err()
	return
}

func (d *Dao) UpdateTribe(c context.Context, id int64, appKey, name, cName, owners, description string, noHost bool, priority int64, isBuildIn bool) (r int64, err error) {
	var (
		args []interface{}
		sqls []string
		res  bsql.Result
	)
	if appKey != "" {
		args = append(args, appKey)
		sqls = append(sqls, "app_key=?")
	}
	if name != "" {
		args = append(args, name)
		sqls = append(sqls, "name=?")
	}
	if cName != "" {
		args = append(args, cName)
		sqls = append(sqls, "c_name=?")
	}
	if owners != "" {
		args = append(args, owners)
		sqls = append(sqls, "owners=?")
	}
	if description != "" {
		args = append(args, description)
		sqls = append(sqls, "description=?")
	}
	args = append(args, priority)
	sqls = append(sqls, "priority=?")
	args = append(args, noHost)
	sqls = append(sqls, "no_host=?")
	args = append(args, isBuildIn)
	sqls = append(sqls, "is_build_in=?")
	args = append(args, id)
	log.Info("UpdateTribe slq[%s] args[%#v] error[%v]", fmt.Sprintf(_updateTribe, strings.Join(sqls, ",")), args, err)
	if res, err = d.db.Exec(c, fmt.Sprintf(_updateTribe, strings.Join(sqls, ",")), args...); err != nil {
		log.Errorc(c, "UpdateTribe slq[%s] args[%#v] error[%#v]", fmt.Sprintf(_updateTribe, strings.Join(sqls, ",")), args, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpdateTribeBuildPackDepJobID(c context.Context, buildPackId, depGitlabJobID int64, status int8) (r int64, err error) {
	var (
		args []interface{}
		res  bsql.Result
	)
	args = append(args, depGitlabJobID)
	args = append(args, status)
	args = append(args, buildPackId)
	if res, err = d.db.Exec(c, _updateTribeBuildPackDepJobId, args...); err != nil {
		log.Errorc(c, "UpdateTribe slq[%s] args[%#v] error[%#v]", _updateTribeBuildPackDepJobId, args, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpdateTribeBuildPackGitInfo(c context.Context, buildPackId, gitlabJobId, depGitlabJobId int64, gitPath, commit string, buildStartTime int64, status int8) (r int64, err error) {
	var (
		args []interface{}
		sqls []string
		res  bsql.Result
	)
	if gitlabJobId != 0 {
		args = append(args, gitlabJobId)
		sqls = append(sqls, "gl_job_id=?")
	}
	if depGitlabJobId != 0 {
		args = append(args, depGitlabJobId)
		sqls = append(sqls, "dep_gl_job_id=?")
	}
	if gitPath != "" {
		args = append(args, gitPath)
		sqls = append(sqls, "git_path=?")
	}
	if commit != "" {
		args = append(args, commit)
		sqls = append(sqls, "commit=?")
	}
	if buildStartTime != 0 {
		args = append(args, time.Unix(buildStartTime, 0))
		sqls = append(sqls, "build_start_time=?")
	}
	if status != 0 {
		args = append(args, status)
		sqls = append(sqls, "status=?")
	}
	args = append(args, buildPackId)
	if res, err = d.db.Exec(c, fmt.Sprintf(_updateTribeBuildPack, strings.Join(sqls, ",")), args...); err != nil {
		log.Errorc(c, "UpdateTribe slq[%s] args[%#v] error[%#v]", fmt.Sprintf(_updateTribeBuildPack, strings.Join(sqls, ",")), args, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpdateTribeBuildPackPkgInfo(c context.Context, buildPackId int64, pkgPath, pkgUrl, mappingUrl, bbrUrl, md5, changeLog, depFeature string, versionCode int64, versionName string, buildEndTime, size int64, status int8) (r int64, err error) {
	var (
		args []interface{}
		sqls []string
		res  bsql.Result
	)
	if pkgPath != "" {
		args = append(args, pkgPath)
		sqls = append(sqls, "pkg_path=?")
	}
	if pkgUrl != "" {
		args = append(args, pkgUrl)
		sqls = append(sqls, "pkg_url=?")
	}
	if size != 0 {
		args = append(args, size)
		sqls = append(sqls, "size=?")
	}
	if mappingUrl != "" {
		args = append(args, mappingUrl)
		sqls = append(sqls, "mapping_url=?")
	}
	if bbrUrl != "" {
		args = append(args, bbrUrl)
		sqls = append(sqls, "bbr_url=?")
	}
	if md5 != "" {
		args = append(args, md5)
		sqls = append(sqls, "md5=?")
	}
	if changeLog != "" {
		args = append(args, changeLog)
		sqls = append(sqls, "change_log=?")
	}
	if depFeature != "" {
		args = append(args, depFeature)
		sqls = append(sqls, "dep_feature=?")
	}
	if buildEndTime != 0 {
		args = append(args, time.Unix(buildEndTime, 0))
		sqls = append(sqls, "build_end_time=?")
	}
	if status != 0 {
		args = append(args, status)
		sqls = append(sqls, "status=?")
	}
	if versionCode != 0 {
		args = append(args, versionCode)
		sqls = append(sqls, "version_code=?")
	}
	if versionName != "" {
		args = append(args, versionName)
		sqls = append(sqls, "version_name=?")
	}
	args = append(args, buildPackId)
	if res, err = d.db.Exec(c, fmt.Sprintf(_updateTribeBuildPackPkgInfo, strings.Join(sqls, ",")), args...); err != nil {
		log.Errorc(c, "UpdateTribe slq[%s] args[%#v] error[%#v]", fmt.Sprintf(_updateTribeBuildPackPkgInfo, strings.Join(sqls, ",")), args, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpdateTribePackPkgInfo(ctx context.Context, tribeId int64, gitlabJobId int64, pkgPath string, pkgUrl string, mappingUrl string, bbrUrl string) (effect int64, err error) {
	var (
		args []interface{}
		sqls []string
		res  bsql.Result
	)
	if pkgPath != "" {
		args = append(args, pkgPath)
		sqls = append(sqls, "pack_path=?")
	}
	if pkgUrl != "" {
		args = append(args, pkgUrl)
		sqls = append(sqls, "pack_url=?")
	}
	if mappingUrl != "" {
		args = append(args, mappingUrl)
		sqls = append(sqls, "mapping_url=?")
	}
	if bbrUrl != "" {
		args = append(args, bbrUrl)
		sqls = append(sqls, "bbr_url=?")
	}
	args = append(args, tribeId, gitlabJobId)
	if res, err = d.db.Exec(ctx, fmt.Sprintf(_updateTribePackPkgInfo, strings.Join(sqls, ",")), args...); err != nil {
		log.Errorc(ctx, "UpdateTribe slq[%s] args[%#v] error[%#v]", fmt.Sprintf(_updateTribePackPkgInfo, strings.Join(sqls, ",")), args, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) AddTribeBuildPack(ctx context.Context, tribeId, depGitlabJobId, pkgType, gitType int64, appKey, gitName, operator, ciEnvVars, description string, shouldNotify bool, appId, gitPath, gitlabPrjId string, status int8) (r int64, err error) {
	var row bsql.Result
	if row, err = d.db.Exec(ctx, _addTribeBuildPack, tribeId, depGitlabJobId, pkgType, gitType, appKey, gitName, operator, ciEnvVars, description, shouldNotify, appId, gitPath, gitlabPrjId, status); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	return row.LastInsertId()
}

// SelectTribeBuildPackByArg tribe_build_pack
func (d *Dao) SelectTribeBuildPackByArg(c context.Context, appKey string, tribeId, gitlabJobId, depGitlabJobId int64, pkgType, status, state, gitType int32, gitName, commit, operator, pushCD, orderBy, sort string, ps, pn int64) (buildPacks []*tribe.BuildPack, err error) {
	sqlAnd, args := tribeBuildPackFilter(appKey, tribeId, gitlabJobId, depGitlabJobId, pkgType, status, state, gitName, commit, operator, gitType, pushCD)
	if len(sqlAnd) == 0 {
		return
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_tribeBuildPacksSelect, strings.Join(sqlAnd, " AND "), "ORDER BY "+orderBy+" "+sort), args...)
	if err != nil {
		log.Errorc(c, "sql[%s], args[%#v], SelectTribeBuildPackByArg error: %#v", _tribeBuildPacksSelect, args, err)
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &buildPacks); err != nil {
		log.Errorc(c, "ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) SelectTribePackByArg(ctx context.Context, appKey string, env string, ps int64, pn int64) (packs []*tribe.Pack, err error) {
	var (
		sqlAnd []string
		args   []interface{}
	)
	if appKey != "" {
		args = append(args, appKey)
		sqlAnd = append(sqlAnd, "app_key=?")
	}
	if env != "" {
		args = append(args, env)
		sqlAnd = append(sqlAnd, "env=?")
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(ctx, fmt.Sprintf(_tribePacksSelect, strings.Join(sqlAnd, " AND ")), args...)
	if err != nil {
		log.Errorc(ctx, "sql[%s], args[%#v], SelectTribePackByArg error: %#v", _tribePacksSelect, args, err)
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &packs); err != nil {
		log.Errorc(ctx, "ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) SelectTribePackById(ctx context.Context, id int64) (pack *tribe.Pack, err error) {
	rows, err := d.db.Query(ctx, _tribePacksSelectById, id)
	if err != nil {
		log.Errorc(ctx, "sql[%s], id[%d], SelectTribePackById error: %#v", _tribePacksSelectById, id, err)
		return
	}
	defer rows.Close()
	err = rows.Err()
	var l []*tribe.Pack
	if err = xsql.ScanSlice(rows, &l); err != nil {
		log.Errorc(ctx, "ScanSlice Error: %v", err)
		return
	}
	if len(l) == 0 {
		return
	}
	return l[0], err
}

func (d *Dao) SelectTribePackVersionById(ctx context.Context, id int64) (version *tribe.PackVersion, err error) {
	rows, err := d.db.Query(ctx, _selectTribePackVersionById, id)
	if err != nil {
		log.Errorc(ctx, "sql[%s], id[%d], SelectTribePackVersionById error: %#v", _selectTribePackVersionById, id, err)
		return
	}
	defer rows.Close()
	err = rows.Err()
	var l []*tribe.PackVersion
	if err = xsql.ScanSlice(rows, &l); err != nil {
		log.Errorc(ctx, "ScanSlice Error: %v", err)
		return
	}
	if len(l) == 0 {
		return
	}
	return l[0], err
}

func (d *Dao) SelectTribePackVersionByArgs(ctx context.Context, tribeId int64, env string, versionCode int64) (version *tribe.PackVersion, err error) {
	rows, err := d.db.Query(ctx, _selectTribePackVersionByArgs, tribeId, env, versionCode)
	if err != nil {
		log.Errorc(ctx, "sql[%s], SelectTribePackVersionByArgs error: %#v", _selectTribePackVersionByArgs, err)
		return
	}
	defer rows.Close()
	err = rows.Err()
	var l []*tribe.PackVersion
	if err = xsql.ScanSlice(rows, &l); err != nil {
		log.Errorc(ctx, "ScanSlice Error: %v", err)
		return
	}
	if len(l) == 0 {
		return
	}
	return l[0], err
}

func (d *Dao) SelectTribePackByVersions(ctx context.Context, tribeId int64, env string, versionIds []int64) (packs []*tribe.Pack, err error) {
	if len(versionIds) == 0 {
		return
	}
	var (
		sqls []string
		args []interface{}
	)
	args = append(args, tribeId, env)
	for _, vid := range versionIds {
		sqls = append(sqls, "?")
		args = append(args, vid)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_tribePacksSelectByVersions, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(ctx, "sql[%s], args[%#v], SelectTribePackByArg error: %#v", _tribePacksSelectByVersions, args, err)
		return
	}
	defer rows.Close()
	var list []*tribe.Pack
	if err = xsql.ScanSlice(rows, &list); err != nil {
		log.Errorc(ctx, "ScanSlice Error: %v", err)
		return
	}
	packs = list
	err = rows.Err()
	return
}

// CountTribeBuildPack get the total counts
func (d *Dao) CountTribeBuildPack(c context.Context, appKey string, tribeId, gitlabJobId, depGitlabJobId int64, pkgType, status, state int32, gitName, commit, operator string, gitType int32, pushCD string) (r int, err error) {
	sqlAnd, args := tribeBuildPackFilter(appKey, tribeId, gitlabJobId, depGitlabJobId, pkgType, status, state, gitName, commit, operator, gitType, pushCD)
	if len(sqlAnd) != 0 {
		row := d.db.QueryRow(c, fmt.Sprintf(_countTribeBuildPacks, strings.Join(sqlAnd, " AND ")), args...)
		if err = row.Scan(&r); err != nil && err != sql.ErrNoRows {
			log.Errorc(c, "d.TribeBuildPackCount row.Scan error(%v)", err)
		}
	}
	return
}

func (d *Dao) CountTribePack(ctx context.Context, appKey string, env string) (r int, err error) {
	var (
		sqlAnd []string
		args   []interface{}
	)
	if appKey != "" {
		args = append(args, appKey)
		sqlAnd = append(sqlAnd, "app_key=?")
	}
	if env != "" {
		args = append(args, env)
		sqlAnd = append(sqlAnd, "env=?")
	}
	row := d.db.QueryRow(ctx, fmt.Sprintf(_countTribePacks, strings.Join(sqlAnd, " AND ")), args...)
	if err = row.Scan(&r); err != nil && err != sql.ErrNoRows {
		log.Errorc(ctx, "d.CountTribePack row.Scan error(%v)", err)
	}
	return
}

func (d *Dao) SelectTribeBuildPacksByIds(ctx context.Context, ids []int64) (tribeInfos []*tribe.BuildPack, err error) {
	if len(ids) == 0 {
		return
	}
	var (
		sqls = make([]string, 0, len(ids))
		args = make([]interface{}, 0)
	)
	for _, v := range ids {
		sqls = append(sqls, "?")
		args = append(args, v)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selectTribeBuildPackByIds, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(ctx, "SelectTribeBuildPacksByIds error: %v", err)
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &tribeInfos); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) SelectTribeBuildPackById(ctx context.Context, id int64) (buildPack *tribe.BuildPack, err error) {
	rows, err := d.db.Query(ctx, _selectTribeBuildPackById, id)
	if err != nil {
		log.Errorc(ctx, "SelectTribeBuildPackById error: %v", err)
		return
	}
	defer rows.Close()
	l := make([]*tribe.BuildPack, 0)
	if err = xsql.ScanSlice(rows, &l); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if len(l) != 0 {
		buildPack = l[0]
	}
	err = rows.Err()
	return
}

func (d *Dao) TxSelectTribePackVersionForUpdate(tx *sql.Tx, tribeId int64, env string, depGitlabJobId int64) (tribePackVersion *tribe.PackVersion, err error) {
	rows, err := tx.Query(_selectTribePackVersionForUpdate, tribeId, env, depGitlabJobId)
	if err != nil {
		log.Error("SelectTribePackVersionForUpdate error: %v", err)
		return
	}
	defer rows.Close()
	l := make([]*tribe.PackVersion, 0)
	if err = xsql.ScanSlice(rows, &l); err != nil {
		log.Error("SelectTribePackVersionForUpdate scan error: %v", err)
		return
	}
	if len(l) == 0 {
		return
	}
	tribePackVersion = l[0]
	err = rows.Err()
	return
}

func (d *Dao) TxSetTribePackVersion(tx *sql.Tx, tribeId int64, env string, versionCode int64, versionName string, isActive bool) (id int64, err error) {
	var (
		result      bsql.Result
		isActiveInt int8
	)
	if isActive {
		isActiveInt = 1
	}
	if result, err = tx.Exec(_setTribePackVersion, tribeId, env, versionCode, versionName, isActiveInt); err != nil {
		log.Error("SetTribePackVersionTx error: %v", err)
		return
	}
	return result.LastInsertId()
}

func (d *Dao) UpdatePackVersionStatus(ctx context.Context, versionID int64, isActive int8, op string) (id int64, err error) {
	var result bsql.Result
	if result, err = d.db.Exec(ctx, _updateTribePackVersionStatus, isActive, op, versionID); err != nil {
		log.Errorc(ctx, "SetTribePackVersionTx error: %v", err)
		return
	}
	return result.LastInsertId()
}

func (d *Dao) TxAddTribePackFromBuild(tx *sql.Tx, buildPack *tribe.BuildPack, versionID int64, env, cdnUrl, desc, sender string) (id int64, err error) {
	exec, err := tx.Exec(_AddTribePack, buildPack.AppId, buildPack.AppKey, env, buildPack.TribeId, buildPack.GlJobId, buildPack.DepGlJobId, buildPack.DepFeature,
		versionID, buildPack.GitType, buildPack.GitName, buildPack.Commit, buildPack.PkgType, buildPack.ChangeLog, buildPack.Operator, buildPack.Size,
		buildPack.Md5, buildPack.PkgPath, buildPack.PkgUrl, buildPack.MappingUrl, buildPack.BbrUrl, cdnUrl, desc, sender)
	if err != nil {
		log.Error("d.TxAddTribePack tx.Exec error(%v)", err)
		return
	}
	return exec.LastInsertId()
}

func (d *Dao) TxCopyTribePack(tx *sql.Tx, pack *tribe.Pack, versionID int64, env, desc, sender string) (id int64, err error) {
	exec, err := tx.Exec(_AddTribePack, pack.AppId, pack.AppKey, env, pack.TribeId, pack.GlJobId, pack.DepGlJobId, pack.DepFeature,
		versionID, pack.GitType, pack.GitName, pack.Commit, pack.PackType, pack.ChangeLog, pack.Operator, pack.Size,
		pack.Md5, pack.PackPath, pack.PackUrl, pack.MappingUrl, pack.BbrUrl, pack.CdnUrl, desc, sender)
	if err != nil {
		log.Error("d.TxAddTribePack tx.Exec error(%v)", err)
		return
	}
	return exec.LastInsertId()
}

// TxUpdateTribeBuildPackDidPush update did push flag.
func (d *Dao) TxUpdateTribeBuildPackDidPush(tx *sql.Tx, buildPackID int64, didPush bool) (r int64, err error) {
	res, err := tx.Exec(_tribeBuildPackDidPush, didPush, buildPackID)
	if err != nil {
		log.Error("d.TxTribeBuildPackDidPush tx.Exec error(%v)", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpdateTribeStatus(c context.Context, id int64, status int8, errMsg string) (r int64, err error) {
	var (
		args []interface{}
		res  bsql.Result
	)
	args = append(args, status, errMsg, id)
	if res, err = d.db.Exec(c, _updateTribeStatus, args...); err != nil {
		log.Errorc(c, "UpdateTribe slq[%s] args[%#v] error[%#v]", _updateTribeStatus, args, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) UpdateTribeMavenStatus(c context.Context, id, pushMavenStatus int64) (r int64, err error) {
	var (
		args []interface{}
		res  bsql.Result
	)
	args = append(args, pushMavenStatus, id)
	if res, err = d.db.Exec(c, _updateTribeMavenStatus, args...); err != nil {
		log.Errorc(c, "_updateTribeMavenStatus slq[%s] args[%#v] error[%#v]", _updateTribeMavenStatus, args, err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) CountPackVersionByOptions(ctx context.Context, id int64, env string) (total int64, err error) {
	row := d.db.QueryRow(ctx, _countPackVersionByOptions, id, env)
	if err = row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(ctx, "CountPackVersionByOptions error: %v", err)
		}
	}
	return
}

func (d *Dao) BatchAddPackFlowConfig(ctx context.Context, tribeId int64, env string, versionId int64, op string, flows *tribe.FlowInfo) (err error) {
	var (
		sqls = make([]string, 0, len(flows.GitlabJobIds))
		args = make([]interface{}, 0)
	)
	if len(flows.GitlabJobIds) == 0 {
		return
	}
	for i, v := range flows.GitlabJobIds {
		sqls = append(sqls, "(?,?,?,?,?)")
		args = append(args, tribeId, env, versionId, v, flows.Flows[i])
	}
	if _, err = d.db.Exec(ctx, fmt.Sprintf(_batchInsertPackFlowConfig, strings.Join(sqls, ",")), args...); err != nil {
		log.Errorc(ctx, "SetTribePackVersionTx error: %v", err)
	}
	return
}

func (d *Dao) SelectTribePackConfigFlow(ctx context.Context, tribeId int64, env string, gitlabJobIds []int64) (flows []*tribe.ConfigFlow, err error) {
	if len(gitlabJobIds) == 0 {
		return
	}
	var (
		sqls []string
		args []interface{}
	)
	args = append(args, tribeId, env)
	for _, glIds := range gitlabJobIds {
		sqls = append(sqls, "?")
		args = append(args, glIds)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selectTribePackConfigFlow, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(ctx, "sql[%s], args[%#v], SelectTribePackByArg error: %#v", _selectTribePackConfigFlow, args, err)
		return
	}
	defer rows.Close()
	var list []*tribe.ConfigFlow
	if err = xsql.ScanSlice(rows, &list); err != nil {
		log.Errorc(ctx, "ScanSlice Error: %v", err)
		return
	}
	flows = list
	err = rows.Err()
	return
}

func (d *Dao) TribePackVersionListByOptions(ctx context.Context, id int64, env string, pn int64, ps int64) (versions []*tribe.PackVersion, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, id, env)
	sqlAdd += " ORDER BY tpv.version_code DESC"
	args = append(args, (pn-1)*ps, ps)
	sqlAdd += " LIMIT ?,?"
	rows, err := d.db.Query(ctx, fmt.Sprintf(_listPackVersionByOptions, sqlAdd), args...)
	if err != nil {
		log.Errorc(ctx, "TribePackVersionListByOptions error: %v", err)
		return
	}
	defer rows.Close()
	var list []*tribe.PackVersion
	if err = xsql.ScanSlice(rows, &list); err != nil {
		log.Errorc(ctx, "scan error: %#v", err)
		return
	}
	versions = list
	err = rows.Err()
	return
}

func (d *Dao) AddTribeConfigUpgrade(ctx context.Context, tribeId int64, env string, tribePackId int64, startingCode string, chosenCode string) (id int64, err error) {
	var row bsql.Result
	if row, err = d.db.Exec(ctx, _addTribeConfigUpgrade, tribeId, env, tribePackId, startingCode, chosenCode); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	return row.LastInsertId()
}

func (d *Dao) SelectTribeConfigUpgrade(ctx context.Context, tribeId int64, env string, versionId int64) (upgrade *tribe.PackUpgrade, err error) {
	rows, err := d.db.Query(ctx, _selectTribeConfigUpgrade, tribeId, env, versionId)
	if err != nil {
		log.Errorc(ctx, "sql[%s], SelectTribeConfigUpgrade error: %#v", _selectTribeConfigUpgrade, err)
		return
	}
	defer rows.Close()
	var l []*tribe.PackUpgrade
	if err = xsql.ScanSlice(rows, &l); err != nil {
		log.Errorc(ctx, "ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	if len(l) == 0 {
		return
	}
	return l[0], err
}

func (d *Dao) AddTribePackFilterConfig(ctx context.Context, tribeId int64, env string, buildId int64, isp, network, channel, city, deviceId string, upgradeType, percent int64, salt, excludesSys, op string) (id int64, err error) {
	var row bsql.Result
	if row, err = d.db.Exec(ctx, _insertTribePackFilterConfig, tribeId, env, buildId, network, isp, channel, city, percent, salt, deviceId, upgradeType, excludesSys, op); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	return row.LastInsertId()
}

func (d *Dao) SelectTribeConfigPackFilter(ctx context.Context, tribeId int64, env string, tribePackId int64) (filter *tribe.ConfigFilter, err error) {
	var (
		args []interface{}
		sqls []string
	)
	if tribeId != 0 {
		args = append(args, tribeId)
		sqls = append(sqls, "tribe_id=?")
	}
	if env != "" {
		args = append(args, env)
		sqls = append(sqls, "env=?")
	}
	if tribePackId != 0 {
		args = append(args, tribePackId)
		sqls = append(sqls, "tribe_pack_id=?")
	}
	if len(args) == 0 {
		return
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selectTribePackFilterConfig, strings.Join(sqls, " AND ")), args...)
	if err != nil {
		log.Errorc(ctx, "TribePackVersionListByOptions error: %v", err)
		return
	}
	defer rows.Close()
	var list []*tribe.ConfigFilter
	if err = xsql.ScanSlice(rows, &list); err != nil {
		log.Errorc(ctx, "scan error: %#v", err)
		return
	}
	err = rows.Err()
	if len(list) == 0 {
		return
	}
	filter = list[0]
	return
}

func (d *Dao) SelectTribeConfigPackFilterByPackIds(ctx context.Context, tribeId int64, env string, tribePackId []int64) (filter []*tribe.ConfigFilter, err error) {
	var (
		args []interface{}
		sqls []string
	)
	if len(tribePackId) == 0 {
		return
	}
	args = append(args, tribeId, env)
	for _, pid := range tribePackId {
		sqls = append(sqls, "?")
		args = append(args, pid)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selectTribeConfigPackFilterByPackIds, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Errorc(ctx, "sql[%s], args[%#v], SelectTribePackByArg error: %#v", fmt.Sprintf(_selectTribeConfigPackFilterByPackIds, strings.Join(sqls, ",")), args, err)
		return
	}
	err = rows.Err()
	defer rows.Close()
	var l []*tribe.ConfigFilter
	if err = xsql.ScanSlice(rows, &l); err != nil {
		log.Errorc(ctx, "ScanSlice Error: %v", err)
		return
	}
	return l, err
}

func (d *Dao) SelectTribeHostRelation(ctx context.Context, appKey, feature string) (relations []*tribe.HostRelations, err error) {
	var (
		args []interface{}
		sqls []string
	)
	if len(appKey) != 0 {
		sqls = append(sqls, "app_key=?")
		args = append(args, appKey)
	}
	if len(feature) != 0 {
		sqls = append(sqls, "feature=?")
		args = append(args, feature)
	}
	rows, err := d.db.Query(ctx, fmt.Sprintf(_selectTribeHostRelation, strings.Join(sqls, " AND ")), args...)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	err = rows.Err()
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &relations); err != nil {
		log.Errorc(ctx, "scan error: %#v", err)
		return
	}
	return
}

func (d *Dao) BatchAddTribeHostRelation(ctx context.Context, appKey string, currentBuildId int64, compatibles []*cimdl.Feature) (err error) {
	var (
		sqls = make([]string, 0, len(compatibles))
		args = make([]interface{}, 0)
	)
	if len(compatibles) == 0 {
		return
	}
	for _, v := range compatibles {
		sqls = append(sqls, "(?,?,?,?)")
		args = append(args, appKey, currentBuildId, v.CompatibleVersion, v.Name)
	}
	if _, err = d.db.Exec(ctx, fmt.Sprintf(_batchInsertTribeHostRelation, strings.Join(sqls, ",")), args...); err != nil {
		log.Errorc(ctx, "SetTribePackVersionTx error: %v", err)
	}
	return
}

func tribeBuildPackFilter(appKey string, tribeId, gitlabJobId, depGitlabJobId int64, pkgType, status, state int32, gitName, commit, operator string, gitType int32, pushCD string) (sqlAnd []string, args []interface{}) {
	if appKey != "" {
		args = append(args, appKey)
		sqlAnd = append(sqlAnd, "app_Key=?")
	}
	if state != 0 {
		args = append(args, state)
		sqlAnd = append(sqlAnd, "state=?")
	}
	if tribeId != 0 {
		args = append(args, tribeId)
		sqlAnd = append(sqlAnd, "tribe_id=?")
	}
	if gitlabJobId != 0 {
		args = append(args, gitlabJobId)
		sqlAnd = append(sqlAnd, "gl_job_id=?")
	}
	if depGitlabJobId != 0 {
		args = append(args, depGitlabJobId)
		sqlAnd = append(sqlAnd, "dep_gl_job_id=?")
	}
	if pkgType != 0 {
		args = append(args, pkgType)
		sqlAnd = append(sqlAnd, "pkg_Type=?")
	}
	if status != 0 {
		args = append(args, status)
		sqlAnd = append(sqlAnd, "status=?")
	}
	if operator != "" {
		args = append(args, operator)
		sqlAnd = append(sqlAnd, "operator=?")
	}
	if pushCD != "" {
		args = append(args, pushCD)
		sqlAnd = append(sqlAnd, "did_push=?")
	}
	if gitName != "" || commit != "" {
		sqlAnd = append(sqlAnd, "git_type=?")
		args = append(args, gitType)
		if gitType != cimdl.GitTypeCommit {
			sqlAnd = append(sqlAnd, "git_name=?")
			args = append(args, gitName)
		} else {
			sqlAnd = append(sqlAnd, "commit=?")
			args = append(args, commit)
		}
	}
	return
}
