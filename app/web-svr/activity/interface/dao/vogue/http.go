package dao

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"
)

func (d *Dao) WinList(c context.Context, sid string) (data []*model.WinListItem, err error) {
	var req *http.Request
	val := url.Values{}
	val.Add("sid", sid)
	val.Add("num", "30")
	if req, err = d.client.NewRequest(http.MethodGet, d.winListURL, metadata.String(c, metadata.RemoteIP), val); err != nil {
		return
	}
	var res struct {
		Code int                  `json:"code"`
		Data []*model.WinListItem `json:"data"`
	}
	if err = d.client.Do(c, req, &res, d.winListURL); err != nil {
		log.Error("winList(%v)", err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("DrawList res.code(%v)", res.Code)
		return nil, ecode.New(res.Code)
	}
	data = res.Data
	return
}

func (d *Dao) FavList(c context.Context, mediaId int64) (data []*model.FavInfo, err error) {
	var req *http.Request
	val := url.Values{}
	val.Add("media_id", fmt.Sprint(mediaId))
	val.Add("pn", "1")
	if req, err = d.client.NewRequest(http.MethodGet, d.favURL, "", val); err != nil {
		return
	}
	var res struct {
		Code int              `json:"code"`
		Data []*model.FavInfo `json:"data"`
	}
	if err = d.client.Do(c, req, &res, d.favURL); err != nil {
		log.Error("winList(%v,%v)", mediaId, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("DrawList res.code(%v,%v)", mediaId, res.Code)
		return nil, ecode.New(res.Code)
	}
	data = res.Data
	return
}
