package cd

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/ecode"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	bizapkmdl "go-gateway/app/app-svr/fawkes/service/model/bizapk"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	mailmdl "go-gateway/app/app-svr/fawkes/service/model/mail"
	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	"go-gateway/app/app-svr/fawkes/service/model/template"
	channel2 "go-gateway/app/app-svr/fawkes/service/tools/channel"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

// Keys sort vids desc.
type Keys []int64

func (k Keys) Len() int           { return len(k) }
func (k Keys) Less(i, j int) bool { return k[i] > k[j] }
func (k Keys) Swap(i, j int)      { k[i], k[j] = k[j], k[i] }

type Vers []*model.Version

func (v Vers) Len() int { return len(v) }
func (v Vers) Less(i, j int) bool {
	var iv, jv int64
	if v[i] != nil {
		iv = v[i].VersionCode
	}
	if v[j] != nil {
		jv = v[j].VersionCode
	}
	return iv > jv
}
func (v Vers) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

// PackVersionByAppKey get version.
func (s *Service) PackVersionByAppKey(c context.Context, appKey, env, filterKey string, ps, pn int) (res []*model.Version, err error) {
	var vs map[int64]*model.Version
	if vs, err = s.fkDao.PackVersionByAppKey(c, appKey, env, filterKey, ps, pn); err != nil {
		log.Error("%v", err)
		return
	}

	var keys []int64
	for key := range vs {
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		log.Error("PackVersionByAppKey %v %v keys is 0", appKey, env)
		return
	}
	sort.Sort(Keys(keys))
	for _, key := range keys {
		res = append(res, vs[key])
	}
	return
}

// AppPortalTest portal ci to cd test.
func (s *Service) AppPortalTest(c context.Context, appKey, desc, sender string, buildID int64) (err error) {
	// get ci info
	var ci *cimdl.BuildPack
	if ci, err = s.fkDao.BuildPack(c, appKey, buildID); err != nil {
		log.Error("%v", err)
		return
	}
	if ci == nil {
		log.Error("%v %v %v %v ci is nil", appKey, desc, sender, buildID)
		return
	}
	// insert to cd.
	var version *model.Version
	if version, err = s.fkDao.PackVersion(c, appKey, "test", ci.Version, ci.VersionCode); err != nil {
		log.Error("%v", err)
		return
	}
	var app *appmdl.APP
	if app, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	var (
		url              string
		filename         = path.Base(ci.PkgPath)
		relativeFilePath = strings.Replace(ci.PkgPath, s.c.LocalPath.LocalDir, "", -1)
		folder           = strings.Replace(relativeFilePath, filename, "", -1)
	)
	if url, _, _, err = s.fkDao.FilePutOss(c, folder, filename, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var versionID int64
	if version == nil {
		if versionID, err = s.fkDao.TxSetPackVersion(tx, ci.AppID, ci.AppKey, "test", ci.Version, ci.VersionCode, cdmdl.NotUpgrade); err != nil {
			log.Error("%v", err)
			return
		}
	} else {
		versionID = version.ID
	}
	if sender == "" {
		sender = "ci"
	}
	var packID int64
	if packID, err = s.fkDao.TxSetPack(tx, ci.AppID, ci.AppKey, "test", versionID, ci.InternalVersionCode, ci.GitlabJobID,
		ci.GitType, ci.GitName, ci.Commit, ci.PkgType, ci.Operator, ci.Size, ci.Md5, ci.PkgPath, ci.PkgURL, ci.MappingURL,
		ci.RURL, ci.RMappingURL, url, desc, sender, ci.ChangeLog, ci.DepGitlabJobID, ci.IsCompatible, ci.BbrURL, ci.Features); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = s.fkDao.TxBuildPackDidPush(tx, ci.BuildID); err != nil {
		log.Error("%v", err)
		return
	}
	// 只有 android 做增量包
	if app.Platform == "android" {
		_ = s.GenerateAllPatchesTest(ci.AppKey, url, versionID, ci.VersionCode, ci.GitlabJobID)
		// nolint:biligowordcheck
		go func() {
			_ = s.bizApkToCDN(context.Background(), ci.AppKey, ci.GitlabJobID)
		}()
	}
	// 只有 ios 做包上传
	if app.Platform == "ios" {
		var appTFInfo *cdmdl.TestFlightAppInfo
		if appTFInfo, _ = s.fkDao.TFAppInfo(context.Background(), appKey); appTFInfo != nil {
			// 包上传 app store
			s.AddHandlerProc(func() {
				_ = s.UploadToAppleStoreConnect(appKey, ci.PkgPath, packID, ci.PkgType, ci.GitlabJobID)
			})
			// 上传bugly符号表
			s.AddHandlerProc(func() {
				_ = s.uploadBugly(appTFInfo, ci.PkgPath, app.AppID, ci.VersionCode)
			})
		}
	}
	// add log
	_, _ = s.fkDao.AddLog(c, appKey, "test", mngmdl.ModelCI, mngmdl.OperationCIPush, fmt.Sprintf("构建ID: %v", ci.GitlabJobID), sender)
	return
}

func (s *Service) bizApkToCDN(c context.Context, appKey string, packBuildID int64) (err error) {
	var (
		builds []*bizapkmdl.Build
	)
	if builds, err = s.fkDao.BizApkBuilds(c, appKey, "test", packBuildID); err != nil {
		log.Error("bizApkToCDN: %v", err)
		return
	}
	for _, build := range builds {
		_ = s.putBuildToCDN(build)
	}
	return
}

func (s *Service) putBuildToCDN(build *bizapkmdl.Build) (err error) {
	var (
		tx                                *sql.Tx
		apkFilename, apkCdnURL, apkFolder string
	)
	apkFilename = path.Base(build.ApkPath)
	apkRelativePath := strings.Replace(build.ApkPath, s.c.LocalPath.LocalDir, "", -1)
	apkFolder = strings.Replace(apkRelativePath, apkFilename, "", -1)
	if apkCdnURL, _, _, err = s.fkDao.FilePutOss(context.Background(), apkFolder, apkFilename, build.AppKey); err != nil {
		log.Error("putBuildToCDN: %v", err)
		return
	}
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if err = s.fkDao.TxUpdateBizApkBuildCDN(tx, apkCdnURL, build.ID); err != nil {
		log.Error("putBuildToCDN: %v", err)
		return
	}
	return
}

// PackBuilds get builds by version_id.
func (s *Service) PackBuilds(c context.Context, appKey, env string, versionID int64) (res []int64, err error) {
	var pv []*cdmdl.Pack
	if pv, err = s.fkDao.PackByVersion(c, appKey, env, versionID); err != nil {
		log.Error("%v", err)
		return
	}
	for _, p := range pv {
		res = append(res, p.BuildID)
	}
	return
}

// PackVersionForHotfix get version.
func (s *Service) PackVersionForHotfix(c context.Context, appKey, env, filter string, pn, ps int) (res []*model.Version, err error) {
	var vs map[int64]*cdmdl.PackItem
	if vs, err = s.fkDao.PackVersionList(c, appKey, env, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	var vids []int64
	for vid := range vs {
		vids = append(vids, vid)
	}
	if len(vids) == 0 {
		log.Error("%v %v %v vids is nil", appKey, env, filter)
		return
	}
	var packs map[int64][]*cdmdl.Pack
	if packs, err = s.fkDao.PackByVersions(c, appKey, env, vids); err != nil {
		log.Error("%v", err)
		return
	}
	var gitlabJobIds []int64
	for _, v := range packs {
		for _, pack := range v {
			gitlabJobIds = append(gitlabJobIds, pack.BuildID)
		}
	}
	var ciPacks []*cimdl.BuildPack
	if ciPacks, err = s.fkDao.SelectBuildPackByJobIds(c, appKey, gitlabJobIds); err != nil {
		log.Errorc(c, "get ci build pack error: %v", err)
		return
	}
	ciMap := groupByBuildID(ciPacks)
	sort.Sort(Keys(vids))
	for _, vid := range vids {
		if ps, ok := packs[vid]; ok {
			for _, p := range ps {
				if filter != "" {
					if !(strings.Contains(vs[vid].Version.Version, filter)) &&
						!(strings.Contains(strconv.FormatInt(vs[vid].VersionCode, 10), filter)) &&
						!(strings.Contains(strconv.FormatInt(p.BuildID, 10), filter)) {
						continue
					}
				}
				var ciEnvVars string
				if v, ok := ciMap[p.BuildID]; ok {
					ciEnvVars = v.CIEnvVars
				}
				re := &model.Version{
					ID:          vid,
					BuildID:     p.BuildID,
					Version:     vs[vid].Version.Version,
					VersionCode: vs[vid].VersionCode,
					CIEnvVars:   ciEnvVars,
				}
				res = append(res, re)
			}
		}
	}
	return
}

// AppCDList get pack list.
func (s *Service) AppCDList(c context.Context, appKey, env string, pn, ps int) (res *cdmdl.PackResult, err error) {
	var (
		total int
		vs    map[int64]*cdmdl.PackItem
		packs map[int64][]*cdmdl.Pack
	)
	if total, err = s.fkDao.PackVersionCount(c, appKey, env); err != nil || total == 0 {
		log.Error("%v or total is 0", err)
		return
	}
	if vs, err = s.fkDao.PackVersionList(c, appKey, env, pn, ps); err != nil || len(vs) == 0 {
		log.Error("%v or version is 0", err)
		return
	}
	var (
		vids []int64
		vers []*model.Version
	)
	for vid, ver := range vs {
		vids = append(vids, vid)
		vers = append(vers, ver.Version)
	}
	if len(vids) == 0 {
		log.Error("AppCDList %v %v %v %v vids is 0", appKey, env, pn, ps)
		return
	}
	if packs, err = s.fkDao.PackByVersions(c, appKey, env, vids); err != nil || len(packs) == 0 {
		log.Error("AppCDList %v or packs is 0", err)
		return
	}
	// get pack config
	var bs []int64
	for _, ps := range packs {
		for _, p := range ps {
			bs = append(bs, p.BuildID)
		}
	}
	if len(bs) == 0 {
		log.Error("AppCDList %v %v %v %v bs is 0", appKey, env, pn, ps)
		return
	}
	// filter config
	var filter map[int64]*cdmdl.FilterConfig
	if filter, err = s.fkDao.PackFilterConfig(c, appKey, env, bs); err != nil {
		log.Error("%v", err)
		return
	}
	var flow map[int64]*cdmdl.FlowConfig
	if flow, err = s.fkDao.PackFlowConfig(c, appKey, env, bs); err != nil {
		log.Error("%v", err)
		return
	}
	appInfo, err := s.fkDao.AppPass(c, appKey)
	for vid, ps := range packs {
		v, ok := vs[vid]
		if !ok {
			log.Error("%v version not matched %v", vid, vs)
			continue
		}
		for _, p := range ps {
			if f, ok := flow[p.BuildID]; ok {
				p.Flow = f.Flow
			}
			if fc, ok := filter[p.BuildID]; ok {
				p.Config = fc
			}
			p.GlJobURL = s.MakeGitPath(appInfo.GitPath, p.BuildID)
			v.Items = append(v.Items, p)
		}
	}
	res = &cdmdl.PackResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
	}
	sort.Sort(Vers(vers))
	for _, ver := range vers {
		res.Items = append(res.Items, vs[ver.ID])
	}
	return
}

// AppCDListFilter get pack list by options.
func (s *Service) AppCDListFilter(c context.Context, appKey, env, filterKey string, steadyState int, hasBbrUrl bool, pn, ps int) (res *cdmdl.PackResult, err error) {
	var (
		total int
		vs    map[int64]*cdmdl.PackItem
		packs map[int64][]*cdmdl.Pack
	)
	if total, err = s.fkDao.PackVersionCountByOptions(c, appKey, env, filterKey, steadyState, hasBbrUrl); err != nil || total == 0 {
		log.Error("%v or total is 0", err)
		return
	}
	if vs, err = s.fkDao.PackVersionListByOptions(c, appKey, env, filterKey, steadyState, hasBbrUrl, pn, ps); err != nil || len(vs) == 0 {
		log.Error("%v or version is 0", err)
		return
	}
	var (
		vids []int64
		vers []*model.Version
	)
	for vid, ver := range vs {
		vids = append(vids, vid)
		vers = append(vers, ver.Version)
	}
	if len(vids) == 0 {
		log.Error("AppCDListByOptions %v %v %v %v vids is 0", appKey, env, pn, ps)
		return
	}
	if packs, err = s.fkDao.PackByVersions(c, appKey, env, vids); err != nil || len(packs) == 0 {
		log.Error("AppCDListByOptions %v or packs is 0", err)
		return
	}
	// get pack config
	var bs []int64
	for _, ps := range packs {
		for _, p := range ps {
			bs = append(bs, p.BuildID)
		}
	}
	if len(bs) == 0 {
		log.Error("AppCDListByOptions %v %v %v %v bs is 0", appKey, env, pn, ps)
		return
	}
	// filter config
	var filter map[int64]*cdmdl.FilterConfig
	if filter, err = s.fkDao.PackFilterConfig(c, appKey, env, bs); err != nil {
		log.Error("%v", err)
		return
	}
	var flow map[int64]*cdmdl.FlowConfig
	if flow, err = s.fkDao.PackFlowConfig(c, appKey, env, bs); err != nil {
		log.Error("%v", err)
		return
	}
	// testflight info
	var tfInfos map[int64]*cdmdl.TestFlightPackInfo
	if tfInfos, err = s.fkDao.TFPackByVersions(c, vids); err != nil {
		log.Error("%v", err)
		return
	}
	appInfo, err := s.fkDao.AppPass(c, appKey)
	for vid, ps := range packs {
		v, ok := vs[vid]
		if !ok {
			log.Error("%v version not matched %v", vid, vs)
			continue
		}
		for _, p := range ps {
			if f, ok := flow[p.BuildID]; ok {
				p.Flow = f.Flow
			}
			if fc, ok := filter[p.BuildID]; ok {
				p.Config = fc
			}
			if tf, ok := tfInfos[p.BuildID]; ok {
				if tf.ExpireTime < 0 {
					tf.ExpireTime = 0
				}
				if tf.RemindUpdTime < 0 {
					tf.RemindUpdTime = 0
				}
				if tf.ForceupdTime < 0 {
					tf.ForceupdTime = 0
				}
				p.TestFlightInfo = tf
			}
			p.GlJobURL = s.MakeGitPath(appInfo.GitPath, p.BuildID)
			v.Items = append(v.Items, p)
		}
	}
	res = &cdmdl.PackResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
	}
	sort.Sort(Vers(vers))
	for _, ver := range vers {
		res.Items = append(res.Items, vs[ver.ID])
	}
	return
}

// AppCDConfigSwitchSet set app config switch.
func (s *Service) AppCDConfigSwitchSet(c context.Context, appKey, env string, versionID int64, isUpgrade bool, userName string) (err error) {
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		var isUp int8
		if isUpgrade {
			isUp = cdmdl.Upgrade
		} else {
			isUp = cdmdl.NotUpgrade
		}
		if _, err = s.fkDao.TxSetPackConfigSwitch(tx, versionID, isUp); err != nil {
			log.Errorc(c, "%v", err)
		}
		return err
	})
	if err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	// databus push
	s.event.Publish(PackGreyPushEvent, PackGreyArgs{Context: utils.CopyTrx(c), AppKey: appKey, Env: env, VersionId: versionID, Operator: userName})
	return
}

// UpgradConfigSet set app upgrad config.
func (s *Service) UpgradConfigSet(c context.Context, appKey, env string, versionID int64, normal, exnormal, force, exforce, system, exSystem string, cycle int, title, content, userName, policyURL, iconURL, confirmBtnText, cancelBtnText string, policy, silent int) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxSetPackUpgradConfig(tx, appKey, env, versionID, normal, exnormal, force, exforce, system, exSystem,
		cycle, title, content, policyURL, iconURL, confirmBtnText, cancelBtnText, policy, silent); err != nil {
		log.Error("%v", err)
		return
	}
	// add log
	var version *model.Version
	if version, err = s.fkDao.PackVersionByID(c, appKey, versionID); err != nil {
		log.Error("%v", err)
		err = nil
		return
	}
	_, _ = s.fkDao.AddLog(c, appKey, env, mngmdl.ModelCD, mngmdl.OperationCDUpgradeConfig,
		fmt.Sprintf("版本: %v(%v)", version.Version, version.VersionCode), userName)
	return
}

// UpgradConfig get app upgrad config.
func (s *Service) UpgradConfig(c context.Context, appKey, env string, versionID int64) (uconfig *cdmdl.UpgradConfig, err error) {
	if uconfig, err = s.fkDao.PackUpgradConfig(c, appKey, env, versionID); err != nil {
		log.Error("%v", err)
	}
	return
}

// FilterConfigSet set filter config.
func (s *Service) FilterConfigSet(c context.Context, appKey, env string, buildID int64, network, isp, channel,
	city string, percent int, device, userName, phoneModel, brand string, status int) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	kinds := [][]int{{10, 48}, {26, 97}}
	keyb := make([]byte, 8)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 8; i++ {
		ikind := rand.Intn(2)
		scope, base := kinds[ikind][0], kinds[ikind][1]
		keyb[i] = uint8(base + rand.Intn(scope))
	}
	salt := string(keyb)
	if _, err = s.fkDao.TxSetPackFilterConfig(tx, appKey, env, buildID, network, isp, channel, city, percent, salt, device, phoneModel, brand, status); err != nil {
		log.Error("%v", err)
		return
	}
	// add log
	_, _ = s.fkDao.AddLog(c, appKey, env, mngmdl.ModelCD, mngmdl.OperationCDFilterConfig, fmt.Sprintf("构建ID: %v", buildID), userName)
	return
}

// FilterConfig get app filter.
func (s *Service) FilterConfig(c context.Context, appKey, env string, buildID int64) (fconfig *cdmdl.FilterConfig, err error) {
	var fm map[int64]*cdmdl.FilterConfig
	if fm, err = s.fkDao.PackFilterConfig(c, appKey, env, []int64{buildID}); err != nil {
		log.Error("%v", err)
		return
	}
	fconfig = fm[buildID]
	return
}

// FlowConfigSet set flow set.
func (s *Service) FlowConfigSet(c context.Context, appKey, env, userName string, flow map[int64]string, versionId int64) (err error) {
	var buildids []string
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		for b, fs := range flow {
			buildids = append(buildids, strconv.FormatInt(b, 10))
			if _, err = s.fkDao.TxSetPackFlowConfig(tx, appKey, env, fs, b); err != nil {
				log.Errorc(c, "%v", err)
			}
		}
		return err
	})
	if err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	// add log
	_, _ = s.fkDao.AddLog(c, appKey, env, mngmdl.ModelCD, mngmdl.OperationCDFlowConfig, fmt.Sprintf("构建ID: %v", strings.Join(buildids, ",")), userName)
	// databus push
	s.event.Publish(PackGreyPushEvent, PackGreyArgs{Context: utils.CopyTrx(c), AppKey: appKey, Env: env, VersionId: versionId, Operator: userName})
	return
}

// FlowConfig get flow config.
func (s *Service) FlowConfig(c context.Context, appKey, env string, versionID int64) (flows map[int64]string, err error) {
	var packs []*cdmdl.Pack
	if packs, err = s.fkDao.PackByVersion(c, appKey, env, versionID); err != nil {
		log.Error("%v", err)
		return
	}
	var buildIDs []int64
	flows = make(map[int64]string)
	for _, p := range packs {
		flows[p.BuildID] = ""
		buildIDs = append(buildIDs, p.BuildID)
	}
	if len(buildIDs) == 0 {
		log.Error("FlowConfig %v %v %v buildIDs is 0", appKey, env, versionID)
		return
	}
	var flow map[int64]*cdmdl.FlowConfig
	if flow, err = s.fkDao.PackFlowConfig(c, appKey, env, buildIDs); err != nil {
		log.Error("%v", err)
		return
	}
	for _, fliter := range flow {
		flows[fliter.BuildID] = fliter.Flow
	}
	return
}

// CDEvolution change cd env.
func (s *Service) CDEvolution(c context.Context, appKey, env, userName string, dispermil int, disLimit, buildID int64) (packID, versionId int64, err error) {
	var (
		pack    *cdmdl.Pack
		version *model.Version
	)
	if pack, err = s.fkDao.PackByBuild(c, appKey, env, buildID); err != nil {
		log.Error("%v", err)
		return
	}
	if pack == nil {
		log.Error("pack is nil appkey: %v;build: %v", appKey, buildID)
		return
	}
	if version, err = s.fkDao.PackVersionByID(c, appKey, pack.VersionID); err != nil {
		log.Error("%v", err)
		return
	}
	if version == nil {
		log.Error("version is nil appkey: %v;build: %v", appKey, buildID)
		return
	}
	var (
		nextEnv       = model.EvolutionEnv(pack.Env)
		nextVersion   *model.Version
		nextVersionID int64
	)
	err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		if nextVersion, err = s.fkDao.PackVersion(c, appKey, nextEnv, version.Version, version.VersionCode); err != nil {
			log.Errorc(c, "PackVersion error %v", err)
			return err
		}
		if nextVersion == nil {
			if nextVersionID, err = s.fkDao.TxSetPackVersion(tx, version.AppID, version.AppKey, nextEnv, version.Version, version.VersionCode, cdmdl.NotUpgrade); err != nil {
				log.Errorc(c, "TxSetPackVersion error %v", err)
				return err
			}
		} else {
			nextVersionID = nextVersion.ID
		}
		if packID, err = s.fkDao.TxSetPack(tx, pack.AppID, pack.AppKey, nextEnv, nextVersionID, pack.InternalVersionCode,
			pack.BuildID, pack.GitType, pack.GitName, pack.Commit, pack.PackType, userName, pack.Size, pack.MD5, pack.PackPath,
			pack.PackURL, pack.MappingURL, pack.RURL, pack.RMappingURL, pack.CDNURL, pack.Desc, userName, pack.ChangeLog, pack.DepGitJobId, pack.IsCompatible, pack.BbrUrl, pack.Features); err != nil {
			log.Errorc(c, "TxSetPack error %v", err)
			return err
		}
		return err
	})
	if err != nil {
		log.Errorc(c, "Transact error %v", err)
		return
	}
	// push eventbus
	log.Infoc(c, "pack env %v,model.EvolutionEnv %v,nextVersionID %v", pack.Env, nextEnv, nextVersionID)
	s.event.Publish(PackGreyPushEvent, PackGreyArgs{Context: utils.CopyTrx(c), AppKey: appKey, Env: "prod", VersionId: nextVersionID})
	_, _ = s.fkDao.AddLog(c, appKey, "test", mngmdl.ModelCD, mngmdl.OperationCDPushProd, fmt.Sprintf("构建ID: %v", buildID), userName)
	defer func() {
		// 推送到正式环境时 modules配置表复制一份最新配置
		if nextEnv == "prod" {
			log.Warn("ModulesConfCopy start appKey %v version %v", appKey, version.Version)
			if err = s.mdlSvr.ModulesConfCopy(c, appKey, version.Version); err != nil {
				log.Error("ModulesConfCopy: %v", err)
			}
			log.Warn("ModulesConfCopy end")
		}
	}()
	var (
		tfInfo *cdmdl.TestFlightPackInfo
	)
	// 如果 app 含有 testflight 信息，则推送正式的时候也需要复制一份 pack_tf_attr 信息
	if tfInfo, err = s.fkDao.TFPackByPackID(context.Background(), pack.ID); err != nil {
		log.Error("UpdateOnlineVersProc TFPackByPackID: %v", err)
		return
	}
	if tfInfo != nil {
		if err = s.setTFPackInfoProd(tfInfo, BetaStateTesting, dispermil, disLimit, packID); err != nil {
			log.Error("setTFPackInfoProd: %v", err)
			return
		}
		//nolint:gomnd
		if pack.PackType != 9 {
			return
		}
		// tf 包需要加到正式环境的测试组里
		var buildsIDs []string
		buildsIDs = append(buildsIDs, tfInfo.BetaBuildID)
		if _, err = s.appstoreClient.BetaGroups.AddBuilds(appKey, tfInfo.BetaGroupID, buildsIDs); err != nil {
			log.Error("s.appstoreClient.BetaGroups.AddBuilds: %v", err)
			return
		}
	}
	versionId = nextVersionID
	return
}

// GenerateList get generate list
func (s *Service) GenerateList(c context.Context, appKey, env, filterKey, order, sort string, buildID, groupID int64, pn, ps int) (result *cdmdl.ChannelResult, err error) {
	var (
		acs   []*appmdl.Channel
		total int
		res   []*cdmdl.ChannelGenerate
	)
	if total, err = s.fkDao.AppChannelListCount(c, appKey, filterKey, groupID); err != nil || total == 0 {
		log.Error("%v or total is 0", err)
		return
	}
	if acs, err = s.fkDao.AppChannelList(c, appKey, filterKey, order, sort, pn, ps, groupID); err != nil {
		log.Error("s.fdDao.AppChannelList failed. %v", err)
		return
	}
	var generates []*cdmdl.Generate
	if generates, err = s.fkDao.GenerateList(c, appKey, buildID); err != nil {
		log.Error("%v", err)
		return
	}
	for _, ac := range acs {
		re := &cdmdl.ChannelGenerate{
			Channel: &appmdl.Channel{},
		}
		*re.Channel = *ac
		re.State = -1
		for _, g := range generates {
			if g.ChannelID == ac.ID {
				re.Generate = &cdmdl.Generate{}
				*re.Generate = *g
				re.State = g.Status
			}
		}
		res = append(res, re)
	}
	result = &cdmdl.ChannelResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
	}
	result.Items = res
	return
}

// AppCDGenerateAdd create generate.
func (s *Service) AppCDGenerateAdd(c context.Context, appKey, userName string, buildID, channelID int64) (err error) {
	var (
		pack     *cdmdl.Pack
		version  *model.Version
		channel  *appmdl.Channel
		size     int64
		fmd5     string
		pkgFile  *os.File
		fileInfo os.FileInfo
	)
	if pack, err = s.fkDao.PackByBuild(c, appKey, "prod", buildID); err != nil {
		log.Error("%v", err)
		return
	}
	if pack == nil {
		log.Error("pack is nil")
		return
	}
	if version, err = s.fkDao.PackVersionByID(c, appKey, pack.VersionID); err != nil {
		log.Error("%v", err)
		return
	}
	if version == nil {
		log.Error("version is nil")
		return
	}
	if channel, err = s.fkDao.GetChannelByID(c, channelID); err != nil {
		log.Error("%v", err)
		return
	}
	if channel == nil {
		log.Error("channel is nil")
		return
	}
	folder := path.Join("pack", pack.AppKey, strconv.FormatInt(pack.BuildID, 10), "channel")
	dest := path.Join(s.c.LocalPath.LocalDir, folder)
	_, err = os.Stat(dest)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dest, 0755); err != nil {
			log.Error("%v", err)
			return
		}
		err = nil
	}
	// 若没有构建过渠道包. 但是存在渠道包. 则进行移除操作
	if generate, _ := s.fkDao.GenerateByOptions(c, appKey, channelID, buildID); generate == nil {
		// 待生成渠道包的名称
		name, ext := channel2.FileNameAndExt(pack.PackPath)
		channelPackPath := filepath.Join(dest, name+"-"+channel.Code+ext)
		// 判断文件是否存在
		isExist := false
		if _, fileError := os.Stat(channelPackPath); fileError == nil {
			isExist = true
		} else {
			if os.IsExist(fileError) {
				isExist = true
			}
		}
		// 若文件存在. 移除渠道包
		if isExist {
			_ = os.Rename(channelPackPath, fmt.Sprintf("%v-delete-%v", channelPackPath, rand.Int()))
		}
	}
	var localPath string
	if localPath, err = channel2.GenerateChannelApk(dest, channel.Code, nil, pack.PackPath, false, false); err != nil {
		log.Error("%v", err)
		return
	}
	var filename, inetPath string
	filename = path.Base(localPath)
	inetPath = s.c.LocalPath.LocalDomain + "/" + path.Join(folder, filename)
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	// 文件大小
	if fileInfo, err = os.Stat(localPath); err != nil {
		log.Error("os.Stat(%s) error(%v)", localPath, err)
		return
	}
	size = fileInfo.Size()
	// md5
	buf := new(bytes.Buffer)
	if pkgFile, err = os.Open(localPath); err != nil {
		log.Error("os.Open(%s) error(%v)", localPath, err)
		return
	}
	if _, err = io.Copy(buf, pkgFile); err != nil {
		log.Error("error(%v)", err)
		return
	}
	md5Bs := md5.Sum(buf.Bytes())
	fmd5 = hex.EncodeToString(md5Bs[:])
	if _, err = s.fkDao.TxSetGenerate(tx, pack.AppKey, pack.BuildID, channelID, size, filename, folder, localPath, inetPath, fmd5, userName); err != nil {
		log.Error("%v", err)
	}
	return
}

func (s *Service) AppCDGeneratePackUpload(ctx context.Context, appKey, userName string, buildId int64, file multipart.File, header *multipart.FileHeader) (re *cdmdl.UploadResult, err error) {
	var (
		cd       *cdmdl.Pack
		filePath string
	)
	log.Infoc(ctx, fmt.Sprintf("开始上传渠道包产物 appkey[%v], userName[%v], buildId[%v], fileName[%v]", appKey, userName, buildId, header.Filename))
	if cd, err = s.fkDao.PackByBuild(ctx, appKey, string(cdmdl.EnvProd), buildId); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("PackByBuild error: %v, build_id: %v", err, buildId))
		log.Errorc(ctx, "%v", err)
		return
	}
	if cd == nil {
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("pack is nil, build_id = %v", buildId))
		log.Errorc(ctx, "%v", err)
		return
	}
	destFileDir := path.Join(conf.Conf.LocalPath.LocalDir, "pack", cd.AppKey, strconv.FormatInt(cd.BuildID, 10), "channel")
	if filePath, err = utils.MultipartFileCopy(file, header, destFileDir); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("文件拷贝失败，目标路径:%v, error: %v,", path.Join(destFileDir, header.Filename), err))
		log.Errorc(ctx, "%v", err)
		return
	}
	re = &cdmdl.UploadResult{
		FilePath: filePath,
	}
	return
}

// AppCDGenerateAddGit create generate to git pipeline.
func (s *Service) AppCDGenerateAddGit(c context.Context, appKey, channels, userName string, buildID int64) (err error) {
	var (
		pack                 *cdmdl.Pack
		channelList          []*cdmdl.ChannelGeneParam
		needAddNewChannels   []*cdmdl.ChannelGeneParam
		generGitPipeChannels []*cdmdl.ChannelGitPipe
		appInfo              *appmdl.APP
		version              *model.Version
		channelIDs           []int64
		sqls                 []string
		args                 []interface{}
		triggerAppkey        = "android"
	)
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Errorc(c, "app pass error %v", err)
		return
	}
	if appInfo.Platform != "android" {
		triggerAppkey = appKey
	}
	if pack, err = s.fkDao.PackByBuild(c, appKey, "prod", buildID); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if pack == nil {
		log.Errorc(c, "pack is nil")
		return
	}
	if version, err = s.fkDao.PackVersionByID(c, appKey, pack.VersionID); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if version == nil {
		log.Errorc(c, "version is nil")
		return
	}
	if err = json.Unmarshal([]byte(channels), &channelList); err != nil {
		log.Errorc(c, "AppCDGenerateAddGit json.Unmarshal(%s) error(%v)", string(channels), err)
		return
	}
	// 过滤出需要新增的数据， 和需要重新打包的数据
	for _, channel := range channelList {
		channelIDs = append(channelIDs, channel.ChannelID)
	}
	existGenerated, _ := s.fkDao.GeneratedByChannel(c, channelIDs, appKey, buildID)
	if len(existGenerated) != 0 {
		for _, channel := range channelList {
			isExist := false
			for _, egene := range existGenerated {
				if channel.ChannelID == egene.ChannelID {
					isExist = true
					break
				}
			}
			// 需要新增的渠道数据
			if !isExist {
				needAddNewChannels = append(needAddNewChannels, channel)
			}
		}
	} else {
		needAddNewChannels = channelList
	}
	for _, ngchannel := range needAddNewChannels {
		sqls = append(sqls, "(?,?,?,?,?,?,?,?,?)")
		args = append(args, appKey, pack.BuildID, ngchannel.ChannelID, -2, "", "", "", "", userName)
	}
	if len(args) != 0 {
		if err = s.setGeneratesByGit(c, sqls, args); err != nil {
			log.Errorc(c, "setGeneratesByGit error: %v", err)
			return
		}
	}
	// 获取符合条件的渠道ID 和 提交的渠道名匹配
	insertGenerates, _ := s.fkDao.GeneratedByChannel(c, channelIDs, appKey, buildID)
	// 需要批量更新状态的数据
	var needUpStatusIDs []string
	for _, gen := range insertGenerates {
		for _, chl := range channelList {
			// 只有处于 失败 -4， 队列中 -2， 初始状态 -1下才需要重新打包
			if gen.ChannelID == chl.ChannelID && (gen.Status == cdmdl.GenerateUnhandle || gen.Status == cdmdl.GenerateQueue || gen.Status == cdmdl.GenerateFailed) {
				param := &cdmdl.ChannelGitPipe{}
				param.ID = gen.ID
				param.Channel = chl.Channel
				generGitPipeChannels = append(generGitPipeChannels, param)
				needUpStatusIDs = append(needUpStatusIDs, strconv.FormatInt(gen.ID, 10))
			}
		}
	}
	if len(generGitPipeChannels) == 0 {
		err = errors.New("渠道已生成或在生成中，刷新后查看，请勿重复操作")
		return
	}
	// 重置状态为排队中
	if _, err = s.fkDao.TxUpGenerateStatusByIDs(c, appKey, strings.Join(needUpStatusIDs, ","), cdmdl.GenerateQueue); err != nil {
		log.Errorc(c, "TxUpGenerateStatusByIDs %v", err)
		return
	}
	generGitPipeChannelsJSON, err := json.Marshal(generGitPipeChannels)
	if err != nil {
		log.Errorc(c, "generGitPipeChannels Marshal %v", err)
		return
	}
	var variables = map[string]string{
		"APP_KEY":        appKey,
		"TASK":           "CHANNEL",
		"BUILD_ID":       strconv.FormatInt(buildID, 10),
		"ORIGIN_APK_URL": pack.CDNURL,
		"LOCAL_APK_URL":  pack.PackURL,
		"CHANNELS":       string(generGitPipeChannelsJSON),
	}
	if _, err = s.gitSvr.TriggerPipeline(c, triggerAppkey, 0, cdmdl.GenerateGitName, variables); err != nil {
		if _, err = s.fkDao.TxUpGenerateStatusByIDs(c, appKey, strings.Join(needUpStatusIDs, ","), cdmdl.GenerateFailed); err != nil {
			log.Errorc(c, "TxUpGenerateStatusByIDs %v", err)
			return
		}
	}
	return
}

func (s *Service) setGeneratesByGit(c context.Context, sqls []string, args []interface{}) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxSetGeneratesByGit(tx, sqls, args); err != nil {
		log.Error("TxSetGeneratesByGit %v", err)
	}
	return
}

// GeneratesUpdate update generate info.
func (s *Service) GeneratesUpdate(c context.Context, appKey, channelFileInfoStr string, jobID int64, channelStatus int) (err error) {
	var (
		channelFileInfo []*cdmdl.ChannelFileInfo
		sqls            []string
		args            []interface{}
	)
	appInfo, err := s.fkDao.AppPass(c, appKey)
	if err != nil {
		log.Error("GeneratesUpdate appPass error: %v", err)
		return
	}
	glJobURL := s.MakeGitPath(appInfo.GitPath, jobID)
	if err = json.Unmarshal([]byte(channelFileInfoStr), &channelFileInfo); err != nil {
		log.Error("GeneratesUpdate json.Unmarshal(%s) error(%v)", string(channelFileInfoStr), err)
		return
	}
	// id,name,folder,patch_path,patch_url,status,size,md5
	for _, cfi := range channelFileInfo {
		if channelStatus == cdmdl.GenerateSuccess {
			if cfi.Path == "" {
				err = errors.New("ID:" + strconv.FormatInt(cfi.ID, 10) + "路径为空，无法解析")
				return
			}
			fileName := path.Base(cfi.Path)
			fileFolder := strings.Replace(strings.Replace(cfi.Path, s.c.LocalPath.LocalDir, "", -1), fileName, "", -1)
			fileURL := s.c.LocalPath.LocalDomain + "/" + path.Join(fileFolder, fileName)
			args = append(args, cfi.ID, fileName, fileFolder[1:len(fileFolder)-1], cfi.Path, fileURL, channelStatus, cfi.Size, cfi.MD5, glJobURL)
		} else {
			args = append(args, cfi.ID, "", "", cfi.Path, "", channelStatus, cfi.Size, cfi.MD5, glJobURL)
		}
		sqls = append(sqls, `(?,?,?,?,?,?,?,?,?)`)
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
		s.event.Publish(PackGenerateUpdateEvent, PackGenerateArgs{Ctx: utils.CopyTrx(c), AppKey: appKey, ChannelStatus: channelStatus, ChannelPacks: channelFileInfo})
	}()
	if _, err = s.fkDao.TxUpGeneratesByGit(tx, sqls, args); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppCDGenerateAdds create generates.
func (s *Service) AppCDGenerateAdds(c context.Context, appKey, userName string, buildID int64) (err error) {
	var (
		pack     *cdmdl.Pack
		version  *model.Version
		channels []*appmdl.Channel
	)
	if pack, err = s.fkDao.PackByBuild(c, appKey, "prod", buildID); err != nil {
		log.Error("%v", err)
		return
	}
	if pack == nil {
		log.Error("pack is nil")
		return
	}
	if version, err = s.fkDao.PackVersionByID(c, appKey, pack.VersionID); err != nil {
		log.Error("%v", err)
		return
	}
	if version == nil {
		log.Error("version is nil")
		return
	}
	if channels, err = s.fkDao.AppChannelList(c, pack.AppKey, "", "", "", -1, -1, -1); err != nil {
		log.Error("%v", err)
		return
	}
	if len(channels) == 0 {
		log.Error("AppCDGenerateAdds %v %v %v channels is 0", appKey, userName, buildID)
		return
	}
	var (
		sqls []string
		args []interface{}
	)
	for _, channel := range channels {
		folder := path.Join("pack", pack.AppKey, strconv.FormatInt(pack.BuildID, 10), "channel")
		dest := path.Join(s.c.LocalPath.LocalDir, folder)
		_, err = os.Stat(dest)
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dest, 0755); err != nil {
				log.Error("%v", err)
				return
			}
			err = nil
		}
		var localPath string
		if localPath, err = channel2.GenerateChannelApk(dest, channel.Code, nil, pack.PackPath, false, false); err != nil {
			log.Error("%v", err)
			return
		}
		var filename, inetPath string
		filename = path.Base(localPath)
		inetPath = s.c.LocalPath.LocalDomain + "/" + path.Join(folder, filename)
		sqls = append(sqls, "?,?,?,?,?,?,?,?")
		args = append(args, pack.AppKey, pack.BuildID, channel.ID, filename, folder, localPath, inetPath, userName)
	}
	if len(args) > 0 {
		var tx *sql.Tx
		if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
			log.Error("s.fkDao.BeginTran() error(%v)", err)
			return
		}
		defer func() {
			if r := recover(); r != nil {
				//nolint:errcheck
				tx.Rollback()
				log.Error("%v", r)
			}
			if err != nil {
				if err1 := tx.Rollback(); err1 != nil {
					log.Error("tx.Rollback() error(%v)", err1)
				}
				return
			}
			if err = tx.Commit(); err != nil {
				log.Error("tx.Commit() error(%v)", err)
			}
		}()
		if _, err = s.fkDao.TxSetGenerates(tx, sqls, args); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

// AppCDGenerateStatus update app generate status.
func (s *Service) AppCDGenerateStatus(c context.Context, appKey, userName string, id int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxUpGenerateStatus(tx, appKey, userName, id, cdmdl.GenerateTest); err != nil {
		log.Error("%v", err)
		return
	}
	// add log
	var generate *cdmdl.Generate
	if generate, err = s.fkDao.Generate(c, appKey, id); err != nil {
		log.Error("%v", err)
		return
	}
	var p *cdmdl.Pack
	if p, err = s.fkDao.PackByBuild(c, appKey, "prod", generate.BuildID); err != nil {
		log.Error("%v", err)
		err = nil
		return
	}
	var ch *appmdl.Channel
	if ch, err = s.fkDao.GetChannelByID(c, generate.ChannelID); err != nil {
		log.Error("%v", err)
		err = nil
		return
	}
	logInfo := fmt.Sprintf("版本：%v(%v)  渠道：%v(%v)", p.Version, p.VersionCode, ch.Name, ch.Code)
	_, _ = s.fkDao.AddLog(c, appKey, "prod", mngmdl.ModelChannelPack, mngmdl.OperationChannelPackTest, logInfo, userName)
	return
}

// AppCDGenerateUpload upload generate file.
func (s *Service) AppCDGenerateUpload(c context.Context, appKey, userName string, id int64) (err error) {
	var (
		generate *cdmdl.Generate
		url      string
	)
	if generate, err = s.fkDao.Generate(c, appKey, id); err != nil {
		log.Error("%v", err)
		return
	}
	if generate == nil {
		log.Error("%v %v generate is nil", appKey, id)
		return
	}
	if generate.Status == cdmdl.GenerateUpload {
		log.Error("%v %v 包已经上传到CDN", appKey, id)
		return
	}
	if url, _, _, err = s.fkDao.FilePutOss(c, generate.Folder, generate.Name, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxUpGenerateCDN(tx, appKey, url, userName, cdmdl.GenerateUpload, id); err != nil {
		log.Error("%v", err)
		return
	}
	// 保存独占的cdn url
	if _, err = s.fkDao.TxUpGenerateSoleCDN(tx, appKey, url, id); err != nil {
		log.Error("%v", err)
		return
	}
	// add log
	var p *cdmdl.Pack
	if p, err = s.fkDao.PackByBuild(c, appKey, "prod", generate.BuildID); err != nil {
		log.Error("%v", err)
		err = nil
		return
	}
	var ch *appmdl.Channel
	if ch, err = s.fkDao.GetChannelByID(c, generate.ChannelID); err != nil {
		log.Error("%v", err)
		err = nil
		return
	}
	logInfo := fmt.Sprintf("版本: %v(%v)  渠道: %v(%v)", p.Version, p.VersionCode, ch.Name, ch.Code)
	_, _ = s.fkDao.AddLog(c, appKey, "prod", mngmdl.ModelChannelPack, mngmdl.OperationChannelPackPushCDN, logInfo, userName)
	return
}

// AppCDGeneratePublish publish generate.
func (s *Service) AppCDGeneratePublish(c context.Context, appKey, userName string, id int64) (err error) {
	var (
		generate        *cdmdl.Generate
		url, from, dest string
	)
	if generate, err = s.fkDao.Generate(c, appKey, id); err != nil {
		log.Error("%v", err)
		return
	}
	if generate == nil {
		log.Error("%v %v generate is nil", appKey, id)
		return
	}
	var (
		appChannels []*appmdl.Channel
		appChannel  *appmdl.Channel
	)
	if appChannels, err = s.fkDao.AppChannelList(context.Background(), appKey, "", "", "", -1, -1, -1); err != nil {
		log.Error("%v", err)
		return
	}
	if len(appChannels) == 0 {
		log.Error("AppCDGeneratePublish appkey(%v) channel is nil", appKey)
		return
	}
	for _, as := range appChannels {
		if as.ID == generate.ChannelID {
			appChannel = as
			break
		}
	}
	if appChannel == nil {
		log.Error("AppCDGeneratePublish appkey(%v) channel_id(%v) not exist", appKey, generate.ChannelID)
		return
	}
	from = path.Join(generate.Folder, generate.Name)
	ns := strings.Split(strings.TrimRight(generate.Name, ".apk"), "-")
	//nolint:gomnd
	if len(ns) < 2 {
		log.Error("AppCDGeneratePublish filename(%v) is err", generate.Name)
		return
	}
	filename := ns[0] + "-" + appChannel.Code + ".apk"
	// android 粉板特殊处理
	if appKey == "android" {
		dest = "iBiliPlayer-" + appChannel.Code + ".apk"
	} else if appKey == "android_i" {
		// android 国际版特殊处理
		dest = "iBiliPlayer-internation-" + appChannel.Code + ".apk"
	} else {
		dest = path.Join(generate.AppKey, filename)
	}
	// android 国际版特殊处理
	if appKey == "android_i" {
		// 国际版第一版线上地址不能变，否则影响线上
		if url, err = s.fkDao.PublishWithDir(from, dest, s.c.Oss.Inland.OriginDir+"/upload", appKey); err != nil {
			log.Error("%v", err)
			return
		}
	} else {
		// 正常处理逻辑
		if url, err = s.fkDao.Publish(from, dest, appKey); err != nil {
			log.Error("%v", err)
			return
		}
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxUpGenerateCDN(tx, appKey, url, userName, cdmdl.GeneratePublish, id); err != nil {
		log.Error("%v", err)
		return
	}
	// 记录线上文件的基础信息
	if _, err = s.fkDao.TxSetGeneratePublish(tx, appKey, generate.ChannelID, generate.ID); err != nil {
		log.Error("%v", err)
		return
	}
	// 刷新CDN文件
	_ = s.fkDao.AppCDRefreshCDN(strings.Split(url, ","))
	// 追加操作记录
	var p *cdmdl.Pack
	if p, err = s.fkDao.PackByBuild(c, appKey, "prod", generate.BuildID); err != nil {
		log.Error("%v", err)
		err = nil
		return
	}
	var ch *appmdl.Channel
	if ch, err = s.fkDao.GetChannelByID(c, generate.ChannelID); err != nil {
		log.Error("%v", err)
		err = nil
		return
	}
	logInfo := fmt.Sprintf("版本: %v(%v)  渠道: %v(%v)", p.Version, p.VersionCode, ch.Name, ch.Code)
	_, _ = s.fkDao.AddLog(c, appKey, "prod", mngmdl.ModelChannelPack, mngmdl.OperationChannelPackPushProd, logInfo, userName)
	return
}

// AppCDGeneratePublishList get publish generate list.
func (s *Service) AppCDGeneratePublishList(c context.Context, cdnUrl string) (gp *cdmdl.GeneratePublishLastest, err error) {
	if gp, err = s.fkDao.GetGeneratePublish(c, cdnUrl); err != nil {
		log.Error("%v", err)
	}
	return
}

// SyncMacross sync data to macross
func (s *Service) SyncMacross(c context.Context, appKey string, buildID int64, isGray int) (err error) {
	var (
		pack      *cdmdl.Pack
		v, tv, ov *model.Version
		patches   []*cdmdl.Patch
	)
	// 其他仓库不需要同步 macross
	if appKey != "android" && appKey != "w19e" {
		return
	}
	if pack, err = s.fkDao.PackByBuild(c, appKey, "prod", buildID); err != nil {
		log.Error("%v", err)
		return
	}
	if v, err = s.fkDao.PackVersionByID(c, appKey, pack.VersionID); err != nil {
		log.Error("%v", err)
		return
	}
	pack.CDNURL = strings.Replace(pack.CDNURL, "https://", "http://", 1)
	pack.PackURL = strings.Replace(pack.PackURL, "https://", "http://", 1)
	pack.MappingURL = strings.Replace(pack.MappingURL, "https://", "http://", 1)
	if err = s.syncApk(pack, isGray, v.Version, v.VersionCode); err != nil {
		log.Error("%v", err)
		return
	}
	if patches, err = s.fkDao.PatchList(context.Background(), appKey, pack.BuildID, 1, 50); err != nil {
		log.Error("%v", err)
		return
	}
	for _, patch := range patches {
		if tv, err = s.fkDao.PackVersionByID(context.Background(), appKey, patch.TargetVersionID); err != nil {
			log.Error("%v", err)
			return
		}
		if ov, err = s.fkDao.PackVersionByID(context.Background(), appKey, patch.OriginVersionID); err != nil {
			log.Error("%v", err)
			return
		}
		patch.CDNURL = strings.Replace(patch.CDNURL, "https://", "http://", 1)
		patch.PatchURL = strings.Replace(patch.PatchURL, "https://", "http://", 1)
		if err = s.syncPatch(patch, tv.Version, ov.Version, tv.VersionCode, ov.VersionCode, tv.BuildID, ov.BuildID); err != nil {
			log.Error("%v", err)
			return
		}
	}
	// add log
	_, _ = s.fkDao.AddLog(c, appKey, "prod", mngmdl.ModelCD, mngmdl.OperationCDSyncMacross, fmt.Sprintf("build_id：%v", buildID), "system")
	return
}

func (s *Service) syncApk(pack *cdmdl.Pack, isGray int, version string, versionCode int64) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginMacrossTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginMacrossTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxAddApk(tx, versionCode, pack.BuildID, version, pack.CDNURL, pack.PackURL, pack.MappingURL,
		pack.PackPath, pack.MD5, pack.Size, isGray); err != nil {
		log.Error("TxAddApk() error(%v)", err)
	}
	return
}

func (s *Service) syncPatch(patch *cdmdl.Patch, targetVersion, originVersion string, targetVersionCode, originVersionCode,
	targetBuildID, originBuildID int64) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginMacrossTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginMacrossTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxAddDiffPatch(tx, targetVersion, originVersion, targetBuildID, originBuildID, targetVersionCode,
		originVersionCode, patch.CDNURL, patch.PatchURL, patch.PatchPath, patch.MD5, patch.Size); err != nil {
		log.Error("TxAddDiffPatch() error(%v)", err)
	}
	return
}

// SyncManager sync data to manager.
func (s *Service) SyncManager(c context.Context, appKey, md5, channel, userName string, buildID int64, isGray, isPush int) (err error) {
	var (
		tx      *sql.Tx
		pack    *cdmdl.Pack
		v       *model.Version
		upgrad  *cdmdl.UpgradConfig
		appInfo *appmdl.APP
		vid     int64
		upID    int64
	)
	if tx, err = s.fkDao.BeginShowTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginShowTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	// 只有设置了 app_attribute 表中的 manager_plat 才可以进行同步操作
	if appInfo.ManagerPlat == -1 {
		return
	}
	if pack, err = s.fkDao.PackByBuild(c, appKey, "prod", buildID); err != nil {
		log.Error("%v", err)
		return
	}
	if v, err = s.fkDao.PackVersionByID(c, appKey, pack.VersionID); err != nil {
		log.Error("%v", err)
		return
	}
	if upgrad, err = s.fkDao.PackUpgradConfig(c, appKey, "prod", v.ID); err != nil {
		log.Error("%v", err)
		return
	}
	if vid, err = s.fkDao.TxInManagerVersion(tx, appInfo.ManagerPlat, isGray, v.VersionCode, upgrad.Content, v.Version); err != nil {
		log.Error("%v", err)
		return
	}
	if upID, err = s.fkDao.TxInManagerVersionUpdate(tx, 100, 1, 0, 0, 0, isPush,
		0, 100, vid, pack.Size, channel, pack.CDNURL, md5, "", "指定版本导入更新", "", ""); err != nil {
		log.Error("%v", err)
		return
	}
	// 安卓粉特殊版本预处理
	if appKey == "android" {
		if _, err = s.fkDao.TxInManagerVersionUpdateLimit(tx, upID, []int{5230000, 5230001}, "ne"); err != nil {
			log.Error("%v", err)
			return
		}
	}
	// add log
	_, _ = s.fkDao.AddLog(c, appKey, "prod", mngmdl.ModelCD, mngmdl.OperationCDSyncManager,
		fmt.Sprintf("版本：%v(%v), build_id：%v, 是否灰度：%d", v.Version, v.VersionCode, buildID, isGray), userName)
	return
}

// AppCDGenerateTestStateSet set test state of app cd
//func (s *Service) AppCDGenerateTestStateSet(c context.Context, appKey, idsStr, userName string, testState int) (err error) {
//	var tx *sql.Tx
//	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
//		log.Error("s.fkDao.BeginTran() error(%v)", err)
//		return
//	}
//	defer func() {
//		if r := recover(); r != nil {
//			//nolint:errcheck
//			tx.Rollback()
//			log.Error("%v", r)
//		}
//		if err != nil {
//			if err1 := tx.Rollback(); err1 != nil {
//				log.Error("tx.Rollback() error(%v)", err1)
//			}
//			return
//		}
//		if err = tx.Commit(); err != nil {
//			log.Error("tx.Commit() error(%v)", err)
//		}
//	}()
//	if _, err = s.fkDao.TxAppCDGenerateTestStateSet(tx, appKey, idsStr, testState); err != nil {
//		log.Error("%v", err)
//	}
//	return
//}

// AppCDPackSteadyStateSet set steady state of app cd
func (s *Service) AppCDPackSteadyStateSet(c context.Context, appKey, description string, buildID int64, steadyState int, autoGenChannelPack int64) (err error) {
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
		if autoGenChannelPack == 1 && steadyState == 1 {
			// 稳定版本&&自动构建
			s.event.Publish(SteadyPackAutoGenChannelPack, AutoGenChannelPackArgs{Ctx: utils.CopyTrx(c), AppKey: appKey, BuildId: buildID})
		}
	}()
	if _, err = s.fkDao.TxUpPackSteadyState(tx, appKey, description, buildID, steadyState); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppCDRefreshCDN refresh app package cdn url
func (s *Service) AppCDRefreshCDN(c context.Context, cdnUrls string) (err error) {
	if err = s.fkDao.AppCDRefreshCDN(strings.Split(cdnUrls, ",")); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppCDCustomChannelAdd add custom channel package
func (s *Service) AppCDCustomChannelAdd(c context.Context, appKey, userName string, buildID int64) (err error) {
	var (
		file *os.File
		bs   []byte
	)
	folder := path.Join("pack", appKey, strconv.FormatInt(buildID, 10))
	dest := path.Join(s.c.LocalPath.LocalDir, folder)
	configPath := path.Join(dest, "channel_packs.txt")
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		return
	}
	if file, err = os.Open(configPath); err != nil {
		log.Error("%v", err)
		return
	}
	defer file.Close()
	if bs, err = ioutil.ReadAll(file); err != nil {
		log.Error("%v", err)
		return
	}
	textContent := string(bs)
	packNames := strings.Split(textContent, "\n")
	var (
		sqls []string
		args []interface{}
	)
	for _, packName := range packNames {
		if packName == "" {
			continue
		}
		packPath := path.Join(dest, packName)
		searchPack, _ := s.fkDao.CustomChannelPack(c, appKey, packPath, buildID)
		if searchPack != nil {
			continue
		}
		packURL := strings.Replace(packPath, s.c.LocalPath.LocalDir, s.c.LocalPath.LocalDomain, -1)
		sqls = append(sqls, "(?,?,?,?,?,?)")
		args = append(args, appKey, buildID, packName, packPath, packURL, userName)
	}
	if len(args) > 0 {
		var tx *sql.Tx
		if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
			log.Error("s.fkDao.BeginTran() error(%v)", err)
			return
		}
		defer func() {
			if r := recover(); r != nil {
				//nolint:errcheck
				tx.Rollback()
				log.Error("%v", r)
			}
			if err != nil {
				if err1 := tx.Rollback(); err1 != nil {
					log.Error("tx.Rollback() error(%v)", err1)
				}
				return
			}
			if err = tx.Commit(); err != nil {
				log.Error("tx.Commit() error(%v)", err)
			}
		}()
		if _, err = s.fkDao.TxAddCustomChannelPacks(tx, sqls, args); err != nil {
			log.Error("%v", err)
		}
	}
	return
}

// AppCDCustomChannelList show custom app channel pacakge list
func (s *Service) AppCDCustomChannelList(c context.Context, appKey string, buildID int64) (res []*cdmdl.CustomChannelPack, err error) {
	if res, err = s.fkDao.CustomChannelPacks(c, appKey, buildID); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppCDCustomChannelUpload upload custom app channel pacakge
func (s *Service) AppCDCustomChannelUpload(c context.Context, appKey, userName string, id int64) (err error) {
	var (
		tx       *sql.Tx
		pack     *cdmdl.CustomChannelPack
		url, md5 string
		size     int64
	)
	if pack, err = s.fkDao.CustomChannelPackByID(c, appKey, id); err != nil {
		log.Error("%v", err)
		return
	}
	folder := path.Join("pack", appKey, strconv.FormatInt(pack.BuildID, 10))
	if url, md5, size, err = s.fkDao.FilePutOss(c, folder, pack.PackName, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxUpCustomChannelPack(tx, appKey, url, md5, userName, 1, id, size); err != nil {
		log.Error("%v", err)
	}
	return
}

// ReleaseNotify
func (s *Service) ReleaseNotify(c context.Context, buildId int64, appKey, env, bots string, notifyGroup bool) (err error) {
	var (
		app                                  *appmdl.APP
		pack                                 *cdmdl.Pack
		upgrad                               *cdmdl.UpgradConfig
		tfInfo                               *cdmdl.TestFlightPackInfo
		fconfigs                             map[int64]*cdmdl.FlowConfig
		flow, tfFlowSum                      int
		mailTemp, weChatTemp, upgradeContent string
	)
	if app, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if pack, err = s.fkDao.PackByBuild(c, appKey, env, buildId); err != nil {
		log.Error("%v", err)
		return
	}
	if upgrad, err = s.fkDao.PackUpgradConfig(c, appKey, env, pack.VersionID); err != nil {
		log.Error("%v", err)
		return
	}
	if upgrad != nil {
		upgradeContent = upgrad.Content
	}
	if fconfigs, err = s.fkDao.PackFlowConfig(c, appKey, env, []int64{buildId}); err != nil {
		log.Error("%v", err)
		return
	}
	if tfInfo, err = s.fkDao.TFPackByPackID(context.Background(), pack.ID); err != nil {
		log.Error("%v", err)
		return
	}
	if tfInfo != nil {
		tfFlowSum = tfInfo.DisPermil
	}
	if len(fconfigs) > 0 {
		f := fconfigs[buildId]
		comps := strings.Split(f.Flow, ",")
		var num0, num1 int
		if num1, err = strconv.Atoi(comps[1]); err != nil {
			log.Error("%v", err)
			return
		}
		if num0, err = strconv.Atoi(comps[0]); err != nil {
			log.Error("%v", err)
			return
		}
		flow = num1 - num0 + 1
	}
	// 推送邮件
	templateModel := template.CDReleaseTemplate{
		AppName:            app.Name,
		Version:            pack.Version,
		VersionCode:        pack.VersionCode,
		PackType:           pack.PackType,
		UpgradeContent:     upgradeContent,
		UpgradeContentHTML: strings.Replace(upgradeContent, "\n", "<br/>", -1),
		FlowSum:            flow,
		TFFlowSum:          tfFlowSum / 10,
	}
	if app.Platform == "ios" {
		mailTemp = template.IOSCDReleaseTemplate_Mail
		weChatTemp = template.IOSCDReleaseTemplate_WeChat
	} else if app.Platform == "android" {
		mailTemp = template.CDReleaseTemplate_Mail
		weChatTemp = template.CDReleaseTemplate_WeChat
	}
	subject := fmt.Sprintf("【版本发布】%v %v 灰度发布", templateModel.AppName, templateModel.Version)
	mailContent, _ := s.fkDao.TemplateAlter(templateModel, mailTemp)
	wxContent, _ := s.fkDao.TemplateAlter(templateModel, weChatTemp)
	// 推送邮件至用户
	var (
		toUsers, ccUsers         []string
		toAddresses, ccAddresses []*mailmdl.Address
	)
	if toUsers, err = s.fkDao.AppMailtoList(c, appKey, mailmdl.CDReleaseNotifyMail, mailmdl.ReceiverWithTo); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if notifyGroup {
		if ccUsers, err = s.fkDao.AppMailtoList(c, appKey, mailmdl.CDReleaseNotifyMail, mailmdl.ReceiverWithCC); err != nil {
			log.Errorc(c, "%v", err)
			return
		}
		for _, ccUser := range ccUsers {
			ccAddrStr := ccUser + mailmdl.AddressSuffix
			ccAddress := &mailmdl.Address{Address: ccAddrStr}
			ccAddresses = append(ccAddresses, ccAddress)
		}
	}
	for _, toUser := range toUsers {
		toAddrStr := toUser + mailmdl.AddressSuffix
		toAddress := &mailmdl.Address{Address: toAddrStr}
		toAddresses = append(toAddresses, toAddress)
	}

	mail := &mailmdl.Mail{Subject: subject, Body: mailContent, Type: mailmdl.TypeTextHTML, ToAddresses: toAddresses, CcAddresses: ccAddresses}
	if err = s.fkDao.SendMail(c, mail, nil, appKey, mailmdl.CDReleaseNotifyMail); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	// 推送企业微信机器人
	botUrls := strings.Split(bots, ",")
	botMessage := wxContent
	for _, botUrl := range botUrls {
		_ = s.fkDao.RobotNotify(botUrl, &appmdl.Text{
			Content:       botMessage,
			MentionedList: []string{"@all"},
		})
	}
	return
}

// WindowsAppinstallerUpload
func (s *Service) WindowsAppinstallerUpload(c context.Context, appKey string, buildID int64) (err error) {
	var (
		pack *cimdl.BuildPack
	)
	if pack, err = s.fkDao.BuildPackByJobId(c, appKey, buildID); err != nil {
		log.Error("%v", err)
		return
	}
	if pack == nil {
		log.Error("WindowsUpload pack is not found; app_key:%v, build_id:%v", appKey, buildID)
		return
	}
	// 获取文件路径
	var (
		_PUBLIC_DIR_NAME       = "public"
		_APP_INSTALLER         = "app.appinstaller"
		_APP_INSTALLER_V2      = fmt.Sprintf("bilibili_%s.appinstaller", appKey)
		_APP_INSTALLER_RELEASE = "release.appinstaller"
	)
	dirPath := path.Dir(pack.PkgPath)
	publicDir := path.Join(dirPath, _PUBLIC_DIR_NAME)
	installerPath := path.Join(publicDir, _APP_INSTALLER)
	installerPathV2 := path.Join(publicDir, _APP_INSTALLER_V2)
	installerPathRelease := path.Join(publicDir, _APP_INSTALLER_RELEASE)
	ossRelationPath := strings.Replace(publicDir, s.c.LocalPath.LocalDir, fmt.Sprintf("%v/%v", s.c.Oss.Inland.CDNDomain, s.c.Oss.Inland.OriginDir), -1)

	// 若存在原始文件. 则进行文件替换； 若不存在原始文件，则忽略复制操作
	if utils.FileExists(installerPath) {
		// 重写发布配置项 - 路径替换
		if err = utils.ReplaceFileText(installerPath, installerPathV2, "http://localhost", ossRelationPath, -1); err != nil {
			return
		}
		// 重写发布配置 - url内部文件名
		if err = utils.ReplaceFileText(installerPathV2, installerPathV2, _APP_INSTALLER, _APP_INSTALLER_V2, -1); err != nil {
			return
		}
		// 生成待发布的配置文件
		// https://dl.hdslb.com/mobile/pack/win/6494770/public/app.appinstaller
		// -> https://dl.hdslb.com/mobile/appinstaller/win.appinstaller
		if err = utils.ReplaceFileText(
			installerPathV2,
			installerPathRelease,
			strings.Replace(installerPathV2, s.c.LocalPath.LocalDir, fmt.Sprintf("%v/%v", s.c.Oss.Inland.CDNDomain, s.c.Oss.Inland.OriginDir), -1),
			fmt.Sprintf("%v/%v/appinstaller/%s", s.c.Oss.Inland.CDNDomain, s.c.Oss.Inland.OriginDir, _APP_INSTALLER_V2),
			-1); err != nil {
			return
		}
	}
	// 遍历目录
	err = filepath.Walk(publicDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				relationPath := strings.Replace(path, s.c.LocalPath.LocalDir, "", -1)
				var uri string
				if uri, _, _, err = s.fkDao.FileUploadOss(context.Background(), path, relationPath, appKey); err != nil {
					log.Warn(fmt.Sprintf("filepath.Walk err %v", err))
					return err
				}
				log.Warn(fmt.Sprintf("filepath.Walk path=%v uri=%v", path, uri))
			}
			return nil
		})
	return
}

// WindowsAppinstallerPublish
func (s *Service) WindowsAppinstallerPublish(c context.Context, appKey string, buildID int64) (err error) {
	var (
		pack *cimdl.BuildPack
		url  string
	)
	if pack, err = s.fkDao.BuildPackByJobId(c, appKey, buildID); err != nil {
		log.Error("%v", err)
		return
	}
	if pack == nil {
		log.Error("WindowsUpload pack is not found; app_key:%v, build_id:%v", appKey, buildID)
		return
	}
	// 获取文件路径
	dirPath := path.Dir(pack.PkgPath)
	publicDir := path.Join(dirPath, "public")
	installerPathRelease := path.Join(publicDir, "release.appinstaller")
	// 发布至线上
	if url, _, _, err = s.fkDao.FileUploadOss(context.Background(), installerPathRelease, fmt.Sprintf("appinstaller/bilibili_%v.appinstaller", appKey), appKey); err != nil {
		log.Warn(fmt.Sprintf("err %v", err))
		return
	}
	// 刷新CDN
	url = strings.Replace(url, s.c.Oss.Inland.Bucket, s.c.Oss.Inland.CDNDomain, -1)
	_ = s.fkDao.AppCDRefreshCDN(strings.Split(url, ","))
	return
}

func (s *Service) AssetsEvolution(c context.Context, appKey string, buildID int64, file multipart.File, header *multipart.FileHeader, unzip bool, pkgName string) (err error) {
	testPack, err := s.fkDao.PackByBuild(c, appKey, "test", buildID)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("查询appkey:%s 构建号%d 出错", appKey, buildID))
		return
	}
	if testPack == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("请先将 appkey:%s 构建号%d 推送到测试环境", appKey, buildID))
		return
	}
	prodPack, err := s.fkDao.PackByBuild(c, appKey, "prod", buildID)
	if err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("查询appkey:%s 构建号%d 出错", appKey, buildID))
		return
	}
	if prodPack != nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("appkey:%s 构建号%d 不可重复推送生产环境", appKey, buildID))
		return
	}

	var (
		pkgPath, pkgURL, fmd5 string
		destFile, pkgFile     *os.File
		fileInfo              os.FileInfo
	)
	destFileDir := filepath.Join(conf.Conf.LocalPath.LocalDir, "pack", appKey, strconv.FormatInt(buildID, 10))
	pkgPath = filepath.Join(destFileDir, pkgName)
	// 文件存在则直接返回成功
	if _, err = os.Stat(pkgPath); err == nil {
		log.Errorc(c, "文件已存在 %v", pkgPath)
		return
	}
	if err = os.MkdirAll(destFileDir, 0755); err != nil {
		log.Errorc(c, "os.MkdirAll error(%v)", err)
		return
	}
	destFilePath := filepath.Join(destFileDir, header.Filename)

	if destFile, err = os.Create(destFilePath); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if _, err = io.Copy(destFile, file); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	defer file.Close()
	defer destFile.Close()
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if unzip {
		if err = utils.Unzip(destFilePath, destFileDir); err != nil {
			log.Errorc(c, "unzip(%s, %s) error(%v)", destFilePath, destFileDir, err)
			return
		}
		if err = os.Remove(destFilePath); err != nil {
			log.Errorc(c, "os.Remove(%s) error(%v)", destFilePath, err)
			return
		}
	}
	saveDir := conf.Conf.LocalPath.LocalDomain + "/pack/" + appKey + "/" + strconv.FormatInt(buildID, 10)
	pkgURL = saveDir + "/" + pkgName

	if fileInfo, err = os.Stat(pkgPath); err != nil {
		log.Errorc(c, "os.Stat(%s) error(%v)", pkgPath, err)
		return
	}
	buf := new(bytes.Buffer)
	if pkgFile, err = os.Open(pkgPath); err != nil {
		log.Errorc(c, "os.Open(%s) error(%v)", pkgPath, err)
		return
	}
	if _, err = io.Copy(buf, pkgFile); err != nil {
		log.Errorc(c, "error(%v)", err)
		return
	}
	md5Bs := md5.Sum(buf.Bytes())
	fmd5 = hex.EncodeToString(md5Bs[:])

	var url string
	if url, _, _, err = s.fkDao.FilePutOss(c, path.Join("pack", appKey, strconv.FormatInt(buildID, 10)), pkgName, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	var testVersion *model.Version
	if testVersion, err = s.fkDao.PackVersionByID(c, appKey, testPack.VersionID); err != nil {
		log.Error("%v", err)
		return
	}
	if testVersion == nil {
		log.Error("version is nil appkey: %v;build: %v", appKey, buildID)
		return
	}
	var packVersionId int64
	if err = s.fkDao.Transact(c, func(tx *sql.Tx) error {
		var version *model.Version
		if version, err = s.fkDao.PackVersion(c, appKey, "prod", testVersion.Version, testVersion.VersionCode); err != nil {
			log.Errorc(c, "PackVersion error %v", err)
			return err
		}
		if version == nil || version.ID == 0 {
			if packVersionId, err = s.fkDao.TxSetPackVersion(tx, testPack.AppID, testPack.AppKey, "prod", testPack.Version, testPack.VersionCode, cdmdl.NotUpgrade); err != nil {
				log.Errorc(c, "error: %v", err)
				return err
			}
		} else {
			packVersionId = version.ID
		}
		// 同步到prod环境
		if _, err = s.fkDao.TxSetPack(tx, testPack.AppID, testPack.AppKey, "prod", packVersionId, testPack.InternalVersionCode,
			testPack.BuildID, testPack.GitType, testPack.GitName, testPack.Commit, testPack.PackType, testPack.Operator, fileInfo.Size(), fmd5, pkgPath,
			pkgURL, testPack.MappingURL, testPack.RURL, testPack.RMappingURL, url, testPack.Desc, testPack.Sender, testPack.ChangeLog, testPack.DepGitJobId, testPack.IsCompatible, testPack.BbrUrl, testPack.Features); err != nil {
			log.Errorc(c, "TxSetPack error %v", err)
			return err
		}
		return err
	}); err != nil {
		log.Errorc(c, "error: %v", err)
		return
	}
	// 发送EP通知
	contents := CombineWxNotifyWithCD(&testPack.AppID, &testPack.AppKey, &testPack.Version, &testPack.Commit, &pkgURL, &url, &testPack.ChangeLog, &testVersion.VersionCode, &testPack.BuildID, &err)
	if err = s.fkDao.WechatEPNotify(contents, testPack.Sender); err != nil {
		log.Error("WechatEPNotify err %v", err)
	}
	return
}

func (s *Service) PackGreyList(c context.Context, appKey, version string, versionCode, glJobId, startTime, endTime int64, pn, ps int) (res *cdmdl.PackGreyDataResp, err error) {
	var (
		packGreyHistoryList []*cdmdl.PackGreyHistory
		packGreyDataList    []*cdmdl.PackGreyData
		filterCfg           map[int64]*cdmdl.FilterConfig
		count               int
		glJobIds            []int64
		app                 *appmdl.APP
	)
	if count, err = s.fkDao.PackGreyHistoryCount(c, appKey, version, versionCode, glJobId, startTime, endTime); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	page := &model.PageInfo{Total: count, Pn: pn, Ps: ps}
	res = &cdmdl.PackGreyDataResp{PageInfo: page}
	if count < 1 {
		return
	}
	if packGreyHistoryList, err = s.fkDao.PackGreyHistoryList(c, appKey, version, versionCode, glJobId, startTime, endTime, pn, ps); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	if len(packGreyHistoryList) < 1 {
		return
	}
	if app, err = s.fkDao.AppPass(c, appKey); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	for _, greyHistory := range packGreyHistoryList {
		glJobIds = append(glJobIds, greyHistory.GlJobID)
	}
	if filterCfg, err = s.fkDao.PackFilterConfig(c, appKey, "prod", glJobIds); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	for _, greyHistory := range packGreyHistoryList {
		packGreyData := &cdmdl.PackGreyData{
			Id:              greyHistory.ID,
			AppKey:          appKey,
			DatacenterAppId: app.DataCenterAppID,
			Platform:        app.Platform,
			MobiApp:         app.MobiApp,
			Version:         greyHistory.Version,
			VersionCode:     greyHistory.VersionCode,
			GlJobID:         greyHistory.GlJobID,
			IsUpgrade:       greyHistory.IsUpgrade,
			Flow:            greyHistory.Flow,
			Config:          filterCfg[greyHistory.GlJobID],
			GreyStartTime:   greyHistory.GreyStartTime,
			GreyCloseTime:   greyHistory.GreyCloseTime,
			GreyFinishTime:  greyHistory.GreyFinishTime,
			CTime:           greyHistory.CTime,
			MTime:           greyHistory.MTime,
		}
		packGreyDataList = append(packGreyDataList, packGreyData)
	}
	res.Items = packGreyDataList
	return
}

// AppCDCDNPublish 将cd产物上传到cdn上的固定地址
func (s *Service) AppCDCDNPublish(c context.Context, appKey string, buildId int64, filename string) (cdnUrl interface{}, err error) {
	var (
		build   *cdmdl.Pack
		appInfo *appmdl.APP
		uri     string
	)
	if build, err = s.fkDao.PackByBuild(c, appKey, string(cdmdl.EnvProd), buildId); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("build_id:%v, query pack error: %v.", buildId, err))
		log.Errorc(c, err.Error())
		return
	}
	if build == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("build_id:%v not exists", buildId))
		log.Warnc(c, err.Error())
		return
	}
	if build.SteadyState != 1 {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("build_id:%v, 非稳定版本，不可以推送到CDN.", buildId))
		log.Warnc(c, err.Error())
		return
	}
	if appInfo, err = s.fkDao.AppPass(c, appKey); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("get app info (appkey: %v) error: %v", appKey, err))
		log.Errorc(c, err.Error())
		return
	}
	if appInfo == nil {
		err = ecode.Error(ecode.RequestErr, fmt.Sprintf("app info not exists"))
		log.Warnc(c, err.Error())
		return
	}
	if uri, _, _, err = s.ossDao.FileUploadOss(c, build.PackPath, path.Join("fixed", appKey, filename), appInfo.ServerZone); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("read file(path:%v) error: %v", build.PackPath, err))
		log.Errorc(c, err.Error())
		return
	}
	if err = s.AppCDRefreshCDN(c, uri); err != nil {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("已经推送到CDN，但CDN刷新失败 error: %v", err))
		log.Errorc(c, err.Error())
		return
	}
	return struct {
		CDNUrl string `json:"cdn_url"`
	}{
		CDNUrl: uri,
	}, err
}

func groupByBuildID(packs []*cimdl.BuildPack) map[int64]*cimdl.BuildPack {
	m := make(map[int64]*cimdl.BuildPack, len(packs))
	for _, p := range packs {
		m[p.GitlabJobID] = p
	}
	return m
}

func CombineWxNotifyWithCD(appId, appKey, version, commit, packUrl, cdnUrl, changeLog *string, versionCode, buildId *int64, err *error) (contents string) {
	contents = fmt.Sprintf("[App ID]: %s\n"+
		"[App Key]: %v\n"+
		"[Version]: %v(%v)\n"+
		"[Commit]: %v\n"+
		"[构建号]: %v\n"+
		"[Archive]: %v\n%v\n"+
		"[Change Log]: %v", *appId, *appKey, *version, *versionCode, *commit, *buildId, *packUrl, *cdnUrl, *changeLog)
	if (*err) != nil {
		contents = fmt.Sprintf("【Fawkes】发布资源包失败\n"+
			"%v\n"+
			"[Error]: %v", contents, (*err).Error())
	} else {
		contents = fmt.Sprintf("【Fawkes】发布资源包成功\n"+
			"%v", contents)
	}
	return
}
