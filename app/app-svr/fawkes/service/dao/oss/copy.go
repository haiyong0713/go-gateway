package oss

import (
	"path"
	"strings"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// copyObjectWrapper
func (d *Dao) copyObjectWrapper(ossClient *oss.Client, bucketName, originDir, objectKey, publishDir, destKey, cdnDomain string) (uri string, err error) {
	bucket, err := ossClient.Bucket(bucketName)
	if err != nil {
		log.Error("bucket.copyObject(%s) error(%v)", uri, err)
		return
	}
	_, err = bucket.CopyObject(originDir+"/"+objectKey, publishDir+"/"+destKey)
	if err != nil {
		log.Error("bucket.copyObject(%s, %s) error(%v)", originDir+"/"+objectKey, publishDir+"/"+destKey, err)
		return
	}
	uri = path.Join(bucketName, publishDir, destKey)
	uri = strings.Replace(uri, bucketName, cdnDomain, -1)
	return
}

// 复制Oss上的文件到Publish目录下某个指定路径
// ${destOssDir}/${destKey}
func (d *Dao) Publish(objectKey, destKey string, ossChannel int64) (uri string, err error) {
	var (
		putConfig *PutConfig
	)
	if putConfig, err = d.getOssConfig(ossChannel); err != nil {
		return
	}
	if uri, err = d.copyObjectWrapper(putConfig.client, putConfig.bucketName, putConfig.originDir, objectKey, putConfig.publishDir, destKey, putConfig.cdnDomain); err != nil {
		return
	}
	return
}

// 复制Oss上的文件到 指定目录下某个指定路径
// ${destOssDir}/${destKey}
func (d *Dao) PublishWithDir(objectKey, destKey, destOssDir string, ossChannel int64) (uri string, err error) {
	var (
		putConfig *PutConfig
	)
	if putConfig, err = d.getOssConfig(ossChannel); err != nil {
		return
	}
	if uri, err = d.copyObjectWrapper(putConfig.client, putConfig.bucketName, putConfig.originDir, objectKey, destOssDir, destKey, putConfig.cdnDomain); err != nil {
		return
	}
	return
}
