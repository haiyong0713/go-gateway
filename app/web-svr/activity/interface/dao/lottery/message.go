package lottery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"go-common/library/log"
	lott "go-gateway/app/web-svr/activity/interface/model/lottery"
)

type msgKey struct {
	MsgKey uint64 `json:"msg_key"`
}

const (
	_success         = 0
	_whiteCheckError = 29001
	_writeDBError    = 29002
)

func (d *Dao) getMsgKey(c context.Context, l *lott.LetterParam) (res *msgKey, code int, err error) {
	getMsgKeyURL := d.c.Host.Dynamic + "/biz_msg_svr/v0/biz_msg_svr/get_msg_key" //业务方请求获取msg_key
	params := make(map[string]interface{})
	params["sender_uid"] = l.SenderUID //发送方uid
	params["msg_type"] = l.MsgType     //消息类型：文本类型 type = 1，通知卡片10
	if l.Params != "" {
		params["params"] = l.Params //通知卡片内容的可配置参数,分隔符是`||
	}
	if l.Content != "" {
		ctn, _ := json.Marshal(map[string]interface{}{"content": l.Content}) //消息实体内容
		params["content"] = string(ctn)
	}
	if l.NotifyCode != "" {
		params["notify_code"] = l.NotifyCode
	}
	if l.JumpURL != "" {
		params["jump_text"] = l.JumpText
		params["jump_uri"] = l.JumpURL
	}
	if l.JumpURL2 != "" {
		params["jump_text_2"] = l.JumpText2
		params["jump_uri_2"] = l.JumpURL2
	}
	bytesData, err := json.Marshal(params)
	if err != nil {
		log.Error("getMsgKey json.Marshal getMsgKeyURL(%s) params(%+v) error(%v)", getMsgKeyURL, params, err)
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
	if req, err = http.NewRequest("POST", getMsgKeyURL, strings.NewReader(paramStr)); err != nil {
		log.Error("getMsgKey http.NewRequest url(%s) error(%v)", getMsgKeyURL+"?"+paramStr, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.client.Do(c, req, &resp); err != nil {
		log.Error("getMsgKey d.httpClient.Post getMsgKeyURL(%s) error(%v)", getMsgKeyURL+"?"+paramStr, err)
		return
	}
	if resp.Code != 0 {
		log.Error("getMsgKey getMsgKeyURL(%s) error code(%v)", getMsgKeyURL+"?"+paramStr, resp.Code)
		return
	}
	log.Info("getMsgKey success by getMsgKeyURL(%s)", getMsgKeyURL+"?"+paramStr)
	res = resp.Data
	code = resp.Code
	return
}

// SendLetter send private letter notify user
func (d *Dao) SendLetter(c context.Context, l *lott.LetterParam) (code int, err error) {
	var (
		msgKeyResp *msgKey
		msgKeycode int
	)
	for i := 1; i < 3; i++ {
		msgKeyResp, msgKeycode, err = d.getMsgKey(context.Background(), l)
		if err != nil { //报错不重试，直接返回，程序不往下执行
			return
		}
		switch msgKeycode {
		case _success: //获取成功，直接跳出循环，程序不往下执行
			break
		case _whiteCheckError: //白名单校验失败，直接返回，程序继续往下执行
			err = errors.New("白名单校验失败")
			log.Error("getMsgKey code(%d) msg(白名单校验失败) error(%v)", msgKeycode, err)
			return
		case _writeDBError:
			err = errors.New("写入数据库错误") //会导致获取msgkey失败，建议重试
			log.Error("getMsgKey code(%d) msg(写入数据库错误) error(%v) retry 3 times", msgKeycode, err)
			continue
		}
	}
	if msgKeyResp == nil || msgKeyResp.MsgKey == 0 {
		log.Error("getMsgKey error msg key(%+v) error", msgKeyResp)
		return
	}
	log.Infoc(c, "getMsgKey get msg_key(%d) success", msgKeyResp.MsgKey)
	sendMsgURL := d.c.Host.Dynamic + "/biz_msg_svr/v0/biz_msg_svr/send_msg" //业务方发送私信功能
	params := make(map[string]interface{})
	params["sender_uid"] = l.SenderUID    //发送方uid
	params["recver_ids"] = l.RecverIDs    //多人消息，列表型，限定每次客户端发送100个
	params["msg_key"] = msgKeyResp.MsgKey //消息唯一标识
	if l.JumpURL != "" {
		params["jump_text"] = l.JumpText
		params["jump_uri"] = l.JumpURL
	}
	if l.JumpURL2 != "" {
		params["jump_text_2"] = l.JumpText2
		params["jump_uri_2"] = l.JumpURL2
	}
	if l.Params != "" {
		params["params"] = l.Params //通知卡片内容的可配置参数,分隔符是`||
	}
	bytesData, err := json.Marshal(params)
	if err != nil {
		log.Errorc(c, "SendLetter json.Marshal sendMsgURL(%s) params(%+v) error(%v)", sendMsgURL, params, err)
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
		log.Errorc(c, "SendLetter http.NewRequest url(%s) error(%v)", sendMsgURL+"?"+paramStr, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.client.Do(c, req, &resp); err != nil {
		log.Errorc(c, "SendLetter d.httpClient.Post sendMsgURL(%s) error(%v)", sendMsgURL+"?"+paramStr, err)
		return
	}
	if resp.Code != 0 {
		log.Errorc(c, "SendLetter sendMsgURL(%s) error code(%v)", sendMsgURL+"?"+paramStr, resp.Code)
		return
	}

	log.Infoc(c, "SendLetter success by sendMsgURL(%s)", sendMsgURL+"?"+paramStr)
	code = resp.Code
	return
}
