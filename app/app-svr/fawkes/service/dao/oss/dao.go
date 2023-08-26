package oss

import (
	"errors"

	"go-gateway/app/app-svr/fawkes/service/conf"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Dao
type Dao struct {
	c            *conf.Config
	inlandClient *oss.Client
	abroadClient *oss.Client
}

// PutConfig
type PutConfig struct {
	client     *oss.Client
	bucketName string
	originDir  string
	publishDir string
	cdnDomain  string
}

// New
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	// 初始化国内client
	inlandClient, err := oss.New(d.c.Oss.Inland.Endpoint, d.c.Oss.Inland.AccessKeyID, d.c.Oss.Inland.AccessKeySecret)
	if err != nil {
		panic(err)
	}
	// 初始化海外client
	abroadClient, err := oss.New(d.c.Oss.Abroad.Endpoint, d.c.Oss.Abroad.AccessKeyID, d.c.Oss.Abroad.AccessKeySecret)
	if err != nil {
		panic(err)
	}
	d.inlandClient = inlandClient
	d.abroadClient = abroadClient
	return
}

// 获取配置信息
func (d *Dao) getOssConfig(ossChannel int64) (putConfig *PutConfig, err error) {
	switch ossChannel {
	// 国内配置
	case appmdl.AppServerZone_Inland:
		putConfig = &PutConfig{
			client:     d.inlandClient,
			bucketName: d.c.Oss.Inland.Bucket,
			originDir:  d.c.Oss.Inland.OriginDir,
			publishDir: d.c.Oss.Inland.PublishDir,
			cdnDomain:  d.c.Oss.Inland.CDNDomain,
		}
	// 海外配置
	case appmdl.AppServerZone_Abroad:
		putConfig = &PutConfig{
			client:     d.abroadClient,
			bucketName: d.c.Oss.Abroad.Bucket,
			originDir:  d.c.Oss.Abroad.OriginDir,
			publishDir: d.c.Oss.Abroad.PublishDir,
			cdnDomain:  d.c.Oss.Abroad.CDNDomain,
		}
	// 默认引用国内配置
	default:
		err = errors.New("getPutConfig not found")
	}
	return
}

////nolint:unused
//func (d *Dao) darkness(seed string) (uri string) {
//	seed += `\u76d8\u53e4\u6709\u8bad`
//	seed += `\u7eb5\u6a2a\u516d\u754c`
//	seed += `\u8bf8\u4e8b\u7686\u6709\u7f18\u6cd5`
//	seed += `\u51e1\u4eba\u4ef0\u89c2\u82cd\u5929`
//	seed += `\u65e0\u660e\u65e5\u6708\u6f5c\u606f`
//	seed += `\u56db\u65f6\u66f4\u66ff`
//	seed += `\u5e7d\u51a5\u4e4b\u95f4`
//	seed += `\u4e07\u7269\u5df2\u5faa\u56e0\u7f18`
//	seed += `\u6052\u5927\u8005\u5219\u4e3a\u5929\u9053`
//	seed += `\u76d8\u53e4\u6709\u8bad`
//	seed += `\u7eb5\u6a2a\u516d\u754c`
//	seed += `\u8bf8\u4e8b\u7686\u6709\u7f18\u6cd5`
//	seed += `\u51e1\u4eba\u4ef0\u89c2\u82cd\u5929`
//	seed += `\u65e0\u660e\u65e5\u6708\u6f5c\u606f`
//	seed += `\u56db\u65f6\u66f4\u66ff`
//	seed += `\u5e7d\u51a5\u4e4b\u95f4`
//	seed += `\u4e07\u7269\u5df2\u5faa\u56e0\u7f18`
//	seed += `\u6052\u5927\u8005\u5219\u4e3a\u5929\u9053`
//	sUnicodev := strings.Split(seed, "\\u")
//	for _, v := range sUnicodev {
//		if len(v) < 1 {
//			continue
//		}
//		temp, err := strconv.ParseInt(v, 16, 32)
//		if err != nil {
//			panic(err)
//		}
//		uri += fmt.Sprintf("%c", temp)
//	}
//	return
//}
