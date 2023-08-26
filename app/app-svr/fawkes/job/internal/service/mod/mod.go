package mod

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"go-common/library/conf/env"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/railgun"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/fawkes/job/internal/model/mod"

	"github.com/pkg/errors"
)

const _tableModVersion = "mod_version"

func (s *Service) initModRailgun(cfg *railgun.DatabusV1Config, pcfg *railgun.SingleConfig) {
	inputer := railgun.NewDatabusV1Inputer(cfg)
	processor := railgun.NewSingleProcessor(pcfg, s.modRailgunUnpack, s.modRailgunDo)
	g := railgun.NewRailGun("稿件状态变更", nil, inputer, processor)
	s.modRailgun = g
	g.Start()
}

func (s *Service) modRailgunUnpack(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	var v *mod.BinlogMsg
	if err := json.Unmarshal(msg.Payload(), &v); err != nil {
		log.Error("日志告警 data:%s,error:%+v", msg.Payload(), err)
		return nil, err
	}
	if v.New == nil {
		return nil, nil
	}
	switch v.Table {
	case _tableModVersion:
		var data *mod.BinlogModVersion
		err := json.Unmarshal(v.New, &data)
		if err != nil {
			log.Error("日志告警 data:%s,error:%+v", v.New, err)
			return nil, err
		}
		version := &mod.Version{
			ID:        data.ID,
			ModuleID:  data.ModuleID,
			Env:       mod.Env(data.Env),
			Version:   data.Version,
			FromVerID: data.FromVerID,
			State:     mod.VersionState(data.State),
		}
		if version.State != mod.VersionProcessing {
			return nil, nil
		}
		return &railgun.SingleUnpackMsg{
			Group: version.ID,
			Item:  version,
		}, nil
	default:
		log.Error("未知的表名:%v", v.Table)
		return nil, nil
	}
}

func (s *Service) modRailgunDo(ctx context.Context, item interface{}) railgun.MsgPolicy {
	version := item.(*mod.Version)
	if err := s.patch(ctx, version); err != nil {
		log.Errorc(ctx, "patch error: %v", err)
		return railgun.MsgPolicyRetryInfinite
	}
	return railgun.MsgPolicyNormal
}

func (s *Service) patch(ctx context.Context, version *mod.Version) error {
	if version == nil {
		return nil
	}
	timeout, err := s.ac.Get("patchTimeout").Duration()
	if err != nil {
		log.Error("日志告警 %+v", err)
		timeout = time.Minute * 1
	}
	key := fmt.Sprintf("patch_lock_%d", version.ID)
	locked, err := s.dao.TryLock(ctx, key, int32(timeout*10/time.Second))
	if err != nil {
		return err
	}
	if !locked {
		log.Warn("lock failed,version:%+v", version)
		return nil
	}
	defer func() {
		if err1 := s.dao.UnLock(ctx, key); err1 != nil {
			log.Error("%+v", err1)
		}
	}()
	if version.Env == mod.EnvProd {
		testVersion, err := s.dao.VersionByID(ctx, version.FromVerID)
		if err != nil {
			return err
		}
		if testVersion.Env != mod.EnvTest {
			log.Error("日志告警 数据错误,查询的资源版本不是测试环境,version:%+v,test_version:%+v", version, testVersion)
			return nil
		}
		switch testVersion.State {
		case mod.VersionSucceeded:
			if err = s.dao.VersionSucceed(ctx, version.ID); err != nil {
				return err
			}
		default:
		}
		return nil
	}
	if version.Version == 1 {
		if err = s.dao.VersionSucceed(ctx, version.ID); err != nil {
			return err
		}
		return nil
	}
	defer func() {
		if err != nil {
			switch errors.Cause(err) {
			case context.DeadlineExceeded:
				log.Error("日志告警 增量包构建超时,version:%+v,err:%+v", version, err)
			case context.Canceled:
				log.Error("日志告警 增量包构建取消,version:%+v,err:%+v", version, err)
			case xsql.ErrNoRows:
				log.Error("日志告警 增量包构建数据有误,version:%+v,err:%+v", version, err)
				err = nil
			default:
				log.Error("日志告警 增量包构建错误,version:%+v,err:%+v", version, err)
			}
			return
		}
		if err = s.dao.VersionSucceed(ctx, version.ID); err != nil {
			return
		}
	}()
	limit, err := s.ac.Get("patchLimit").Int64()
	if err != nil {
		log.Error("日志告警 %+v", err)
		limit = 1
	}
	file, err := s.dao.OriginalFile(ctx, version.ID)
	if err != nil {
		return err
	}
	lastVersionsTest, err := s.dao.LastVersionList(ctx, version.ModuleID, version.Version, limit, mod.EnvTest)
	if err != nil {
		return err
	}
	lastVersionsProd, err := s.dao.LastVersionList(ctx, version.ModuleID, version.Version, limit, mod.EnvProd)
	if err != nil {
		return err
	}
	targetVersions := union(lastVersionsTest, lastVersionsProd)
	lastVersions, err := s.dao.VersionList(ctx, version.ModuleID, targetVersions, mod.EnvTest)
	if err != nil {
		return err
	}
	if len(lastVersions) == 0 {
		log.Infoc(ctx, "moduleId:%d, version:%d lastVersion is empty.", version.ModuleID, version.Version)
		return nil
	}
	var versionIDs []int64
	fromVerm := map[int64]int64{}
	for _, lastVersion := range lastVersions {
		versionIDs = append(versionIDs, lastVersion.ID)
		fromVerm[lastVersion.ID] = lastVersion.Version
	}
	lastFiles, err := s.dao.OriginalFileList(ctx, versionIDs)
	if err != nil {
		return err
	}
	patchFiles, err := s.patchFileList(ctx, file, lastFiles, fromVerm)
	if err != nil {
		return err
	}
	if err = s.dao.PatchAdd(ctx, version, patchFiles); err != nil {
		return err
	}
	return nil
}

func (s *Service) patchFileList(ctx context.Context, file *mod.File, lastFiles []*mod.File, fromVerm map[int64]int64) ([]*mod.File, error) {
	newFilePath, err := s.downloadFile(ctx, file)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err1 := deleteFile(newFilePath); err1 != nil {
			log.Error(" %+v", err1)
		}
	}()
	var (
		res  []*mod.File
		lock sync.Mutex
	)
	g := errgroup.WithContext(ctx)
	for _, val := range lastFiles {
		lastFile := val
		g.Go(func(ctx context.Context) error {
			fromVer, ok := fromVerm[lastFile.VersionID]
			if !ok {
				return nil
			}
			oldFilePath, err := s.downloadFile(ctx, lastFile)
			if err != nil {
				return err
			}
			defer func() {
				if err1 := deleteFile(oldFilePath); err1 != nil {
					log.Error("%+v", err1)
				}
			}()
			patchFilename := fmt.Sprintf("%d_%d_%s.bspatch", file.ID, lastFile.ID, file.Name)
			patchFilePath := fmt.Sprintf("%s/%s", s.folder, patchFilename)
			patchFileData, err := s.genPatchFile(ctx, newFilePath, oldFilePath, patchFilePath)
			if err != nil {
				return err
			}
			defer func() {
				if err1 := deleteFile(patchFilePath); err1 != nil {
					log.Error("%+v", err1)
				}
			}()
			contentType, md5, size, err := parseFileData(patchFileData)
			if err != nil {
				return err
			}
			url, err := s.fileUpload(ctx, patchFilename, md5, patchFileData)
			if err != nil {
				return err
			}
			patch := &mod.File{
				Name:        patchFilename,
				Size:        size,
				Md5:         md5,
				URL:         url,
				ContentType: contentType,
				IsPatch:     true,
				FromVer:     fromVer,
			}
			lock.Lock()
			res = append(res, patch)
			lock.Unlock()
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Service) downloadFile(ctx context.Context, file *mod.File) (string, error) {
	filename := fmt.Sprintf("%d_%s", file.ID, file.Name)
	filePath := fmt.Sprintf("%s/%s", s.folder, strings.Replace(filename, "/", "*", -1))
	url, ok := s.url(file.URL)
	if !ok {
		log.Error("日志告警 file url 未识别前缀,file:%+v", file)
		return "", errors.New(fmt.Sprintf("file url wrong,file:%+v", file))
	}
	if err := s.dao.DownloadFile(ctx, url, filePath); err != nil {
		// 特殊逻辑，电商在uat环境使用的是线上的下载地址
		if env.DeployEnv != env.DeployEnvUat {
			return "", err
		}
		if !strings.Contains(url, "uat-") {
			return "", err
		}
		url := strings.Replace(url, "uat-", "", 1)
		if err := s.dao.DownloadFile(ctx, url, filePath); err != nil {
			return "", err
		}
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", errors.Wrap(err, filePath)
	}
	return filePath, nil
}

func (s *Service) genPatchFile(ctx context.Context, newFilePath, oldFilePath, patchFilePath string) ([]byte, error) {
	timeout, err := s.ac.Get("patchTimeout").Duration()
	if err != nil {
		log.Error("日志告警 %+v", err)
		timeout = time.Minute * 1
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	// nolint:biligowordcheck
	go func() {
		select {
		case <-s.closeChan:
			cancel()
		case <-ctx.Done():
		}
	}()
	cmd := exec.CommandContext(ctx, "bsdiff", oldFilePath, newFilePath, patchFilePath)
	if err = cmd.Start(); err != nil {
		return nil, errors.Wrapf(err, "cmd:%v %v %v %v", "bsdiff", oldFilePath, newFilePath, patchFilePath)
	}
	if err = cmd.Wait(); err != nil {
		if err1, ok := err.(*exec.ExitError); ok {
			// If the process exited by itself, just return the error to the caller.
			if err1.Exited() {
				return nil, errors.Wrapf(err1, "cmd:%v %v %v %v", "bsdiff", oldFilePath, newFilePath, patchFilePath)
			}
			// We know now that the process could be started, but didn't exit
			// by itself. Something must have killed it. If the context is done,
			// we can *assume* that it has been killed by the exec.Command.
			// Let's return ctx.Err() so our user knows that this *might* be
			// the case.
			select {
			case <-ctx.Done():
				return nil, errors.Wrapf(ctx.Err(), "cmd:%v %v %v %v", "bsdiff", oldFilePath, newFilePath, patchFilePath)
			default:
				return nil, errors.Wrapf(err1, "cmd:%v %v %v %v", "bsdiff", oldFilePath, newFilePath, patchFilePath)
			}
		}
		return nil, errors.Wrapf(err, "cmd:%v %v %v %v", "bsdiff", oldFilePath, newFilePath, patchFilePath)
	}
	if _, err = os.Stat(patchFilePath); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "patch_file:%v", patchFilePath)
	}
	return ioutil.ReadFile(patchFilePath)
}

func parseFileData(data []byte) (contentType, md5Value string, size int64, err error) {
	h := md5.New()
	if _, err := io.Copy(h, bytes.NewReader(data)); err != nil {
		return "", "", 0, err
	}
	return http.DetectContentType(data), hex.EncodeToString(h.Sum(nil)), int64(len(data)), nil
}

func (s *Service) fileUpload(ctx context.Context, filename, md5 string, fileData []byte) (string, error) {
	const _bossBucket = "appstaticboss"
	if len(fileData) == 0 {
		return "", errors.New(fmt.Sprintf("文件内容不能为空,filename:%s", filename))
	}
	path := fmt.Sprintf("%s/%s", md5, filename)
	if _, err := s.boss.PutObject(ctx, _bossBucket, path, fileData); err != nil {
		return "", err
	}
	return fmt.Sprintf("/%s/%s", _bossBucket, path), nil
}

func deleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return errors.Wrapf(err, "file_path:%v", filePath)
	}
	return nil
}

func (s *Service) url(path string) (string, bool) {
	var modCDN map[string]string
	if path == "" {
		return "", false
	}
	if err := s.ac.Get("modCDN").UnmarshalTOML(&modCDN); err != nil {
		log.Error("日志告警 error:%+v", err)
		return "", false
	}
	for prefix, host := range modCDN {
		if strings.HasPrefix(path, prefix) {
			return host + path, true
		}
	}
	return "", false
}

// 两个版本的并集
func union(test []*mod.Version, prod []*mod.Version) []int64 {
	testMap := make(map[int64]*mod.Version)
	prodMap := make(map[int64]*mod.Version)
	var union []*mod.Version
	for _, v := range test {
		testMap[v.Version] = v
		union = append(union, v)
	}
	for _, v := range prod {
		prodMap[v.Version] = v
	}
	for k, v := range prodMap {
		if _, ok := testMap[k]; !ok {
			union = append(union, v)
		}
	}
	var targetVersion []int64
	for _, v := range union {
		targetVersion = append(targetVersion, v.Version)
	}
	return targetVersion
}
