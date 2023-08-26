package bizapk

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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

	"go-gateway/app/app-svr/fawkes/service/conf"
	bizapkmdl "go-gateway/app/app-svr/fawkes/service/model/bizapk"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"
)

// BizApkBuildsList list business apk builds
func (s *Service) BizApkBuildsList(c context.Context, appKey, env string, packBuildID int64) (res []*bizapkmdl.BuildListResp, err error) {
	var (
		builds       []*bizapkmdl.Build
		bids         []int64
		filterConfig map[int64]*bizapkmdl.FilterConfig
		gitlabPrjID  string
	)
	if _, _, gitlabPrjID, err = s.fkDao.AppBasicInfo(c, appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if builds, err = s.fkDao.BizApkBuilds(c, appKey, env, packBuildID); err != nil {
		log.Error("%v", err)
		return
	}
	groupMap := make(map[string][]*bizapkmdl.Build)
	for _, build := range builds {
		if build.GitlabJobID != 0 {
			build.GitlabJobURL = conf.Conf.Gitlab.Host + "/" + gitlabPrjID + "/-/jobs/" + strconv.FormatInt(build.GitlabJobID, 10)
		}
		if build.GitlabPipelineID != 0 {
			build.GitlabPplURL = conf.Conf.Gitlab.Host + "/" + gitlabPrjID + "/pipelines/" + strconv.FormatInt(build.GitlabPipelineID, 10)
		}
		groupMap[build.Name] = append(groupMap[build.Name], build)
		bids = append(bids, build.ID)
	}
	if len(bids) > 0 {
		if filterConfig, err = s.fkDao.BizApkFilterConfig(context.Background(), env, bids); err != nil {
			if err == sql.ErrNoRows {
				err = nil
			} else {
				log.Error("%v", err)
				return
			}
		}
	}
	for _, oneGroup := range groupMap {
		group := &bizapkmdl.BuildListResp{
			Name:        oneGroup[0].Name,
			Cname:       oneGroup[0].Cname,
			Description: oneGroup[0].Description,
			BizApkID:    oneGroup[0].BizApkID,
			SettingsID:  oneGroup[0].SettingsID,
			Active:      oneGroup[0].Active,
			Priority:    oneGroup[0].Priority,
			Builds:      oneGroup,
		}
		var flowConfig map[int64]*bizapkmdl.FlowConfig
		if flowConfig, err = s.fkDao.BizApkFlowConfig(context.Background(), env, packBuildID, group.BizApkID); err != nil {
			if err == sql.ErrNoRows {
				err = nil
			} else {
				log.Error("%v", err)
				return
			}
		}
		for _, build := range oneGroup {
			if _, ok := filterConfig[build.ID]; ok {
				build.Config = filterConfig[build.ID]
			}
			if _, ok := flowConfig[build.ID]; ok {
				build.Flow = flowConfig[build.ID].Flow
			}
		}
		res = append(res, group)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].BizApkID < res[j].BizApkID
	})
	return
}

// BizApkList list business apks for a package
func (s *Service) BizApkList(c context.Context, appKey string, packBuildID int64, env string) (r []*bizapkmdl.ApkPackSettings, err error) {
	if r, err = s.fkDao.BizApks(c, appKey, packBuildID, env); err != nil {
		log.Error("%v", err)
		return
	}
	return
}

// SetBizApkActive set business apk active
func (s *Service) SetBizApkActive(c context.Context, active int, operator string, settingsID int64) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if err = s.fkDao.TxSetBizapkActive(tx, active, operator, settingsID); err != nil {
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// SetBizapkPriority update business apk priority
func (s *Service) SetBizapkPriority(c context.Context, priority int, operator string, settingsID int64) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if err = s.fkDao.TxSetBizapkPriority(tx, priority, operator, settingsID); err != nil {
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// BizApkBuildCreate create a business apk build
func (s *Service) BizApkBuildCreate(c context.Context, appKey, name string, packBuildID int64, gitType int, gitName, userName string) (bizapkBuildID int64, err error) {
	var (
		tx       *sql.Tx
		bizApkID int64
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if bizApkID, err = s.fkDao.BizApkID(c, appKey, name); err != nil {
		log.Error("BizApkBuildCreate error(%v)", err)
		return
	}
	if bizApkID == 0 {
		if bizApkID, err = s.fkDao.TxAddBizApk(tx, appKey, name, "", ""); err != nil {
			log.Error("BizApkBuildCreate error(%v)", err)
			return
		}
	}
	if bizapkBuildID, err = s.fkDao.TxCreateBizApkBuild(tx, bizApkID, packBuildID, gitType, gitName, userName); err != nil {
		log.Error("BizApkBuildCreate error(%v)", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// BizApkBuildUpdatePpl update pipeline info for business apk build
func (s *Service) BizApkBuildUpdatePpl(c context.Context, pplID int, commit string, packBuildID int64) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if err = s.fkDao.TxUpdateBizApkBuildPpl(tx, pplID, commit, packBuildID); err != nil {
		log.Error("BizApkBuildUpdatePpl error(%v)", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// UploadBizApkBuildFromCI upload a business apk build from CI
func (s *Service) UploadBizApkBuildFromCI(c context.Context, file multipart.File, header *multipart.FileHeader, apk, mapping, meta, appKey, name string, priority, builtIn int, packBuildID int64) (res *bizapkmdl.UploadResp, err error) {
	var (
		tx                                       *sql.Tx
		buildPacks                               []*cimdl.BuildPack
		bizApkID, bizapkBuildID, bizApkSettingID int64
	)
	// 获取 CI 原始包的基础信息来填充 bizapk 第一版的信息
	if buildPacks, err = s.fkDao.BuildPacks(c, appKey, 1, 20, 0, 0, 0, "", "", "", "", packBuildID, 0, "", false); err != nil {
		log.Error("UploadBizApkBuildFromCI error(%v)", err)
		return
	}
	if len(buildPacks) == 0 {
		err = errors.New("can not find original pack from ci")
		return
	}
	if bizApkID, err = s.fkDao.BizApkID(c, appKey, name); err != nil {
		log.Error("UploadBizApkBuildFromCI error(%v)", err)
		return
	}
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if bizApkID == 0 {
		if bizApkID, err = s.fkDao.TxAddBizApk(tx, appKey, name, "", ""); err != nil {
			log.Error("UploadBizApkBuildFromCI error(%v)", err)
			return
		}
	}
	if bizApkSettingID, err = s.fkDao.BizApkSettingID(c, packBuildID, bizApkID, "test"); err != nil {
		log.Error("UploadBizApkBuildFromCI get BizApkSettingID(%v)", err)
		return
	}
	if bizApkSettingID == 0 {
		if _, err = s.fkDao.TxAddBizapkSettings(tx, packBuildID, bizApkID, "test", priority, "fawkes"); err != nil {
			log.Error("UploadBizApkBuildFromCI error(%v)", err)
			return
		}
	}
	if bizapkBuildID, err = s.fkDao.BizApkBuildID(c, packBuildID, bizApkID, "test"); err != nil {
		log.Error("UploadBizApkBuildFromCI get BizApkBuildID(%v)", err)
		return
	}
	if bizapkBuildID == 0 {
		if bizapkBuildID, err = s.fkDao.TxCreateBizApkBuild(tx, bizApkID, packBuildID, int(buildPacks[0].GitType), buildPacks[0].GitName, "fawkes"); err != nil {
			log.Error("UploadBizApkBuildFromCI error(%v)", err)
			return
		}
	}
	if err = s.fkDao.TxUpdateBizApkBuildPpl(tx, 0, buildPacks[0].Commit, bizapkBuildID); err != nil {
		log.Error("UploadBizApkBuildFromCI error(%v)", err)
		return
	}
	if err = s.fkDao.TxStartBizApkBuild(tx, packBuildID, bizapkBuildID); err != nil {
		log.Error("UploadBizApkBuildFromCI error(%v)", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
		return
	}
	if res, err = s.handleBizApkZipFile(context.Background(), file, header, apk, mapping, meta, appKey, name, bizapkBuildID, packBuildID, packBuildID, builtIn); err != nil {
		log.Error("UploadBizApkBuildFromCI error(%v)", err)
		return
	}
	return
}

// UploadBizApkBuildFromCD upload a business apk build from CD
func (s *Service) UploadBizApkBuildFromCD(c context.Context, file multipart.File, header *multipart.FileHeader, apk, mapping, meta string, bizapkBuildID int64, builtIn, needUploadToCDN int) (res *bizapkmdl.UploadResp, err error) {
	var (
		tx        *sql.Tx
		build     *bizapkmdl.Build
		apkCdnURL string
	)
	if build, err = s.fkDao.BizApkBuildWithID(c, bizapkBuildID); err != nil {
		log.Error("UploadBizApkBuildFromCD error(%v)", err)
		return
	}
	if res, err = s.handleBizApkZipFile(c, file, header, apk, mapping, meta, build.AppKey, build.Name, bizapkBuildID, build.PackBuildID, build.BundleVer, builtIn); err != nil {
		log.Error("UploadBizApkBuildFromCD error(%v)", err)
		return
	}
	apkFilename := path.Base(res.ApkPath)
	relativeFilePath := strings.Replace(res.ApkPath, s.c.LocalPath.LocalDir, "", -1)
	apkFolder := strings.Replace(relativeFilePath, apkFilename, "", -1)
	// 标记是否要上传CDN, 目前从ci进入组件列表构建的，不需要上传cdn
	if needUploadToCDN == bizapkmdl.NEED_UPLOAD {
		if apkCdnURL, _, _, err = s.fkDao.FilePutOss(c, apkFolder, apkFilename, build.AppKey); err != nil {
			log.Error("Put oss error: %v (folder:%v name:%v)", err, apkFolder, apkFilename)
			return
		}
	}
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if err = s.fkDao.TxUpdateBizApkBuildCDN(tx, apkCdnURL, build.ID); err != nil {
		log.Error("%v", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
		return
	}
	return
}

func (s *Service) handleBizApkZipFile(c context.Context, file multipart.File, header *multipart.FileHeader, apk, mapping, meta, appKey, name string, bizapkBuildID, packBuildID, bundleVer int64, builtIn int) (res *bizapkmdl.UploadResp, err error) {
	var (
		tx                                         *sql.Tx
		size                                       int64
		destFile, apkFile                          *os.File
		fileInfo                                   os.FileInfo
		mappingPath, metaPath, mappingURL, metaURL string
	)
	// 文件复制和解压
	destFileDir := filepath.Join(conf.Conf.LocalPath.LocalDir, "pack", appKey, strconv.FormatInt(packBuildID, 10), "bizapk", name, strconv.FormatInt(bundleVer, 10))
	if err = os.MkdirAll(destFileDir, 0755); err != nil {
		log.Error("os.MkdirAll error(%v)", err)
		return
	}
	apkPath := filepath.Join(destFileDir, apk)
	// 文件存在则直接返回成功
	if _, err = os.Stat(apkPath); err == nil {
		return
	}
	if mapping != "" {
		mappingPath = filepath.Join(destFileDir, mapping)
	}
	if meta != "" {
		metaPath = filepath.Join(destFileDir, meta)
	}
	destFilePath := filepath.Join(destFileDir, header.Filename)
	// 由于偶现的超时问题，当接口重试发现文件已下载，则任务返回成功
	if _, err = os.Stat(destFilePath); err == nil {
		return
	}
	if destFile, err = os.Create(destFilePath); err != nil {
		log.Error("%v", err)
		return
	}
	if _, err = io.Copy(destFile, file); err != nil {
		log.Error("%v", err)
		return
	}
	defer file.Close()
	defer destFile.Close()
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		log.Error("%v", err)
		return
	}
	if err = utils.Unzip(destFilePath, destFileDir); err != nil {
		log.Error("unzip(%s, %s) error(%v)", destFilePath, destFileDir, err)
		return
	}
	if err = os.Remove(destFilePath); err != nil {
		log.Error("os.Remove(%s) error(%v)", destFilePath, err)
		return
	}
	urlPrefix := conf.Conf.LocalPath.LocalDomain + "/pack/" + appKey + "/" + strconv.FormatInt(packBuildID, 10) + "/bizapk/" + name + "/" + strconv.FormatInt(bundleVer, 10)
	apkURL := urlPrefix + "/" + apk
	if mapping != "" {
		mappingURL = urlPrefix + "/" + mapping
	}
	if meta != "" {
		metaURL = urlPrefix + "/" + meta
	}
	if fileInfo, err = os.Stat(apkPath); err != nil {
		log.Error("os.Stat(%s) error(%v)", apkPath, err)
		return
	}
	size = fileInfo.Size()
	buf := new(bytes.Buffer)
	if apkFile, err = os.Open(apkPath); err != nil {
		log.Error("os.Open(%s) error(%v)", apkPath, err)
		return
	}
	if _, err = io.Copy(buf, apkFile); err != nil {
		log.Error("error(%v)", err)
		return
	}
	md5Bs := md5.Sum(buf.Bytes())
	fmd5 := hex.EncodeToString(md5Bs[:])
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if err = s.fkDao.TxUploadBizApk(tx, fmd5, size, apkPath, mappingPath, metaPath, apkURL, mappingURL, metaURL, bizapkBuildID, builtIn); err != nil {
		log.Error("UploadBizApkBuildFromCI error(%v)", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	res = &bizapkmdl.UploadResp{
		ApkURL:   apkURL,
		MapURL:   mappingURL,
		MetaURL:  metaURL,
		ApkPath:  apkPath,
		MapPath:  mappingPath,
		MetaPath: metaPath,
	}
	return
}

// UpdateBizApkBuildInfo update business apk build info
func (s *Service) UpdateBizApkBuildInfo(c context.Context, gitlabJobID, bizapkBuildID int64) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if err = s.fkDao.TxStartBizApkBuild(tx, gitlabJobID, bizapkBuildID); err != nil {
		log.Error("UploadBizApkBuildFromCI error(%v)", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// BizApkBuildEvolution push business apk build to the production enviroment
func (s *Service) BizApkBuildEvolution(c context.Context, bizApkBuildID int64, operator string) (err error) {
	var (
		tx  *sql.Tx
		b   *bizapkmdl.Build
		row int
	)
	if b, err = s.fkDao.BizApkBuildWithID(c, bizApkBuildID); err != nil {
		log.Error("BizApkBuildEvolution error(%v)", err)
		return
	}
	// 唯一性检查，检查 prod 环境是否已经有该版本的包存在
	if row, err = s.fkDao.BundleVerHasProd(c, b.BundleVer, b.PackBuildID, b.BizApkID); err != nil {
		log.Error("BizApkBuildEvolution error(%v)", err)
		return
	}
	if row > 0 {
		err = errors.New("there is a bundle build already in prod env")
		log.Error("BizApkBuildEvolution error(%v)", err)
		return
	}
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	// 复制一份新的构建数据到 prod 环境
	if _, err = s.fkDao.TxAddBizApkBuild(tx, b.BizApkID, "prod", b.PackBuildID, b.BundleVer, b.MD5, b.Size, b.GitlabPipelineID, b.GitlabJobID, b.GitType, b.GitName, b.Commit,
		b.ApkPath, b.MapPath, b.MetaPath, b.ApkURL, b.MapURL, b.MetaURL, b.ApkCdnURL, b.Status, operator); err != nil {
		log.Error("BizApkBuildEvolution error(%v)", err)
		return
	}
	// 检查是否 prod 环境已经有 settings
	if _, err = s.fkDao.BizApk(context.Background(), b.AppKey, b.PackBuildID, "prod", b.BizApkID); err != nil {
		if err == sql.ErrNoRows {
			// 复制一份新的 settings 数据到 prod 环境
			if _, err = s.fkDao.TxAddBizapkSettings(tx, b.PackBuildID, b.BizApkID, "prod", int(b.Priority), operator); err != nil {
				log.Error("BizApkBuildEvolution error(%v)", err)
				return
			}
		} else {
			log.Error("BizApkBuildEvolution error(%v)", err)
			return
		}
	}
	// test 环境的原数据更新字段
	if err = s.fkDao.TxDidPushProd(tx, bizApkBuildID); err != nil {
		log.Error("BizApkBuildEvolution error(%v)", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// BizApkBuildCancel cancel a businesss apk build
func (s *Service) BizApkBuildCancel(c context.Context, bizApkBuildID int64) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if err = s.fkDao.TxUpdateBizApkBuildStatus(tx, -2, bizApkBuildID); err != nil {
		log.Error("BizApkBuildCancel error(%v)", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// BizApkBuildDelete delete a businesss apk build
func (s *Service) BizApkBuildDelete(c context.Context, bizApkBuildID int64) (err error) {
	var (
		tx *sql.Tx
	)
	if tx, err = s.fkDao.BeginTran(context.Background()); err != nil {
		log.Error("s.dao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
	}()
	if err = s.fkDao.TxDelBizApkBuild(tx, bizApkBuildID); err != nil {
		log.Error("BizApkBuildDelete error(%v)", err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
	}
	return
}

// OrgPackBuildByJobID get original package from pack build ID
func (s *Service) OrgPackBuildByJobID(c context.Context, appKey string, buildID int64) (res *bizapkmdl.OrgPackURLResp, err error) {
	if res, err = s.fkDao.OrgPackURL(c, appKey, buildID); err != nil {
		log.Error("%v", err)
		return
	}
	return
}

// FilterConfigSet set filter config.
func (s *Service) FilterConfigSet(c context.Context, appKey, env string, buildID int64, network, isp, channel,
	city, excludesSystem string, percent int, device, userName string, status int) (err error) {
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
	if _, err = s.fkDao.TxSetBizApkFilterConfig(tx, env, buildID, network, isp, channel, city, excludesSystem, percent, salt, device, status); err != nil {
		log.Error("%v", err)
		return
	}
	// add log
	_, _ = s.fkDao.AddLog(c, appKey, env, mngmdl.ModelBizApk, mngmdl.OperationBizApkFilterConfig, fmt.Sprintf("组件构建ID: %v", buildID), userName)
	return
}

// FilterConfig get app filter.
func (s *Service) FilterConfig(c context.Context, env string, buildID int64) (fconfig *bizapkmdl.FilterConfig, err error) {
	var fm map[int64]*bizapkmdl.FilterConfig
	if fm, err = s.fkDao.BizApkFilterConfig(c, env, []int64{buildID}); err != nil {
		log.Error("%v", err)
		return
	}
	fconfig = fm[buildID]
	return
}

// FlowConfigSet set flow set.
func (s *Service) FlowConfigSet(c context.Context, appKey, env, userName string, packBuildID, apkID int64, flow map[int64]string) (err error) {
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
	var buildids []string
	for b, fs := range flow {
		buildids = append(buildids, strconv.FormatInt(b, 10))
		if _, err = s.fkDao.TxSetBizApkFlowConfig(tx, env, fs, packBuildID, apkID, b); err != nil {
			log.Error("%v", err)
			return
		}
	}
	// add log
	_, _ = s.fkDao.AddLog(c, appKey, env, mngmdl.ModelBizApk, mngmdl.OperationBizApkFlowConfig, fmt.Sprintf("组件ID: %v，组件构建ID：%v", apkID, strings.Join(buildids, ",")), userName)
	return
}

// FlowConfig get flow config.
func (s *Service) FlowConfig(c context.Context, env string, packBuildID, apkID int64) (map[int64]string, error) {
	flow, err := s.fkDao.BizApkFlowConfig(c, env, packBuildID, apkID)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	if len(flow) == 0 {
		return nil, nil
	}
	flows := make(map[int64]string, len(flow))
	for _, fliter := range flow {
		flows[fliter.BuildID] = fliter.Flow
	}
	return flows, nil
}
