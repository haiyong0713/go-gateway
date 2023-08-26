package fawkes

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/database/sql"

	bizapkmdl "go-gateway/app/app-svr/fawkes/service/model/bizapk"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_bizApkBuilds = `SELECT b.id,a.app_key,a.name,a.c_name,a.description,s.id,s.active,s.priority,b.biz_apk_id,b.env,b.pack_build_id,b.bundle_ver,b.md5,b.size,b.gl_ppl_id,b.gl_job_id,b.git_type,b.git_name,b.git_commit,b.apk_path,b.map_path,b.meta_path,b.apk_url,b.map_url,b.meta_url,b.apk_cdn_url,b.job_status,b.operator,b.did_push,b.built_in,unix_timestamp(b.ctime),unix_timestamp(b.mtime) 
	FROM biz_apk AS a,biz_apk_build AS b,biz_apk_pack_settings AS s WHERE a.id=b.biz_apk_id AND b.biz_apk_id=s.biz_apk_id AND b.pack_build_id=s.pack_build_id AND b.env=s.env AND a.app_key=? AND b.env=? AND b.pack_build_id=? AND state=0 ORDER BY b.id DESC`
	_bizApkBuildsWithID = `SELECT b.id,a.app_key,a.name,a.c_name,a.description,s.id,s.active,s.priority,b.biz_apk_id,b.env,b.pack_build_id,b.bundle_ver,b.md5,b.size,b.gl_ppl_id,b.gl_job_id,b.git_type,b.git_name,b.git_commit,b.apk_path,b.map_path,b.meta_path,b.apk_url,b.map_url,b.meta_url,b.apk_cdn_url,b.job_status,b.operator,b.did_push,unix_timestamp(b.ctime),unix_timestamp(b.mtime) 
	FROM biz_apk AS a,biz_apk_build AS b,biz_apk_pack_settings AS s WHERE b.id=? AND b.biz_apk_id=a.id AND b.biz_apk_id=s.biz_apk_id AND b.pack_build_id=s.pack_build_id AND b.env=s.env AND state=0`

	_bizApks           = `SELECT ba.id,ba.name,ba.c_name,ba.description,s.id,s.pack_build_id,s.env,s.active,s.priority,s.operator,unix_timestamp(s.mtime) FROM biz_apk AS ba, biz_apk_pack_settings AS s WHERE s.biz_apk_id=ba.id AND ba.app_key=? AND s.pack_build_id=? AND s.env=? %v`
	_addBizApkSetting  = `INSERT INTO biz_apk_pack_settings (pack_build_id,biz_apk_id,env,priority,operator) VALUES (?,?,?,?,?)`
	_bizApkSettingID   = `SELECT id from biz_apk_pack_settings where pack_build_id=? AND biz_apk_id=? AND env=?`
	_setBizApkActive   = `UPDATE biz_apk_pack_settings SET active=?,operator=? WHERE id=?`
	_setBizApkPriority = `UPDATE biz_apk_pack_settings SET priority=?,operator=? WHERE id=?`

	_getBizApkID = `SELECT id FROM biz_apk WHERE app_key=? AND BINARY name=?`
	_addBizApk   = `INSERT INTO biz_apk (app_key,name,c_name,description) VALUES (?,?,?,?)`

	_createBizApkBuild       = `INSERT INTO biz_apk_build (biz_apk_id,env,pack_build_id,git_type,git_name,operator) VALUES (?,'test',?,?,?,?)`
	_bizApkBuildID           = `SELECT id from biz_apk_build where pack_build_id=? AND biz_apk_id=? AND env=?`
	_updateBizApkBuildPpl    = `UPDATE biz_apk_build SET gl_ppl_id=?,git_commit=? WHERE id=? AND env='test' AND state=0`
	_startBizApkBuild        = `UPDATE biz_apk_build SET gl_job_id=?,bundle_ver=?,job_status=2 WHERE id=? AND env='test' AND state=0`
	_uploadBizApkBuild       = `UPDATE biz_apk_build SET md5=?,size=?,apk_path=?,map_path=?,meta_path=?,apk_url=?,map_url=?,meta_url=?,job_status=3,built_in=? WHERE id=? AND env='test' AND state=0`
	_updateBizApkBuildStatus = `UPDATE biz_apk_build SET job_status=? WHERE id=? AND env='test' AND state=0`
	_updateBizApkBuildCDN    = `UPDATE biz_apk_build SET apk_cdn_url=? WHERE id=? AND env='test' AND state=0`
	_delBizApkBuild          = `UPDATE biz_apk_build SET state=1 WHERE id=? AND env='test' AND state=0`

	_getBizApkJobRefresh = `SELECT biz_apk_build.id,git_prj_id,gl_ppl_id,gl_job_id,job_status FROM biz_apk_build,biz_apk,app,app_attribute WHERE biz_apk.id=biz_apk_build.biz_apk_id AND biz_apk.app_key=app.app_key AND app.id=app_attribute.app_table_id AND app_attribute.state=1 AND job_status in (1,2)`
	_getTribeJobRefresh  = `SELECT tribe_build_pack.id,git_prj_id,gl_job_id, status FROM tribe_build_pack,tribe,app,app_attribute WHERE tribe.id=tribe_build_pack.tribe_id AND tribe.app_key=app.app_key AND app.id=app_attribute.app_table_id AND app_attribute.state=1 AND gl_job_id != 0 AND status in (1,2)`

	_didPushProd      = `UPDATE biz_apk_build SET did_push=1 WHERE id=? AND env='test' AND state=0`
	_bundleVerHasProd = `SELECT count(*) FROM biz_apk_build WHERE bundle_ver=? AND env='prod' AND pack_build_id=? AND biz_apk_id=?`
	_addBizApkBuild   = `INSERT INTO biz_apk_build (biz_apk_id,env,pack_build_id,bundle_ver,md5,size,gl_ppl_id,gl_job_id,git_type,git_name,git_commit,apk_path,map_path,meta_path,apk_url,map_url,meta_url,apk_cdn_url,job_status,operator) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

	_orgPackURL = `SELECT pkg_url,mapping_url FROM build_pack WHERE app_key=? AND gl_job_id=?`

	_bizApkFilter    = `SELECT env,biz_apk_build_id,network,isp,channel,city,percent,salt,device,state,excludes_system FROM biz_apk_build_filter WHERE biz_apk_build_id IN (%v) AND env=?`
	_setBizApkFilter = `INSERT INTO biz_apk_build_filter (env,biz_apk_build_id,network,isp,channel,city,excludes_system,percent,salt,device,state) VALUES(?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE network=VALUES(network),isp=VALUES(isp),channel=VALUES(channel),city=VALUES(city),excludes_system=VALUES(excludes_system),percent=VALUES(percent),device=VALUES(device),state=VALUES(state)`

	_bizApkFlow    = `SELECT env,biz_apk_build_id,flow FROM biz_apk_flow WHERE pack_build_id=? AND biz_apk_id=? AND env=?`
	_setBizApkFlow = `INSERT INTO biz_apk_flow (env,pack_build_id,biz_apk_id,biz_apk_build_id,flow) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE flow=VALUES(flow)`
)

// BizApkBuilds get business apk builds of a package
func (d *Dao) BizApkBuilds(c context.Context, appKey, env string, packBuildID int64) (builds []*bizapkmdl.Build, err error) {
	rows, err := d.db.Query(c, _bizApkBuilds, appKey, env, packBuildID)
	if err != nil {
		log.Error("d.BizApkBuilds d.dbQuery() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var b = &bizapkmdl.Build{}
		if err = rows.Scan(&b.ID, &b.AppKey, &b.Name, &b.Cname, &b.Description, &b.SettingsID, &b.Active, &b.Priority, &b.BizApkID, &b.Env, &b.PackBuildID, &b.BundleVer, &b.MD5, &b.Size, &b.GitlabPipelineID, &b.GitlabJobID, &b.GitType, &b.GitName, &b.Commit, &b.ApkPath, &b.MapPath, &b.MetaPath, &b.ApkURL, &b.MapURL, &b.MetaURL, &b.ApkCdnURL, &b.Status, &b.Operator, &b.DidPush, &b.BuiltIn, &b.CTime, &b.MTime); err != nil {
			log.Error("d.BizApkBuilds rows.Scan error(%v)", err)
			return
		}
		builds = append(builds, b)
	}
	err = rows.Err()
	return
}

// BizApkBuildWithID get a business apk build from id
func (d *Dao) BizApkBuildWithID(c context.Context, bizapkBuildID int64) (b *bizapkmdl.Build, err error) {
	row := d.db.QueryRow(c, _bizApkBuildsWithID, bizapkBuildID)
	b = &bizapkmdl.Build{}
	if err = row.Scan(&b.ID, &b.AppKey, &b.Name, &b.Cname, &b.Description, &b.SettingsID, &b.Active, &b.Priority, &b.BizApkID, &b.Env, &b.PackBuildID, &b.BundleVer, &b.MD5, &b.Size, &b.GitlabPipelineID, &b.GitlabJobID, &b.GitType, &b.GitName, &b.Commit, &b.ApkPath, &b.MapPath, &b.MetaPath, &b.ApkURL, &b.MapURL, &b.MetaURL, &b.ApkCdnURL, &b.Status, &b.Operator, &b.DidPush, &b.CTime, &b.MTime); err != nil {
		log.Error("d.BizApkBuilds rows.Scan error(%v)", err)
	}
	return
}

// BizApks get business apks list from a package
func (d *Dao) BizApks(c context.Context, appKey string, packBuildID int64, env string) (r []*bizapkmdl.ApkPackSettings, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_bizApks, ""), appKey, packBuildID, env)
	if err != nil {
		log.Error("d.BizApks d.dbQuery() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var s = &bizapkmdl.ApkPackSettings{}
		if err = rows.Scan(&s.BizApkID, &s.Name, &s.Cname, &s.Description, &s.SettingsID, &s.PackBuildID, &s.Env, &s.Active, &s.Priority, &s.Operator, &s.MTime); err != nil {
			log.Error("d.BizApks rows.Scan error(%v)", err)
			return
		}
		r = append(r, s)
	}
	err = rows.Err()
	return
}

// BizApk get one business apk
func (d *Dao) BizApk(c context.Context, appKey string, packBuildID int64, env string, bizapkID int64) (s *bizapkmdl.ApkPackSettings, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	sqlAdd += " AND biz_apk_id=?"
	args = append(args, appKey, packBuildID, env, bizapkID)
	row := d.db.QueryRow(c, fmt.Sprintf(_bizApks, sqlAdd), args...)
	s = &bizapkmdl.ApkPackSettings{}
	if err = row.Scan(&s.BizApkID, &s.Name, &s.Cname, &s.Description, &s.SettingsID, &s.PackBuildID, &s.Env, &s.Active, &s.Priority, &s.Operator, &s.MTime); err != nil {
		log.Error("d.BizApk row.Scan error(%v)", err)
		return
	}
	return
}

// TxAddBizapkSettings add business apk settings
func (d *Dao) TxAddBizapkSettings(tx *sql.Tx, packBuildID, bizApkID int64, env string, priority int, operator string) (r int64, err error) {
	res, err := tx.Exec(_addBizApkSetting, packBuildID, bizApkID, env, priority, operator)
	if err != nil {
		log.Error("d.TxAddBizApk tx.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

// BizApkSettingID get BizApkSetting id
func (d *Dao) BizApkSettingID(c context.Context, packBuildID, bizApkID int64, env string) (r int64, err error) {
	res := d.db.QueryRow(c, _bizApkSettingID, packBuildID, bizApkID, env)
	if err = res.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Error("d.BizApkSettingID d.db.QueryRow() error(%v)", err)
	}
	return
}

// TxSetBizapkActive set business apk active
func (d *Dao) TxSetBizapkActive(tx *sql.Tx, active int, operator string, settingsID int64) (err error) {
	_, err = tx.Exec(_setBizApkActive, active, operator, settingsID)
	if err != nil {
		log.Error("d.TxUploadBizApk tx.Exec error(%v)", err)
		return
	}
	return
}

// TxSetBizapkPriority update business apk priority
func (d *Dao) TxSetBizapkPriority(tx *sql.Tx, priority int, operator string, settingsID int64) (err error) {
	_, err = tx.Exec(_setBizApkPriority, priority, operator, settingsID)
	if err != nil {
		log.Error("d.TxUploadBizApk tx.Exec error(%v)", err)
		return
	}
	return
}

// BizApkID whether biz apk did exist
func (d *Dao) BizApkID(c context.Context, appKey, name string) (r int64, err error) {
	res := d.db.QueryRow(c, _getBizApkID, appKey, name)
	if err = res.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Error("d.BizApkDidExist d.dbQuery() error(%v)", err)
	}
	return
}

// TxAddBizApk add business apk Info
func (d *Dao) TxAddBizApk(tx *sql.Tx, appKey, name, cName, description string) (r int64, err error) {
	res, err := tx.Exec(_addBizApk, appKey, name, cName, description)
	if err != nil {
		log.Error("d.TxAddBizApk tx.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

// TxCreateBizApkBuild build a business apk from pipeline
func (d *Dao) TxCreateBizApkBuild(tx *sql.Tx, bizApkID, packBuildID int64, gitType int, gitName, userName string) (r int64, err error) {
	res, err := tx.Exec(_createBizApkBuild, bizApkID, packBuildID, gitType, gitName, userName)
	if err != nil {
		log.Error("d.TxCreateBizApkBuild tx.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

// BizApkBuildID get BizApkBuildID id
func (d *Dao) BizApkBuildID(c context.Context, packBuildID, bizApkID int64, env string) (r int64, err error) {
	res := d.db.QueryRow(c, _bizApkBuildID, packBuildID, bizApkID, env)
	if err = res.Scan(&r); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Error("d.BizApkBuildID d.db.QueryRow() error(%v)", err)
	}
	return
}

// TxDidPushProd update push status for test env
func (d *Dao) TxDidPushProd(tx *sql.Tx, bizApkID int64) (err error) {
	_, err = tx.Exec(_didPushProd, bizApkID)
	if err != nil {
		log.Error("d.TxUploadBizApk tx.Exec error(%v)", err)
		return
	}
	return
}

// BundleVerHasProd check if there is a bundle ver already existing in prod env
func (d *Dao) BundleVerHasProd(c context.Context, bundleVer, packBuildID, bizApkID int64) (r int, err error) {
	row := d.db.QueryRow(c, _bundleVerHasProd, bundleVer, packBuildID, bizApkID)
	if err = row.Scan(&r); err != nil {
		log.Error("d.BuildPacksCount row.Scan error(%v)", err)
	}
	return
}

// TxAddBizApkBuild add a business apk's result directly
func (d *Dao) TxAddBizApkBuild(tx *sql.Tx, bizApkID int64, env string, packBuildID, bundleVer int64, md5 string, size, gitlabPipelineID, gitlabJobID int64, gitType int8, gitName, commit, apkPath, mapPath, metaPath, apkURL, mapURL, metaURL, apkCdnURL string, status int8, operator string) (r int64, err error) {
	res, err := tx.Exec(_addBizApkBuild, bizApkID, env, packBuildID, bundleVer, md5, size, gitlabPipelineID, gitlabJobID, gitType, gitName, commit, apkPath, mapPath, metaPath, apkURL, mapURL, metaURL, apkCdnURL, status, operator)
	if err != nil {
		log.Error("d.TxAddBizApkBuild tx.Exec error(%v)", err)
		return
	}
	r, err = res.LastInsertId()
	return
}

// TxUploadBizApk update info after business apk uploaded
func (d *Dao) TxUploadBizApk(tx *sql.Tx, md5 string, size int64, apkPath, mapPath, metaPath, apkURL, mapURL, metaURL string, buildID int64, builtIn int) (err error) {
	_, err = tx.Exec(_uploadBizApkBuild, md5, size, apkPath, mapPath, metaPath, apkURL, mapURL, metaURL, builtIn, buildID)
	if err != nil {
		log.Error("d.TxUploadBizApk tx.Exec error(%v)", err)
		return
	}
	return
}

// TxUpdateBizApkBuildPpl update pipeline id of a business apk build
func (d *Dao) TxUpdateBizApkBuildPpl(tx *sql.Tx, pplID int, commit string, buildID int64) (err error) {
	_, err = tx.Exec(_updateBizApkBuildPpl, pplID, commit, buildID)
	if err != nil {
		log.Error("d.TxUpdateBizApkBuildPpl tx.Exec error(%v)", err)
		return
	}
	return
}

// TxStartBizApkBuild update info of the business apk build while the job starting
func (d *Dao) TxStartBizApkBuild(tx *sql.Tx, jobID int64, buildID int64) (err error) {
	_, err = tx.Exec(_startBizApkBuild, jobID, jobID, buildID)
	if err != nil {
		log.Error("d.TxStartBizApkBuild tx.Exec error(%v)", err)
		return
	}
	return
}

// TxUpdateBizApkBuildStatus update business apk build pipeline status
func (d *Dao) TxUpdateBizApkBuildStatus(tx *sql.Tx, status int, buildID int64) (err error) {
	_, err = tx.Exec(_updateBizApkBuildStatus, status, buildID)
	if err != nil {
		log.Error("d.TxUpdateBizApkBuildStatus tx.Exec error(%v)", err)
		return
	}
	return
}

// TxUpdateBizApkBuildCDN update cdn url for business apk
func (d *Dao) TxUpdateBizApkBuildCDN(tx *sql.Tx, apkCdnURL string, buildID int64) (err error) {
	_, err = tx.Exec(_updateBizApkBuildCDN, apkCdnURL, buildID)
	if err != nil {
		log.Error("d.TxUpdateBizApkBuildCDN tx.Exec error(%v)", err)
		return
	}
	return
}

// TxDelBizApkBuild delete a build of business apk
func (d *Dao) TxDelBizApkBuild(tx *sql.Tx, buildID int64) (err error) {
	_, err = tx.Exec(_delBizApkBuild, buildID)
	if err != nil {
		log.Error("d.TxDelBizApkBuild tx.Exec error(%v)", err)
		return
	}
	return
}

// GetBizApkJobRefresh get the business apk builds which should refresh
func (d *Dao) GetBizApkJobRefresh(c context.Context) (jobsInfo []*bizapkmdl.JobInfo, err error) {
	rows, err := d.db.Query(c, _getBizApkJobRefresh)
	if err != nil {
		log.Error("d.BizApkBuilds d.dbQuery() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var info = &bizapkmdl.JobInfo{}
		if err = rows.Scan(&info.ID, &info.GitlabProjectID, &info.GitlabPipelineID, &info.GitlabJobID, &info.Status); err != nil {
			log.Error("d.BizApkBuilds rows.Scan error(%v)", err)
			return
		}
		jobsInfo = append(jobsInfo, info)
	}
	err = rows.Err()
	return
}

// GetTribeJobRefresh get the business apk builds which should refresh
func (d *Dao) GetTribeJobRefresh(c context.Context) (jobsInfo []*tribemdl.JobInfo, err error) {
	rows, err := d.db.Query(c, _getTribeJobRefresh)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var info = &tribemdl.JobInfo{}
		if err = rows.Scan(&info.ID, &info.GitlabProjectID, &info.GitlabJobID, &info.Status); err != nil {
			return
		}
		jobsInfo = append(jobsInfo, info)
	}
	err = rows.Err()
	return
}

// OrgPackURL get original package URL from gitlab job ID
func (d *Dao) OrgPackURL(c context.Context, appKey string, jobID int64) (res *bizapkmdl.OrgPackURLResp, err error) {
	row := d.db.QueryRow(c, _orgPackURL, appKey, jobID)
	res = &bizapkmdl.OrgPackURLResp{}
	if err = row.Scan(&res.PkgURL, &res.MappingURL); err != nil {
		log.Error("d.OrgPackURL d.dbQuery() error(%v)", err)
	}
	return
}

// TxSetBizApkFilterConfig insert app filter config.
func (d *Dao) TxSetBizApkFilterConfig(tx *sql.Tx, env string, buildID int64, network, isp, channel, city, excludesSystem string, percent int, salt, device string, status int) (r int64, err error) {
	res, err := tx.Exec(_setBizApkFilter, env, buildID, network, isp, channel, city, excludesSystem, percent, salt, device, status)
	if err != nil {
		log.Error("TxSetBizApkFilterConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// BizApkFilterConfig get filter config.
func (d *Dao) BizApkFilterConfig(c context.Context, env string, bids []int64) (fconfig map[int64]*bizapkmdl.FilterConfig, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, bid := range bids {
		sqls = append(sqls, "?")
		args = append(args, bid)
	}
	args = append(args, env)
	rows, err := d.db.Query(c, fmt.Sprintf(_bizApkFilter, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("BizApkFilterConfig %v", err)
		return
	}
	defer rows.Close()
	fconfig = make(map[int64]*bizapkmdl.FilterConfig)
	for rows.Next() {
		fc := &bizapkmdl.FilterConfig{}
		if err = rows.Scan(&fc.Env, &fc.BuildID, &fc.Network, &fc.ISP, &fc.Channel, &fc.City, &fc.Percent, &fc.Salt, &fc.Device, &fc.Status, &fc.ExcludesSystem); err != nil {
			log.Error("BizApkFilterConfig %v", err)
			return
		}
		fconfig[fc.BuildID] = fc
	}
	err = rows.Err()
	return
}

// TxSetBizApkFlowConfig insert app filter flow config.
func (d *Dao) TxSetBizApkFlowConfig(tx *sql.Tx, env, flow string, packBuildID, apkID, buildID int64) (r int64, err error) {
	res, err := tx.Exec(_setBizApkFlow, env, packBuildID, apkID, buildID, flow)
	if err != nil {
		log.Error("TxSetBizApkFlowConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// BizApkFlowConfig get flow config.
func (d *Dao) BizApkFlowConfig(c context.Context, env string, packBuildID, apkID int64) (fconfig map[int64]*bizapkmdl.FlowConfig, err error) {
	rows, err := d.db.Query(c, _bizApkFlow, packBuildID, apkID, env)
	if err != nil {
		log.Error("BizApkFlowConfig %v", err)
		return
	}
	defer rows.Close()
	fconfig = make(map[int64]*bizapkmdl.FlowConfig)
	for rows.Next() {
		fc := &bizapkmdl.FlowConfig{}
		if err = rows.Scan(&fc.Env, &fc.BuildID, &fc.Flow); err != nil {
			log.Error("BizApkFlowConfig %v", err)
			return
		}
		fconfig[fc.BuildID] = fc
	}
	err = rows.Err()
	return
}
