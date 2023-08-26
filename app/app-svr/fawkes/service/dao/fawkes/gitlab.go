package fawkes

import (
	"context"
	"fmt"

	"go-common/library/database/sql"

	gitmdl "go-gateway/app/app-svr/fawkes/service/model/gitlab"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	_gitlabProjIDSQL  = `SELECT a.git_prj_id FROM app AS a,app_attribute AS aa WHERE a.id=aa.app_table_id AND aa.state=1 AND a.app_key=?`
	_buildPackJobInfo = `SELECT gl_prj_id,gl_job_id FROM build_pack WHERE id=? AND state=0`
	_hotfixJobInfo    = `SELECT gl_prj_id,gl_job_id FROM hotfix WHERE id=? AND state=0`
	_getBizApkJobInfo = `SELECT git_prj_id,gl_ppl_id,gl_job_id FROM biz_apk_build,biz_apk,app,app_attribute WHERE biz_apk_build.id=? AND biz_apk.id=biz_apk_build.biz_apk_id AND biz_apk.app_key=app.app_key AND app.id=app_attribute.app_table_id AND app_attribute.state=1`
)

// GitlabProjectID get project id from gitlab
func (d *Dao) GitlabProjectID(c context.Context, appKey string) (gitlabPrjID string, err error) {
	row := d.db.QueryRow(c, _gitlabProjIDSQL, appKey)
	if err = row.Scan(&gitlabPrjID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("d.gitlabProjectID row.Scan error(%v)", err)
		}
	}
	return
}

// GitlabJobInfo get gitlab job info from bulid id
func (d *Dao) GitlabJobInfo(c context.Context, buildID int64) (jobInfo *gitmdl.BuildPackJobInfo, err error) {
	row := d.db.QueryRow(c, _buildPackJobInfo, buildID)
	jobInfo = &gitmdl.BuildPackJobInfo{}
	if err = row.Scan(&jobInfo.GitlabProjectID, &jobInfo.GitlabJobID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			jobInfo = nil
		} else {
			log.Error("d.GitlabJobInfo row.Scan error(%v)", err)
		}
	}
	return
}

// HotfixJobInfo get gitlab job info from bulid id
func (d *Dao) HotfixJobInfo(c context.Context, buildID int64) (jobInfo *gitmdl.BuildPackJobInfo, err error) {
	row := d.db.QueryRow(c, _hotfixJobInfo, buildID)
	jobInfo = &gitmdl.BuildPackJobInfo{}
	if err = row.Scan(&jobInfo.GitlabProjectID, &jobInfo.GitlabJobID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			jobInfo = nil
		} else {
			log.Error("d.GitlabJobInfo row.Scan error(%v)", err)
		}
	}
	return
}

// BizApkJobInfo get business apk job info from build id
func (d *Dao) BizApkJobInfo(c context.Context, buildID int64) (jobInfo *gitmdl.BuildPackJobInfo, err error) {
	row := d.db.QueryRow(c, _getBizApkJobInfo, buildID)
	jobInfo = &gitmdl.BuildPackJobInfo{}
	if err = row.Scan(&jobInfo.GitlabProjectID, &jobInfo.GitlabPipelineID, &jobInfo.GitlabJobID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			jobInfo = nil
		} else {
			log.Error("d.GitlabJobInfo row.Scan error(%v)", err)
		}
	}
	return
}

// GetFawkesToken get fawkes token
func (d *Dao) GetFawkesToken(c context.Context, gitlabProjectID string) (value string, err error) {
	key := fmt.Sprintf("fawkestoken:%s", gitlabProjectID)
	do, err := d.redis.Do(c, "GET", key)
	if err != nil {
		return
	}
	if do == nil {
		return
	}
	value = string(do.([]uint8))
	return
}

// SetFawkesToken set fawkes token
func (d *Dao) SetFawkesToken(c context.Context, gitlabProjectID string, value string) (err error) {
	key := fmt.Sprintf("fawkestoken:%s", gitlabProjectID)
	if _, err = d.redis.Do(c, "SET", key, value); err != nil {
		return err
	}
	if err != nil {
		return
	}
	return
}
