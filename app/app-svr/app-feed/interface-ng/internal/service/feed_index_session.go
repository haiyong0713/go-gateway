package service

import (
	"net/http"
	"net/url"
	"strconv"

	"go-common/component/metadata/device"
	"go-common/library/net/http/blademaster/binding"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/api/session"

	"github.com/pkg/errors"
)

func parseDevice(fakeReq *http.Request) *feedcard.CtxDevice {
	dev := new(device.Device)
	header := fakeReq.Header
	dev.TfISP = header.Get("X-Tf-Isp")
	dev.UserAgent = header.Get("User-Agent")
	dev.Buvid = header.Get("Buvid")
	dev.FawkesAppKey = header.Get("App-key")
	dev.FawkesEnv = header.Get("Env")
	if dev.FawkesEnv == "" {
		dev.FawkesEnv = "prod"
	}
	if buvid3, err := fakeReq.Cookie("buvid3"); err == nil && buvid3 != nil {
		dev.Buvid3 = buvid3.Value
	}
	if sid, err := fakeReq.Cookie("sid"); err == nil && sid != nil {
		dev.Sid = sid.Value
	}
	query := fakeReq.URL.Query()
	if build, err := strconv.ParseInt(query.Get("build"), 10, 64); err == nil {
		dev.Build = build
	}
	dev.Channel = query.Get("channel")
	dev.Device = query.Get("device")
	dev.RawMobiApp = query.Get("mobi_app")
	dev.RawPlatform = query.Get("platform")
	dev.Model = query.Get("model")
	dev.Brand = query.Get("brand")
	dev.Osver = query.Get("osver")
	dev.Network = query.Get("network")
	return feedcard.NewCtxDevice(dev)
}

func parseFeedParam(session *session.IndexSession) (*feedcard.IndexParam, *feedcard.CtxDevice, error) {
	fakeReq, _ := http.NewRequest("GET", "http://api.bilibili.com/x/v2/feed/index", nil)
	query := url.Values(session.Request.Query)
	header := http.Header(session.Request.Header)
	fakeReq.URL.RawQuery = query.Encode()
	fakeReq.Header = header

	param := &feedcard.IndexParam{}
	if err := binding.Form.Bind(fakeReq, param); err != nil {
		return nil, nil, err
	}
	param.AppList = header.Get("AppList")
	param.DeviceInfo = header.Get("DeviceInfo")
	if _, ok := cdm.Columnm[param.Column]; !ok {
		return nil, nil, errors.Errorf("invalid column: %d", param.Column)
	}
	style := int(cdm.Columnm[param.Column])
	if style == 1 {
		//nolint:ineffassign
		style = 3
	}
	device := parseDevice(fakeReq)
	return param, device, nil
}
