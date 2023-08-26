package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	xhttp "net/http"
	"net/url"

	"go-common/library/database/elastic"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus/report"
	"go-gateway/app/web-svr/activity/admin/model"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

// AddTags add tags from http request.
func (d *Dao) AddTags(c context.Context, tags string, ip string) (err error) {
	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	params := url.Values{}
	params.Set("tag_name", tags)
	if err = d.client.Post(c, d.actURLAddTags, ip, params, &res); err != nil {
		err = errors.Wrapf(err, "d.client.Post(%s)", d.actURLAddTags)
		return
	}
	if res.Code != 0 {
		err = fmt.Errorf("res code(%v)", res)
	}
	return
}

type FeatureImportOid struct {
	UserName string `json:"user_name"`
	GroupID  int    `json:"group_id"`
	Type     int    `json:"type"`
	Oid      string `json:"oid"`
	Cover    int    `json:"cover"`
}

func (d *Dao) SubjectAuditFeatureImport(c context.Context, sid int64, username string) (err error) {
	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	params := struct {
		Oids []*FeatureImportOid `json:"oids"`
	}{
		Oids: []*FeatureImportOid{
			{
				UserName: username,
				GroupID:  d.c.Subject.AuditGroupID,
				Type:     3,
				Oid:      fmt.Sprint(sid),
				Cover:    1,
			},
		},
	}
	buf, _ := json.Marshal(params)
	var req *xhttp.Request
	req, err = xhttp.NewRequest(xhttp.MethodPost, d.ContentFeatureSingleImportURL, bytes.NewReader(buf))
	if err != nil {
		err = errors.Wrapf(err, "d.client.NewRequest(%s)", d.ContentFeatureSingleImportURL)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.client.Do(c, req, &res); err != nil {
		err = errors.Wrapf(err, "d.client.Do(%s)", d.ContentFeatureSingleImportURL)
		return
	}
	if res.Code != 0 {
		err = fmt.Errorf("res code(%v)", res)
	}
	return
}

// SubStat .
func (d *Dao) SubStat(c context.Context, sid int64) (rly *model.SubjectStat, err error) {
	rly = new(model.SubjectStat)
	if err = d.DB.Where("sid = ?", sid).First(rly).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}
	return
}

func (d *Dao) UserAwardLog(c context.Context, oid, mid, pn, ps int64) (list []*report.UserActionLog, total int64, err error) {
	var res struct {
		Page struct {
			Num   int   `json:"num"`
			Size  int   `json:"size"`
			Total int64 `json:"total"`
		} `json:"page"`
		Result []*report.UserActionLog `json:"result"`
	}
	req := d.es.NewRequest("log_user_action").Index("log_user_action_250_all")
	if oid > 0 {
		req = req.WhereEq("oid", oid)
	}
	if mid > 0 {
		req = req.WhereEq("mid", mid)
	}
	if ps > 0 {
		req = req.Ps(int(ps))
	}
	if err = req.Pn(int(pn)).Order("ctime", elastic.OrderDesc).Scan(c, &res); err != nil {
		log.Error("UserAwardLog oid(%d) mid(%d), err: %v", oid, mid, err)
		return
	}
	list = res.Result
	total = res.Page.Total
	return
}

// TagInfoByName get tagInfo from grpc
func (d *Dao) TagInfoByName(c context.Context, tagName string) (res *tagrpc.Tag, err error) {
	tagRes, err := d.tagGRPC.TagByName(c, &tagrpc.TagByNameReq{Tname: tagName})
	if err != nil {
		err = errors.Wrapf(err, "s.tagRPC.TagInfoByName err(%v)", err)
		return nil, err
	}
	if tagRes == nil || tagRes.Tag == nil {
		return nil, nil
	}
	return tagRes.Tag, nil
}

// TagUpdateByID get tagInfo from grpc
func (d *Dao) TagUpdateByID(c context.Context, id int64, tagType tagrpc.TagType) (err error) {
	_, err = d.tagGRPC.UpdateTagType(c, &tagrpc.UpdateTagTypeReq{Tid: id, Type: tagType})
	if err != nil {
		err = errors.Wrapf(err, "s.tagRPC.TagInfoByName err(%v)", err)
		return err
	}
	return nil
}

// AddTagNew add tags from http request.
func (d *Dao) AddTagNew(c context.Context, tags string) (tag *tagrpc.Tag, err error) {
	tagReply, err := d.tagGRPC.AddTag(c, &tagrpc.AddTagReq{Name: tags})
	if err != nil || tagReply == nil || tagReply.Tag == nil {
		err = errors.Wrapf(err, "s.tagRPC.AddTag err(%v)", err)
		return nil, err
	}

	return tagReply.Tag, nil
}

// AddNativePage .
func (d *Dao) AddNativePage(c context.Context, tagName, ip string) (err error) {
	var res struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	params := url.Values{}
	params.Set("topic", tagName)
	params.Set("source", "video")
	if err = d.client.Post(c, d.actNativeURL, ip, params, &res); err != nil {
		log.Errorc(c, "AddNativePage d.client.Post(%s)", d.actNativeURL)
		return
	}
	if res.Code == 16001 || res.Code == 176000 {
		err = xecode.Code(res.Code)
		return
	}
	if res.Code != 0 {
		err = fmt.Errorf("AddNativePage res code(%v)", res)
	}
	return
}
