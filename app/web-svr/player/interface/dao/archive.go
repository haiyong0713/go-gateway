package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	"github.com/pkg/errors"
)

const (
	_pcdnLoaderURI = "/pcdnd/loader"
)

// PvData get binary data from pvdata url
func (d *Dao) PvData(c context.Context, pvURL string) (res []byte, err error) {
	var (
		req    *http.Request
		resp   *http.Response
		cancel func()
	)
	if req, err = http.NewRequest("GET", pvURL, nil); err != nil {
		err = errors.Wrapf(err, "PvData http.NewRequest(%s)", pvURL)
		return
	}
	c, cancel = context.WithTimeout(c, time.Duration(d.c.Rule.VsTimeout))
	defer cancel()
	req = req.WithContext(c)
	if resp, err = d.vsClient.Do(req); err != nil {
		err = errors.Wrapf(err, "httpClient.Do(%s)", pvURL)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf("PvData url(%s) resp.StatusCode(%v)", pvURL, resp.StatusCode)
		return
	}
	res, err = ioutil.ReadAll(resp.Body)
	return
}

// PcdnLoader get pcdn data.
func (d *Dao) PcdnLoader(c context.Context, cid, mid int64, refer, innerSign string) (data json.RawMessage, err error) {
	params := url.Values{}
	params.Set("cid", strconv.FormatInt(cid, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	var req *http.Request
	req, err = d.client.NewRequest(http.MethodGet, d.pcdnLoaderURL, metadata.String(c, metadata.RemoteIP), params)
	if err != nil {
		return
	}
	req.Header.Set("Referer", refer)
	req.Header.Set("Cookie", "innersign="+innerSign)
	var res struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.pcdnLoaderURL+"?"+params.Encode())
		return
	}
	data = res.Data
	return
}
