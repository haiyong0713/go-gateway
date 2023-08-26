package newyear2021

import (
	"encoding/json"
	"testing"
)

// go test -v -count=1 model_test.go model.go
func TestModelBiz(t *testing.T) {
	t.Run("android testing", androidTesting)
	t.Run("ios testing", iosTesting)
}

func iosTesting(t *testing.T) {
	userAgent := "bili-universal/61500090 CFNetwork/1197 Darwin/20.0.0 os/ios model/iPhone 12 Pro mobi_app/iphone_b build/61500090 osVer/14.1 network/2 channel/AppStore"
	info, err := ParseUserAgent2UserAppInfo(userAgent)
	if err != nil {
		t.Error(err)

		return
	}

	if info.Os != Os4Ios || info.Model != "iphone12pro" || info.OsVersion != 14.1 {
		t.Error("parse ios device user_agent failed")

		return
	}

	bs, _ := json.Marshal(info)
	t.Log(string(bs))
}

func androidTesting(t *testing.T) {
	userAgent := "Mozilla/5.0 BiliDroid/5.53.1 (bbcallen@gmail.com) os/android model/Redmi K30 Pro Zoom Edition mobi_app/android build/5531000 channel/xiaomi_cn_tv.danmaku.bili_20190930 innerVer/5531000 osVer/10 network/2"
	info, err := ParseUserAgent2UserAppInfo(userAgent)
	if err != nil {
		t.Error(err)

		return
	}

	if info.Os != Os4Android || info.Model != "redmik30prozoomedition" || info.OsVersion != 10 {
		t.Error("parse android device user_agent failed")

		return
	}

	bs, _ := json.Marshal(info)
	t.Log(string(bs))
}
