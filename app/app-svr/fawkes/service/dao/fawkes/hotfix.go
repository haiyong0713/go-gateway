package fawkes

import (
	"context"

	"go-common/library/database/sql"

	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_txAddOrUpdateHtConfig = `INSERT INTO hotfix_config (app_key,env,build_id,device,channel,city,upgrad_num,gray) VALUES(?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE device=?,channel=?,city=?,upgrad_num=?,gray=?`
	_hotfixConfGet         = `SELECT id,app_key,env,build_id,city,channel,upgrad_num,device,gray,effect FROM hotfix_config WHERE app_key=? AND env=? AND build_id=?`
	_txHfConfEnvUpdate     = `UPDATE hotfix_config SET env=? WHERE app_key=? AND build_id=?`
	_txHfEnvUpdate         = `UPDATE hotfix SET env=?,sender=? WHERE app_key=? AND build_id=?`
	_txHfConfUpdate        = `UPDATE hotfix_config SET channel=?,city=?,upgrad_num=?,device=?,gray=? WHERE app_key=? AND env=? AND build_id=?`
	_getHotfixID           = `SELECT id FROM hotfix WHERE app_key=? AND env=? AND build_id=? AND state=0`
	_getHotfixConfigID     = `SELECT id FROM hotfix_config WHERE app_key=? AND env=? AND build_id=?`
	_getHotfixInfo         = `SELECT app_id,app_key,env,origin_version,origin_version_code,origin_build_id,build_id,gl_job_id,internal_version_code,commit,operator FROM hotfix WHERE app_key=? AND build_id=?`
	_getSingleHotfix       = `SELECT app_id,app_key,env,origin_version,origin_version_code,origin_build_id,build_id,gl_job_id,internal_version_code,commit,operator FROM hotfix WHERE id=? AND state=0`
	_getHfJobRefresh       = `SELECT id,gl_prj_id,gl_job_id,status FROM hotfix WHERE state=0 AND status in (1,2) AND gl_job_id!=0`
	_pushHotfix            = `INSERT INTO hotfix(app_id,app_key,env,gl_prj_id,gl_job_id,origin_version,origin_version_code,origin_build_id,build_id,git_type,git_name,commit,operator,size,md5,hotfix_path,hotfix_url,cdn_url,description,status,sender,state,ctime,mtime,internal_version_code,env_vars) SELECT app_id,app_key,REPLACE(env,?,?),gl_prj_id,gl_job_id,origin_version,origin_version_code,origin_build_id,build_id,git_type,git_name,commit,operator,size,md5,hotfix_path,hotfix_url,cdn_url,description,status,sender,state,ctime,mtime,internal_version_code,env_vars FROM hotfix WHERE app_key=? AND build_id=?`
	_pushHotfixConf        = `INSERT INTO hotfix_config(app_key,env,build_id,device,channel,city,upgrad_num,gray,effect,ctime,mtime) SELECT app_key,REPLACE(env,?,?),build_id,device,channel,city,upgrad_num,gray,1,ctime,mtime FROM hotfix_config WHERE app_key=? AND build_id=?`
	_txHotfixEffet         = `UPDATE hotfix_config SET effect=? WHERE app_key=? AND env=? AND build_id=?`
	_getHfListCount        = `SELECT count(1) FROM hotfix WHERE app_key=? AND env=?`
	_getHfList             = `SELECT app_id,app_key,origin_version,origin_version_code,origin_build_id,build_id,hotfix_url,cdn_url,operator,status,size,md5,unix_timestamp(mtime),git_name,git_type,commit,gl_job_id,gl_prj_id,internal_version_code,env_vars FROM hotfix WHERE app_key=? AND env=? AND state=0 ORDER BY ctime DESC LIMIT ? OFFSET ?`
	_getHfOrigin           = `SELECT size,md5,unix_timestamp(mtime),pack_url,build_id,git_type,git_name,commit,internal_version_code FROM pack WHERE app_key=? AND env=? AND build_id=?`
	_getHfConfig           = `SELECT app_key,env,build_id,city,channel,upgrad_num,device,gray,effect FROM hotfix_config WHERE app_key=? AND env=? AND build_id=?`
	_txAddHotfixBuild      = `INSERT INTO hotfix(app_id,gl_prj_id,app_key,origin_build_id,git_type,git_name,env,origin_version,origin_version_code,internal_version_code,operator,status,env_vars) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_txHotfixBuildIDUpdate = `UPDATE hotfix SET build_id=? WHERE id=?`                          //更新build_id跟id一致
	_txHotfixUpdate        = `UPDATE hotfix SET gl_job_id=?,commit=?,status=2 WHERE build_id=?` //热修复包信息更新（pipeline 回传用）
	_txHotfixUploadInfo    = `UPDATE hotfix SET size=?,md5=?,hotfix_path=?,hotfix_url=?,cdn_url=?,status=3 WHERE app_key=? AND build_id=?`
	_txHotfixCancel        = `UPDATE hotfix SET status=? WHERE app_key=? AND build_id=?`
	_txHotfixUpdateStatus  = `UPDATE hotfix SET status=? WHERE id=?`
	_txHotfixDel           = `UPDATE hotfix SET state=1 WHERE app_key=? AND build_id=?`
	_getOriginURL          = `SELECT p.pack_path,p.pack_url,p.mapping_url,p.r_url,p.r_mapping_url FROM pack AS p, hotfix AS h WHERE h.app_key=? AND h.id=? AND h.origin_build_id=p.build_id`
	_getOriginInfo         = `SELECT a.version_id,b.version,b.version_code FROM pack AS a, pack_version AS b WHERE a.app_key=? AND a.env=? AND a.build_id=? AND a.version_id=b.id`
	_getLastProdHfInterVer = `SELECT internal_version_code from hotfix WHERE origin_version_code=? AND env="prod" AND app_key=? ORDER BY internal_version_code DESC LIMIT 1`
)

// TxAddOrUpdateHtConfig add hotfix config
func (d *Dao) TxAddOrUpdateHtConfig(tx *sql.Tx, appKey, env, channel, city, device string, buildID int64, upgradNum, gray int) (r int64, err error) {
	res, err := tx.Exec(_txAddOrUpdateHtConfig, appKey, env, buildID, device, channel, city, upgradNum, gray, device, channel, city, upgradNum, gray)
	if err != nil {
		log.Error("TxAddHotfixConfig INSERT OR UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// HotfixConfGet get hotfix config by app_key, env and build_id
func (d *Dao) HotfixConfGet(c context.Context, appKey, env string, buildID int64) (conf appmdl.HotfixConf, err error) {
	row := d.db.QueryRow(c, _hotfixConfGet, appKey, env, buildID)
	if err = row.Scan(&conf.ID, &conf.AppKey, &conf.Env, &conf.BuildID, &conf.City, &conf.Channel, &conf.UpgradNum,
		&conf.Device, &conf.Status, &conf.Effect); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("HotfixConfGet rows.Scan() failed. %v", err)
		}
	}
	return
}

// TxHfConfEnvUpdate update hotfix_config's env
func (d *Dao) TxHfConfEnvUpdate(tx *sql.Tx, appKey, env, buildID string) (r int64, err error) {
	res, err := tx.Exec(_txHfConfEnvUpdate, env, appKey, buildID)
	if err != nil {
		log.Error("TxHfConfEnvUpdate UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// TxHfEnvUpdate update hotfix's env
func (d *Dao) TxHfEnvUpdate(tx *sql.Tx, appKey, env, sender string, buildID int64) (r int64, err error) {
	res, err := tx.Exec(_txHfEnvUpdate, env, sender, appKey, buildID)
	if err != nil {
		log.Error("TxHfEnvUpdate UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// TxHfConfUpdate update hotfix_config
func (d *Dao) TxHfConfUpdate(tx *sql.Tx, appKey, env, channel, city, device string, buildID int64, upgradNum, gray int) (r int64, err error) {
	res, err := tx.Exec(_txHfConfUpdate, channel, city, upgradNum, device, gray, appKey, env, buildID)
	if err != nil {
		log.Error("TxHfConfUpdate UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// GetHotfixID get hotfix id by appKey, env and buildID
func (d *Dao) GetHotfixID(c context.Context, appKey, env string, buildID int64) (hotfixID int64, err error) {
	row := d.db.QueryRow(c, _getHotfixID, appKey, env, buildID)
	if err = row.Scan(&hotfixID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("GetHotfixID row.Scan() failed. %v", err)
		}
	}
	return
}

// GetHotfixConfigID get hotfix_config id by appKey, env and buildID
func (d *Dao) GetHotfixConfigID(c context.Context, appKey, env string, buildID int64) (hotfixConfigID int64, err error) {
	row := d.db.QueryRow(c, _getHotfixConfigID, appKey, env, buildID)
	if err = row.Scan(&hotfixConfigID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("GetHotfixID row.Scan() failed. %v", err)
		}
	}
	return
}

// GetHotfixInfo get hotfix information
func (d *Dao) GetHotfixInfo(c context.Context, appKey string, buildID int64) (hfInfos []*appmdl.HotfixInfo, err error) {
	rows, err := d.db.Query(c, _getHotfixInfo, appKey, buildID)
	if err != nil {
		log.Error("GetHotfixInfo SELECT failed. %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		hfInfo := &appmdl.HotfixInfo{}
		if err = rows.Scan(&hfInfo.AppID, &hfInfo.AppKey, &hfInfo.Env, &hfInfo.OrigVersion, &hfInfo.OrigVersionCode,
			&hfInfo.OrigBuildID, &hfInfo.BuildID, &hfInfo.GlJobID, &hfInfo.InternalVersionCode, &hfInfo.Commit, &hfInfo.Operator); err != nil {
			log.Error("GetHotfixInfo rows.Scan failed. %v", err)
			return
		}
		hfInfos = append(hfInfos, hfInfo)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	//err = rows.Err()
	//for i := 0; i < 2; i++ {
	//	hfInfo := &appmdl.HotfixInfo{}
	//	hfInfo.AppId = "233"
	//	hfInfo.AppKey = "233"
	//	hfInfo.Env = "test"
	//	hfInfo.OrigBuildId = 233
	//	hfInfo.OrigVersionCode = "2333"
	//	hfInfo.OrigBuildId = 2333
	//	hfInfo.BuildId = 233
	//	hfInfos = append(hfInfos, hfInfo)
	//}
	return
}

// GetSingleHotfixInfo get single hotfix information
func (d *Dao) GetSingleHotfixInfo(c context.Context, patchBuildID int64) (hfInfo *appmdl.HotfixInfo, err error) {
	hfInfo = &appmdl.HotfixInfo{}
	row := d.db.QueryRow(c, _getSingleHotfix, patchBuildID)
	if err = row.Scan(&hfInfo.AppID, &hfInfo.AppKey, &hfInfo.Env, &hfInfo.OrigVersion, &hfInfo.OrigVersionCode,
		&hfInfo.OrigBuildID, &hfInfo.BuildID, &hfInfo.GlJobID, &hfInfo.InternalVersionCode, &hfInfo.Commit, &hfInfo.Operator); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("_getSingleHotfix row.Scan error(%v)", err)
		}
	}
	return
}

// GetHfJobRefresh get hotfix job info which should be refreshed
func (d *Dao) GetHfJobRefresh(c context.Context) (jobs []*appmdl.HotfixJobInfo, err error) {
	rows, err := d.db.Query(c, _getHfJobRefresh)
	if err != nil {
		log.Error("d.GetHfJobRefresh d.dbQuery() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var job = &appmdl.HotfixJobInfo{}
		if err = rows.Scan(&job.BuildPatchID, &job.GitlabProjectID, &job.GitlabJobID, &job.Status); err != nil {
			log.Error("d.GetHfJobRefresh rows.Scan error(%v)", err)
			return
		}
		jobs = append(jobs, job)
	}
	err = rows.Err()
	return
}

// PushHotfix push hotfix to nextEnv
func (d *Dao) PushHotfix(tx *sql.Tx, appKey, env, nextEnv string, buildID int64) (r int64, err error) {
	res, err := tx.Exec(_pushHotfix, env, nextEnv, appKey, buildID)
	if err != nil {
		log.Error("PushHotfix INSERT failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// PushHotfixConf push hotfix config to nextEnv
func (d *Dao) PushHotfixConf(tx *sql.Tx, appKey, env, nextEnv string, buildID int64) (r int64, err error) {
	res, err := tx.Exec(_pushHotfixConf, env, nextEnv, appKey, buildID)
	if err != nil {
		log.Error("PushHotfixConf INSERT failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// TxHotfixEffect update hotfix_config effect
func (d *Dao) TxHotfixEffect(tx *sql.Tx, appKey, env string, buildID int64, effect int) (r int64, err error) {
	res, err := tx.Exec(_txHotfixEffet, effect, appKey, env, buildID)
	if err != nil {
		log.Error("TxHotfixEffect UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// GetHotfixListCount get the ci listcount.
func (d *Dao) GetHotfixListCount(c context.Context, appKey, env string) (count int, err error) {
	rows := d.db.QueryRow(c, _getHfListCount, appKey, env)
	if err = rows.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("GetHotfixID row.Scan() failed. %v", err)
		}
	}
	return
}

// GetHotfixList get the ci list.
func (d *Dao) GetHotfixList(c context.Context, appKey, env string, pn, ps int, order, sort string) (items []*appmdl.HfListItem, err error) {
	rows, err := d.db.Query(c, _getHfList, appKey, env, ps, (pn-1)*ps)
	if err != nil {
		log.Error("GetHotfixList row.Scan() failed. %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item = &appmdl.HfListItem{}
		if err = rows.Scan(&item.AppID, &item.AppKey, &item.Version, &item.VersionCode, &item.OriginBuildID, &item.BuildID, &item.PatchURL, &item.CdnURL, &item.Operator, &item.Status,
			&item.Size, &item.Md5, &item.Mtime, &item.GitName, &item.GitType, &item.Commit, &item.GlJobID, &item.GlPrjID, &item.InternalVersionCode, &item.EnvVars); err != nil {
			log.Error("AppFollow %v", err)
			return
		}
		items = append(items, item)
	}
	err = rows.Err()
	return
}

// GetHotfixOrigin get the ci origin's list.
func (d *Dao) GetHotfixOrigin(c context.Context, appKey, env string, buildID int64) (origin appmdl.HfOrigin, err error) {
	row := d.db.QueryRow(c, _getHfOrigin, appKey, env, buildID)
	if err = row.Scan(&origin.Size, &origin.Md5, &origin.Mtime, &origin.PkgURL, &origin.BuildID, &origin.GitType, &origin.GitName, &origin.Commit, &origin.InternalVersionCode); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("GetHotfixOrigin %v", err)
		}
	}
	return
}

// GetHotfixConfig get the ci config's list.
func (d *Dao) GetHotfixConfig(c context.Context, appKey, env string, buildID int64) (config appmdl.HfConfig, err error) {
	row := d.db.QueryRow(c, _getHfConfig, appKey, env, buildID)
	if err = row.Scan(&config.AppKey, &config.Env, &config.BuildID, &config.City, &config.Channel, &config.UpgradNum, &config.Device, &config.Status, &config.Effect); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("GetHotfixConfig %v", err)
		}
	}
	return
}

// GetOriginInfo get origin version information
func (d *Dao) GetOriginInfo(c context.Context, appKey, env string, buildID int64) (res appmdl.HfOriginVersion, err error) {
	row := d.db.QueryRow(c, _getOriginInfo, appKey, env, buildID)
	if err = row.Scan(&res.VersionID, &res.Version, &res.VersionCode); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("GetOriginInfo %v", err)
		}
	}
	return
}

// TxAddHotfixBuild add hotfix build
func (d *Dao) TxAddHotfixBuild(tx *sql.Tx, appKey, appID, gitlabPrjID string, buildID, versionCode, internalVersionCode int64, gitType int, gitName, env, version, name, envVars string) (r int64, err error) {
	res, err := tx.Exec(_txAddHotfixBuild, appID, gitlabPrjID, appKey, buildID, gitType, gitName, env, version, versionCode, internalVersionCode, name, 1, envVars)
	if err != nil {
		log.Error("TxAddHotfixBuild INSERT failed. %v", err)
		return
	}
	return res.LastInsertId()
}

// TxHotfixBuildIDUpdate update hotfix's build_id as same to id
func (d *Dao) TxHotfixBuildIDUpdate(tx *sql.Tx, id int64) (r int64, err error) {
	res, err := tx.Exec(_txHotfixBuildIDUpdate, id, id)
	if err != nil {
		log.Error("_txHotfixBuildIDUpdate UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// TxHotfixUpdate update hotfix's env
func (d *Dao) TxHotfixUpdate(tx *sql.Tx, buildID int64, glJobID int64, commit string) (r int64, err error) {
	res, err := tx.Exec(_txHotfixUpdate, glJobID, commit, buildID)
	if err != nil {
		log.Error("TxHfConfEnvUpdate UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// TxHotfixUpload upload hotfix's pipeline return
func (d *Dao) TxHotfixUpload(tx *sql.Tx, size, buildID int64, md5, hfPath, hfURL, cdnURL, appKey string) (r int64, err error) {
	res, err := tx.Exec(_txHotfixUploadInfo, size, md5, hfPath, hfURL, cdnURL, appKey, buildID)
	if err != nil {
		log.Error("TxHotfixUpload UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// TxHotfixCancel cancel the hotfix's pipeline
func (d *Dao) TxHotfixCancel(tx *sql.Tx, status int, appKey string, buildID int64) (r int64, err error) {
	res, err := tx.Exec(_txHotfixCancel, status, appKey, buildID)
	if err != nil {
		log.Error("TxHotfixUpload UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// TxHotfixUpdateStatus update the hotfix's status
func (d *Dao) TxHotfixUpdateStatus(tx *sql.Tx, patchBuildID int64, status int) (r int64, err error) {
	res, err := tx.Exec(_txHotfixUpdateStatus, status, patchBuildID)
	if err != nil {
		log.Error("TxHotfixUpdateStatus UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// GetOriginURL get hotfix origin package's URL information
func (d *Dao) GetOriginURL(c context.Context, appKey string, patchID int64) (originInfo appmdl.HfOrigURLInfo, err error) {
	row := d.db.QueryRow(c, _getOriginURL, appKey, patchID)
	if err = row.Scan(&originInfo.PackPath, &originInfo.PackURL, &originInfo.MappingURL, &originInfo.RURL, &originInfo.RMappingURL); err != nil {
		if err != sql.ErrNoRows {
			log.Error("GetOriginURL SELECT failed. %v", err)
		}
	}
	return
}

// TxHotfixDel cancel the hotfix's pipeline
func (d *Dao) TxHotfixDel(tx *sql.Tx, appKey string, buildID int64) (r int64, err error) {
	res, err := tx.Exec(_txHotfixDel, appKey, buildID)
	if err != nil {
		log.Error("TxHotfixDel UPDATE failed. %v", err)
		return
	}
	return res.RowsAffected()
}

// GetLastProdHfInterVer get last hotfix internal versioncode in Production env
func (d *Dao) GetLastProdHfInterVer(c context.Context, appKey string, orgVersionCode int64) (hfInternalVersionCode int64, err error) {
	row := d.db.QueryRow(c, _getLastProdHfInterVer, orgVersionCode, appKey)
	if err = row.Scan(&hfInternalVersionCode); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			hfInternalVersionCode = 0
		}
	}
	return
}
