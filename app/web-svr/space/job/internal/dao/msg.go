package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/space/job/internal/model"

	"github.com/pkg/errors"
)

const (
	_success         = 0
	_whiteCheckError = 29001
	_writeDBError    = 29002
	_msgKeyURI       = "/biz_msg_svr/v0/biz_msg_svr/get_msg_key"
	_sendMsgURI      = "/biz_msg_svr/v0/biz_msg_svr/send_msg"
)

func (d *dao) getMsgKey(ctx context.Context, l *model.LetterParam) (res int64, code int, err error) {
	getMsgKeyURL := d.msgKeyURL //业务方请求获取msg_key
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
	if l.Title != "" && l.Text != "" {
		p, _ := json.Marshal(map[string]interface{}{"title": l.Title, "text": l.Text})
		params["params"] = string(p)
	}
	bytesData, err := json.Marshal(params)
	if err != nil {
		log.Error("getMsgKey json.Marshal getMsgKeyURL(%s) params(%+v) error(%v)", getMsgKeyURL, params, err)
		return 0, 0, err
	}
	paramStr := string(bytesData)
	var (
		req  *http.Request
		resp = struct {
			Code int          `json:"code"`
			Msg  string       `json:"msg"`
			Data model.MsgKey `json:"data"`
		}{}
	)
	if req, err = http.NewRequest("POST", getMsgKeyURL, strings.NewReader(paramStr)); err != nil {
		log.Error("getMsgKey http.NewRequest url(%s) error(%v)", getMsgKeyURL+"?"+paramStr, err)
		return 0, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.httpClient.Do(ctx, req, &resp); err != nil {
		log.Error("getMsgKey d.httpClient.Post getMsgKeyURL(%s) error(%v)", getMsgKeyURL+"?"+paramStr, err)
		return 0, 0, err
	}
	if resp.Code != 0 {
		log.Error("getMsgKey getMsgKeyURL(%s) error code(%v)", getMsgKeyURL+"?"+paramStr, resp.Code)
		return 0, 0, err
	}
	log.Info("getMsgKey success by getMsgKeyURL(%s)", getMsgKeyURL+"?"+paramStr)
	return resp.Data.MsgKey, resp.Code, nil
}

func (d *dao) SendLetter(ctx context.Context, arg *model.LetterParam) error {
	var (
		msgKey     int64
		msgKeycode int
		err        error
	)
	for i := 1; i < 3; i++ {
		msgKey, msgKeycode, err = d.getMsgKey(context.Background(), arg)
		if err != nil { //报错不重试，直接返回，程序不往下执行
			return err
		}
		switch msgKeycode {
		case _success: //获取成功，直接跳出循环，程序不往下执行
			break
		case _whiteCheckError: //白名单校验失败，直接返回，程序继续往下执行
			return fmt.Errorf("getMsgKey code(%d) msg(白名单校验失败) error(%v)", msgKeycode, err)
		case _writeDBError:
			log.Error("SendLetter getMsgKey code(%d) msg(写入数据库错误) error(%v) retry 3 times", msgKeycode, err)
			continue
		}
	}
	if msgKey <= 0 {
		return fmt.Errorf("getMsgKey error msg key(%+v) error", msgKey)
	}
	log.Info("SendLetter getMsgKey get msg_key(%d) success", msgKey)
	sendMsgURL := d.sendMsgURL //业务方发送私信功能
	params := make(map[string]interface{})
	params["sender_uid"] = arg.SenderUID //发送方uid
	params["recver_ids"] = arg.RecverIDs //多人消息，列表型，限定每次客户端发送100个
	params["msg_key"] = msgKey           //消息唯一标识
	bytesData, err := json.Marshal(params)
	if err != nil {
		log.Error("SendLetter json.Marshal sendMsgURL(%s) params(%+v) error(%v)", sendMsgURL, params, err)
		return err
	}
	paramStr := string(bytesData)
	var (
		req  *http.Request
		resp = struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
		}{}
	)
	if req, err = http.NewRequest("POST", sendMsgURL, strings.NewReader(paramStr)); err != nil {
		log.Error("SendLetter http.NewRequest url(%s) error(%v)", sendMsgURL+"?"+paramStr, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.httpClient.Do(ctx, req, &resp); err != nil {
		log.Error("SendLetter d.httpClient.Post sendMsgURL(%s) error(%v)", sendMsgURL+"?"+paramStr, err)
		return err
	}
	if resp.Code != 0 {
		return errors.Wrapf(ecode.Int(resp.Code), "SendLetter sendMsgURL(%s) error code(%v)", sendMsgURL+"?"+paramStr, resp.Code)
	}
	log.Info("SendLetter success by url:%s", sendMsgURL+"?"+paramStr)
	return nil
}
