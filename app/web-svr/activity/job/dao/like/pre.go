package like

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"

	"github.com/pkg/errors"
)

const (
	_preUpURI        = "/x/internal/activity/prediction/up"
	_preItemUpURI    = "/x/internal/activity/prediction/item/up"
	_preSetUpURI     = "/x/internal/activity/prediction/set"
	_preItemSetUpURI = "/x/internal/activity/prediction/set/item"
)

// UpPre .
func (d *Dao) UpPre(c context.Context, id int64) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("id", strconv.FormatInt(id, 10))
	if err = d.httpClient.Get(c, d.preUpURL, "", params, &res); err != nil {
		log.Error("UpPre:d.httpClient.Get(%s,%d) error(%v)", d.preUpURL, id, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.preUpURL+"?"+params.Encode())
	}
	return
}

// UpItemPre .
func (d *Dao) UpItemPre(c context.Context, id int64) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("id", strconv.FormatInt(id, 10))
	if err = d.httpClient.Get(c, d.preItemUpURL, "", params, &res); err != nil {
		log.Error("UpItemPre:d.httpClient.Get(%s,%d) error(%v)", d.preItemUpURL, id, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.preItemUpURL+"?"+params.Encode())
	}
	return
}

// PreSetUp .
func (d *Dao) PreSetUp(c context.Context, id, sid int64, state int) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("id", strconv.FormatInt(id, 10))
	params.Set("sid", strconv.FormatInt(sid, 10))
	params.Set("state", strconv.Itoa(state))
	if err = d.httpClient.Get(c, d.preSetUpURL, "", params, &res); err != nil {
		log.Error("PreSetUp:d.httpClient.Get(%s,%d) error(%v)", d.preSetUpURL, id, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.preSetUpURL+"?"+params.Encode())
	}
	return
}

// PreItemSetUp .
func (d *Dao) PreItemSetUp(c context.Context, id, pid int64, state int) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("id", strconv.FormatInt(id, 10))
	params.Set("pid", strconv.FormatInt(pid, 10))
	params.Set("state", strconv.Itoa(state))
	if err = d.httpClient.Get(c, d.preItemSetUpURL, "", params, &res); err != nil {
		log.Error("PreItemSetUp:d.httpClient.Get(%s,%d) error(%v)", d.preItemSetUpURL, id, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.preItemSetUpURL+"?"+params.Encode())
	}
	return
}
