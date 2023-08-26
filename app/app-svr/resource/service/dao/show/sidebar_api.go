package show

import (
	"context"
	"net/url"
	"strconv"
	"unicode/utf8"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/resource/service/model"

	"github.com/pkg/errors"
)

// WhiteCheck 各种入口白名单
// https://www.tapd.cn/20055921/prong/stories/view/1120055921001066980  动态互推TAPD在此！！
func (d *Dao) WhiteCheck(c context.Context, checkURL string, mid int64, buvid string) (ok bool, err error) {
	if checkURL == "" || (mid == 0 && buvid == "") {
		return false, nil
	}
	params := url.Values{}

	if uri, err := url.Parse(checkURL); err == nil {
		params = uri.Query()
		checkURL = "http://" + uri.Host + uri.Path
	}

	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)

	var res struct {
		Code int `json:"code"`
		Data struct {
			Status int `json:"status"`
		} `json:"data"`
	}
	if err = d.httpClient.Get(c, checkURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), checkURL+"?"+params.Encode())
		return
	}
	if res.Data.Status == 1 {
		ok = true
		//} else {
		//	log.Warn("white check: %v res: %v", checkURL+"?"+params.Encode(), res)
	}
	return
}

// RedDot 我的页小红点逻辑
func (d *Dao) RedDot(c context.Context, mid int64, redDotURL string) (ok bool, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			RedDot bool `json:"red_dot"`
		} `json:"data"`
	}
	if err = d.httpClient.Get(c, redDotURL, "", params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), redDotURL+"?"+params.Encode())
		return
	}
	//log.Warn("reddot response mid(%d) url(%s) res(%t)", mid, redDotURL+"?"+params.Encode(), res.Data.RedDot)
	ok = res.Data.RedDot
	return
}

const NameMaxLen = 10

// 获取动态配置
func (d *Dao) FetchDynamicConf(c context.Context, checkURL string, mid int64, buvid string) (conf *model.DynamicConf, err error) {
	if checkURL == "" || (mid == 0 && buvid == "") {
		return nil, nil
	}
	params := url.Values{}

	if uri, err := url.Parse(checkURL); err == nil {
		params = uri.Query()
		checkURL = "http://" + uri.Host + uri.Path
	}

	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)

	var res struct {
		Code int               `json:"code"`
		Data model.DynamicConf `json:"data"`
	}

	if err = d.httpClient.Get(c, checkURL, "", params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), checkURL+"?"+params.Encode())
		return nil, err
	}
	if utf8.RuneCountInString(res.Data.Name) > NameMaxLen {
		res.Data.Name = ""
	}

	if res.Data.Name == "" && res.Data.Param == "" {
		return nil, nil
	}
	return &res.Data, nil
}
