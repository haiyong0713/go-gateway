package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/appstatic/admin/model"

	"github.com/pkg/errors"
)

var (
	_uploadPrefix = "whitelist_"
)

// UploadBfs can upload a file object: store the info in Redis, and transfer the file to Bfs
func (s *Service) UploadBfs(c context.Context, fileName string, fileType string, timing int64, body []byte) (location string, err error) {
	if len(body) > model.BfsMaxSize {
		return "", fmt.Errorf("bfs最大允许上传20M文件")
	}
	if location, err = s.dao.Upload(c, fileName, fileType, timing, body); err != nil { // bfs
		log.Error("Upload bfs name(%s) error(%v)", fileName, err)
		return
	}
	log.Info("Upload bfs name(%s) success", fileName)
	return
}

// UploadBigFile .
func (s *Service) UploadBigFile(c context.Context, content []byte, fInfo *model.FileInfo) (path string, err error) {
	if _, err = s.dao.UploadBoss(c, fInfo.Name, content); err != nil {
		log.Error("UploadBigFile UploadBoss name(%s) error(%v)", fInfo.Name, err)
		return "", errors.Wrapf(err, "文件上传失败")
	}
	path = s.c.Host.Boss + "/" + model.BossBucket + "/" + fInfo.Name
	log.Info("UploadBigFile (%s) success", path)
	// 去掉预热
	// if env.DeployEnv == env.DeployEnvProd || env.DeployEnv == env.DeployEnvPre {
	// 	//只有正式环境才支持cdn预热
	// 	if err = s.dao.CdnDoPreload(c, []string{path}); err != nil {
	// 		log.Error("UploadBigFile CdnDoPreload url(%s) name(%s) error(%v)", path, fInfo.Name, err)
	// 		return "", errors.Wrapf(err, "文件上传成功但是CDN预热失败 url(%s)", path)
	// 	}
	// }
	return
}

// UploadGray .
func (s *Service) UploadGray(c context.Context, content []byte) (fInfo *model.FileInfo, err error) {
	var (
		location string
	)
	if _, err = xstr.SplitInts(string(content)); err != nil {
		return nil, fmt.Errorf("上传的mid文件不合法 %s", err.Error())
	}
	if len(content) == 0 {
		return nil, fmt.Errorf("文件内容不能为空")
	}
	t := time.Now().Unix()
	if fInfo, err = s.ParseFile(content); err != nil {
		return
	}
	fileName := _uploadPrefix + fInfo.Md5
	if location, err = s.dao.Upload(c, fileName, fInfo.Type, t, content); err != nil {
		log.Error("UploadGray bfs error(%v)", err)
		return
	}
	fInfo.URL = location
	return
}

// AddFile inserts file info into DB and updates its resource version+1
func (s *Service) AddFile(c context.Context, file *model.ResourceFile, version int) (err error) {
	if err = s.DB.Create(file).Error; err != nil {
		log.Error("resSrv.DB.Create error(%v)", err)
		return
	}
	// the resource containing the file updates its version
	if err = s.DB.Model(&model.Resource{}).Where("id = ?", file.ResourceID).Update("version", version+1).Error; err != nil {
		log.Error("resSrv.Update Version error(%v)", err)
		return
	}
	return nil
}

// ParseFile analyses file info
func (s *Service) ParseFile(content []byte) (file *model.FileInfo, err error) {
	fType := http.DetectContentType(content)
	// file md5
	md5hash := md5.New()
	if _, err = io.Copy(md5hash, bytes.NewReader(content)); err != nil {
		log.Error("resource uploadFile.Copy error(%v)", err)
		return
	}
	md5 := md5hash.Sum(nil)
	fMd5 := hex.EncodeToString(md5[:])
	file = &model.FileInfo{
		Md5:  fMd5,
		Type: fType,
		Size: int64(len(content)),
	}
	return
}

// TypeCheck checks whether the file type is allowed
func (s *Service) TypeCheck(fType string) (canAllow bool) {
	allowed := s.c.Cfg.Filetypes
	if len(allowed) == 0 {
		return true
	}
	for _, v := range allowed {
		if v == fType {
			return true
		}
	}
	return false
}
