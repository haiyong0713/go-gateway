package fawkes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	xsql "go-common/library/database/sql"
	xxsql "go-common/library/database/xsql"
	"go-common/library/ecode"

	"go-gateway/app/app-svr/fawkes/service/model"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	taskmdl "go-gateway/app/app-svr/fawkes/service/model/task"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/pkg/errors"
)

const (
	_packVersionByAppkey       = `SELECT id,app_key,version,version_code,is_upgrade FROM pack_version WHERE app_key=? AND env=? %s ORDER BY version_code DESC LIMIT ?,?`
	_packVersionByAppKeys      = `SELECT id,app_key,version,version_code,is_upgrade FROM pack_version WHERE app_key IN(%s) AND env=? ORDER BY version_code DESC`
	_packVersionCount          = `SELECT count(*) FROM pack_version WHERE app_key=? AND env=?`
	_packVersionList           = `SELECT id,version,version_code,is_upgrade FROM pack_version WHERE app_key=? %s ORDER BY version_code DESC LIMIT ?,?`
	_setVersion                = `INSERT INTO pack_version (app_id,app_key,env,version,version_code,is_upgrade) VALUES(?,?,?,?,?,?)`
	_packVersionByID           = `SELECT id,app_id,app_key,env,version,version_code,is_upgrade,unix_timestamp(ctime),unix_timestamp(mtime) FROM pack_version WHERE app_key=? AND id=?`
	_packVersion               = `SELECT id,app_id,app_key,env,version,version_code,is_upgrade FROM pack_version WHERE app_key=? AND env=? AND version=? AND version_code=?`
	_setPackConfigSwitch       = `UPDATE pack_version SET is_upgrade=? WHERE id=?`
	_packVersionCountByOptions = `SELECT count(*) FROM pack AS p,pack_version AS pv WHERE pv.app_key=? AND pv.env=? AND p.version_id=pv.id %s`
	_packVersionListByOptions  = `SELECT pv.id,pv.version,pv.version_code,pv.is_upgrade FROM pack AS p,pack_version AS pv WHERE pv.app_key=? AND pv.env=? AND p.version_id=pv.id %s`

	_packByVersion     = `SELECT id,app_id,app_key,env,version_id,internal_version_code,build_id,git_type,git_name,commit,pack_type,operator,size,md5,pack_path,pack_url,mapping_url,r_url,r_mapping_url,cdn_url,description,sender,steady_state,change_log,dep_gl_job_id,is_compatible,unix_timestamp(ctime),unix_timestamp(mtime),mtime FROM pack WHERE app_key=? AND env=? AND version_id=? ORDER BY id DESC`
	_packByVersions    = `SELECT id,app_id,app_key,env,version_id,internal_version_code,build_id,git_type,git_name,commit,pack_type,operator,size,md5,pack_path,pack_url,mapping_url,r_url,r_mapping_url,cdn_url,description,sender,steady_state,change_log,dep_gl_job_id,is_compatible,bbr_url,unix_timestamp(ctime),unix_timestamp(mtime),mtime FROM pack WHERE app_key=? %s AND version_id IN(%s) ORDER BY id DESC`
	_packByBuildID     = `SELECT p.id,p.app_id,p.app_key,p.env,p.version_id,p.internal_version_code,p.build_id,p.git_type,p.git_name,p.commit,p.pack_type,p.operator,p.size,p.md5,p.pack_path,p.pack_url,p.mapping_url,p.r_url,p.r_mapping_url,p.cdn_url,p.description,p.sender,p.steady_state,p.change_log,p.dep_gl_job_id,p.is_compatible,p.bbr_url,p.features,unix_timestamp(p.ctime),unix_timestamp(p.mtime),p.mtime,pv.version,pv.version_code FROM pack AS p,pack_version AS pv WHERE p.app_key=? AND p.env=? AND p.build_id=? AND p.version_id=pv.id ORDER BY p.id DESC`
	_packByGlJobId     = `SELECT build_id,env,steady_state FROM pack WHERE app_key=? AND build_id IN (%s)`
	_upPackVersionID   = `UPDATE pack SET version_id=? WHERE app_key=? AND id=?`
	_upPackSteadyState = `UPDATE pack SET steady_state=?,description=? WHERE app_key=? AND build_id=?`
	_upPackSender      = `UPDATE pack SET sender=? WHERE app_key=? AND build_id=? AND env=?`
	_setPack           = `INSERT INTO pack (app_id,app_key,env,version_id,internal_version_code,build_id,git_type,git_name,commit,pack_type,operator,size,md5,pack_path,pack_url,mapping_url,r_url,r_mapping_url,cdn_url,description,sender,change_log,dep_gl_job_id,is_compatible,bbr_url,features) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_lastPack          = `SELECT p.id,p.app_id,p.app_key,p.env,p.version_id,pv.version_code,p.internal_version_code,p.build_id,p.git_type,p.git_name,p.commit,p.pack_type,p.operator,p.size,p.md5,p.pack_path,p.pack_url,p.mapping_url,p.r_url,p.r_mapping_url,p.cdn_url,p.description,p.sender,p.change_log,unix_timestamp(p.ctime),unix_timestamp(p.mtime),p.mtime FROM pack AS p, pack_version AS pv WHERE p.env='prod' AND p.version_id=pv.id AND p.app_key=? AND pv.version_code<? AND p.steady_state=? ORDER BY pv.version_code DESC LIMIT ?`
	_packUpgrad        = "SELECT app_key,env,version_id,normal_upgrad,force_upgrad,exclude_normal,exclude_force,`system`,exclude_system,cycle,title,content,is_silent,policy,policy_url,icon_url, confirm_btn_text,cancel_btn_text FROM pack_upgrad WHERE app_key=? AND env=? AND version_id=?"
	_setPackUpgrad     = "INSERT INTO pack_upgrad (app_key,env,version_id,normal_upgrad,force_upgrad,exclude_normal,exclude_force,`system`,exclude_system,cycle,title,content,is_silent,policy,policy_url,icon_url,confirm_btn_text,cancel_btn_text) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE normal_upgrad=?,force_upgrad=?,exclude_normal=?,exclude_force=?,`system`=?,exclude_system=?,cycle=?,title=?,content=?,is_silent=?,policy=?,policy_url=?,icon_url=?,confirm_btn_text=?,cancel_btn_text=?"
	_packFilter        = `SELECT app_key,env,build_id,network,isp,channel,city,percent,salt,device,status,phone_model,brand FROM pack_filter WHERE app_key=? AND env=? AND build_id IN (%v)`
	_setPackFilter     = `INSERT INTO pack_filter (app_key,env,build_id,network,isp,channel,city,percent,salt,device,status,phone_model,brand) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE network=?,isp=?,channel=?,city=?,percent=?,device=?,status=?,phone_model=?,brand=?`

	_packFlow    = `SELECT app_key,env,build_id,flow,unix_timestamp(ctime),unix_timestamp(mtime) FROM pack_flow WHERE app_key=? AND env=? AND build_id IN (%v)`
	_setPackFlow = `INSERT INTO pack_flow (app_key,env,build_id,flow) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE flow=?`

	_patchCount = `SELECT count(*) FROM patch AS p,pack_version AS pv1,pack_version AS pv2 WHERE p.app_key=? AND p.build_id=? AND p.target_version_id=pv1.id AND p.origin_version_id=pv2.id`
	_patchs     = `SELECT p.id,p.app_key,p.build_id,p.target_build_id,p.target_version_id,pv1.version_code AS target_version_code,pv1.version AS target_version,p.origin_build_id,p.origin_version_id,pv2.version_code AS origin_version_code,pv2.version AS origin_version,p.size,p.status,p.gl_job_id,p.md5,p.pack_url,p.patch_path,p.patch_url,p.cdn_url,p.ctime,p.mtime FROM patch AS p,pack_version AS pv1,pack_version AS pv2 WHERE p.app_key=? AND p.build_id=? AND p.target_version_id=pv1.id AND p.origin_version_id=pv2.id LIMIT ?,?`
	_addPatch   = `INSERT INTO patch (app_key,build_id,target_build_id,target_version_id,origin_build_id,origin_version_id,size,status,md5,patch_path,patch_url,cdn_url,pack_url) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`

	_generateCount              = `SELECT count(*) FROM pack_generate WHERE app_key=? AND build_id=?`
	_generates                  = `SELECT id,app_key,build_id,channel_id,name,folder,patch_path,patch_url,cdn_url,status,size,md5,operator,channel_test_state,unix_timestamp(ctime),unix_timestamp(mtime),gl_job_url FROM pack_generate WHERE app_key=? AND build_id=?`
	_generatesByIDs             = `SELECT id,app_key,build_id,channel_id,name,folder,patch_path,patch_url,cdn_url,status,size,md5,operator,channel_test_state,unix_timestamp(ctime),unix_timestamp(mtime),gl_job_url FROM pack_generate WHERE id IN (%v)`
	_generate                   = `SELECT id,app_key,build_id,channel_id,name,folder,patch_path,patch_url,cdn_url,status,size,md5,operator,channel_test_state,unix_timestamp(ctime),unix_timestamp(mtime) FROM pack_generate WHERE app_key=? AND id=?`
	_generatesByAppKeyAndStatus = `SELECT id,app_key,build_id,channel_id,name,folder,patch_path,patch_url,cdn_url,status,size,md5,operator,channel_test_state,pack_state,unix_timestamp(ctime),unix_timestamp(mtime) FROM pack_generate WHERE app_key=? AND status IN (%s) AND pack_state=?`
	_generateByOptions          = `SELECT id,app_key,build_id,channel_id,name,folder,patch_path,patch_url,cdn_url,status,size,md5,operator,channel_test_state,unix_timestamp(ctime),unix_timestamp(mtime) FROM pack_generate WHERE app_key=? AND channel_id=? AND build_id=?`
	_generatedByChannels        = `SELECT id,app_key,build_id,channel_id,name,folder,patch_path,patch_url,cdn_url,status,size,md5,operator,channel_test_state,unix_timestamp(ctime),unix_timestamp(mtime) FROM pack_generate WHERE channel_id in (%s) AND app_key=? AND build_id=?`
	_setGenerate                = `INSERT INTO pack_generate (app_key,build_id,channel_id,name,folder,patch_path,patch_url,size,md5,operator) VALUES (?,?,?,?,?,?,?,?,?,?)`
	_setGenerates               = `INSERT INTO pack_generate (app_key,build_id,channel_id,name,folder,patch_path,patch_url,operator) VALUES %s`
	_setGeneratesByGit          = `INSERT INTO pack_generate (app_key,build_id,channel_id,status,name,folder,patch_path,patch_url,operator) VALUES %s`
	_upGeneratesByGit           = `INSERT INTO pack_generate(id,name,folder,patch_path,patch_url,status,size,md5,gl_job_url) VALUES %s ON DUPLICATE KEY UPDATE name=VALUES(name),folder=VALUES(folder),patch_path=VALUES(patch_path),patch_url=VALUES(patch_url),status=VALUES(status),size=VALUES(size),md5=VALUES(md5),gl_job_url=VALUES(gl_job_url)`
	_upGenerateStatusByIDs      = `UPDATE pack_generate SET status=? WHERE app_key=? AND id IN (%v)`
	_upGenerateState            = `UPDATE pack_generate SET pack_state=? WHERE id IN (%s)`
	_upGenerateStatus           = `UPDATE pack_generate SET status=?,operator=? WHERE app_key=? AND id=?`
	_upGenerateCDN              = `UPDATE pack_generate SET cdn_url=?,operator=?,status=? WHERE app_key=? AND id=?`
	_upGenerateSoleCDN          = `UPDATE pack_generate SET sole_cdn_url=? WHERE app_key=? AND id=?`
	//_upGenerateTestStates  = `UPDATE pack_generate SET channel_test_state=? WHERE app_key=? AND id IN (%v)`
	_setGeneratePublish = `INSERT INTO pack_generate_publish(app_key,channel_id,pack_generate_id) VALUES(?, ?, ?) ON DUPLICATE KEY UPDATE pack_generate_id=?`
	_getGeneratePublish = `SELECT pg.app_key,pg.build_id,pg.sole_cdn_url,pg.cdn_url,pv.version,pv.version_code,pg.size,pg.md5,unix_timestamp(pgp.mtime) FROM pack_generate_publish AS pgp,pack_generate AS pg,pack AS p,pack_version AS pv WHERE pg.id=pgp.pack_generate_id AND pg.build_id=p.build_id AND p.app_key = pg.app_key AND p.env='prod' AND pv.id=p.version_id AND pg.cdn_url=?;`

	_inManagerVersion            = `INSERT INTO version (plat,description,version,build,state,ctime,mtime,ptime) VALUES (?,?,?,?,?,?,?,?)`
	_inManagerVersionUpdate      = `INSERT INTO version_update (vid,channel,coverage,size,url,md5,state,sdkint,model,policy,is_force,policy_name,is_push,policy_url,sdkint_list,buvid_start,buvid_end,ctime,mtime) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_inManagerVersionUpdateLimit = `INSERT INTO version_update_limit (up_id,condi,value) VALUES %v`

	_customChannelPack     = `SELECT id,app_key,build_id,operator,size,md5,pack_name,pack_path,pack_url,cdn_url,sender,state,unix_timestamp(ctime),unix_timestamp(mtime) FROM custom_channel_pack WHERE app_key=? AND build_id=? AND pack_path=?`
	_customChannelPackByID = `SELECT id,app_key,build_id,operator,size,md5,pack_name,pack_path,pack_url,cdn_url,sender,state,unix_timestamp(ctime),unix_timestamp(mtime) FROM custom_channel_pack WHERE app_key=? AND id=?`
	_customChannelPacks    = `SELECT id,app_key,build_id,operator,size,md5,pack_name,pack_path,pack_url,cdn_url,sender,state,unix_timestamp(ctime),unix_timestamp(mtime) FROM custom_channel_pack WHERE app_key=? AND build_id=?`
	_inCustomChannelPack   = `INSERT INTO custom_channel_pack (app_key,build_id,pack_name,pack_path,pack_url,sender) VALUES %s`
	_upCustomChannelPack   = `UPDATE custom_channel_pack SET cdn_url=?,md5=?,size=?,operator=?,state=? WHERE app_key=? AND id=?`

	_setTFAppInfo  = `INSERT INTO app_tf_attr (app_key,store_app_id,issuer_id,key_id,tag_prefix,bugly_app_id,bugly_app_key) VALUES (?,?,?,?,?,?,?)`
	_upBetaGroup   = `UPDATE app_tf_attr SET beta_group_id=?,public_link=?,beta_group_id_test=?,public_link_test=? WHERE app_key=?`
	_TFAppInfo     = `SELECT app_key,store_app_id,beta_group_id,public_link,beta_group_id_test,public_link_test,issuer_id,key_id,tag_prefix,bugly_app_id,bugly_app_key,online_version,online_version_code,online_build_id FROM app_tf_attr where app_key=?`
	_TFAllAppsInfo = `SELECT app_key,store_app_id,beta_group_id,public_link,beta_group_id_test,public_link_test,issuer_id,key_id,tag_prefix,bugly_app_id,bugly_app_key,online_version,online_version_code,online_build_id FROM app_tf_attr`
	_upTFAppInfo   = `UPDATE app_tf_attr SET store_app_id=?,issuer_id=?,key_id=?,tag_prefix=?,bugly_app_id=?,bugly_app_key=? WHERE app_key=?`
	_upOnlineInfo  = `UPDATE app_tf_attr SET online_version=?,online_version_code=?,online_build_id=? WHERE app_key=?`
	_onlineBuildID = `SELECT p.id,p.build_id,p.commit FROM pack AS p,pack_tf_attr AS pt WHERE p.id=pt.pack_id AND p.env='test' AND pt.beta_build_id=?`

	_TFPackInfo       = `SELECT pt.id,pt.app_key,pack_id,version,version_code,p.env,pack_path,build_id,store_app_id,beta_group_id,public_link,beta_group_id_test,public_link_test,beta_build_id,expire_time,pack_state,review_state,beta_state,dis_permil,dis_num,dis_limit,remind_upd_time,force_upd_time,guide_tf_txt,remind_upd_txt,force_upd_txt,pt.ctime FROM pack AS p, pack_version AS pv, pack_tf_attr AS pt, app_tf_attr AS at  WHERE p.version_id=pv.id AND p.id=pt.pack_id AND pt.app_key=at.app_key AND beta_state=?`
	_TFPackByPackID   = `SELECT pt.id,pt.app_key,pack_id,version,version_code,p.env,pack_path,build_id,store_app_id,beta_group_id,public_link,beta_group_id_test,public_link_test,beta_build_id,expire_time,pack_state,review_state,beta_state,dis_permil,dis_num,dis_limit,remind_upd_time,force_upd_time,guide_tf_txt,remind_upd_txt,force_upd_txt,pt.ctime FROM pack AS p, pack_version AS pv, pack_tf_attr AS pt, app_tf_attr AS at  WHERE p.version_id=pv.id AND p.id=pt.pack_id AND pt.app_key=at.app_key AND p.id=?`
	_TFPackByPackTFID = `SELECT pt.id,pt.app_key,pack_id,version,version_code,p.env,pack_path,build_id,store_app_id,beta_group_id,public_link,beta_group_id_test,public_link_test,beta_build_id,expire_time,pack_state,review_state,beta_state,dis_permil,dis_num,dis_limit,remind_upd_time,force_upd_time,guide_tf_txt,remind_upd_txt,force_upd_txt,pt.ctime FROM pack AS p, pack_version AS pv, pack_tf_attr AS pt, app_tf_attr AS at  WHERE p.version_id=pv.id AND p.id=pt.pack_id AND pt.app_key=at.app_key AND pt.id=?`
	_TFPackInfoValid  = `SELECT pt.id,pt.app_key,pack_id,version,version_code,p.env,pack_path,build_id,store_app_id,beta_group_id,public_link,beta_group_id_test,public_link_test,beta_build_id,expire_time,pack_state,review_state,beta_state,dis_permil,dis_num,dis_limit,remind_upd_time,force_upd_time,guide_tf_txt,remind_upd_txt,force_upd_txt,pt.ctime FROM pack AS p, pack_version AS pv, pack_tf_attr AS pt, app_tf_attr AS at  WHERE p.version_id=pv.id AND p.id=pt.pack_id AND pt.app_key=at.app_key AND beta_state IN(4,5,6)`
	_TFPackInVersions = `SELECT pt.id,pt.app_key,pack_id,version,version_code,p.env,pack_path,build_id,store_app_id,beta_group_id,public_link,beta_group_id_test,public_link_test,beta_build_id,expire_time,pack_state,review_state,beta_state,dis_permil,dis_num,dis_limit,remind_upd_time,force_upd_time,guide_tf_txt,remind_upd_txt,force_upd_txt,pt.ctime FROM pack AS p, pack_version AS pv, pack_tf_attr AS pt, app_tf_attr AS at  WHERE p.version_id=pv.id AND p.id=pt.pack_id AND pt.app_key=at.app_key AND pv.id IN(%s)`
	_setTFPackInfo    = `INSERT INTO pack_tf_attr (app_key,pack_id,guide_tf_txt,remind_upd_txt,force_upd_txt) VALUES (?,?,?,?,?)`
	_setTFProdPack    = `INSERT INTO pack_tf_attr (app_key,pack_id,beta_build_id,expire_time,pack_state,review_state,beta_state,dis_permil,dis_num,dis_limit,remind_upd_time,force_upd_time,guide_tf_txt,remind_upd_txt,force_upd_txt) VALUES (?,?,?,FROM_UNIXTIME(?),?,?,?,?,?,?,FROM_UNIXTIME(?),FROM_UNIXTIME(?),?,?,?)`
	_upTFPackInfo     = `UPDATE pack_tf_attr SET beta_build_id=?,pack_state=?,expire_time=FROM_UNIXTIME(?),beta_state=1 WHERE id=?`
	_upTFPackState    = `UPDATE pack_tf_attr SET pack_state=?,beta_state=? WHERE id=?`
	_upTFReviewState  = `UPDATE pack_tf_attr SET review_state=?,beta_state=? WHERE id=?`
	_upTFBetaState    = `UPDATE pack_tf_attr SET beta_state=? WHERE id=?`
	_TFPackDistribute = `UPDATE pack_tf_attr SET dis_permil=?,dis_limit=?,beta_state=5 WHERE id=?`
	_upTFPackDisNum   = `UPDATE pack_tf_attr SET dis_num=? WHERE id=?`
	_upRemindUpdTime  = `UPDATE pack_tf_attr SET remind_upd_time=FROM_UNIXTIME(?) WHERE id=?`
	_upForceUpdTime   = `UPDATE pack_tf_attr SET force_upd_time=FROM_UNIXTIME(?) WHERE id=?`
	_upTFUpdTxt       = `UPDATE pack_tf_attr SET guide_tf_txt=?,remind_upd_txt=?,force_upd_txt=? WHERE id=?`
	_setTFBlackWhite  = `INSERT INTO tf_black_white (app_key,env,mid,nick,operator,list_type) VALUES (?,?,?,?,?,?)`
	_TFBlackWhiteList = `SELECT id,mid,nick,operator,ctime FROM tf_black_white WHERE app_key=? AND list_type=? AND env=?`
	_delTFBlackWhite  = `DELETE FROM tf_black_white WHERE id=?`

	_addPackGreyHistory   = `INSERT INTO pack_grey_history (app_key,version,version_code,gl_job_id,is_upgrade,flow,grey_start_time,grey_finish_time,grey_close_time,operator) VALUES %v`
	_lastPackGreyHistory  = `SELECT h1.id,h1.app_key,h1.version,h1.version_code,h1.gl_job_id,h1.is_upgrade,h1.flow,h1.grey_start_time,h1.grey_finish_time,h1.grey_close_time,h1.operator,h1.ctime,h1.mtime FROM pack_grey_history as h1 INNER JOIN (SELECT MAX(h3.id) as id FROM pack_grey_history as h3 GROUP BY gl_job_id) as h2 ON h1.id=h2.id WHERE app_key=? AND gl_job_id IN (%v)`
	_packGeryHistoryCount = `SELECT COUNT(*) FROM pack_grey_history WHERE app_key=? %s `
	_packGeryHistoryList  = `SELECT id,app_key,version,version_code,gl_job_id,is_upgrade,flow,grey_start_time,grey_finish_time,grey_close_time,operator,ctime,mtime FROM pack_grey_history WHERE app_key=? %s `
)

// PackVersionByAppKey get pack version by appkey.
func (d *Dao) PackVersionByAppKey(c context.Context, appKey, env, filterKey string, ps, pn int) (res map[int64]*model.Version, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey, env)
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey, filterKey)
		sqlAdd = "AND (version LIKE ? OR version_code LIKE ?)"
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_packVersionByAppkey, sqlAdd), args...)
	if err != nil {
		log.Error("PackVersionByAppKey %v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*model.Version)
	for rows.Next() {
		re := &model.Version{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.Version, &re.VersionCode, &re.IsUpgrade); err != nil {
			log.Error("PackVersionByAppKey %v", err)
			return
		}
		res[re.ID] = re
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// PackVersionByAppKeys get cd list.
func (d *Dao) PackVersionByAppKeys(c context.Context, env string, appKeys []string) (res map[string][]*model.Version, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, appKey := range appKeys {
		sqls = append(sqls, "?")
		args = append(args, appKey)
	}
	args = append(args, env)
	rows, err := d.db.Query(c, fmt.Sprintf(_packVersionByAppKeys, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("PackVersionByAppKeys %v", err)
		return
	}
	defer rows.Close()
	res = make(map[string][]*model.Version)
	for rows.Next() {
		pv := &model.Version{}
		if err = rows.Scan(&pv.ID, &pv.AppKey, &pv.Version, &pv.VersionCode, &pv.IsUpgrade); err != nil {
			log.Error("PackVersionByAppKeys %v", err)
			return
		}
		res[pv.AppKey] = append(res[pv.AppKey], pv)
	}
	err = rows.Err()
	return
}

// PackVersionCount get cd version list count.
func (d *Dao) PackVersionCount(c context.Context, appKey, env string) (count int, err error) {
	row := d.db.QueryRow(c, _packVersionCount, appKey, env)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("PackVersionCount %v", err)
		}
	}
	return
}

// PackVersionList get cd version list.
func (d *Dao) PackVersionList(c context.Context, appKey, env string, pn, ps int) (res map[int64]*cdmdl.PackItem, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if env != "" {
		sqlAdd = "AND env=?"
		args = append(args, env)
	}
	args = append(args, (pn-1)*ps, ps)
	rows, err := d.db.Query(c, fmt.Sprintf(_packVersionList, sqlAdd), args...)
	if err != nil {
		log.Error("PackVersionList %v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*cdmdl.PackItem)
	for rows.Next() {
		v := &cdmdl.PackItem{Version: &model.Version{}}
		if err = rows.Scan(&v.Version.ID, &v.Version.Version, &v.Version.VersionCode, &v.Version.IsUpgrade); err != nil {
			log.Error("VersionList %v", err)
			return
		}
		res[v.ID] = v
	}
	err = rows.Err()
	return
}

// PackVersionCountByOptions get cd version list count by options.
func (d *Dao) PackVersionCountByOptions(c context.Context, appKey, env, filterKey string, steadyState int, hasBbrUrl bool) (count int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey, env)
	if steadyState != 0 {
		args = append(args, steadyState)
		sqlAdd += " AND p.steady_state=?"
	}
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey)
		sqlAdd += " AND (pv.version_code LIKE ?)"
	}
	if hasBbrUrl {
		sqlAdd += " AND LENGTH(trim(p.bbr_url))>0"
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_packVersionCountByOptions, sqlAdd), args...)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("PackVersionCountByOptions error: %v", err)
		}
	}
	return
}

// PackVersionListByOptions get cd version list by options.
func (d *Dao) PackVersionListByOptions(c context.Context, appKey, env, filterKey string, steadyState int, hasBbrUrl bool, pn, ps int) (res map[int64]*cdmdl.PackItem, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey, env)
	if steadyState != 0 {
		args = append(args, steadyState)
		sqlAdd += " AND p.steady_state=?"
	}
	if filterKey != "" {
		filterKey = "%" + filterKey + "%"
		args = append(args, filterKey)
		sqlAdd += " AND (pv.version_code LIKE ?)"
	}
	if hasBbrUrl {
		sqlAdd += " AND LENGTH(trim(p.bbr_url))>0"
	}
	sqlAdd += " ORDER BY pv.version_code DESC"
	args = append(args, (pn-1)*ps, ps)
	sqlAdd += " LIMIT ?,?"
	rows, err := d.db.Query(c, fmt.Sprintf(_packVersionListByOptions, sqlAdd), args...)
	if err != nil {
		log.Error("PackVersionListByOptions error: %v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*cdmdl.PackItem)
	for rows.Next() {
		v := &cdmdl.PackItem{Version: &model.Version{}}
		if err = rows.Scan(&v.Version.ID, &v.Version.Version, &v.Version.VersionCode, &v.Version.IsUpgrade); err != nil {
			log.Error("PackVersionListByOptions VersionList %v", err)
			return
		}
		res[v.ID] = v
	}
	err = rows.Err()
	return
}

// TxSetPackVersion set version.
func (d *Dao) TxSetPackVersion(tx *xsql.Tx, appID, appKey, env, version string, versionCode, isUpgrade int64) (id int64, err error) {
	res, err := tx.Exec(_setVersion, appID, appKey, env, version, versionCode, isUpgrade)
	if err != nil {
		log.Error("TxSetPackVersion %v", err)
		return
	}
	return res.LastInsertId()
}

// PackVersionByID get version by id.
func (d *Dao) PackVersionByID(c context.Context, appKey string, id int64) (re *model.Version, err error) {
	row := d.db.QueryRow(c, _packVersionByID, appKey, id)
	re = &model.Version{}
	if err = row.Scan(&re.ID, &re.AppID, &re.AppKey, &re.Env, &re.Version, &re.VersionCode, &re.IsUpgrade, &re.CTime, &re.MTime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("PackVersionByID %v", err)
		}
	}
	return
}

// PackVersion if cd version.
func (d *Dao) PackVersion(c context.Context, appKey, env, version string, versionCode int64) (re *model.Version, err error) {
	row := d.db.QueryRow(c, _packVersion, appKey, env, version, versionCode)
	re = &model.Version{}
	if err = row.Scan(&re.ID, &re.AppID, &re.AppKey, &re.Env, &re.Version, &re.VersionCode, &re.IsUpgrade); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("PackVersion %v", err)
		}
	}
	return
}

// PackByVersion get pack by version version_cdoe.
func (d *Dao) PackByVersion(c context.Context, appKey, env string, versionID int64) (res []*cdmdl.Pack, err error) {
	rows, err := d.db.Query(c, _packByVersion, appKey, env, versionID)
	if err != nil {
		log.Error("PackByVersion %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		p := &cdmdl.Pack{}
		if err = rows.Scan(&p.ID, &p.AppID, &p.AppKey, &p.Env, &p.VersionID, &p.InternalVersionCode, &p.BuildID, &p.GitType, &p.GitName, &p.Commit,
			&p.PackType, &p.Operator, &p.Size, &p.MD5, &p.PackPath, &p.PackURL, &p.MappingURL, &p.RURL, &p.RMappingURL, &p.CDNURL, &p.Desc, &p.Sender, &p.SteadyState, &p.ChangeLog, &p.DepGitJobId, &p.IsCompatible, &p.CTime, &p.MTime, &p.PTime); err != nil {
			log.Error("AppCDList %v", err)
			return
		}
		res = append(res, p)
	}
	err = rows.Err()
	return
}

// PackByVersions get pack by versions.
func (d *Dao) PackByVersions(c context.Context, appKey, env string, versionIDs []int64) (res map[int64][]*cdmdl.Pack, err error) {
	var (
		sqls   []string
		args   []interface{}
		sqlAdd string
	)
	args = append(args, appKey)
	if env != "" {
		sqlAdd = "AND env=?"
		args = append(args, env)
	}
	for _, versionID := range versionIDs {
		sqls = append(sqls, "?")
		args = append(args, versionID)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_packByVersions, sqlAdd, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("Packs %v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64][]*cdmdl.Pack)
	for rows.Next() {
		re := &cdmdl.Pack{}
		if err = rows.Scan(&re.ID, &re.AppID, &re.AppKey, &re.Env, &re.VersionID, &re.InternalVersionCode, &re.BuildID, &re.GitType, &re.GitName, &re.Commit,
			&re.PackType, &re.Operator, &re.Size, &re.MD5, &re.PackPath, &re.PackURL, &re.MappingURL, &re.RURL, &re.RMappingURL, &re.CDNURL, &re.Desc, &re.Sender, &re.SteadyState, &re.ChangeLog, &re.DepGitJobId, &re.IsCompatible, &re.BbrUrl, &re.CTime, &re.MTime, &re.PTime); err != nil {
			log.Error("Packs %v", err)
			return
		}
		res[re.VersionID] = append(res[re.VersionID], re)
	}
	err = rows.Err()
	return
}

// PackByBuild get version by id.
func (d *Dao) PackByBuild(c context.Context, appKey, env string, buildID int64) (re *cdmdl.Pack, err error) {
	row := d.db.QueryRow(c, _packByBuildID, appKey, env, buildID)
	re = &cdmdl.Pack{}
	if err = row.Scan(&re.ID, &re.AppID, &re.AppKey, &re.Env, &re.VersionID, &re.InternalVersionCode, &re.BuildID, &re.GitType, &re.GitName, &re.Commit,
		&re.PackType, &re.Operator, &re.Size, &re.MD5, &re.PackPath, &re.PackURL, &re.MappingURL, &re.RURL, &re.RMappingURL, &re.CDNURL, &re.Desc, &re.Sender,
		&re.SteadyState, &re.ChangeLog, &re.DepGitJobId, &re.IsCompatible, &re.BbrUrl, &re.Features, &re.CTime, &re.MTime, &re.PTime, &re.Version, &re.VersionCode); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("PackByBuild %v", err)
		}
	}
	return
}

func (d *Dao) SelectPackByGlJobId(c context.Context, appKey string, glJobId []int64) (res []*cdmdl.Pack, err error) {
	if len(glJobId) == 0 {
		return
	}
	var (
		sqls = make([]string, 0, len(glJobId))
		args = make([]interface{}, 0)
	)
	args = append(args, appKey)
	for _, v := range glJobId {
		sqls = append(sqls, "?")
		args = append(args, v)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_packByGlJobId, strings.Join(sqls, ",")), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xxsql.ScanSlice(rows, &res); err != nil {
		log.Error("ScanSlice Error: %v", err)
		return
	}
	err = rows.Err()
	return
}

// TxUpPackVersionID update pack version_id by id.
func (d *Dao) TxUpPackVersionID(tx *xsql.Tx, appKey string, id, versionID int64) (r int64, err error) {
	res, err := tx.Exec(_upPackVersionID, versionID, appKey, id)
	if err != nil {
		log.Error("TxUpPackVersionID %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpPackSteadyState update pack steady_state, description by id.
func (d *Dao) TxUpPackSteadyState(tx *xsql.Tx, appKey, description string, buildID int64, steadyState int) (r int64, err error) {
	res, err := tx.Exec(_upPackSteadyState, steadyState, description, appKey, buildID)
	if err != nil {
		log.Error("TxUpPackVersionID %v", err)
		return
	}
	return res.RowsAffected()
}

func (d *Dao) TxUpdatePackSender(tx *xsql.Tx, sender, appKey, env string, buildId int64) (r int64, err error) {
	res, err := tx.Exec(_upPackSender, sender, appKey, buildId, env)
	if err != nil {
		return
	}
	return res.RowsAffected()
}

// TxSetPack set pack.
func (d *Dao) TxSetPack(tx *xsql.Tx, appID, appKey, env string, versionID, internalVersionCode, buildID int64, gitType int8, gitName, commit string, packType int8, operator string, size int64, md5, packPath, PackURL, mapppingURL, rURL, rMappingURL, url, desc, sender, changeLog string, depGitJobID int64, isCompatible int, bbrUlr, features string) (r int64, err error) {
	res, err := tx.Exec(_setPack, appID, appKey, env, versionID, internalVersionCode, buildID, gitType, gitName, commit, packType, operator, size, md5, packPath, PackURL, mapppingURL, rURL, rMappingURL, url, desc, sender, changeLog, depGitJobID, isCompatible, bbrUlr, features)
	if err != nil {
		log.Error("TxSetPack %v", err)
		return
	}
	return res.LastInsertId()
}

// LastPack get last N pack.
func (d *Dao) LastPack(c context.Context, appKey string, versionCode int64, steadyState, limit int) (res []*cdmdl.Pack, err error) {
	var (
		args []interface{}
	)
	args = append(args, appKey, versionCode, steadyState, limit)
	rows, err := d.db.Query(c, _lastPack, args...)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &cdmdl.Pack{}
		if err = rows.Scan(&re.ID, &re.AppID, &re.AppKey, &re.Env, &re.VersionID, &re.VersionCode, &re.InternalVersionCode, &re.BuildID, &re.GitType, &re.GitName, &re.Commit, &re.PackType, &re.Operator, &re.Size, &re.MD5, &re.PackPath, &re.PackURL, &re.MappingURL, &re.RURL, &re.RMappingURL, &re.CDNURL, &re.Desc, &re.Sender, &re.ChangeLog, &re.CTime, &re.MTime, &re.PTime); err != nil {
			log.Error("%v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxSetPackConfigSwitch update update switch.
func (d *Dao) TxSetPackConfigSwitch(tx *xsql.Tx, versionID int64, isUp int8) (r int64, err error) {
	res, err := tx.Exec(_setPackConfigSwitch, isUp, versionID)
	if err != nil {
		log.Error("TxSetPackConfigSwitch %v", err)
		return
	}
	return res.RowsAffected()
}

// TxSetPackUpgradConfig insert app upgrad config.
func (d *Dao) TxSetPackUpgradConfig(tx *xsql.Tx, appKey, env string, versionID int64, normal, exnormal, force, exforce, system, exSystem string, cycle int, title, content, policyURL, iconURL, confirmBtnText, cancelBtnText string, policy, silent int) (r int64, err error) {
	res, err := tx.Exec(_setPackUpgrad, appKey, env, versionID, normal, force, exnormal, exforce, system, exSystem, cycle, title, content, silent, policy, policyURL, iconURL, confirmBtnText, cancelBtnText, normal, force, exnormal, exforce, system, exSystem, cycle, title, content, silent, policy, policyURL, iconURL, confirmBtnText, cancelBtnText)
	if err != nil {
		log.Error("TxSetPackUpgradConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// PackUpgradConfig get app upgrad config.
func (d *Dao) PackUpgradConfig(c context.Context, appKey, env string, versionID int64) (uconfig *cdmdl.UpgradConfig, err error) {
	row := d.db.QueryRow(c, _packUpgrad, appKey, env, versionID)
	uconfig = &cdmdl.UpgradConfig{}
	if err = row.Scan(&uconfig.AppKey, &uconfig.Env, &uconfig.VersionID, &uconfig.Normal, &uconfig.Force, &uconfig.ExNormal, &uconfig.ExForce, &uconfig.System, &uconfig.ExcludeSystem, &uconfig.Cycle, &uconfig.Title, &uconfig.Content, &uconfig.IsSilent, &uconfig.Policy, &uconfig.PolicyURL, &uconfig.IconURL, &uconfig.ConfirmBtnText, &uconfig.CancelBtnText); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			uconfig = nil
		} else {
			log.Error("PackUpgradConfig %v", err)
		}
	}
	return
}

// TxSetPackFilterConfig insert app filter config.
func (d *Dao) TxSetPackFilterConfig(tx *xsql.Tx, appKey, env string, buildID int64, network, isp, channel, city string, percent int, salt, device, phoneModel, brand string, status int) (r int64, err error) {
	res, err := tx.Exec(_setPackFilter, appKey, env, buildID, network, isp, channel, city, percent, salt, device, status, phoneModel, brand, network, isp, channel, city, percent, device, status, phoneModel, brand)
	if err != nil {
		log.Error("TxSetPackFilterConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// PackFilterConfig get filter config.
func (d *Dao) PackFilterConfig(c context.Context, appKey, env string, bids []int64) (fconfig map[int64]*cdmdl.FilterConfig, err error) {
	var (
		sqls []string
		args []interface{}
	)
	args = append(args, appKey, env)
	for _, bid := range bids {
		sqls = append(sqls, "?")
		args = append(args, bid)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_packFilter, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("PackFilterConfig %v", err)
		return
	}
	defer rows.Close()
	fconfig = make(map[int64]*cdmdl.FilterConfig)
	for rows.Next() {
		fc := &cdmdl.FilterConfig{}
		if err = rows.Scan(&fc.AppKey, &fc.Env, &fc.BuildID, &fc.Network, &fc.ISP, &fc.Channel, &fc.City, &fc.Percent, &fc.Salt, &fc.Device, &fc.Status, &fc.PhoneModel, &fc.Brand); err != nil {
			log.Error("PackFilterConfig %v", err)
			return
		}
		fconfig[fc.BuildID] = fc
	}
	err = rows.Err()
	return
}

// TxSetPackFlowConfig insert app filter flow config.
func (d *Dao) TxSetPackFlowConfig(tx *xsql.Tx, appKey, env, flow string, buildID int64) (r int64, err error) {
	res, err := tx.Exec(_setPackFlow, appKey, env, buildID, flow, flow)
	if err != nil {
		log.Error("TxSetPackFlowConfig %v", err)
		return
	}
	return res.RowsAffected()
}

// PackFlowConfig get flow config.
func (d *Dao) PackFlowConfig(c context.Context, appKey, env string, bids []int64) (fconfig map[int64]*cdmdl.FlowConfig, err error) {
	var (
		sqls []string
		args []interface{}
	)
	args = append(args, appKey, env)
	for _, bid := range bids {
		sqls = append(sqls, "?")
		args = append(args, bid)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_packFlow, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("PackFlowConfig %v", err)
		return
	}
	defer rows.Close()
	fconfig = make(map[int64]*cdmdl.FlowConfig)
	for rows.Next() {
		fc := &cdmdl.FlowConfig{}
		if err = rows.Scan(&fc.AppKey, &fc.Env, &fc.BuildID, &fc.Flow, &fc.CTime, &fc.MTime); err != nil {
			log.Error("PackFlowConfig %v", err)
			return
		}
		fconfig[fc.BuildID] = fc
	}
	err = rows.Err()
	return
}

// PatchListCount get patch list count.
func (d *Dao) PatchListCount(c context.Context, appKey string, buildID int64) (count int, err error) {
	row := d.db.QueryRow(c, _patchCount, appKey, buildID)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("PatchListCount %v", err)
		}
	}
	return
}

// PatchList get patch list.
func (d *Dao) PatchList(c context.Context, appKey string, buildID int64, pn, ps int) (res []*cdmdl.Patch, err error) {
	rows, err := d.db.Query(c, _patchs, appKey, buildID, (pn-1)*ps, ps)
	if err != nil {
		log.Error("PatchList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		p := &cdmdl.Patch{}
		if err = rows.Scan(&p.ID, &p.AppKey, &p.BuildID, &p.TargetBuildID, &p.TargetVersionID, &p.TargetVersionCode, &p.TargetVersion, &p.OriginBuildID, &p.OriginVersionID, &p.OriginVersionCode, &p.OriginVersion, &p.Size, &p.Status, &p.GlJobID, &p.MD5, &p.PackURL, &p.PatchPath, &p.PatchURL, &p.CDNURL, &p.CTime, &p.MTime); err != nil {
			log.Error("PatchList %v", err)
			return
		}
		res = append(res, p)
	}
	err = rows.Err()
	return
}

// GenerateListCount get generate count.
func (d *Dao) GenerateListCount(c context.Context, appKey, env string, buildID int64) (count int, err error) {
	row := d.db.QueryRow(c, _generateCount, appKey, buildID)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Error("GenerateListCount %v", err)
		}
	}
	return
}

// GenerateList get Generate list.
func (d *Dao) GenerateList(c context.Context, appKey string, buildID int64) (res []*cdmdl.Generate, err error) {
	rows, err := d.db.Query(c, _generates, appKey, buildID)
	if err != nil {
		log.Error("GenerateList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &cdmdl.Generate{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.BuildID, &re.ChannelID, &re.Name, &re.Folder, &re.GeneratePath, &re.GenerateURL, &re.CDNURL, &re.Status, &re.Size, &re.MD5, &re.Operator, &re.ChannelTestState, &re.CTime, &re.PTime, &re.GlJobURL); err != nil {
			log.Error("GenerateList %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// GenerateListByIds get Generate list.
func (d *Dao) GenerateListByIds(c context.Context, ids []int64) (generateInfoMap map[int64]*cdmdl.Generate, err error) {
	var (
		sqls []string
		args []interface{}
	)
	if len(ids) <= 0 {
		return
	}
	for _, bid := range ids {
		sqls = append(sqls, "?")
		args = append(args, bid)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_generatesByIDs, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("PackFlowConfig %v", err)
		return
	}
	defer rows.Close()
	generateInfoMap = make(map[int64]*cdmdl.Generate)
	for rows.Next() {
		re := &cdmdl.Generate{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.BuildID, &re.ChannelID, &re.Name, &re.Folder, &re.GeneratePath, &re.GenerateURL, &re.CDNURL, &re.Status, &re.Size, &re.MD5, &re.Operator, &re.ChannelTestState, &re.CTime, &re.PTime, &re.GlJobURL); err != nil {
			log.Error("GenerateList %v", err)
			return
		}
		generateInfoMap[re.ID] = re
	}
	err = rows.Err()
	return
}

// Generate get generate info
func (d *Dao) Generate(c context.Context, appKey string, gid int64) (re *cdmdl.Generate, err error) {
	row := d.db.QueryRow(c, _generate, appKey, gid)
	re = &cdmdl.Generate{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.BuildID, &re.ChannelID, &re.Name, &re.Folder, &re.GeneratePath, &re.GenerateURL, &re.CDNURL, &re.Status, &re.Size, &re.MD5, &re.Operator, &re.ChannelTestState, &re.CTime, &re.PTime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("Generate %v", err)
		}
	}
	return
}

// GenerateByOptions get generate info by app_key &&channel_id && build_id
func (d *Dao) GenerateByOptions(c context.Context, appKey string, channelID, buildID int64) (re *cdmdl.Generate, err error) {
	row := d.db.QueryRow(c, _generateByOptions, appKey, channelID, buildID)
	re = &cdmdl.Generate{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.BuildID, &re.ChannelID, &re.Name, &re.Folder, &re.GeneratePath, &re.GenerateURL, &re.CDNURL, &re.Status, &re.Size, &re.MD5, &re.Operator, &re.ChannelTestState, &re.CTime, &re.PTime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("GenerateByOptions %v", err)
		}
	}
	return
}

// GenerateByAppKeyAndStatus get generate info by app_key && status
func (d *Dao) GenerateByAppKeyAndStatus(c context.Context, appKey string, status []int, active taskmdl.PackDeleteState) (res []*cdmdl.Generate, err error) {
	var (
		args   []interface{}
		sqlAdd []string
	)
	args = append(args, appKey)
	for _, s := range status {
		sqlAdd = append(sqlAdd, "?")
		args = append(args, s)
	}
	args = append(args, active)
	rows, err := d.db.Query(c, fmt.Sprintf(_generatesByAppKeyAndStatus, strings.Join(sqlAdd, ",")), args...)
	if err != nil {
		log.Error("PackFlowConfig %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := new(cdmdl.Generate)
		if err = rows.Scan(&re.ID, &re.AppKey, &re.BuildID, &re.ChannelID, &re.Name, &re.Folder, &re.GeneratePath, &re.GenerateURL, &re.CDNURL, &re.Status, &re.Size, &re.MD5, &re.Operator, &re.ChannelTestState, &re.PackState, &re.CTime, &re.Mtime); err != nil {
			log.Error("GenerateList %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// UpdateChannelPackState 更新pack_state
func (d *Dao) UpdateChannelPackState(c context.Context, ids []int64, deleted taskmdl.PackDeleteState) (r int64, err error) {
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
	res, err := d.db.Exec(c, fmt.Sprintf(_upGenerateState, strings.Join(sqlAdd, ",")), args...)
	if err != nil {
		log.Errorc(c, "d.UpdateChannelPackState error(%v)", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// GeneratedByChannel get generate already ids
func (d *Dao) GeneratedByChannel(c context.Context, channelIDs []int64, appKey string, buildID int64) (existGenera []*cdmdl.Generate, err error) {
	var (
		args          []interface{}
		channelIDStrs []string
	)
	for _, cID := range channelIDs {
		idStr := strconv.FormatInt(cID, 10)
		channelIDStrs = append(channelIDStrs, idStr)
	}
	args = append(args, appKey, buildID)
	rows, err := d.db.Query(c, fmt.Sprintf(_generatedByChannels, strings.Join(channelIDStrs, ",")), args...)
	if err != nil {
		log.Error("GeneratedByChannel %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var re = &cdmdl.Generate{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.BuildID, &re.ChannelID, &re.Name, &re.Folder, &re.GeneratePath, &re.GenerateURL, &re.CDNURL, &re.Status, &re.Size, &re.MD5, &re.Operator, &re.ChannelTestState, &re.CTime, &re.PTime); err != nil {
			log.Error("GeneratedByChannel Scan %v", err)
			return
		}
		existGenera = append(existGenera, re)
	}
	err = rows.Err()
	return
}

// TxUpGenerateCDN update generate cdn.
func (d *Dao) TxUpGenerateCDN(tx *xsql.Tx, appKey, url, userName string, status int, id int64) (r int64, err error) {
	res, err := tx.Exec(_upGenerateCDN, url, userName, status, appKey, id)
	if err != nil {
		log.Error("TxUpGenerateCDN %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpGenerateSoleCDN update generate sole cdn.
func (d *Dao) TxUpGenerateSoleCDN(tx *xsql.Tx, appKey, url string, id int64) (r int64, err error) {
	res, err := tx.Exec(_upGenerateSoleCDN, url, appKey, id)
	if err != nil {
		log.Error("TxUpGenerateSoleCDN %v", err)
		return
	}
	return res.RowsAffected()
}

// TxSetGenerate set generate.
func (d *Dao) TxSetGenerate(tx *xsql.Tx, appKey string, buildID, channelID, size int64, filename, folder, localPath, inetPath, md5, userName string) (r int64, err error) {
	res, err := tx.Exec(_setGenerate, appKey, buildID, channelID, filename, folder, localPath, inetPath, size, md5, userName)
	if err != nil {
		log.Error("TxSetGenerate %v", err)
		return
	}
	return res.RowsAffected()
}

// TxSetGenerates set generates.
func (d *Dao) TxSetGenerates(tx *xsql.Tx, sqls []string, sqlAdd []interface{}) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_setGenerates, strings.Join(sqls, ",")), sqlAdd...)
	if err != nil {
		log.Error("TxSetGenerates %v", err)
		return
	}
	return res.RowsAffected()
}

// TxSetGeneratesByGit set generates by git pipeline channel pack.
func (d *Dao) TxSetGeneratesByGit(tx *xsql.Tx, sqls []string, sqlAdd []interface{}) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_setGeneratesByGit, strings.Join(sqls, ",")), sqlAdd...)
	if err != nil {
		log.Error("TxSetGeneratesByGit %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpGeneratesByGit update generates info by git pipeline channel pack.
func (d *Dao) TxUpGeneratesByGit(tx *xsql.Tx, sqls []string, args []interface{}) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_upGeneratesByGit, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TxUpGeneratesByGit %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpGenerateStatus update generate status.
func (d *Dao) TxUpGenerateStatus(tx *xsql.Tx, appKey, userName string, id int64, status int) (r int64, err error) {
	res, err := tx.Exec(_upGenerateStatus, status, userName, appKey, id)
	if err != nil {
		log.Error("TxUpGenerateStatus %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpGenerateStatusByIDs update generate status by ids.
func (d *Dao) TxUpGenerateStatusByIDs(c context.Context, appKey, idsStr string, status int) (r int64, err error) {
	res, err := d.db.Exec(c, fmt.Sprintf(_upGenerateStatusByIDs, idsStr), status, appKey)
	if err != nil {
		log.Error("TxUpGenerateStatusByIDs %v", err)
		return
	}
	return res.RowsAffected()
}

// GetGeneratePublish set generate lastest publish.
func (d *Dao) GetGeneratePublish(c context.Context, cdnUrl string) (re *cdmdl.GeneratePublishLastest, err error) {
	row := d.db.QueryRow(c, _getGeneratePublish, cdnUrl)
	re = &cdmdl.GeneratePublishLastest{}
	if err = row.Scan(&re.AppKey, &re.BuildID, &re.SoleCDNURL, &re.CDNURL, &re.Version, &re.VersionCode, &re.Size, &re.MD5, &re.MTime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("GetGeneratePublish %v", err)
		}
	}
	return
}

// TxSetGeneratePublish set generate lastest publish.
func (d *Dao) TxSetGeneratePublish(tx *xsql.Tx, appKey string, channelId, generateId int64) (r int64, err error) {
	res, err := tx.Exec(_setGeneratePublish, appKey, channelId, generateId, generateId)
	if err != nil {
		log.Error("TxSetGeneratePublish %v", err)
		return
	}
	return res.RowsAffected()
}

// TxAppCDGenerateTestStateSet update generate channel_test_state.
//func (d *Dao) TxAppCDGenerateTestStateSet(tx *xsql.Tx, appKey, idStr string, testState int) (r int64, err error) {
//	var (
//		sqls []string
//		args []interface{}
//	)
//	args = append(args, testState, appKey)
//	for _, id := range strings.Split(idStr, ",") {
//		sqls = append(sqls, "?")
//		args = append(args, id)
//	}
//	res, err := tx.Exec(fmt.Sprintf(_upGenerateTestStates, strings.Join(sqls, ",")), args...)
//	if err != nil {
//		log.Error("TxUpGenerateStatus %v", err)
//		return
//	}
//	return res.RowsAffected()
//}

// TxAddPatch set generates.
func (d *Dao) TxAddPatch(tx *xsql.Tx, appKey string, buildID, targetBuildID, targetVersionID, originBuildID, originVersionID, size int64, status int, md5, localPath, inetPath, URL, packURL string) (r int64, err error) {
	res, err := tx.Exec(_addPatch, appKey, buildID, targetBuildID, targetVersionID, originBuildID, originVersionID, size, status, md5, localPath, inetPath, URL, packURL)
	if err != nil {
		log.Error("TxAddPatch %v", err)
		return
	}
	return res.LastInsertId()
}

// TxInManagerVersion add version
func (d *Dao) TxInManagerVersion(tx *xsql.Tx, plat, state int, build int64, description, version string) (r int64, err error) {
	nowTime := time.Now()
	res, err := tx.Exec(_inManagerVersion, plat, description, version, build, state, nowTime, nowTime, nowTime)
	if err != nil {
		log.Error("TxInManagerVersion %v", err)
		return
	}
	return res.LastInsertId()
}

// TxInManagerVersionUpdate add version_update
func (d *Dao) TxInManagerVersionUpdate(tx *xsql.Tx, coverage, state, sdkint, policy, isForce, isPush, buvidStart, buvidEnd int, vid, size int64, channel, url, md5, model, policyName, policyURL, sdkintList string) (r int64, err error) {
	nowTime := time.Now()
	res, err := tx.Exec(_inManagerVersionUpdate, vid, channel, coverage, size, url, md5, state, sdkint, model, policy, isForce, policyName, isPush, policyURL, sdkintList, buvidStart, buvidEnd, nowTime, nowTime)
	if err != nil {
		log.Error("TxInManagerVersionUpdate %v", err)
		return
	}
	return res.LastInsertId()
}

// TxInManagerVersionUpdateLimit add version_update_limit
func (d *Dao) TxInManagerVersionUpdateLimit(tx *xsql.Tx, upID int64, values []int, condi string) (r int64, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, value := range values {
		sqls = append(sqls, "(?,?,?)")
		args = append(args, upID, condi, value)
	}
	res, err := tx.Exec(fmt.Sprintf(_inManagerVersionUpdateLimit, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TxInManagerVersionUpdateLimit %v", err)
		return
	}
	return res.RowsAffected()
}

//nolint:unused
func (d *Dao) darkness(seed string) (uri string) {
	seed += `\u76d8\u53e4\u6709\u8bad`
	seed += `\u7eb5\u6a2a\u516d\u754c`
	seed += `\u8bf8\u4e8b\u7686\u6709\u7f18\u6cd5`
	seed += `\u51e1\u4eba\u4ef0\u89c2\u82cd\u5929`
	seed += `\u65e0\u660e\u65e5\u6708\u6f5c\u606f`
	seed += `\u56db\u65f6\u66f4\u66ff`
	seed += `\u5e7d\u51a5\u4e4b\u95f4`
	seed += `\u4e07\u7269\u5df2\u5faa\u56e0\u7f18`
	seed += `\u6052\u5927\u8005\u5219\u4e3a\u5929\u9053`
	seed += `\u76d8\u53e4\u6709\u8bad`
	seed += `\u7eb5\u6a2a\u516d\u754c`
	seed += `\u8bf8\u4e8b\u7686\u6709\u7f18\u6cd5`
	seed += `\u51e1\u4eba\u4ef0\u89c2\u82cd\u5929`
	seed += `\u65e0\u660e\u65e5\u6708\u6f5c\u606f`
	seed += `\u56db\u65f6\u66f4\u66ff`
	seed += `\u5e7d\u51a5\u4e4b\u95f4`
	seed += `\u4e07\u7269\u5df2\u5faa\u56e0\u7f18`
	seed += `\u6052\u5927\u8005\u5219\u4e3a\u5929\u9053`
	sUnicodev := strings.Split(seed, "\\u")
	for _, v := range sUnicodev {
		if len(v) < 1 {
			continue
		}
		temp, err := strconv.ParseInt(v, 16, 32)
		if err != nil {
			panic(err)
		}
		uri += fmt.Sprintf("%c", temp)
	}
	return
}

// AppCDRefreshCDN app cd refersh cdn
func (d *Dao) AppCDRefreshCDN(cdnUrls []string) (err error) {
	var (
		req       *http.Request
		reqMdl    *cdmdl.CDNRefreshReq
		res       *cdmdl.CDNRefreshRes
		data      []byte
		accountID int64
	)
	refreshAccountIDs := strings.Split(d.c.CDN.RefreshAccountIDs, ",")
	for _, refreshAccountID := range refreshAccountIDs {
		if accountID, err = strconv.ParseInt(refreshAccountID, 10, 64); err != nil {
			return
		}
		reqMdl = &cdmdl.CDNRefreshReq{
			Action:    d.c.CDN.RefreshAction,
			AccountID: accountID,
			Urls:      cdnUrls,
		}
		if data, err = json.Marshal(reqMdl); err != nil {
			log.Error("s.AppCDRefreshCDN json marshal error(%v)", err)
			return
		}
		// Request AccessToken
		if req, err = http.NewRequest(http.MethodPost, d.c.CDN.RefreshURL, strings.NewReader(string(data))); err != nil {
			log.Error("d.AppCDRefreshCDN call http.NewRequest error(%v)", err)
			return
		}
		req.Header.Add("x-secretid", d.c.CDN.SecretID)
		req.Header.Add("x-signature", d.c.CDN.Signature)
		req.Header.Add("content-type", "application/json")
		if err = d.httpClient.Do(context.Background(), req, &res); err != nil {
			log.Error("AppCDRefreshCDN error(%v)", res)
			return
		}
		if res.Code != 0 {
			err = errors.Wrap(ecode.Int(res.Code), res.Message)
			return
		}
	}
	return
}

// CustomChannelPack single get info
func (d *Dao) CustomChannelPack(c context.Context, appKey, packPath string, buildID int64) (re *cdmdl.CustomChannelPack, err error) {
	row := d.db.QueryRow(c, _customChannelPack, appKey, buildID, packPath)
	re = &cdmdl.CustomChannelPack{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.BuildID, &re.Operator, &re.Size, &re.MD5, &re.PackName, &re.PackPath, &re.PackURL, &re.CDNURL, &re.Sender, &re.State, &re.CTime, &re.MTime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("CustomChannelPack %v", err)
		}
	}
	return
}

// CustomChannelPackByID single get info by id
func (d *Dao) CustomChannelPackByID(c context.Context, appKey string, id int64) (re *cdmdl.CustomChannelPack, err error) {
	row := d.db.QueryRow(c, _customChannelPackByID, appKey, id)
	re = &cdmdl.CustomChannelPack{}
	if err = row.Scan(&re.ID, &re.AppKey, &re.BuildID, &re.Operator, &re.Size, &re.MD5, &re.PackName, &re.PackPath, &re.PackURL, &re.CDNURL, &re.Sender, &re.State, &re.CTime, &re.MTime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			re = nil
		} else {
			log.Error("CustomChannelPackByID %v", err)
		}
	}
	return
}

// CustomChannelPacks get infos
func (d *Dao) CustomChannelPacks(c context.Context, appKey string, buildID int64) (res []*cdmdl.CustomChannelPack, err error) {
	rows, err := d.db.Query(c, _customChannelPacks, appKey, buildID)
	if err != nil {
		log.Error("CustomChannelPacks %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &cdmdl.CustomChannelPack{}
		if err = rows.Scan(&re.ID, &re.AppKey, &re.BuildID, &re.Operator, &re.Size, &re.MD5, &re.PackName, &re.PackPath, &re.PackURL, &re.CDNURL, &re.Sender, &re.State, &re.CTime, &re.MTime); err != nil {
			log.Error("CustomChannelPacks %v", err)
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}

// TxAddCustomChannelPacks add CustomChannelPack.
func (d *Dao) TxAddCustomChannelPacks(tx *xsql.Tx, sqls []string, sqlAdd []interface{}) (r int64, err error) {
	res, err := tx.Exec(fmt.Sprintf(_inCustomChannelPack, strings.Join(sqls, ",")), sqlAdd...)
	if err != nil {
		log.Error("TxAddCustomChannelPacks %v", err)
		return
	}
	return res.RowsAffected()
}

// TxUpCustomChannelPack update customChannelPack info.
func (d *Dao) TxUpCustomChannelPack(tx *xsql.Tx, appKey, cdnURL, md5, operator string, state int, id, size int64) (r int64, err error) {
	res, err := tx.Exec(_upCustomChannelPack, cdnURL, md5, size, operator, state, appKey, id)
	if err != nil {
		log.Error("TxUpCustomChannelPack %v", err)
		return
	}
	return res.RowsAffected()
}

// TxSetTFAppInfo set app testflight info.
func (d *Dao) TxSetTFAppInfo(tx *xsql.Tx, appKey, storeAppID, IssuerID, keyID, tagPrefix, buglyAppID, buglyAppKey string) (err error) {
	_, err = tx.Exec(_setTFAppInfo, appKey, storeAppID, IssuerID, keyID, tagPrefix, buglyAppID, buglyAppKey)
	if err != nil {
		log.Error("TxSetAppTFInfo %v", err)
		return
	}
	return
}

// TxUpBetaGroup update app beta grouop info.
func (d *Dao) TxUpBetaGroup(tx *xsql.Tx, appKey, betaGroupID, publicLink, betaGroupIDTest, publicLinkTest string) (err error) {
	_, err = tx.Exec(_upBetaGroup, betaGroupID, publicLink, betaGroupIDTest, publicLinkTest, appKey)
	if err != nil {
		log.Error("TxUpBetaGroup %v", err)
		return
	}
	return
}

// TFAppInfo testflight app info.
func (d *Dao) TFAppInfo(c context.Context, appKey string) (i *cdmdl.TestFlightAppInfo, err error) {
	row := d.db.QueryRow(c, _TFAppInfo, appKey)
	i = &cdmdl.TestFlightAppInfo{}
	if err = row.Scan(&i.AppKey, &i.StoreAppID, &i.BetaGroupID, &i.PublicLink, &i.BetaGroupIDTest, &i.PublicLinkTest, &i.IssuerID, &i.KeyID, &i.TagPrefix, &i.BuglyAppID, &i.BuglyAppKey, &i.OnlineVersion, &i.OnlineVersionCode, &i.OnlineBuildID); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
			i = nil
		} else {
			log.Error("TFAppInfo %v", err)
		}
	}
	return
}

// TFAllAppsInfo testflight all apps info
func (d *Dao) TFAllAppsInfo(c context.Context) (res []*cdmdl.TestFlightAppInfo, err error) {
	rows, err := d.db.Query(c, _TFAllAppsInfo)
	if err != nil {
		log.Error("TFAllAppsInfo %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &cdmdl.TestFlightAppInfo{}
		if err = rows.Scan(&r.AppKey, &r.StoreAppID, &r.BetaGroupID, &r.PublicLink, &r.BetaGroupIDTest, &r.PublicLinkTest, &r.IssuerID, &r.KeyID, &r.TagPrefix, &r.BuglyAppID, &r.BuglyAppKey, &r.OnlineVersion, &r.OnlineVersionCode, &r.OnlineBuildID); err != nil {
			log.Error("TFAllAppsInfo %v", err)
			return
		}
		res = append(res, r)
	}
	err = rows.Err()
	return
}

// TxUpTFAppInfo update testflight app info.
func (d *Dao) TxUpTFAppInfo(tx *xsql.Tx, appKey, storeAppID, issuerID, keyID, tagPrefix, buglyAppID, buglyAppKey string) (err error) {
	_, err = tx.Exec(_upTFAppInfo, storeAppID, issuerID, keyID, tagPrefix, buglyAppID, buglyAppKey, appKey)
	if err != nil {
		log.Error("TxUpTFAppInfo %v", err)
	}
	return
}

// TxUpOnlineInfo update app online version info.
func (d *Dao) TxUpOnlineInfo(tx *xsql.Tx, appKey, onlineVer string, onlineVerCode, buildID int64) (err error) {
	_, err = tx.Exec(_upOnlineInfo, onlineVer, onlineVerCode, buildID, appKey)
	if err != nil {
		log.Error("TxUpOnlineInfo error(%v)", err)
	}
	return
}

// OnlineBuildID get online build ID from App Store build ID.
func (d *Dao) OnlineBuildID(c context.Context, betaBuildID string) (packID, buildID int64, commit string, err error) {
	row := d.db.QueryRow(c, _onlineBuildID, betaBuildID)
	if err = row.Scan(&packID, &buildID, &commit); err != nil {
		log.Error("OnlineBuildID %v", err)
	}
	return
}

// TFPackInfoWithState get testflight packs with state.
func (d *Dao) TFPackInfoWithState(c context.Context, betaState int) (res []*cdmdl.TestFlightPackInfo, err error) {
	rows, err := d.db.Query(c, _TFPackInfo, betaState)
	if err != nil {
		log.Error("TFPackInfo %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &cdmdl.TestFlightPackInfo{}
		if err = rows.Scan(&r.ID, &r.AppKey, &r.PackID, &r.Version, &r.VersionCode, &r.Env, &r.PackPath, &r.BuildID, &r.StoreAppID, &r.BetaGroupID, &r.PublicLink, &r.BetaGroupIDTest, &r.PublicLinkTest, &r.BetaBuildID, &r.ExpireTime, &r.PackState, &r.ReviewState, &r.BetaState, &r.DisPermil, &r.DisNum, &r.DisLimit, &r.RemindUpdTime, &r.ForceupdTime, &r.GuideTFTxt, &r.RemindUpdTxt, &r.ForceUpdTxt, &r.CTime); err != nil {
			log.Error("TFPackInfo %v", err)
			return
		}
		res = append(res, r)
	}
	err = rows.Err()
	return
}

// TFPackByPackID get one testflight pack info with pack ID
func (d *Dao) TFPackByPackID(c context.Context, packID int64) (r *cdmdl.TestFlightPackInfo, err error) {
	row := d.db.QueryRow(c, _TFPackByPackID, packID)
	r = &cdmdl.TestFlightPackInfo{}
	if err = row.Scan(&r.ID, &r.AppKey, &r.PackID, &r.Version, &r.VersionCode, &r.Env, &r.PackPath, &r.BuildID, &r.StoreAppID, &r.BetaGroupID, &r.PublicLink, &r.BetaGroupIDTest, &r.PublicLinkTest, &r.BetaBuildID, &r.ExpireTime, &r.PackState, &r.ReviewState, &r.BetaState, &r.DisPermil, &r.DisNum, &r.DisLimit, &r.RemindUpdTime, &r.ForceupdTime, &r.GuideTFTxt, &r.RemindUpdTxt, &r.ForceUpdTxt, &r.CTime); err != nil {
		if err == xsql.ErrNoRows {
			return nil, nil
		}
		log.Error("TFPackByPackID %v", err)
	}
	return
}

// TFPackByPackTFID get one testflight pack info with pack testflight ID
func (d *Dao) TFPackByPackTFID(c context.Context, packTFID int64) (r *cdmdl.TestFlightPackInfo, err error) {
	row := d.db.QueryRow(c, _TFPackByPackTFID, packTFID)
	r = &cdmdl.TestFlightPackInfo{}
	if err = row.Scan(&r.ID, &r.AppKey, &r.PackID, &r.Version, &r.VersionCode, &r.Env, &r.PackPath, &r.BuildID, &r.StoreAppID, &r.BetaGroupID, &r.PublicLink, &r.BetaGroupIDTest, &r.PublicLinkTest, &r.BetaBuildID, &r.ExpireTime, &r.PackState, &r.ReviewState, &r.BetaState, &r.DisPermil, &r.DisNum, &r.DisLimit, &r.RemindUpdTime, &r.ForceupdTime, &r.GuideTFTxt, &r.RemindUpdTxt, &r.ForceUpdTxt, &r.CTime); err != nil {
		if err == xsql.ErrNoRows {
			return nil, nil
		}
		log.Error("TFPackByPackTFID %v", err)
	}
	return
}

// TFPackInfoValid get testflight packs which is valid.
func (d *Dao) TFPackInfoValid(c context.Context) (res []*cdmdl.TestFlightPackInfo, err error) {
	rows, err := d.db.Query(c, _TFPackInfoValid)
	if err != nil {
		log.Error("TFPackInfo %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &cdmdl.TestFlightPackInfo{}
		if err = rows.Scan(&r.ID, &r.AppKey, &r.PackID, &r.Version, &r.VersionCode, &r.Env, &r.PackPath, &r.BuildID, &r.StoreAppID, &r.BetaGroupID, &r.PublicLink, &r.BetaGroupIDTest, &r.PublicLinkTest, &r.BetaBuildID, &r.ExpireTime, &r.PackState, &r.ReviewState, &r.BetaState, &r.DisPermil, &r.DisNum, &r.DisLimit, &r.RemindUpdTime, &r.ForceupdTime, &r.GuideTFTxt, &r.RemindUpdTxt, &r.ForceUpdTxt, &r.CTime); err != nil {
			log.Error("TFPackInfo %v", err)
			return
		}
		res = append(res, r)
	}
	err = rows.Err()
	return
}

// TFPackByVersions get testflight packs in versions.
func (d *Dao) TFPackByVersions(c context.Context, versionIDs []int64) (res map[int64]*cdmdl.TestFlightPackInfo, err error) {
	var (
		sqls []string
		args []interface{}
	)
	for _, versionID := range versionIDs {
		sqls = append(sqls, "?")
		args = append(args, versionID)
	}
	for _, versionID := range versionIDs {
		sqls = append(sqls, "?")
		args = append(args, versionID)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_TFPackInVersions, strings.Join(sqls, ",")), args...)
	if err != nil {
		log.Error("TFPackByVersions %v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*cdmdl.TestFlightPackInfo)
	for rows.Next() {
		r := &cdmdl.TestFlightPackInfo{}
		if err = rows.Scan(&r.ID, &r.AppKey, &r.PackID, &r.Version, &r.VersionCode, &r.Env, &r.PackPath, &r.BuildID, &r.StoreAppID, &r.BetaGroupID, &r.PublicLink, &r.BetaGroupIDTest, &r.PublicLinkTest, &r.BetaBuildID, &r.ExpireTime, &r.PackState, &r.ReviewState, &r.BetaState, &r.DisPermil, &r.DisNum, &r.DisLimit, &r.RemindUpdTime, &r.ForceupdTime, &r.GuideTFTxt, &r.RemindUpdTxt, &r.ForceUpdTxt, &r.CTime); err != nil {
			log.Error("TFPackByVersions %v", err)
			return
		}
		res[r.BuildID] = r
	}
	err = rows.Err()
	return
}

// TxSetTFPackInfo set testflight pack info.
func (d *Dao) TxSetTFPackInfo(tx *xsql.Tx, appKey string, packID int64, guideTFTxt, RemindUpdTxt, ForceUpdTxt string) (err error) {
	_, err = tx.Exec(_setTFPackInfo, appKey, packID, guideTFTxt, RemindUpdTxt, ForceUpdTxt)
	if err != nil {
		log.Error("TxSetTFPackInfo %v", err)
	}
	return
}

// TxSetTFProdPack set app store prod pack info.
func (d *Dao) TxSetTFProdPack(tx *xsql.Tx, appKey string, packID int64, betaBuildID string, expireTime int64, packState, reviewState string, betaState, disPermil int, disNum, disLimit, remindUpdTime, forceUpdTime int64, guideTFTxt, remindUpdTxt, forceUpdTxt string) (err error) {
	_, err = tx.Exec(_setTFProdPack, appKey, packID, betaBuildID, expireTime, packState, reviewState, betaState, disPermil, disNum, disLimit, remindUpdTime, forceUpdTime, guideTFTxt, remindUpdTxt, forceUpdTxt)
	if err != nil {
		log.Error("TxSetTFProdPack %v", err)
	}
	return
}

// TxUpTFPackInfo update testflight pack info.
func (d *Dao) TxUpTFPackInfo(tx *xsql.Tx, ID int64, betaBuildID, packState string, expireTime int64) (err error) {
	_, err = tx.Exec(_upTFPackInfo, betaBuildID, packState, expireTime, ID)
	if err != nil {
		log.Error("TxUpTFPackInfo %v", err)
	}
	return
}

// TxUpTFPackState update the processing state of a package in appstore connect.
func (d *Dao) TxUpTFPackState(tx *xsql.Tx, ID int64, packState string, betaState int) (err error) {
	_, err = tx.Exec(_upTFPackState, packState, betaState, ID)
	if err != nil {
		log.Error("TxUpTFPackState %v", err)
	}
	return
}

// TxUpTFReviewState update the beta review state of a package in appstore connect.
func (d *Dao) TxUpTFReviewState(tx *xsql.Tx, ID int64, reviewState string, betaState int) (err error) {
	_, err = tx.Exec(_upTFReviewState, reviewState, betaState, ID)
	if err != nil {
		log.Error("TxUpTFReviewState %v", err)
	}
	return
}

// TxUpTFBetaState update the state of a testflight package in fawkes.
func (d *Dao) TxUpTFBetaState(tx *xsql.Tx, ID int64, betaState int) (err error) {
	_, err = tx.Exec(_upTFBetaState, betaState, ID)
	if err != nil {
		log.Error("TxUpTFBetaState %v", err)
	}
	return
}

// TxTFPackDistribute distribute a testflight package to external.
func (d *Dao) TxTFPackDistribute(tx *xsql.Tx, ID int64, disPermil int, disLimit int64) (err error) {
	_, err = tx.Exec(_TFPackDistribute, disPermil, disLimit, ID)
	if err != nil {
		log.Error("TxTFPackDistribute %v", err)
	}
	return
}

// TxUpTFPackDisNum update the user number of a testflight package.
func (d *Dao) TxUpTFPackDisNum(tx *xsql.Tx, ID, disNum int64) (err error) {
	_, err = tx.Exec(_upTFPackDisNum, disNum, ID)
	if err != nil {
		log.Error("TxUpTFPackDisNum %v", err)
	}
	return
}

// TxUpRemindUpdTime update the remind update time of a testflight package.
func (d *Dao) TxUpRemindUpdTime(tx *xsql.Tx, ID, remindUpdTime int64) (err error) {
	_, err = tx.Exec(_upRemindUpdTime, remindUpdTime, ID)
	if err != nil {
		log.Error("TxUpRemindUpdTime %v", err)
	}
	return
}

// TxUpForceUpdTime update the force update time of a testflight package.
func (d *Dao) TxUpForceUpdTime(tx *xsql.Tx, ID, forceUpdTime int64) (err error) {
	_, err = tx.Exec(_upForceUpdTime, forceUpdTime, ID)
	if err != nil {
		log.Error("TxUpForceUpdTime %v", err)
	}
	return
}

// TxUpTFUpdTxt update testflight update text.
func (d *Dao) TxUpTFUpdTxt(tx *xsql.Tx, ID int64, guideTFTxt, remindUpdTxt, forceUpdTxt string) (err error) {
	_, err = tx.Exec(_upTFUpdTxt, guideTFTxt, remindUpdTxt, forceUpdTxt, ID)
	if err != nil {
		log.Error("TxUpTFUpdTxt %v", err)
	}
	return
}

// TxSetTFBlackWhite set a testflight user to black/white list
func (d *Dao) TxSetTFBlackWhite(tx *xsql.Tx, appKey, env string, mid int64, nick, operator, listType string) (err error) {
	_, err = tx.Exec(_setTFBlackWhite, appKey, env, mid, nick, operator, listType)
	if err != nil {
		log.Error("TxSetTFBlackWhite %v", err)
	}
	return
}

// TFBlackWhiteList get testflight users black/white list
func (d *Dao) TFBlackWhiteList(c context.Context, appKey, listType, env string) (res []*cdmdl.TestFlightBWList, err error) {
	rows, err := d.db.Query(c, _TFBlackWhiteList, appKey, listType, env)
	if err != nil {
		log.Error("TFBlackWhiteList %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &cdmdl.TestFlightBWList{}
		if err = rows.Scan(&r.ID, &r.MID, &r.Nick, &r.Operator, &r.CTime); err != nil {
			log.Error("TFBlackWhiteList %v", err)
			return
		}
		res = append(res, r)
	}
	err = rows.Err()
	return
}

// TxDelBlackWhite delete a testflight user from black/white list
func (d *Dao) TxDelBlackWhite(tx *xsql.Tx, ID int64) (err error) {
	_, err = tx.Exec(_delTFBlackWhite, ID)
	if err != nil {
		log.Error("TxDelBlackWhite %v", err)
	}
	return
}

// AddPackGreyHistory 插入pack灰度信息记录
func (d *Dao) AddPackGreyHistory(c context.Context, operator string, packGreyData []*cdmdl.PackGreyData) (err error) {
	var (
		sqls []string
		args []interface{}
	)
	if len(packGreyData) == 0 {
		return
	}
	for _, packGrey := range packGreyData {
		sqls = append(sqls, "(?,?,?,?,?,?,?,?,?,?)")
		args = append(args, packGrey.AppKey, packGrey.Version, packGrey.VersionCode, packGrey.GlJobID, packGrey.IsUpgrade, packGrey.Flow, packGrey.GreyStartTime, packGrey.GreyFinishTime, packGrey.GreyCloseTime, operator)
	}
	_, err = d.db.Exec(c, fmt.Sprintf(_addPackGreyHistory, strings.Join(sqls, ",")), args...)
	return
}

// LastPackGreyHistory 返回pack最新一条记录
func (d *Dao) LastPackGreyHistory(c context.Context, appKey string, glJobIds []int64) (res map[int64]*cdmdl.PackGreyHistory, err error) {
	var (
		sqls []string
		args []interface{}
	)
	args = append(args, appKey)
	if len(glJobIds) == 0 {
		return
	}
	for _, glJobId := range glJobIds {
		sqls = append(sqls, "?")
		args = append(args, glJobId)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_lastPackGreyHistory, strings.Join(sqls, ",")), args...)
	if err != nil {
		return
	}
	res = make(map[int64]*cdmdl.PackGreyHistory)
	for rows.Next() {
		re := &cdmdl.PackGreyHistory{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			log.Error("%v", err)
			return
		}
		res[re.GlJobID] = re
	}
	err = rows.Err()
	return
}

func (d *Dao) PackGreyHistoryCount(c context.Context, appKey, version string, versionCode, glJobId, startTime, endTime int64) (count int, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if version != "" {
		sqlAdd += " AND version=? "
		args = append(args, version)
	}
	if versionCode != 0 {
		sqlAdd += " AND version_code=? "
		args = append(args, versionCode)
	}
	if glJobId != 0 {
		sqlAdd += " AND gl_job_id=? "
		args = append(args, glJobId)
	}
	if startTime != 0 {
		sqlAdd += " AND unix_timestamp(ctime)>? "
		args = append(args, startTime)
	}
	if endTime != 0 {
		sqlAdd += " AND unix_timestamp(ctime)<? "
		args = append(args, endTime)
	}
	row := d.db.QueryRow(c, fmt.Sprintf(_packGeryHistoryCount, sqlAdd), args...)
	if err = row.Scan(&count); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		}
	}
	return
}

func (d *Dao) PackGreyHistoryList(c context.Context, appKey, version string, versionCode, glJobId, startTime, endTime int64, pn, ps int) (res []*cdmdl.PackGreyHistory, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if version != "" {
		sqlAdd += " AND version=? "
		args = append(args, version)
	}
	if versionCode != 0 {
		sqlAdd += " AND version_code=? "
		args = append(args, versionCode)
	}
	if glJobId != 0 {
		sqlAdd += " AND gl_job_id=? "
		args = append(args, glJobId)
	}
	if startTime != 0 {
		sqlAdd += " AND unix_timestamp(ctime)>? "
		args = append(args, startTime)
	}
	if endTime != 0 {
		sqlAdd += " AND unix_timestamp(ctime)<? "
		args = append(args, endTime)
	}
	sqlAdd += " ORDER BY ctime DESC"
	args = append(args, (pn-1)*ps, ps)
	sqlAdd += " LIMIT ?,?"
	rows, err := d.db.Query(c, fmt.Sprintf(_packGeryHistoryList, sqlAdd), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		re := &cdmdl.PackGreyHistory{}
		if err = d.GetRowsValue(rows.Rows, re); err != nil {
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}
