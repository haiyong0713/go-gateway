package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"go-common/library/log"
	"go-gateway/app/app-svr/macross/service/conf"
	pkgInfo "go-gateway/app/app-svr/macross/service/model/package"
	"go-gateway/app/app-svr/macross/service/tools"
)

// PackageUpload upload package zip file & unzip it
func (s *Service) PackageUpload(file multipart.File, pkgInfo pkgInfo.PkgInfo) (err error) {
	err = os.MkdirAll(pkgInfo.SaveDir, 0755)
	if err != nil {
		log.Error("os.MkdirAll error(%v)", err)
		return
	}

	destFilePath := filepath.Join(pkgInfo.SaveDir, pkgInfo.FileName)
	destFile, err := os.Create(destFilePath)
	if err != nil {
		log.Error("os.Create(%s) error(%v)", destFilePath, err)
		return
	}
	io.Copy(destFile, file)

	err = tools.Unzip(destFilePath, pkgInfo.SaveDir)
	if err != nil {
		log.Error("unzip(%s, %s) error(%v)", destFilePath, pkgInfo.SaveDir, err)
		return
	}

	err = os.Remove(destFilePath)
	if err != nil {
		log.Error("os.Remove(%s) error(%v)", destFilePath, err)
		return
	}

	// generate Android channel package
	if pkgInfo.ClientType == "android" {
		if len(pkgInfo.Channel) > 0 && len(pkgInfo.ApkName) > 0 {
			dest := filepath.Join(pkgInfo.SaveDir, "channel")
			targetFilePath := filepath.Join(pkgInfo.SaveDir, pkgInfo.ApkName)
			_, err = tools.GenerateChannelApk(dest, pkgInfo.Channel, nil, targetFilePath, false, false)
			if err != nil {
				log.Error("tools.GenerateChannelApk(%s, %s, nil, %s, false, false) error(%v)", dest, pkgInfo.Channel, targetFilePath, err)
				return
			}
		}
	}

	return
}

// PackageList list the Package Files
func (s *Service) PackageList(path string) (fileList []string, err error) {
	err = filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Error("filepath.Walk error(%v)", err)
			return err
		}
		if f == nil {
			errMsg := "found no file"
			err = fmt.Errorf(errMsg)
			log.Error(errMsg)
			return err
		}
		if f.IsDir() {
			return nil
		}
		// ignore tmp files
		if strings.HasPrefix(f.Name(), ".") {
			return nil
		}
		fileURL := strings.Replace(path, conf.Conf.Property.Package.SavePath, conf.Conf.Property.Package.URLPrefix, -1)
		fileList = append(fileList, fileURL)
		return err
	})
	return
}
