package fawkes

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	"go-gateway/app/app-svr/fawkes/service/model/mod"
	"go-gateway/app/app-svr/fawkes/service/model/template"
)

var viewData = map[string]map[string][]cimdl.BBRItem{
	"class": {
		"tes1": {{Name: "1", Id: "2", DisplayName: "3"}, {Name: "1", Id: "2", DisplayName: "4"}},
	},
}

const APIInfo = `{{ range $key, $value := . }}
{{ $key }}:{{ range $key1, $value1 := $value}}
	{{ $key1 }}:{{ range $value1}}
		{{.DisplayName}}{{end}}{{end}}{{end}}
`

func TestDao_TemplateAlter(t *testing.T) {
	Convey("template", t, func() {
		alter, err := d.TemplateAlter(viewData, APIInfo)
		So(err, ShouldBeNil)
		fmt.Printf(alter)
	})

}

func TestDao_TemplateAlter1(t *testing.T) {
	Convey("TestDao_TemplateAlter1", t, func() {

		data := &mod.TrafficNotify{
			TrafficDetail: &mod.TrafficDetail{
				AppKey:                       "",
				PoolName:                     "",
				ModName:                      "",
				Operator:                     "",
				VerNum:                       0,
				OriginFileSize:               "",
				PatchFileSize:                "",
				AvgFileSize:                  "",
				DownloadCount:                0,
				DownloadSizeEstimate:         "",
				DownloadSizeOnline:           "",
				DownloadSizeOnlineTotal:      "",
				DownloadCDNBandwidthEstimate: "",
				DownloadCDNBandwidthOnline:   "",
				DownloadCDNBandwidthTotal:    "",
				ModUrl:                       "",
				Percentage:                   "",
				IsManual:                     false,
				Cost:                         0,
				Advice:                       nil,
				ErrorMsg:                     "",
				Doc:                          "",
			},
			OperateType: mod.ConfigChange,
		}

		alter, err := d.TemplateAlter(data, template.ModTrafficNotify)
		So(err, ShouldBeNil)
		fmt.Printf(alter)
	})

}

/*
【Fawkes】通知[Test-Android] master 分支debug 包打包 成功
App ID:tv.danmaku.bili
App Key:w19e
Version:6.69.0(6690000)
Internal Ver:6690000
Commit:9410689af40f9b1cf4439faa337d2759fcdfea6c
构建号:8317104
Job URL:http://git.bilibili.co/android/andruid/-/jobs/8317104
Size:153.87 MB
Archive:https://macross-jks.bilibili.co/archive/fawkes/pack/w19e/8317104
https://macross-jks.bilibili.co/archive/fawkes/pack/w19e/8317104/iBiliPlayer-apinkDebug-6.69.0-b8317104.apk
https://macross-jks.bilibili.co/archive/fawkes/pack/w19e/8317104/install.png
Change Log:04-10 12:19 d81448f2b971 颜正浩: [交易埋点修复] --预售log上传异常
04-10 13:49 02428228109d 颜正浩: [交易埋点修复] --预售log上传变换结构
04-10 13:53 b7a75b860ef1 颜正浩: [交易埋点修复] --删除无用代码
04-10 18:18 3d672b039418 yuyashuai: [main]fix闪屏结束不回调
04-10 22:13 a636d59e7945 lifan02: 【听视频】crash fix
04-06 15:17 775f35a86c54 wangzhichao02: fix fragment not attached to a context
04-11 00:00 9a91d05cef7d zhaoyang: fix bugly:https://bugly.qq.com/v2/crash-reporting/crashes/900028525/127980492\?pid\=1
04-11 10:07 17beaa4e5bbe suntao: [游戏中心]<6.68>修复bigfun bug
04-11 13:16 d25c8fe1f4c7 Jackin: ucrop add error log
04-11 12:56 c2472f807eb1 chihuili: fix: 点击卡片回顶不展示按钮
*/
var weChatContent = *&template.CIBuildInfoTemplate{
	AppID:               "tv.danmaku.bili",
	AppKey:              "w19e",
	AppName:             "Test-Android",
	Version:             "6.69.0",
	VersionCode:         6690000,
	InternalVersionCode: 6690000,
	GitlabJobID:         8317104,
	GitName:             "master",
	PackType:            "测试",
	Commit:              "9410689af40f9b1cf4439faa337d2759fcdfea6c",
	GitlabJobURL:        "http://git.bilibili.co/android/andruid/-/jobs/8317104",
	ReadableSize:        "40M",
	SaveURLDir:          "https://macross-jks.bilibili.co/archive/fawkes/pack/w19e/8317104",
	PkgURL:              "https://macross-jks.bilibili.co/archive/fawkes/pack/w19e/8317104",
	QrURLDir:            "https://macross-jks.bilibili.co/archive/fawkes/pack/w19e/8317104",
	ChangeLog:           "xxxxxx",
	ChangeLogHTML:       "xxxxx",
	FileURL:             "https://macross-jks.bilibili.co/archive/fawkes/pack/w19e/8317104",
	AddFileUrl:          false,
	Result:              "失败",
}

func TestDao_TemplateAlterFile(t *testing.T) {
	Convey("template", t, func() {
		alter, err := d.TemplateAlter(weChatContent, template.CIBuildInfoTemp_WeChat)
		So(err, ShouldBeNil)
		fmt.Printf(alter)
	})
}
