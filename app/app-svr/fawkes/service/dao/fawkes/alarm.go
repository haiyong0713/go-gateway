package fawkes

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go-gateway/app/app-svr/fawkes/service/model"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

/*
	func Broadcast()参数结构示例
	{
		"username": "a,b",
		"hook": {
			"uri": "http://xxxx",
			"method": "post"
		},
		"param": {
			"app_key": "iphone",
			"build_id": 123,
			"ctime": 158393835
		}
		"bots":"http://xxxx,http://xxxx"
	}
*/

// Hook
func (d *Dao) Hook(c context.Context, param interface{}, hook *model.Hook) (err error) {
	var (
		params = url.Values{}
	)
	if hook == nil {
		log.Errorc(c, "hook is nil")
		return
	}
	// 若URL携带参数. 则会携带至最终的url中
	parseResult, err := url.Parse(hook.URI)
	if err != nil {
		log.Errorc(c, "hook url params not found")
	}
	requestURL := strings.Split(hook.URI, "?")[0]
	if parseResult != nil {
		params = parseResult.Query()
	}
	switch paramType := param.(type) {
	case *cimdl.HookParam:
		ciParam := paramType
		params.Set("gl_job_id", strconv.FormatInt(ciParam.GitlabJobID, 10))
		params.Set("build_id", strconv.FormatInt(ciParam.BuildID, 10))
		params.Set("app_key", ciParam.AppKey)
		params.Set("pack_url", ciParam.PackURL)
		log.Infoc(c, "start webhook request - app_key = %v AND gl_job_id = %v AND build_id = %v AND pack_url = %v", ciParam.AppKey, ciParam.GitlabJobID, ciParam.BuildID, ciParam.PackURL)
	default:
		log.Errorc(c, "param(%v) error", param)
		return
	}
	if err = d.HookRequest(c, requestURL, hook.Method, params); err != nil {
		log.Errorc(c, "webhook request error(%v)", err)
		return
	}
	log.Infoc(c, "webhook request success")
	return
}

func (d *Dao) HookRequest(c context.Context, requestUrl, method string, params url.Values) (err error) {
	var (
		req *http.Request
	)
	if params == nil {
		params = url.Values{}
	}
	if req, err = d.httpClient.NewRequest(method, requestUrl, "", params); err != nil {
		log.Errorc(c, "d.httpClient.NewRequest error(%v)", err)
		return
	}
	if err = d.httpClient.Do(c, req, &struct{}{}); err != nil {
		log.Errorc(c, "d.httpClient.Do error(%v)", err)
	}
	return
}
