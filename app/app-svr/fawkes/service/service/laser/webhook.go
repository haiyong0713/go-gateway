package laser

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/fawkes/service/model/app"
)

func (s *Service) LaserPushByWebhook(c context.Context, appInfo *app.APP, laser *app.Laser) (_ int64, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("task_id", strconv.FormatInt(laser.ID, 10))
	params.Set("buvid", laser.Buvid)
	params.Set("mid", strconv.FormatInt(laser.MID, 10))
	params.Set("mobi_app", laser.MobiApp)
	params.Set("log_date", laser.LogDate)
	var res struct {
		Code int `json:"code"`
	}
	if err = s.httpClient.Get(c, appInfo.LaserWebhook, ip, params, &res); err != nil {
		return
	}
	return
}
