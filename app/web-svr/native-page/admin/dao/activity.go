package dao

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/web-svr/native-page/admin/model"
)

const (
	_addActSubjectURI    = "/x/admin/activity/subject/add"
	_updateActSubjectURI = "/x/admin/activity/subject/up"
)

func (d *Dao) AddActSubject(c context.Context, req *model.AddActSubjectReq) (int64, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("type", strconv.FormatInt(req.Type, 10))
	params.Set("stime", req.Stime.Format("2006-01-02 15:04:05"))
	params.Set("etime", req.Etime.Format("2006-01-02 15:04:05"))
	params.Set("author", req.Author)
	params.Set("name", req.Name)
	params.Set("types", req.Types)
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	params.Set("state", "1")
	params.Set("tags", req.Tags)
	var res struct {
		Code int   `json:"code"`
		Data int64 `json:"data"`
	}
	if err := d.actAdminClient.Post(c, d.addActSubjectURL, ip, params, &res); err != nil {
		log.Error("Fail to request AddActSubject, req=%+v error=%+v", params.Encode(), err)
		return 0, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.addActSubjectURL+"?"+params.Encode())
		log.Error("Fail to request AddActSubject, req=%+v error=%+v", params.Encode(), err)
		return 0, err
	}
	return res.Data, nil
}

func (d *Dao) UpdateActSubject(c context.Context, sid int64, req *model.AddActSubjectReq) error {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("id", strconv.FormatInt(sid, 10))
	params.Set("author", req.Author)
	params.Set("types", req.Types)
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int   `json:"code"`
		Data int64 `json:"data"`
	}
	if err := d.actAdminClient.Post(c, d.updateActSubjectURL, ip, params, &res); err != nil {
		log.Error("Fail to request UpdateActSubject, req=%+v error=%+v", params.Encode(), err)
		return err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.updateActSubjectURL+"?"+params.Encode())
		log.Error("Fail to request UpdateActSubject, req=%+v error=%+v", params.Encode(), err)
		return err
	}
	return nil
}
