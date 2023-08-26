package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"go-common/library/log"
)

// Unzip unzip pack.
func (s *Service) Unzip(c context.Context, f io.ReaderAt, size int64, folder string) (zr *zip.Reader, saves []string, err error) {
	if zr, err = zip.NewReader(f, size); err != nil {
		log.Error("zip.NewReader error(%v)", err)
		return
	}
	filePath := path.Join(s.c.APK.LocalDir, folder)
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(filePath, 0755); err != nil {
			log.Error("os.MkDirAll(%s) error(%v)", filePath, err)
			return
		}
		err = nil // NOTE: folder ok~
	} else if !fileInfo.IsDir() {
		err = fmt.Errorf("%s is not folder", filePath)
		return
	}
	var (
		rd   io.ReadCloser
		sf   *os.File
		save string
	)
	for _, zf := range zr.File {
		if strings.Index(zf.Name, "MACOSX") > -1 {
			continue
		}
		//if !strings.HasSuffix(zf.Name, ".apk") && !strings.HasSuffix(zf.Name, ".txt") {
		//	continue
		//}
		if rd, err = zf.Open(); err != nil {
			log.Error("zf.Open error(%v)", err)
			return
		}
		save = path.Join(filePath, zf.Name)
		if sf, err = os.OpenFile(save, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, zf.Mode()); err != nil {
			log.Error("os.OpenFile(%s) error(%v)", save, err)
			rd.Close()
			return
		}
		_, err = io.Copy(sf, rd)
		sf.Close()
		rd.Close()
		if err != nil {
			log.Error("io.Copy error(%v)", err)
			return
		}
		save = strings.Replace(save, s.c.APK.LocalDir, s.c.APK.LocalDomain, -1)
		saves = append(saves, save)
	}
	return
}

// ApkUpload apk local.
func (s *Service) ApkUpload(c context.Context, rd io.Reader, folder, filename string) (save string, err error) {
	var (
		filePath = path.Join(s.c.APK.LocalDir, folder)
		f        *os.File
	)
	_, err = os.Stat(filePath)
	if os.IsExist(err) {
		if err = os.RemoveAll(filePath); err != nil {
			log.Error("os.RemoveAll(%s) error(%v)", filePath, err)
			return
		}
	} else {
		if err = os.MkdirAll(filePath, 0755); err != nil {
			log.Error("os.MkDirAll(%s) error(%v)", filePath, err)
			return
		}
	}
	save = path.Join(filePath, filename)
	if f, err = os.OpenFile(save, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
		log.Error("os.OpenFile(%s) error(%v)", save, err)
		return
	}
	defer f.Close()
	if _, err = io.Copy(f, rd); err != nil {
		log.Error("io.Copy error(%v)", err)
		return
	}
	save = strings.Replace(save, s.c.APK.LocalDir, s.c.APK.LocalDomain, -1)
	return
}

// ApkPutOss put apk object into oss.
func (s *Service) ApkPutOss(c context.Context, folder, filename string) (uri string, err error) {
	var (
		f         *os.File
		filePatch = path.Join(s.c.APK.LocalDir, folder, filename)
	)
	if f, err = os.Open(filePatch); err != nil {
		log.Error("%v", err)
		return
	}
	defer f.Close()
	tmp := new(bytes.Buffer)
	if _, err = io.Copy(tmp, f); err != nil {
		log.Error("io.Copy error(%v)", err)
		return
	}
	if uri, err = s.oss.Put(c, tmp, path.Join(folder, filename)); err != nil {
		log.Error("s.oss.Put(%s) error(%v)", path.Join(folder, filename), err)
		return
	}
	// dl-hdslb-com -> http://dl.hdslb.com
	uri = strings.Replace(uri, s.c.Oss.Bucket, s.c.APK.CDNDomain, -1)
	return
}
