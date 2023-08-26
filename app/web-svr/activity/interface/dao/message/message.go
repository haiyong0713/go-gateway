package message

import (
	"context"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"

	"github.com/pkg/errors"
)

// SendSysMsg 发送系统消息
func (d *dao) SendSysMsg(c context.Context, uids []int64, mc, title string, context string, ip string) (err error) {
	params := url.Values{}
	params.Set("mc", mc)
	params.Set("title", title)
	params.Set("data_type", "4")
	params.Set("context", context)
	params.Set("mid_list", xstr.JoinInts(uids))
	var res struct {
		Code int `json:"code"`
		Data *struct {
			Status int8   `json:"status"`
			Remark string `json:"remark"`
		} `json:"data"`
	}
	if err = d.client.Post(c, d.msgURL, ip, params, &res); err != nil {
		log.Errorc(c, "SendSysMsg d.client.Post(%s) error(%+v)", d.msgURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrapf(ecode.Int(res.Code), "SendSysMsg dao.client.Post(%s,%d)", d.msgURL+"?"+params.Encode(), res.Code)
		return
	}
	log.Info("send msg ok, resdata=%+v", res.Data)
	return
}
