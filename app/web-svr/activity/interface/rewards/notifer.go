package rewards

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/tool"
	"net/http"
	"strings"
	"time"
)

const (
	_success         = 0
	_whiteCheckError = 29001
	_writeDBError    = 29002
)

type msgKey struct {
	MsgKey uint64 `json:"msg_key"`
}

type LetterParam struct {
	RecverIDs  []uint64 `json:"recver_ids"`       //多人消息，列表型，限定每次客户端发送<=100
	SenderUID  uint64   `json:"sender_uid"`       //官号uid：发送方uid
	MsgKey     uint64   `json:"msg_key"`          //消息唯一标识
	MsgType    int32    `json:"msg_type"`         //文本类型 type = 1
	Content    string   `json:"content"`          //{"content":"test" //文本内容}
	NotifyCode string   `json:"notify_code"`      //通知码
	Params     string   `json:"params,omitempty"` //逗号分隔，通知卡片内容的可配置参数
	JumpUri    string   `json:"jump_uri"`         //通知卡片跳转链接
	Title      string   `json:"title"`
	Text       string   `json:"text"`
	JumpText   string   `json:"jump_text"`
}

// https://info.bilibili.co/pages/viewpage.action?pageId=23122720
// getMsgKey: 获取通知卡片message key
func (s *service) getMsgKey(c context.Context, senderId uint64, notifyCode, jumpUri1, jumpUri2 string, params []string) (res *msgKey, code int, err error) {
	getMsgKeyURL := s.c.Host.Dynamic + "/biz_msg_svr/v0/biz_msg_svr/get_msg_key" //业务方请求获取msg_key
	requestParams := make(map[string]interface{})
	requestParams["sender_uid"] = senderId //发送方uid
	requestParams["msg_type"] = 10         //消息类型：文本类型 type = 1，通知卡片10
	requestParams["notify_code"] = notifyCode
	if jumpUri1 != "" {
		requestParams["jump_uri"] = jumpUri1
	}
	if jumpUri2 != "" {
		requestParams["jump_uri_2"] = jumpUri2
	}
	requestParams["params"] = strings.Join(params, "`||")
	bytesData, err := json.Marshal(requestParams)
	if err != nil {
		log.Error("getMsgKey json.Marshal getMsgKeyURL(%s) requestParams(%+v) error(%v)", getMsgKeyURL, requestParams, err)
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
	if err = s.httpClient.Do(c, req, &resp); err != nil {
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

// sendNotifyCard: 发送通知卡片
func (s *service) sendNotifyCard(c context.Context, senderId, mid int64, notifyCode, jumpUri1, jumpUri2 string, params []string) (err error) {
	var (
		msgKeyResp *msgKey
		msgKeycode int
	)
	for i := 1; i < 3; i++ {
		time.Sleep(time.Duration(i*10) * time.Millisecond)
		msgKeyResp, msgKeycode, err = s.getMsgKey(c, uint64(senderId), notifyCode, jumpUri1, jumpUri2, params)
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
		default:
			log.Error("getMsgKey errorcode(%v)", msgKeycode)
			continue
		}
	}
	if msgKeyResp == nil || msgKeyResp.MsgKey == 0 {
		log.Error("getMsgKey error msg key(%+v) error: %v", msgKeyResp, err)
		return
	}
	log.Info("getMsgKey get msg_key(%d) success", msgKeyResp.MsgKey)
	sendMsgURL := s.c.Host.Dynamic + "/biz_msg_svr/v0/biz_msg_svr/send_msg" //业务方发送私信功能
	requestParams := make(map[string]interface{})
	requestParams["sender_uid"] = senderId       //发送方uid
	requestParams["recver_ids"] = []int64{mid}   //多人消息，列表型，限定每次客户端发送100个
	requestParams["msg_key"] = msgKeyResp.MsgKey //消息唯一标识
	if jumpUri1 != "" {
		requestParams["jump_uri"] = jumpUri1
	}
	if jumpUri2 != "" {
		requestParams["jump_uri_2"] = jumpUri2
	}

	bytesData, err := json.Marshal(requestParams)
	if err != nil {
		log.Errorc(c, "SendLetter json.Marshal sendMsgURL(%s) requestParams(%+v) error(%v)", sendMsgURL, requestParams, err)
		return
	}
	paramStr := string(bytesData)
	var (
		req  *http.Request
		resp = struct {
			Code int `json:"code"`
			Data struct {
				Msgs []struct {
					Uid     int64  `json:"uid"`
					Code    int64  `json:"err_code"`
					Message string `json:"err_msg"`
				} `json:"failed_msgs"`
			} `json:"data"`
		}{}
	)
	if req, err = http.NewRequest("POST", sendMsgURL, strings.NewReader(paramStr)); err != nil {
		log.Errorc(c, "SendLetter http.NewRequest url(%s) error(%v)", sendMsgURL+"?"+paramStr, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = s.httpClient.Do(c, req, &resp); err != nil {
		log.Errorc(c, "SendLetter d.httpClient.Post sendMsgURL(%s) error(%v)", sendMsgURL+"?"+paramStr, err)
		return
	}
	if resp.Code != 0 || len(resp.Data.Msgs) != 0 {
		err = fmt.Errorf("SendLetter sendMsgURL(%s) error code(%v), Msgs: %+v", sendMsgURL+"?"+paramStr, resp.Code, resp.Data.Msgs)
		log.Errorc(c, "%v", err)
		return
	}
	log.Infoc(c, "SendLetter success by sendMsgURL(%s)", sendMsgURL+"?"+paramStr)
	return
}

func (s *service) sendAwardNotifyCard(mid int64, ac *api.RewardsAwardInfo, jumpUri1, jumpUri2 string) {
	if ac.NotifySenderId == 0 || !ac.ShouldSendNotify {
		return
	}
	go func() {
		ctx := context.Background()
		var notifyErr error
		for i := 0; i <= 3; i++ {
			notifyErr = s.sendNotifyCard(ctx, ac.NotifySenderId, mid, ac.NotifyCode, jumpUri1, jumpUri2, []string{ac.ActivityName, ac.Name})
			if notifyErr == nil {
				break
			}
			log.Errorc(ctx, "sendNotifyLetter error and waiting next retry: %v", notifyErr)
			time.Sleep(time.Duration(i*10) * time.Millisecond)
		}
		if notifyErr != nil {
			tool.Metric4RewardFail.WithLabelValues([]string{ac.Type, "notify"}...).Inc()
			log.Errorc(ctx, "SendAwardById sendNotifyLetter error after 4 retry: %v", notifyErr)
		}
	}()
}
