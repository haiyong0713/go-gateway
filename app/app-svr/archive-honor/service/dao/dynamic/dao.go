package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/archive-honor/service/api"
	"go-gateway/app/app-svr/archive-honor/service/conf"

	"github.com/pkg/errors"
)

// Dao is dynamic dao
type Dao struct {
	c          *conf.Config
	client     *bm.Client
	msgKeyURL  string
	sendMsgURL string
}

// New new a Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		client:     bm.NewClient(c.HttpClient),
		msgKeyURL:  c.Host.Dynamic + "/biz_msg_svr/v0/biz_msg_svr/get_msg_key", //业务方请求获取msg_key
		sendMsgURL: c.Host.Dynamic + "/biz_msg_svr/v0/biz_msg_svr/send_msg",    //业务方请求发送私信
	}
	return
}

// GetMsgKey get message key
func (d *Dao) GetMsgKey(c context.Context, arcTitle, upName string) (uint64, error) {
	params := make(map[string]interface{})
	params["sender_uid"] = api.HotSenderUID
	params["msg_type"] = api.HotMsgTp
	params["notify_code"] = api.HotNotifyCode
	params["params"] = fmt.Sprintf("%s`||%s", upName, arcTitle)
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return 0, err
	}
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data *struct {
			MsgKey uint64 `json:"msg_key"`
		} `json:"data"`
	}
	req, err := http.NewRequest("POST", d.msgKeyURL, strings.NewReader(string(paramsStr)))
	if err != nil {
		return 0, errors.Wrapf(err, fmt.Sprintf("NewRequest err (%s)", d.msgKeyURL+"?"+string(paramsStr)))
	}
	req.Header.Set("Content-Type", "application/json")
	if err := d.client.Do(c, req, &resp); err != nil {
		return 0, errors.Wrapf(err, fmt.Sprintf("d.client.Do err (%s)", d.msgKeyURL+"?"+string(paramsStr)))
	}
	if resp.Code != 0 {
		return 0, errors.Wrap(ecode.Int(resp.Code), d.msgKeyURL+"?"+string(paramsStr))
	}
	if resp.Data == nil {
		return 0, errors.New("GetMsgKey resp no data")
	}
	return resp.Data.MsgKey, nil
}

// SendMsg send message
func (d *Dao) SendMsg(c context.Context, upid, msgKey uint64) error {
	params := make(map[string]interface{})
	params["sender_uid"] = api.HotSenderUID //发送方uid
	params["recver_ids"] = []uint64{upid}   //多人消息，列表型，限定每次客户端发送100个
	params["msg_key"] = msgKey              //消息唯一标识
	paramsStr, err := json.Marshal(params)
	if err != nil {
		return err
	}
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	req, err := http.NewRequest("POST", d.sendMsgURL, strings.NewReader(string(paramsStr)))
	if err != nil {
		return errors.Wrapf(err, fmt.Sprintf("NewRequest err (%s)", d.sendMsgURL+"?"+string(paramsStr)))
	}
	req.Header.Set("Content-Type", "application/json")
	if err = d.client.Do(c, req, &resp); err != nil {
		return errors.Wrapf(err, fmt.Sprintf("d.client.Do err (%s)", d.sendMsgURL+"?"+string(paramsStr)))
	}
	if resp.Code != 0 {
		return errors.New("GetMsgKey resp no data")
	}
	return nil
}
