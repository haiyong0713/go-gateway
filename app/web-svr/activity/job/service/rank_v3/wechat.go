package rank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"net/http"
	"net/url"
)

// sendWechat 发送微信
func (s *Service) sendWechat(c context.Context, title, message, user, token string) (err error) {

	var (
		req *http.Request
		res struct {
			ErrCode      int64  `json:"errcode"`
			ErrMsg       string `json:"errmsg"`
			InvalidUser  string `json:"invaliduser"`
			InvalidParty string `json:"invalidparty"`
			InvalidTag   string `json:"invalidtag"`
		}
	)
	params := url.Values{}
	params.Set("key", token)
	_url := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?" + params.Encode()
	var buf []byte
	buf, _ = json.Marshal(struct {
		ToUser  []string          `json:"mentioned_list"`
		MsgType string            `json:"msgtype"`
		Text    map[string]string `json:"markdown"`
	}{
		ToUser:  []string{"@all"},
		MsgType: "markdown",
		Text:    map[string]string{"content": fmt.Sprintf("%s\n>%s", title, message)},
	})

	body := bytes.NewBuffer(buf)
	if req, err = http.NewRequest("POST", _url, body); err != nil {
		log.Errorc(c, "sendMessageToUser url(%s) error(%v)", _url, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = s.client.Do(c, req, &res); err != nil {
		log.Errorc(c, "sendMessageToUser Do failed url(%s) response(%+v) error(%v)", _url, res, err)
		return
	}
	log.Infoc(c, "sendMessageToUser res %v", res)
	return
}
