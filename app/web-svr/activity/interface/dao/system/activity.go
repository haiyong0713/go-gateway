package system

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/system"
)

func (d *Dao) GetActivityInfo(ctx context.Context, aid int64) (res *system.Activity, err error) {
	res, err = d.GetActivityInfoFromDB(ctx, aid)
	return
}

func (d *Dao) SendMessage(ctx context.Context, token string, uid string, message string) (err error) {
	res := new(struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	})
	var resp string
	params := system.SendMessage{
		Touser:  uid,
		MsgType: "text",
		AgentID: "1000186",
		Text: system.SendMessageContent{
			Content: message,
		},
		EnableDuplicateCheck:   "1",
		DuplicateCheckInterval: "10",
	}
	reqBody, _ := json.Marshal(params)
	if resp, err = d.HTTPPost(ctx, d.c.System.NotificationUrl+"?access_token="+token, string(reqBody), map[string]string{}); err != nil {
		err = fmt.Errorf("SendMessage HTTPGet Params:%v Resp:%v Err:%v", params, resp, err)
		log.Errorc(ctx, err.Error())
		return
	}
	if err = json.Unmarshal([]byte(resp), res); err != nil {
		err = fmt.Errorf("SendMessage json.Unmarshal Resp:%v Err:%v", resp, err)
		log.Errorc(ctx, err.Error())
		return
	}
	if res.Errcode != 0 {
		err = fmt.Errorf("SendMessage Response Err Res:%v", res)
		log.Errorc(ctx, err.Error())
		return
	}
	return
}
