package steins

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go-common/library/ecode"

	"github.com/pkg/errors"
)

const (
	_wechatAction = "NotifyCreate"
	_wechatType   = "wechat_message"
)

// SendWechat  send wechat work message.
func (d *Dao) SendWechat(c context.Context, title, msg, user string) (err error) {
	var msgBytes []byte
	params := map[string]interface{}{
		"Action":    _wechatAction,
		"SendType":  _wechatType,
		"PublicKey": d.c.Wechat.Wxkey,
		"UserName":  user,
		"Content": map[string]string{
			"subject": title,
			"body":    title + "\n" + msg,
		},
		"TreeId":    "",
		"Signature": "1",
		"Severity":  "P5",
	}
	if msgBytes, err = json.Marshal(params); err != nil {
		return
	}
	var req *http.Request
	if req, err = http.NewRequest(http.MethodPost, d.c.Host.Merak, strings.NewReader(string(msgBytes))); err != nil {
		return
	}
	req.Header.Add("content-type", "application/json; charset=UTF-8")
	res := &struct {
		RetCode int `json:"RetCode"`
	}{}
	if err = d.httpWechatClient.Do(c, req, &res); err != nil {
		err = errors.Wrapf(err, "SendWechat d.client.Do(title:%s,msg:%s,user:%s)", title, msg, user)
		return
	}
	if res.RetCode != 0 {
		err = errors.Wrapf(ecode.Int(res.RetCode), "SendWechat d.client.Do(title:%s,msg:%s,user:%s)", title, msg, user)
		return
	}
	return

}
