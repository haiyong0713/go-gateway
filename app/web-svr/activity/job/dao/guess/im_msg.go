package guess

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/guess"
)

const (
	_getMsgKeyPath = "/biz_msg_svr/v0/biz_msg_svr/get_msg_key"
	_sendMsgPath   = "/biz_msg_svr/v0/biz_msg_svr/send_msg"
)

type msgKey struct {
	MsgKey uint64 `json:"msg_key"`
}

func (d *Dao) imMsgkey(c context.Context, l *guess.ImMsgParam) (res *msgKey, code int, err error) {
	msgKeyURL := d.imKeyURL
	params := make(map[string]interface{})
	params["sender_uid"] = l.SenderUID                                   //发送方uid
	params["msg_type"] = l.MsgType                                       //消息类型：文本类型 type = 1
	ctn, _ := json.Marshal(map[string]interface{}{"content": l.Content}) //消息实体内容
	params["content"] = string(ctn)
	bytesData, err := json.Marshal(params)
	if err != nil {
		log.Error("imMsgkey json.Marshal msgKeyURL(%s) params(%+v) error(%v)", msgKeyURL, params, err)
		return
	}
	paramStr := string(bytesData)
	var (
		req  *http.Request
		resp = struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data *msgKey
		}{}
	)
	if req, err = http.NewRequest("POST", msgKeyURL, strings.NewReader(paramStr)); err != nil {
		log.Error("imMsgkey http.NewRequest url(%s) error(%v)", msgKeyURL+"?"+paramStr, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.client.Do(c, req, &resp); err != nil {
		log.Error("imMsgkey d.client.Do msgKeyURL(%s) error(%v)", msgKeyURL+"?"+paramStr, err)
		return
	}
	if resp.Code != 0 {
		log.Error("imMsgkey msgKeyURL(%s) error code(%v)", msgKeyURL+"?"+paramStr, resp.Code)
		return
	}
	log.Info("imMsgkey success by msgKeyURL(%s)", msgKeyURL+"?"+paramStr)
	res = resp.Data
	code = resp.Code
	return
}

// SendImMsg send im message.
func (d *Dao) SendImMsg(c context.Context, l *guess.ImMsgParam) (code int, err error) {
	var (
		msgKeyResp *msgKey
		msgKeycode int
	)
	msgKeyResp, msgKeycode, err = d.imMsgkey(context.Background(), l)
	if err != nil {
		return
	}
	if msgKeycode != 0 {
		return
	}
	if msgKeyResp == nil || msgKeyResp.MsgKey == 0 {
		log.Error("SendImMsg error msg key(%+v) error", msgKeyResp)
		return
	}
	log.Info("SendImMsg get msg_key(%d) success", msgKeyResp.MsgKey)
	sendMsgURL := d.imSendURL
	res := struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}{}

	params := make(map[string]interface{})
	params["sender_uid"] = l.SenderUID    //发送方uid
	params["recver_ids"] = l.RecverIDs    //多人消息，列表型，限定每次客户端发送100个
	params["msg_key"] = msgKeyResp.MsgKey //消息唯一标识
	bytesData, err := json.Marshal(params)
	if err != nil {
		log.Error("SendImMsg json.Marshal sendMsgURL(%s) params(%+v) error(%v)", sendMsgURL, params, err)
		return
	}
	paramStr := string(bytesData)
	var (
		req  *http.Request
		resp = struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data *msgKey
		}{}
	)
	if req, err = http.NewRequest("POST", sendMsgURL, strings.NewReader(paramStr)); err != nil {
		log.Error("SendImMsg http.NewRequest url(%s) error(%v)", sendMsgURL+"?"+paramStr, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.client.Do(c, req, &resp); err != nil {
		log.Error("SendImMsg d.httpClient.Post sendMsgURL(%s) error(%v)", sendMsgURL+"?"+paramStr, err)
		return
	}
	if resp.Code != 0 {
		log.Error("SendImMsg sendMsgURL(%s) error code(%v)", sendMsgURL+"?"+paramStr, resp.Code)
		return
	}
	log.Info("SendImMsg success by sendMsgURL(%s)", sendMsgURL+"?"+paramStr)
	code = res.Code
	return
}
