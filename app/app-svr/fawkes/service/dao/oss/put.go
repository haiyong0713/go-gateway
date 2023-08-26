package oss

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// 发布文件
// PATH: /${originDir}/${relationPath}
func (d *Dao) putWrapper(c context.Context, rd io.Reader, relationPath string, ossClientType int64) (uri string, err error) {
	var (
		putConfig *PutConfig
	)
	if putConfig, err = d.getOssConfig(ossClientType); err != nil {
		log.Error("putWrapper(%s) error(%v)", uri, err)
		return
	}
	if uri, err = d.putObject(c, rd, putConfig.client, putConfig.bucketName, path.Join(putConfig.originDir, relationPath), putConfig.cdnDomain); err != nil {
		log.Error("putWrapper(%s) error(%v)", uri, err)
		return
	}
	return
}

// Put put object into oss.
func (d *Dao) putObject(c context.Context, rd io.Reader, ossClient *oss.Client, bucketName, objectKey, cdnDomain string) (uri string, err error) {
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		log.Errorc(c, "putObject(%s) error(%v)", uri, err)
		return
	}
	// 添加Content-Disposition
	// .exe .dmg 追加指定下载文件名
	fileName := path.Base(objectKey)
	fileType := path.Ext(fileName)
	if fileType == "exe" || fileType == "dmg" {
		fileContentDisposition := fmt.Sprintf("attachment;filename=%v", path.Base(objectKey))
		opts := []oss.Option{oss.ContentDisposition(fileContentDisposition)}
		if err = bucket.PutObject(objectKey, rd, opts...); err != nil {
			log.Errorc(c, "putObject(%s) error(%v)", uri, err)
			return
		}
		log.Infoc(c, "putObject with Content-Disposition: key:%s，filename:%v", objectKey, path.Base(objectKey))
	} else {
		// 旧方案
		if err = bucket.PutObject(objectKey, rd); err != nil {
			log.Errorc(c, "putObject(%s) error(%v)", uri, err)
			return
		}
	}
	uri = path.Join(bucketName, objectKey)
	uri = strings.Replace(uri, bucketName, cdnDomain, -1)
	return
}

// ------------------------------------------------------------------------

// FilePutOss put file object into oss.
// eg: /mnt/build-archive/archive/fawkes/${folder}/${filename} -> http://dl.hdslb.com/mobile/${folder}/${filename}
func (d *Dao) FilePutOss(c context.Context, folder, filename string, ossChannel int64) (uri, fmd5 string, size int64, err error) {
	var (
		f        *os.File
		filePath = path.Join(d.c.LocalPath.LocalDir, folder, filename)
	)
	if f, err = os.Open(filePath); err != nil {
		log.Error("%v", err)
		return
	}
	defer f.Close()
	tmp := new(bytes.Buffer)
	if _, err = io.Copy(tmp, f); err != nil {
		log.Error("io.Copy error(%v)", err)
		return
	}
	md5Bs := md5.Sum(tmp.Bytes())
	fmd5 = hex.EncodeToString(md5Bs[:])
	size = int64(tmp.Len())
	if uri, err = d.putWrapper(c, tmp, path.Join(folder, filename), ossChannel); err != nil {
		log.Error("s.oss.Put(%s) error(%v)", path.Join(folder, filename), err)
		return
	}
	return
}

// FileUploadOss put file object into oss (推荐使用)
// eg: /mnt/build-archive/archive/fawkes/${filePath} -> http://dl.hdslb.com/mobile/${relativePath}
func (d *Dao) FileUploadOss(c context.Context, filePath, relativePath string, ossChannel int64) (uri, fmd5 string, size int64, err error) {
	var (
		f *os.File
	)
	if f, err = os.Open(filePath); err != nil {
		log.Error("%v", err)
		return
	}
	defer f.Close()
	tmp := new(bytes.Buffer)
	if _, err = io.Copy(tmp, f); err != nil {
		log.Error("io.Copy error(%v)", err)
		return
	}
	md5Bs := md5.Sum(tmp.Bytes())
	fmd5 = hex.EncodeToString(md5Bs[:])
	size = int64(tmp.Len())
	if uri, err = d.putWrapper(c, tmp, relativePath, ossChannel); err != nil {
		log.Error("s.oss.Put(%s) error(%v)", filePath, err)
		return
	}
	return
}
