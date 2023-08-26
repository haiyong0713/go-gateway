package rank

import (
	"context"
	"encoding/json"
	"go-common/library/ecode"
	"go-common/library/log"
	"net/http"
	"strings"
)

// SendWeChat ...
func (d *dao) SendWeChat(c context.Context, publicKey, title, msg, user string) (err error) {
	var msgBytes []byte
	params := map[string]interface{}{
		"Action":    "NotifyCreate",
		"SendType":  "wechat_message",
		"PublicKey": publicKey,
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
	if req, err = http.NewRequest(http.MethodPost, d.c.Host.MerakCo, strings.NewReader(string(msgBytes))); err != nil {
		return
	}
	req.Header.Add("content-type", "application/json; charset=UTF-8")
	res := &struct {
		RetCode int `json:"RetCode"`
	}{}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Errorc(c, "SendWechat d.client.Do(title:%s,msg:%s,user:%s) error(%v)", title, msg, user, err)
		return
	}
	if res.RetCode != 0 {
		err = ecode.Int(res.RetCode)
		log.Errorc(c, "SendWechat d.client.Do(title:%s,msg:%s,user:%s) error(%v)", title, msg, user, err)
		return
	}
	return
}
