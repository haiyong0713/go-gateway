package bangumi

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	"go-gateway/app/app-svr/app-car/interface/model/playurl"

	"github.com/pkg/errors"
)

const (
	_model       = "/pgc/page/car/module"
	_view        = "/pgc/view/v2/app/season"
	_playurlH5   = "/pgc/player/web/playurl/html5/full"
	_playurlProj = "/pgc/player/api/playurlproj"
	_playurlApp  = "/pgc/player/api/playurl"
)

func (d *Dao) Module(c context.Context, id int, mobiApp, buvid string) ([]*bangumi.Module, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mobi_app", mobiApp)
	params.Set("id", strconv.Itoa(id))
	var res struct {
		Code int               `json:"code"`
		Data []*bangumi.Module `json:"data"`
	}
	req, err := d.client.NewRequest("GET", d.module, ip, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Buvid", buvid)
	if err = d.client.Do(c, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.module+"?"+params.Encode())
	}
	return res.Data, nil
}

func (d *Dao) View(c context.Context, mid, seasonid int64, accessKey, cookie, mobiApp, platform, buvid, referer string, build int) (*bangumi.View, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("season_id", strconv.FormatInt(seasonid, 10))
	params.Set("access_key", accessKey)
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("build", strconv.Itoa(build))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int           `json:"code"`
		Data *bangumi.View `json:"data"`
	}
	req, err := d.client.NewRequest("GET", d.view, ip, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Buvid", buvid)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", referer)
	if err = d.client.Do(c, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.view+"?"+params.Encode())
	}
	if res.Data == nil {
		return nil, ecode.NothingFound
	}
	return res.Data, nil
}

func (d *Dao) PlayurlH5(c context.Context, buvid, cookie, referer string, param *playurl.Param) (*bangumi.PlayInfo, string, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("build", strconv.Itoa(param.Build))
	params.Set("ep_id", strconv.FormatInt(param.Cid, 10))
	params.Set("platform", param.Platform)
	params.Set("qn", strconv.FormatInt(param.Qn, 10))
	params.Set("fnver", strconv.Itoa(param.Fnver))
	params.Set("fnval", strconv.Itoa(param.Fnval))
	params.Set("fourk", strconv.Itoa(param.Fourk))
	params.Set("force_host", strconv.Itoa(param.ForceHost))
	params.Set("is_preview", strconv.Itoa(param.IsPreview))
	params.Set("channel", "web_car")
	req, err := d.client.NewRequest("GET", d.playurlH5, ip, params)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Buvid", buvid)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", referer)
	var data struct {
		Code    int               `json:"code"`
		Data    *bangumi.PlayInfo `json:"data"`
		Message string            `json:"message"`
	}
	if err := d.client.Do(c, req, &data); err != nil {
		return nil, "", err
	}
	if data.Code != ecode.OK.Code() {
		return nil, "由于权限等原因无法播放", errors.Wrap(ecode.Int(data.Code), data.Message)
	}
	return data.Data, "", nil
}

func (d *Dao) PlayurlProj(c context.Context, buvid, cookie, referer string, param *playurl.Param) (*bangumi.PlayInfo, string, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("build", strconv.Itoa(param.Build))
	params.Set("ep_id", strconv.FormatInt(param.Cid, 10))
	params.Set("platform", param.Platform)
	params.Set("qn", strconv.FormatInt(param.Qn, 10))
	params.Set("fnver", strconv.Itoa(param.Fnver))
	params.Set("fnval", strconv.Itoa(param.Fnval))
	params.Set("fourk", strconv.Itoa(param.Fourk))
	params.Set("force_host", strconv.Itoa(param.ForceHost))
	params.Set("is_preview", strconv.Itoa(param.IsPreview))
	req, err := d.client.NewRequest("GET", d.playurlProj, ip, params)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Buvid", buvid)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", referer)
	var data struct {
		Code int `json:"code"`
		*bangumi.PlayInfo
		Message string `json:"message"`
	}
	if err := d.client.Do(c, req, &data); err != nil {
		return nil, "", err
	}
	if data.Code != ecode.OK.Code() {
		return nil, "由于权限等原因无法播放", errors.Wrap(ecode.Int(data.Code), data.Message)
	}
	return data.PlayInfo, "", nil
}

func (d *Dao) PlayurlAPP(c context.Context, buvid, referer string, param *playurl.Param) (*bangumi.PlayInfo, string, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("dolby log  err:%+v", r)
		}
	}()
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("access_key", param.AccessKey)
	params.Set("mid", strconv.FormatInt(param.Mid, 10))
	params.Set("mobi_app", param.MobiApp)
	params.Set("device", param.Device)
	params.Set("build", strconv.Itoa(param.Build))
	params.Set("platform", param.Platform)
	params.Set("ep_id", strconv.FormatInt(param.Cid, 10))
	params.Set("qn", strconv.FormatInt(param.Qn, 10))
	params.Set("fnver", strconv.Itoa(param.Fnver))
	params.Set("fnval", strconv.Itoa(param.Fnval))
	params.Set("fourk", strconv.Itoa(param.Fourk))
	params.Set("force_host", strconv.Itoa(param.ForceHost))
	params.Set("is_preview", strconv.Itoa(param.IsPreview))
	params.Set("backup_num", strconv.Itoa(int(d.c.Custom.BackupNum)))
	params.Set("is_dazhongcar", strconv.FormatBool(param.IsDazhongcar))
	req, err := d.client.NewRequest("GET", d.playurlApp, ip, params)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Buvid", buvid)
	req.Header.Set("Referer", referer)
	var data struct {
		Code int `json:"code"`
		*bangumi.PlayInfo
		Message string `json:"message"`
	}
	//see https://git.bilibili.co/bapis/bapis/blob/master/video/vod/playurlpgc/service.proto
	if err := d.client.Do(c, req, &data); err != nil {
		return nil, "", err
	}

	if data.Code != ecode.OK.Code() {
		return nil, data.Message, errors.Wrap(ecode.Int(data.Code), data.Message)
	}
	return data.PlayInfo, "", nil
}
