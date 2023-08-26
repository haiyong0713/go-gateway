package fawkes

import (
	"context"
	dsql "database/sql"
	"fmt"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/database/xsql"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	bizapkmdl "go-gateway/app/app-svr/fawkes/service/model/bizapk"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	"go-gateway/app/app-svr/fawkes/service/model/pcdn"
	taskmdl "go-gateway/app/app-svr/fawkes/service/model/task"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

//go:generate gensql -filter _addPcdnFile
const (
	_newestConfigPublishVersion = `SELECT app_key,env,max(id) FROM config_publish WHERE state=1 GROUP BY app_key,env`
	_newestFFPublishVersion     = `SELECT app_key,env,max(id) FROM ff_publish WHERE state=1 GROUP BY app_key,env`
	_versionAll                 = `SELECT id,app_id,app_key,env,version,version_code FROM pack_version WHERE is_upgrade=0`
	_releaseVersionList         = `SELECT id,app_id,app_key,env,version,version_code FROM pack_version WHERE app_key=? AND env=? ORDER BY version_code DESC`
	_upgradConfigAll            = "SELECT app_key,env,version_id,normal_upgrad,force_upgrad,exclude_normal,exclude_force," +
		"`system`,exclude_system,cycle,title,content,is_silent,policy,policy_url,icon_url,confirm_btn_text,cancel_btn_text FROM pack_upgrad"
	_packAll = `SELECT app_id,app_key,env,version_id,build_id,size,md5,pack_url,cdn_url,
mtime FROM pack ORDER BY version_id DESC,build_id DESC`
	_patchAll = `SELECT app_key,build_id,target_build_id,target_version_id,origin_build_id,origin_version_id,size,
md5,cdn_url FROM patch WHERE status=3 AND cdn_url!='' ORDER BY build_id DESC`
	_patchAppKeys           = `SELECT DISTINCT(app_key) from patch`
	_lastPackVersionIds     = `SELECT id from pack_version where app_key=? AND env=? AND is_upgrade=0 ORDER BY version_code DESC LIMIT 10`
	_packBuildIdsByVersions = `SELECT p.build_id FROM pack p INNER JOIN pack_version pv ON p.version_id=pv.id WHERE p.app_key=? AND p.env=? AND pv.id IN (%s)`
	_patchAll4              = `SELECT id,app_key,build_id,target_build_id,target_version_id,origin_build_id,origin_version_id,size,
md5,cdn_url FROM patch WHERE build_id IN(%s) and status=3 AND cdn_url!='' ORDER BY build_id DESC`
	_filterConfigAll = `SELECT app_key,env,build_id,network,isp,channel,city,percent,salt,device,status,phone_model,brand FROM pack_filter`
	_appChannelAll   = `SELECT ac.app_key,c.id,c.code,c.name,c.plate FROM app_channel AS ac,
channel AS c WHERE ac.channel_id=c.id AND c.channel_status=0`
	_flowConfigAll = `SELECT app_key,env,build_id,flow FROM pack_flow`
	_hotfixAll     = `SELECT app_id,app_key,env,gl_prj_id,gl_job_id,origin_version,origin_version_code,
origin_build_id,build_id,git_type,git_name,commit,size,md5,hotfix_path,hotfix_url,cdn_url,
description,status,state FROM hotfix WHERE status=3 AND state=0 ORDER BY build_id DESC`
	_hotfixConfigAll = `SELECT app_key,env,build_id,device,channel,city,upgrad_num,gray,effect FROM hotfix_config`
	_laserAll        = `SELECT id,app_key,platform,mid,buvid,email,log_date,url,status,silence_url,
silence_status,operator,mobi_app,unix_timestamp(ctime),unix_timestamp(mtime) FROM app_laser WHERE status IN (1,-1)`
	_laserPengingAll = `SELECT id,app_key,platform,mid,buvid,email,log_date,url,status,silence_url,
silence_status,operator,mobi_app,unix_timestamp(ctime),unix_timestamp(mtime) FROM app_laser WHERE status IN (-1,1,2)`
	_laser = `SELECT id,app_key,platform,mid,buvid,email,log_date,url,status,silence_url,silence_status,operator,mobi_app,
unix_timestamp(ctime),unix_timestamp(mtime) FROM app_laser WHERE id=?`
	_laserAllSilence = `SELECT id,app_key,platform,mid,buvid,email,log_date,url,status,silence_url,
silence_status,operator,unix_timestamp(ctime),unix_timestamp(mtime) FROM app_laser WHERE silence_status IN (1,-1)`
	_packLatestStable = `SELECT p.app_id,p.app_key,p.env,p.version_id,pv.version_code,p.internal_version_code,p.build_id,p.git_type,p.git_name,p.commit,
p.pack_type,p.steady_state,p.operator,p.size,p.md5,p.pack_path,p.pack_url,p.mapping_url,p.r_url,p.r_mapping_url,p.cdn_url,p.description,p.sender,
p.change_log,unix_timestamp(p.ctime),unix_timestamp(p.mtime) FROM pack AS p, pack_version AS pv WHERE p.version_id = pv.id AND p.app_key=? AND p.env="prod" AND p.steady_state=1 %s ORDER BY p.version_id DESC,p.build_id DESC limit 0,1`
	_bizApkListAll     = `SELECT a.name,b.id,b.pack_build_id,b.bundle_ver,b.md5,b.apk_cdn_url,b.env,s.priority FROM biz_apk_build AS b,biz_apk AS a,biz_apk_pack_settings AS s WHERE b.biz_apk_id=a.id AND b.state=0 AND b.biz_apk_id=s.biz_apk_id AND b.pack_build_id=s.pack_build_id AND b.env=s.env AND s.active=1`
	_bizApkFilterAll   = "SELECT env,biz_apk_build_id,network,isp,channel,city,percent,salt,device,state,excludes_system FROM biz_apk_build_filter"
	_bizApkFlowAll     = "SELECT env,biz_apk_build_id,flow FROM biz_apk_flow"
	_publish_generates = `SELECT id,app_key,build_id,channel_id,name,folder,patch_path,patch_url,cdn_url,status,size,md5,operator,channel_test_state,unix_timestamp(ctime),unix_timestamp(mtime) FROM pack_generate WHERE app_key=? AND build_id=? AND status>=1`

	_tfAppInfos       = `SELECT at.app_key,a.mobi_app,at.store_app_id,public_link,public_link_test FROM app_tf_attr AS at, app AS a, app_attribute AS aa WHERE aa.app_table_id=a.id AND aa.state=1 AND at.app_key=a.app_key`
	_latestOnline     = `SELECT version,version_code,remind_upd_txt,force_upd_txt FROM pack AS p, pack_version AS pv, pack_tf_attr AS pt WHERE p.version_id=pv.id AND p.id=pt.pack_id AND pack_state='VALID' AND pt.app_key=? AND p.pack_type=4 AND p.env='prod' ORDER BY version_code DESC LIMIT 1`
	_latestTF         = `SELECT version,version_code,dis_permil,guide_tf_txt,remind_upd_txt,force_upd_txt FROM pack AS p, pack_version AS pv, pack_tf_attr AS pt WHERE p.version_id=pv.id AND p.id=pt.pack_id AND pack_state='VALID' AND pt.app_key=? AND p.pack_type=9 AND p.env=? AND dis_permil > 0 ORDER BY version_code DESC LIMIT 1`
	_tfPackList       = `SELECT version,version_code,remind_upd_time,force_upd_time FROM pack AS p, pack_version AS pv, pack_tf_attr AS pt WHERE p.version_id=pv.id AND p.id=pt.pack_id AND pack_state='VALID' AND review_state='APPROVED' AND p.pack_type=9 AND p.env=? AND pt.expire_time > CURRENT_TIMESTAMP AND pt.app_key=? ORDER BY version_code DESC`
	_upPatchStatus    = `UPDATE patch SET status=?,gl_job_id=? WHERE id=? AND app_key=?`
	_patchInfo        = `SELECT app_key,origin_version_id,target_version_id,origin_build_id,build_id,status FROM patch WHERE id=? AND app_key=? LIMIT 1`
	_patchAppKey      = `SELECT id,app_key,origin_version_id,target_version_id,origin_build_id,target_build_id,build_id,status,patch_path,patch_state,ctime,mtime FROM patch WHERE app_key=? AND patch_state=?`
	_updatePatchState = `UPDATE patch SET patch_state=? WHERE id IN (%s)`
	_upPatchFileInfo  = `UPDATE patch SET size=?,md5=?,patch_path=?,patch_url=?,cdn_url=? WHERE id=?`

	_tribeListAll         = `SELECT t.name,t.no_host,t.priority,tp.tribe_id,tp.id,tp.app_key,tp.dep_gl_job_id,tp.md5,tp.cdn_url,tp.env,tpv.version_code,tp.dep_feature FROM tribe_pack AS tp,tribe AS t,tribe_pack_version AS tpv WHERE tp.tribe_id=t.id AND tp.version_id=tpv.id AND tp.env=tpv.env AND tpv.is_active=1 ORDER BY tp.id DESC`
	_tribeFilterAll       = "SELECT tribe_id,tribe_pack_id,env,network,isp,channel,city,percent,salt,device,type,excludes_system FROM tribe_config_filter"
	_tribeHostRelationAll = "SELECT current_build_id,parent_build_id,feature,app_key FROM tribe_host_relations ORDER BY id"
	_tribes               = "SELECT id,app_key,name,no_host FROM tribe"
	_tribeUpgradeAll      = "SELECT id,tribe_id,env,tribe_pack_id,start_version_code,chosen_version_code FROM tribe_pack_upgrade"

	// pcdn_files
	_addPcdnFile               = "INSERT INTO pcdn_files (rid,url,md5,size,business,version_id) VALUES (?,?,?,?,?,?)"
	_batchAddPcdnFile          = "INSERT INTO pcdn_files (rid,url,md5,size,business,version_id) VALUES %s ON DUPLICATE KEY UPDATE id=id"
	_pcdnFileByVersion         = "SELECT id,url,md5,size,business,rid,version_id,ctime,mtime FROM pcdn_files WHERE version_id>?" // pcdn.Files
	_pcdnRidAll                = "SELECT rid FROM pcdn_files"                                                                    // pcdn.Files
	_addPcdnQueryLog           = "INSERT INTO pcdn_query_log (version_id,zone) VALUES (?,?)"
	_PcdnQueryLogLatestVersion = "SELECT MAX(version_id) FROM pcdn_query_log WHERE zone=?"
)

// NewestConfigVersion get newest config version.
func (d *Dao) NewestConfigVersion(c context.Context) (res map[string]map[string]int64, err error) {
	rows, err := d.db.Query(c, _newestConfigPublishVersion)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[string]int64)
	for rows.Next() {
		var (
			appkey, env string
			id          int64
		)
		if err = rows.Scan(&appkey, &env, &id); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[string]int64
			ok bool
		)
		if re, ok = res[env]; !ok {
			re = make(map[string]int64)
			res[env] = re
		}
		re[appkey] = id
	}
	err = rows.Err()
	return
}

// NewestFFVersion get newest FF version.
func (d *Dao) NewestFFVersion(c context.Context) (res map[string]map[string]int64, err error) {
	rows, err := d.db.Query(c, _newestFFPublishVersion)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[string]int64)
	for rows.Next() {
		var (
			appkey, env string
			id          int64
		)
		if err = rows.Scan(&appkey, &env, &id); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[string]int64
			ok bool
		)
		if re, ok = res[env]; !ok {
			re = make(map[string]int64)
			res[env] = re
		}
		re[appkey] = id
	}
	err = rows.Err()
	return
}

// VersionAll get all version.
func (d *Dao) VersionAll(c context.Context) (res map[string]map[int64]*model.Version, err error) {
	rows, err := d.db.Query(c, _versionAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[int64]*model.Version)
	for rows.Next() {
		version := &model.Version{}
		if err = rows.Scan(&version.ID, &version.AppID, &version.AppKey, &version.Env, &version.Version, &version.VersionCode); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[int64]*model.Version
			ok bool
		)
		key := fmt.Sprintf("%v_%v", version.AppKey, version.Env)
		if re, ok = res[key]; !ok {
			re = make(map[int64]*model.Version)
			res[key] = re
		}
		re[version.ID] = version
	}
	err = rows.Err()
	return
}

// AppCDVersionList get version list.
func (d *Dao) AppCDVersionList(c context.Context, appKey, env string) (res []*model.Version, err error) {
	rows, err := d.db.Query(c, _releaseVersionList, appKey, env)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		version := &model.Version{}
		if err = rows.Scan(&version.ID, &version.AppID, &version.AppKey, &version.Env, &version.Version, &version.VersionCode); err != nil {
			log.Error("%v", err)
			return nil, err
		}
		res = append(res, version)
	}
	return res, rows.Err()
}

// UpgradConfigAll get all upgrad config.
func (d *Dao) UpgradConfigAll(c context.Context) (res map[string]map[int64]*cdmdl.UpgradConfig, err error) {
	rows, err := d.db.Query(c, _upgradConfigAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[int64]*cdmdl.UpgradConfig)
	for rows.Next() {
		uc := &cdmdl.UpgradConfig{}
		if err = rows.Scan(&uc.AppKey, &uc.Env, &uc.VersionID, &uc.Normal, &uc.Force, &uc.ExNormal, &uc.ExForce,
			&uc.System, &uc.ExcludeSystem, &uc.Cycle, &uc.Title, &uc.Content, &uc.IsSilent, &uc.Policy, &uc.PolicyURL, &uc.IconURL, &uc.ConfirmBtnText, &uc.CancelBtnText); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[int64]*cdmdl.UpgradConfig
			ok bool
		)
		key := fmt.Sprintf("%v_%v", uc.AppKey, uc.Env)
		if re, ok = res[key]; !ok {
			re = make(map[int64]*cdmdl.UpgradConfig)
			res[key] = re
		}
		re[uc.VersionID] = uc
	}
	err = rows.Err()
	return
}

// PackAll get all pack.
func (d *Dao) PackAll(c context.Context) (res map[string]map[int64][]*cdmdl.Pack, err error) {
	rows, err := d.db.Query(c, _packAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[int64][]*cdmdl.Pack)
	for rows.Next() {
		pack := &cdmdl.Pack{}
		if err = rows.Scan(&pack.AppID, &pack.AppKey, &pack.Env, &pack.VersionID, &pack.BuildID, &pack.Size, &pack.MD5,
			&pack.PackURL, &pack.CDNURL, &pack.PTime); err != nil {
			log.Error("%v", err)
			return
		}
		key := fmt.Sprintf("%v_%v", pack.AppKey, pack.Env)
		var (
			re map[int64][]*cdmdl.Pack
			ok bool
		)
		if re, ok = res[key]; !ok {
			re = make(map[int64][]*cdmdl.Pack)
			res[key] = re
		}
		re[pack.VersionID] = append(re[pack.VersionID], pack)
	}
	err = rows.Err()
	return
}

// PackLatestStable get latest stable pack
func (d *Dao) PackLatestStable(c context.Context, appKey string, versionCode int) (pack *cdmdl.Pack, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if versionCode != 0 {
		args = append(args, versionCode)
		sqlAdd = "AND pv.version_code<=?"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_packLatestStable, sqlAdd), args...)
	pack = &cdmdl.Pack{}
	if err = row.Scan(&pack.AppID, &pack.AppKey, &pack.Env, &pack.VersionID, &pack.VersionCode, &pack.InternalVersionCode, &pack.BuildID,
		&pack.GitType, &pack.GitName, &pack.Commit, &pack.PackType, &pack.SteadyState, &pack.Operator, &pack.Size,
		&pack.MD5, &pack.PackPath, &pack.PackURL, &pack.MappingURL, &pack.RURL, &pack.RMappingURL, &pack.CDNURL,
		&pack.Desc, &pack.Sender, &pack.ChangeLog, &pack.CTime, &pack.MTime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			pack = nil
		} else {
			log.Error("PackLatestStable %v", err)
		}
	}
	return

}

// PatchAll get all patch.
func (d *Dao) PatchAll(c context.Context) (res map[string]map[int64]*cdmdl.Patch, err error) {
	rows, err := d.db.Query(c, _patchAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[int64]*cdmdl.Patch)
	for rows.Next() {
		patch := &cdmdl.Patch{}
		if err = rows.Scan(&patch.AppKey, &patch.BuildID, &patch.TargetBuildID, &patch.TargetVersionID, &patch.OriginBuildID,
			&patch.OriginVersionID, &patch.Size, &patch.MD5, &patch.CDNURL); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[int64]*cdmdl.Patch
			ok bool
		)
		key := fmt.Sprintf("%v", patch.AppKey)
		if re, ok = res[key]; !ok {
			re = make(map[int64]*cdmdl.Patch)
			res[key] = re
		}
		if _, ok = re[patch.BuildID]; ok {
			continue
		}
		re[patch.BuildID] = patch
	}
	err = rows.Err()
	return
}

// PatchAll2 get all patch.
func (d *Dao) PatchAll2(c context.Context) (res map[string]map[string]*cdmdl.Patch, err error) {
	rows, err := d.db.Query(c, _patchAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[string]*cdmdl.Patch)
	for rows.Next() {
		patch := &cdmdl.Patch{}
		if err = rows.Scan(&patch.AppKey, &patch.BuildID, &patch.TargetBuildID, &patch.TargetVersionID, &patch.OriginBuildID,
			&patch.OriginVersionID, &patch.Size, &patch.MD5, &patch.CDNURL); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[string]*cdmdl.Patch
			ok bool
		)
		if re, ok = res[patch.AppKey]; !ok {
			re = make(map[string]*cdmdl.Patch)
			res[patch.AppKey] = re
		}
		key2 := fmt.Sprintf("%v_%v", patch.TargetBuildID, patch.OriginBuildID)
		re[key2] = patch
	}
	err = rows.Err()
	return
}

// PatchAppKeys get patch app_key
func (d *Dao) PatchAppKeys(c context.Context) (res []string, err error) {
	rows, err := d.db.Query(c, _patchAppKeys)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var appKey string
		if err = rows.Scan(&appKey); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, appKey)
	}
	err = rows.Err()
	return
}

// LastPackVersionIds get last 10 pack version ids
func (d *Dao) LastPackVersionIds(c context.Context, appKey, env string) (res []int64, err error) {
	rows, err := d.db.Query(c, _lastPackVersionIds, appKey, env)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var versionId int64
		if err = rows.Scan(&versionId); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, versionId)
	}
	err = rows.Err()
	return
}

// PackBuildIdsByVersions get pack buildIds by versionIds
func (d *Dao) PackBuildIdsByVersions(c context.Context, appKey, env string, versionIds []int64) (res []int64, err error) {
	var (
		sqls []string
		args []interface{}
	)
	args = append(args, appKey, env)
	for _, versionId := range versionIds {
		sqls = append(sqls, "?")
		args = append(args, versionId)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_packBuildIdsByVersions, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var buildId int64
		if err = rows.Scan(&buildId); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, buildId)
	}
	err = rows.Err()
	return
}

// PatchAll4 get patch by buildId.
func (d *Dao) PatchAll4(c context.Context, buildIds []int64) (res map[string]map[string]*cdmdl.Patch, err error) {
	var patch []*cdmdl.Patch
	if patch, err = d.PatchAll3(c, buildIds); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	res = make(map[string]map[string]*cdmdl.Patch)
	for _, p := range patch {
		var (
			re map[string]*cdmdl.Patch
			ok bool
		)
		if re, ok = res[p.AppKey]; !ok {
			re = make(map[string]*cdmdl.Patch)
			res[p.AppKey] = re
		}
		key2 := fmt.Sprintf("%v_%v", p.TargetBuildID, p.OriginBuildID)
		re[key2] = p
	}
	return
}

// PatchAll3 get patch by buildId.
func (d *Dao) PatchAll3(c context.Context, buildIds []int64) (res []*cdmdl.Patch, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, buildId := range buildIds {
		sqls = append(sqls, "?")
		args = append(args, buildId)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_patchAll4, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		patch := &cdmdl.Patch{}
		if err = rows.Scan(&patch.ID, &patch.AppKey, &patch.BuildID, &patch.TargetBuildID, &patch.TargetVersionID, &patch.OriginBuildID,
			&patch.OriginVersionID, &patch.Size, &patch.MD5, &patch.CDNURL); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, patch)
	}
	err = rows.Err()
	return
}

// FilterConfigAll get all filter config.
func (d *Dao) FilterConfigAll(c context.Context) (res map[string]map[int64]*cdmdl.FilterConfig, err error) {
	rows, err := d.db.Query(c, _filterConfigAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[int64]*cdmdl.FilterConfig)
	for rows.Next() {
		fc := &cdmdl.FilterConfig{}
		if err = rows.Scan(&fc.AppKey, &fc.Env, &fc.BuildID, &fc.Network, &fc.ISP, &fc.Channel, &fc.City, &fc.Percent,
			&fc.Salt, &fc.Device, &fc.Status, &fc.PhoneModel, &fc.Brand); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[int64]*cdmdl.FilterConfig
			ok bool
		)
		key := fmt.Sprintf("%v_%v", fc.AppKey, fc.Env)
		if re, ok = res[key]; !ok {
			re = make(map[int64]*cdmdl.FilterConfig)
			res[key] = re
		}
		re[fc.BuildID] = fc
	}
	err = rows.Err()
	return
}

// AppChannelAll get all app channel.
func (d *Dao) AppChannelAll(c context.Context) (res map[string]map[int64]*appmdl.Channel, err error) {
	rows, err := d.db.Query(c, _appChannelAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[int64]*appmdl.Channel)
	for rows.Next() {
		channel := &appmdl.Channel{}
		if err = rows.Scan(&channel.AppKey, &channel.ID, &channel.Code, &channel.Name, &channel.Plate); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[int64]*appmdl.Channel
			ok bool
		)
		key := channel.AppKey
		if re, ok = res[key]; !ok {
			re = make(map[int64]*appmdl.Channel)
			res[key] = re
		}
		re[channel.ID] = channel
	}
	err = rows.Err()
	return
}

// FlowConfigAll get all flow config.
func (d *Dao) FlowConfigAll(c context.Context) (res map[string]map[int64]*cdmdl.FlowConfig, err error) {
	rows, err := d.db.Query(c, _flowConfigAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[int64]*cdmdl.FlowConfig)
	for rows.Next() {
		flow := &cdmdl.FlowConfig{}
		if err = rows.Scan(&flow.AppKey, &flow.Env, &flow.BuildID, &flow.Flow); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[int64]*cdmdl.FlowConfig
			ok bool
		)
		key := fmt.Sprintf("%v_%v", flow.AppKey, flow.Env)
		if re, ok = res[key]; !ok {
			re = make(map[int64]*cdmdl.FlowConfig)
			res[key] = re
		}
		re[flow.BuildID] = flow
	}
	err = rows.Err()
	return
}

// HotfixAll get all hotfix
func (d *Dao) HotfixAll(c context.Context) (res map[string]map[int64][]*appmdl.HfUpgrade, err error) {
	rows, err := d.db.Query(c, _hotfixAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[int64][]*appmdl.HfUpgrade)
	for rows.Next() {
		hf := &appmdl.HfUpgrade{}
		if err = rows.Scan(&hf.AppID, &hf.AppKey, &hf.Env, &hf.GlPrjID, &hf.GlJobID, &hf.OrigVersion, &hf.OrigVersionCode, &hf.OrigBuildID,
			&hf.BuildID, &hf.GitType, &hf.GitName, &hf.Commit, &hf.Size, &hf.Md5, &hf.HotfixPath, &hf.HotfixURL, &hf.CDNURL, &hf.Description,
			&hf.Status, &hf.State); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[int64][]*appmdl.HfUpgrade
			ok bool
		)
		key := fmt.Sprintf("%v_%v", hf.AppKey, hf.Env)
		if re, ok = res[key]; !ok {
			re = make(map[int64][]*appmdl.HfUpgrade)
			res[key] = re
		}
		re[hf.OrigBuildID] = append(re[hf.OrigBuildID], hf)
	}
	err = rows.Err()
	return
}

// HotfixConfigAll get all hotfix config
func (d *Dao) HotfixConfigAll(c context.Context) (res map[string]map[int64]*appmdl.HotfixConfig, err error) {
	rows, err := d.db.Query(c, _hotfixConfigAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[string]map[int64]*appmdl.HotfixConfig)
	for rows.Next() {
		hfc := &appmdl.HotfixConfig{}
		if err = rows.Scan(&hfc.AppKey, &hfc.Env, &hfc.BuildID, &hfc.Device, &hfc.Channel, &hfc.City, &hfc.UpgradNum, &hfc.Gray, &hfc.Effect); err != nil {
			log.Error("%v", err)
			return
		}
		var (
			re map[int64]*appmdl.HotfixConfig
			ok bool
		)
		key := fmt.Sprintf("%v_%v", hfc.AppKey, hfc.Env)
		if re, ok = res[key]; !ok {
			re = make(map[int64]*appmdl.HotfixConfig)
			res[key] = re
		}
		re[hfc.BuildID] = hfc
	}
	err = rows.Err()
	return
}

// LaserAll get all laser.
func (d *Dao) LaserAll(c context.Context) (res []*appmdl.Laser, err error) {
	rows, err := d.db.Query(c, _laserAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &appmdl.Laser{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Platform, &re.MID, &re.Buvid, &re.Email, &re.LogDate, &re.URL,
			&re.Status, &re.SilenceURL, &re.SilenceStatus, &re.Operator, &re.MobiApp, &re.CTime, &re.MTime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// LaserPendingAll get all laser.
func (d *Dao) LaserPendingAll(c context.Context) (res []*appmdl.Laser, err error) {
	rows, err := d.db.Query(c, _laserPengingAll)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &appmdl.Laser{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Platform, &re.MID, &re.Buvid, &re.Email, &re.LogDate, &re.URL,
			&re.Status, &re.SilenceURL, &re.SilenceStatus, &re.Operator, &re.MobiApp, &re.CTime, &re.MTime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// Laser get laser info.
func (d *Dao) Laser(c context.Context, taskID int64) (re *appmdl.Laser, err error) {
	row := d.db.QueryRow(c, _laser, taskID)
	re = &appmdl.Laser{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.Platform, &re.MID, &re.Buvid, &re.Email, &re.LogDate, &re.URL, &re.Status,
		&re.SilenceURL, &re.SilenceStatus, &re.Operator, &re.MobiApp, &re.CTime, &re.MTime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("Laser %v", err)
		}
	}
	return
}

// LaserAllSilence get all laser.
func (d *Dao) LaserAllSilence(c context.Context) (res []*appmdl.Laser, err error) {
	rows, err := d.db.Query(c, _laserAllSilence)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &appmdl.Laser{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Platform, &re.MID, &re.Buvid, &re.Email, &re.LogDate, &re.URL,
			&re.Status, &re.SilenceURL, &re.SilenceStatus, &re.Operator, &re.CTime, &re.MTime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

func (d *Dao) BizApkListAll(c context.Context) ([]*bizapkmdl.Apk, error) {
	rows, err := d.db.Query(c, _bizApkListAll)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*bizapkmdl.Apk
	for rows.Next() {
		re := &bizapkmdl.Apk{}
		if err := rows.Scan(&re.Name, &re.BuildID, &re.PackBuildID, &re.BundleVer, &re.MD5, &re.ApkCdnURL, &re.Env, &re.Priority); err != nil {
			log.Error("%v", err)
			return nil, err
		}
		res = append(res, re)
	}
	return res, rows.Err()
}

func (d *Dao) BizApkFilterAll(c context.Context) ([]*bizapkmdl.FilterConfig, error) {
	rows, err := d.db.Query(c, _bizApkFilterAll)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*bizapkmdl.FilterConfig
	for rows.Next() {
		re := &bizapkmdl.FilterConfig{}
		if err := rows.Scan(&re.Env, &re.BuildID, &re.Network, &re.ISP, &re.Channel, &re.City, &re.Percent, &re.Salt, &re.Device, &re.Status, &re.ExcludesSystem); err != nil {
			log.Error("%v", err)
			return nil, err
		}
		res = append(res, re)
	}
	return res, rows.Err()
}

func (d *Dao) BizApkFlowAll(c context.Context) ([]*bizapkmdl.FlowConfig, error) {
	rows, err := d.db.Query(c, _bizApkFlowAll)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*bizapkmdl.FlowConfig
	for rows.Next() {
		re := &bizapkmdl.FlowConfig{}
		if err := rows.Scan(&re.Env, &re.BuildID, &re.Flow); err != nil {
			log.Error("%v", err)
			return nil, err
		}
		res = append(res, re)
	}
	return res, rows.Err()
}

// TribePackListAll get all tribes
func (d *Dao) TribePackListAll(c context.Context) ([]*tribemdl.TribeApk, error) {
	rows, err := d.db.Query(c, _tribeListAll)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*tribemdl.TribeApk
	for rows.Next() {
		re := &tribemdl.TribeApk{}
		if err := rows.Scan(&re.Name, &re.Nohost, &re.Priority, &re.TribeID, &re.ID, &re.AppKey, &re.TribeHostJobID, &re.MD5, &re.ApkCdnURL, &re.Env, &re.BundleVer, &re.DepFeature); err != nil {
			log.Error("%v", err)
			return nil, err
		}
		res = append(res, re)
	}
	return res, rows.Err()
}

func (d *Dao) TribePackFilterAll(c context.Context) ([]*tribemdl.ConfigFilter, error) {
	rows, err := d.db.Query(c, _tribeFilterAll)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*tribemdl.ConfigFilter
	for rows.Next() {
		re := &tribemdl.ConfigFilter{}
		if err := rows.Scan(&re.TribeId, &re.TribePackId, &re.Env, &re.Network, &re.Isp, &re.Channel, &re.City, &re.Percent, &re.Salt, &re.Device, &re.Type, &re.ExcludesSystem); err != nil {
			log.Error("%v", err)
			return nil, err
		}
		res = append(res, re)
	}
	return res, rows.Err()
}

func (d *Dao) TribePackUpgradeAll(c context.Context) ([]*tribemdl.PackUpgrade, error) {
	rows, err := d.db.Query(c, _tribeUpgradeAll)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*tribemdl.PackUpgrade
	for rows.Next() {
		re := &tribemdl.PackUpgrade{}
		if err := rows.Scan(&re.Id, &re.TribeId, &re.Env, &re.TribePackId, &re.StartVersionCode, &re.ChosenVersionCode); err != nil {
			log.Error("%v", err)
			return nil, err
		}
		res = append(res, re)
	}
	return res, rows.Err()
}

func (d *Dao) TribeHostRelationAll(c context.Context) ([]*tribemdl.TribeHostRelation, error) {
	rows, err := d.db.Query(c, _tribeHostRelationAll)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*tribemdl.TribeHostRelation
	for rows.Next() {
		re := &tribemdl.TribeHostRelation{}
		if err := rows.Scan(&re.CurrentBuildID, &re.ParentBuildID, &re.Feature, &re.AppKey); err != nil {
			log.Error("%v", err)
			return nil, err
		}
		res = append(res, re)
	}
	return res, rows.Err()
}

func (d *Dao) Tribes(c context.Context) ([]*tribemdl.Tribe, error) {
	rows, err := d.db.Query(c, _tribes)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*tribemdl.Tribe
	for rows.Next() {
		re := &tribemdl.Tribe{}
		if err := rows.Scan(&re.Id, &re.AppKey, &re.Name, &re.NoHost); err != nil {
			log.Error("%v", err)
			return nil, err
		}
		res = append(res, re)
	}
	return res, rows.Err()
}

// PublishGenerateList get Generate list.
func (d *Dao) PublishGenerateList(c context.Context, appKey string, buildID int64) (res []*cdmdl.Generate, err error) {
	rows, err := d.db.Query(c, _publish_generates, appKey, buildID)
	if err != nil {
		log.Error("PublishGenerateList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &cdmdl.Generate{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.BuildID, &re.ChannelID, &re.Name, &re.Folder, &re.GeneratePath, &re.GenerateURL, &re.CDNURL, &re.Status, &re.Size, &re.MD5, &re.Operator, &re.ChannelTestState, &re.CTime, &re.PTime); err != nil {
			log.Error("PublishGenerateList %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TFAppInfos TestFlight app infos
func (d *Dao) TFAppInfos(c context.Context) (res []*cdmdl.TFAppBaseInfo, err error) {
	rows, err := d.db.Query(c, _tfAppInfos)
	if err != nil {
		log.Error("TFAppInfos %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &cdmdl.TFAppBaseInfo{}
		if err = rows.Scan(&re.AppKey, &re.MobiApp, &re.StoreAppID, &re.PublicLink, &re.PublicLinkTest); err != nil {
			log.Error("TFAppInfos %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// LatestOnline Get latest online version
func (d *Dao) LatestOnline(c context.Context, appKey string) (res *cdmdl.TestFlightAttribute, err error) {
	row := d.db.QueryRow(c, _latestOnline, appKey)
	res = &cdmdl.TestFlightAttribute{}
	if err = row.Scan(&res.Version, &res.VersionCode, &res.RemindUpdTxt, &res.ForceUpdTxt); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("LatestOnline %v", err)

		}
		return
	}
	res.PackageType = "appstore"
	return
}

// LatestTF Get latest testflight version
func (d *Dao) LatestTF(c context.Context, appKey, env string) (res *cdmdl.TestFlightAttribute, err error) {
	row := d.db.QueryRow(c, _latestTF, appKey, env)
	res = &cdmdl.TestFlightAttribute{}
	if err = row.Scan(&res.Version, &res.VersionCode, &res.DisPermil, &res.GuideTFTxt, &res.RemindUpdTxt, &res.ForceUpdTxt); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("LatestTF %v", err)
		}
		return
	}
	res.PackageType = "testflight"
	return
}

// TFPackList Get Testflight Package list
func (d *Dao) TFPackList(c context.Context, appKey, env string) (res []*cdmdl.TestFlightUpdTimeInfo, err error) {
	rows, err := d.db.Query(c, _tfPackList, env, appKey)
	if err != nil {
		log.Error("TFPackList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &cdmdl.TestFlightUpdTimeInfo{}
		if err = rows.Scan(&re.Version, &re.VersionCode, &re.RemindUpdTime, &re.ForceUpdTime); err != nil {
			log.Error("TFPackList %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxPatchStatus set patch status
func (d *Dao) TxPatchStatus(tx *sql.Tx, id, glJobID int64, status int, appKey string) (err error) {
	_, err = tx.Exec(_upPatchStatus, status, glJobID, id, appKey)
	if err != nil {
		log.Error("TxPatchStatus %v", err)
	}
	return
}

// PatchInfo get patch info
func (d *Dao) PatchInfo(c context.Context, id int64, appKey string) (res *cdmdl.Patch, err error) {
	row := d.db.QueryRow(c, _patchInfo, id, appKey)
	res = &cdmdl.Patch{}
	if err = row.Scan(&res.AppKey, &res.OriginVersionID, &res.TargetVersionID, &res.OriginBuildID, &res.BuildID, &res.Status); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("PatchInfo %v", err)
		}
	}
	return
}

// PatchByAppKey get patch by appKey
func (d *Dao) PatchByAppKey(c context.Context, appKey string, active taskmdl.PackDeleteState) (res []*cdmdl.Patch, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, _patchAppKey, appKey, active); err != nil {
		log.Errorc(c, "select error: %v", err)
		return
	}
	if err = rows.Err(); err != nil {
		return
	}
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "scan error: %v", err)
		return
	}
	return
}

// UpdatePatchState 更新patch_state
func (d *Dao) UpdatePatchState(c context.Context, ids []int64, deleted taskmdl.PackDeleteState) (r int64, err error) {
	var (
		sqlAdd []string
		args   []interface{}
	)
	if len(ids) == 0 {
		return
	}
	args = append(args, deleted)
	for _, id := range ids {
		sqlAdd = append(sqlAdd, "?")
		args = append(args, id)
	}
	res, err := d.db.Exec(c, fmt.Sprintf(_updatePatchState, strings.Join(sqlAdd, ",")), args...)
	if err != nil {
		log.Errorc(c, "d.UpdatePatchState error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// TxUpPatchFileInfo set patch file info
func (d *Dao) TxUpPatchFileInfo(tx *sql.Tx, id, size int64, md5, localPath, inetPath, URL string) (err error) {
	_, err = tx.Exec(_upPatchFileInfo, size, md5, localPath, inetPath, URL, id)
	if err != nil {
		log.Error("TxUpPatchFileInfo %v", err)
	}
	return
}

// TxAddPcdnFile add pcdn file
func (d *Dao) TxAddPcdnFile(tx *sql.Tx, rid, url, md5, business, versionId string, size int64) (err error) {
	_, err = tx.Exec(_addPcdnFile, rid, url, md5, size, business, versionId)
	if err != nil {
		log.Error("TxAddPcdnFile %v", err)
	}
	return
}

// BatchAddPcdnFile add pcdn file
func (d *Dao) BatchAddPcdnFile(ctx context.Context, files []*pcdn.Files) (err error) {
	var (
		sqls = make([]string, 0, len(files))
		args = make([]interface{}, 0, len(files))
	)
	if len(files) == 0 {
		return
	}
	for _, f := range files {
		sqls = append(sqls, "(?,?,?,?,?,?)")
		args = append(args, f.Rid, f.Url, f.Md5, f.Size, f.Business, f.VersionId)
	}
	if _, err = d.db.Exec(ctx, fmt.Sprintf(_batchAddPcdnFile, strings.Join(sqls, ",")), args...); err != nil {
		log.Errorc(ctx, "BatchAddPcdnFile error: %v", err)
	}
	return
}

func (d *Dao) PcdnFileByVersion(ctx context.Context, versionId string) (items []*pcdn.Files, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _pcdnFileByVersion, versionId); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		if err == sql.ErrNoRows {
			return []*pcdn.Files{}, nil
		}
		return
	}
	if err = xsql.ScanSlice(rows, &items); err != nil {
		return
	}
	return
}

func (d *Dao) PcdnRidAll(ctx context.Context) (ridSet *map[string]bool, err error) {
	var rows *sql.Rows
	rmap := make(map[string]bool)
	if rows, err = d.db.Query(ctx, _pcdnRidAll); err != nil {
		return
	}
	if err = rows.Err(); err != nil {
		if err == sql.ErrNoRows {
			return new(map[string]bool), nil
		}
		return
	}
	for rows.Next() {
		var r string
		if err = rows.Scan(&r); err != nil {
			return nil, err
		}
		rmap[r] = true
	}
	ridSet = &rmap
	return
}

func (d *Dao) AddPcdnQueryLog(ctx context.Context, versionId, zone string) (id int64, err error) {
	row, err := d.db.Exec(ctx, _addPcdnQueryLog, versionId, zone)
	if err != nil {
		return
	}
	return row.LastInsertId()
}

func (d *Dao) LatestPcdnQueryLog(ctx context.Context, zone string) (latestVer string, err error) {
	row := d.db.QueryRow(ctx, _PcdnQueryLogLatestVersion, zone)
	var ver dsql.NullString
	if err = row.Scan(&ver); err != nil {
		return "", err
	}
	latestVer = ver.String
	return
}
