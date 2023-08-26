package dao

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"github.com/pkg/errors"
)

const (
	_registGroup = "/gateway/bind"
	_room        = "esport://%d"
	_broadURL    = "/x/internal/broadcast/push/room"
)

// ThirdGet get.
func (d *Dao) ThirdGet(c context.Context, url string) (res []byte, err error) {
	var (
		req    *http.Request
		resp   *http.Response
		cancel func()
	)
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		err = errors.Wrapf(err, "ThirdGet http.NewRequest(%s)", url)
		return
	}
	c, cancel = context.WithTimeout(c, time.Duration(d.c.Leidata.Timeout))
	defer cancel()
	req = req.WithContext(c)
	if resp, err = d.ldClient.Do(req); err != nil {
		log.Errorc(c, "ThirdGet [%s] error err[%v]", url, err)
		err = errors.Wrapf(err, "ThirdGet d.ldClient.Do(%s)", url)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		log.Errorc(c, "ThirdGet [%s] error err[%v]", url, err)
		err = fmt.Errorf("ThirdGet url(%s) resp.StatusCode(%v)", url, resp.StatusCode)
		return
	}
	res, err = ioutil.ReadAll(resp.Body)
	log.Infoc(c, "ThirdGet [%s] response [%s] err[%v]", url, res, err)
	return
}

// ThirdPost post
func (d *Dao) ThirdPost(c context.Context, params url.Values) (res []byte, err error) {
	var (
		req    *http.Request
		resp   *http.Response
		cancel func()
	)
	paramStr := params.Encode()
	if strings.IndexByte(paramStr, '+') > -1 {
		paramStr = strings.Replace(paramStr, "+", "%20", -1)
	}
	u := d.c.Leidata.GroupURL + _registGroup
	if req, err = http.NewRequest("POST", u, strings.NewReader(paramStr)); err != nil {
		err = errors.Wrapf(err, "ThirdPost http.NewRequest(%s)", u)
		return
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	c, cancel = context.WithTimeout(c, time.Duration(d.c.Leidata.Timeout))
	defer cancel()
	req = req.WithContext(c)
	if resp, err = d.ldClient.Do(req); err != nil {
		err = errors.Wrapf(err, "ThirdPost httpClient.Do(%s)", u)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf("ThirdPost url(%s) data(%s) resp.StatusCode(%v)", u, paramStr, resp.StatusCode)
		return
	}
	if res, err = ioutil.ReadAll(resp.Body); err != nil {
		err = fmt.Errorf("ThirdPost url(%s) data(%s) resp.StatusCode(%v) regist(%v) error(%v)", u, paramStr, resp.StatusCode, string(res), err)
	}
	return
}

// PushRoom  broadcast push room.
func (d *Dao) PushRoom(c context.Context, room int64, opt, msg string) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	if room == 0 {
		log.Info("PushRoom room(0) msg(%v)", msg)
		return
	}
	params := url.Values{}
	params.Set("operation", opt)
	params.Set("room", fmt.Sprintf(_room, room))
	params.Set("message", msg)
	ip := metadata.String(c, metadata.RemoteIP)
	url := d.c.Host.API + _broadURL
	if err = d.http.Post(c, url, ip, params, &res); err != nil {
		log.Error("PushRoom url(%s) error(%v)", url+"?"+params.Encode(), err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
		log.Error("PushRoom url(%s) error code(%v)", url+"?"+params.Encode(), err)
	}
	return
}
