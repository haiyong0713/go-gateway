package dao

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/pkg/errors"
)

const (
	_accTagsURI    = "/api/tag/get"
	_accTagsSetURI = "/api/tag/set"
	_isAnsweredURI = "/x/internal/credit/labour/isanswered"
)

// AccTags get account tags.
func (d *Dao) AccTags(c context.Context, mid int64) (data json.RawMessage, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("mids", strconv.FormatInt(mid, 10))
	var res struct {
		Code int             `json:"code"`
		List json.RawMessage `json:"list"`
	}
	if err = d.httpR.Get(c, d.accTagsURL, ip, params, &res); err != nil {
		log.Error("d.httpR.Get(%s) error(%v)", d.accTagsURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("d.httpR.Get(%s) code(%d) error", d.accTagsURL, res.Code)
		err = ecode.Int(res.Code)
		return
	}
	data = res.List
	return
}

// SetAccTags set account tags.
func (d *Dao) SetAccTags(c context.Context, tags, ck string) (err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("tags", tags)
	var req *http.Request
	if req, err = d.httpW.NewRequest(http.MethodGet, d.accTagsSetURL, ip, params); err != nil {
		log.Error("SetAccTags(%s) error(%v)", d.accTagsSetURL, err)
		return
	}
	req.Header.Set("Cookie", ck)
	var res struct {
		Code int `json:"code"`
	}
	if err = d.httpW.Do(c, req, &res); err != nil {
		log.Error("SetAccTags(%s) error(%v)", d.accTagsSetURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("SetAccTags(%s) res(%v)", d.accTagsSetURL, res)
		err = ecode.Int(res.Code)
	}
	return
}

// IsAnswered get if block account answered.
func (d *Dao) IsAnswered(c context.Context, mid, start int64) (status int, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("start", strconv.FormatInt(start, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Status int `json:"status"`
		} `json:"data"`
	}
	if err = d.httpR.Get(c, d.isAnsweredURL, ip, params, &res); err != nil {
		log.Error("d.httpR.Get(%s) error(%v)", d.isAnsweredURL, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		log.Error("d.httpR.Get(%s) code(%d) error", d.isAnsweredURL, res.Code)
		err = ecode.Int(res.Code)
		return
	}
	status = res.Data.Status
	return
}

func (d *Dao) AccountInfo(ctx context.Context, mid int64) (*accgrpc.Info, error) {
	req := &accgrpc.MidReq{Mid: mid}
	reply, err := d.accGRPC.Info3(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", req)
	}
	return reply.GetInfo(), nil
}
