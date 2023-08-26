package bnj

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	xecode "go-common/library/ecode"
	"go-common/library/log"

	"github.com/google/uuid"
)

const (
	_msgKeyURI          = "/biz_msg_svr/v0/biz_msg_svr/get_msg_key"
	_sendMsgURI         = "/biz_msg_svr/v0/biz_msg_svr/send_msg"
	_privateMsgTypeText = 1
)

func (d *Dao) MessageKey(ctx context.Context, sender int64, content string) (key int64, err error) {
	reqParam := struct {
		SenderUid int64  `json:"sender_uid"`
		MsgType   int64  `json:"msg_type"`
		Content   string `json:"content"`
	}{sender, _privateMsgTypeText, content}
	b, _ := json.Marshal(reqParam)
	req, err := http.NewRequest(http.MethodPost, d.msgKeyURL, bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	res := struct {
		Code int `json:"code"`
		Data struct {
			MsgKey int64 `json:"msg_key"`
		}
	}{}
	if err = d.client.Do(ctx, req, &res); err != nil {
		log.Error("send Private Message step(getMessageKey) error(%v)", err)
		return
	}
	if res.Code != xecode.OK.Code() {
		err = fmt.Errorf("send Private Message step(getMessageKey) params(%s) code(%d)", string(b), res.Code)
		log.Errorc(ctx, "%s", err.Error())
		return
	}
	key = res.Data.MsgKey
	return
}

func (d *Dao) AsyncSendNormalMessage(ctx context.Context, msgKey int64, sender int64, receivers []int64) error {
	reqParam := struct {
		SenderUID int64   `json:"sender_uid"`
		RecverIDs []int64 `json:"recver_ids"`
		MsgKey    int64   `json:"msg_key"`
	}{sender, receivers, msgKey}
	return d.messagePub.Send(ctx, uuid.New().String(), reqParam)
}

func (d *Dao) SendNormalMessage(ctx context.Context, msgKey int64, sender int64, receivers []int64) error {
	reqParam := struct {
		SenderUID int64   `json:"sender_uid"`
		RecverIDs []int64 `json:"recver_ids"`
		MsgKey    int64   `json:"msg_key"`
	}{sender, receivers, msgKey}
	b, _ := json.Marshal(reqParam)
	req, err := http.NewRequest(http.MethodPost, d.normalMsgURL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res := struct {
		Code int `json:"code"`
		Data struct {
			MsgKey string `json:"msg_key"`
		}
	}{}
	if err = d.client.Do(ctx, req, &res); err != nil {
		log.Error("sendMessage step(sendMessage) error(%v)", err)
		return err
	}
	if res.Code != xecode.OK.Code() {
		err = fmt.Errorf("sendMessage step(sendMessage) params(%s) code(%d)", string(b), res.Code)
		log.Errorc(ctx, "%s", err.Error())
		return err
	}
	return nil
}
