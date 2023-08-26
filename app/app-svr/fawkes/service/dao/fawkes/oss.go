package fawkes

import (
	"context"

	"go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// 复制Oss上的文件到Publish目录下某个指定路径
// ${destOssDir}/${destKey}
func (d *Dao) Publish(objectKey, destKey, appKey string) (uri string, err error) {
	var appInfo *app.APP
	if appInfo, err = d.AppPass(context.Background(), appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if uri, err = d.ossWrapper.Publish(objectKey, destKey, appInfo.ServerZone); err != nil {
		log.Error("%v", err)
	}
	return
}

// 复制Oss上的文件到 指定目录下某个指定路径
// ${destOssDir}/${destKey}
func (d *Dao) PublishWithDir(objectKey, destKey, destOssDir, appKey string) (uri string, err error) {
	var appInfo *app.APP
	if appInfo, err = d.AppPass(context.Background(), appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if uri, err = d.ossWrapper.PublishWithDir(objectKey, destKey, destOssDir, appInfo.ServerZone); err != nil {
		log.Error("%v", err)
	}
	return
}

// FilePutOss put file object into oss.
// eg: /mnt/build-archive/archive/fawkes/${folder}/${filename} -> http://dl.hdslb.com/mobile/${folder}/${filename}
func (d *Dao) FilePutOss(c context.Context, folder, filename, appKey string) (uri, fmd5 string, size int64, err error) {
	var appInfo *app.APP
	if appInfo, err = d.AppPass(context.Background(), appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if uri, fmd5, size, err = d.ossWrapper.FilePutOss(c, folder, filename, appInfo.ServerZone); err != nil {
		log.Error("%v", err)
	}
	return
}

// FileUploadOss put file object into oss (推荐使用)
// eg: /mnt/build-archive/archive/fawkes/${filePath} -> http://dl.hdslb.com/mobile/${relativePath}
func (d *Dao) FileUploadOss(c context.Context, filePath, relativePath, appKey string) (uri, fmd5 string, size int64, err error) {
	var appInfo *app.APP
	if appInfo, err = d.AppPass(context.Background(), appKey); err != nil {
		log.Error("%v", err)
		return
	}
	if uri, fmd5, size, err = d.ossWrapper.FileUploadOss(c, filePath, relativePath, appInfo.ServerZone); err != nil {
		log.Error("%v", err)
	}
	return
}
