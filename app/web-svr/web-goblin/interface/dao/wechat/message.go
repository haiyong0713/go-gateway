package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
)

const (
	_sendMsgURI = "/cgi-bin/message/custom/send"
)

// SendMessage send message to wechat.
func (d *Dao) SendMessage(c context.Context, accessToken string, arg []byte) (err error) {
	var (
		req     *http.Request
		bs      []byte
		jsonErr error
	)
	params := url.Values{}
	params.Set("access_token", accessToken)
	if req, err = http.NewRequest(http.MethodPost, d.wxSendMsgURL+"?"+params.Encode(), bytes.NewReader(arg)); err != nil {
		log.Error("SendMessage http.NewRequest error(%v)", err)
		return
	}
	if bs, err = d.httpClient.Raw(c, req); err != nil {
		log.Error("SendMessage d.httpClient.Do error(%v)", err)
		return
	}
	var res struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	if jsonErr = json.Unmarshal(bs, &res); jsonErr == nil && res.Errcode != ecode.OK.Code() {
		log.Error("SendMessage errcode error(%d) msg(%s)", res.Errcode, res.Errmsg)
		err = ecode.RequestErr
	}
	return
}
