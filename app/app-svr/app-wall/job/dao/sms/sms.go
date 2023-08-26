package sms

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-wall/job/conf"
)

const (
	_smsSendURL = "/x/internal/sms/send"
)

type Dao struct {
	c          *conf.Config
	client     *httpx.Client
	smsSendURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		client:     httpx.NewClient(c.HTTPSMS),
		smsSendURL: c.Host.APICo + _smsSendURL,
	}
	return
}

// SendSMS
func (d *Dao) SendSMS(c context.Context, phone int, smsCode, dataJSON string) (err error) {
	params := url.Values{}
	params.Set("mobile", strconv.Itoa(phone))
	params.Set("country", "86")
	params.Set("tcode", smsCode)
	if dataJSON != "" {
		params.Set("tparam", dataJSON)
	}
	var res struct {
		Code int `json:"code"`
	}
	if err = d.client.Post(c, d.smsSendURL, "", params, &res); err != nil {
		log.Error("SendSMS hots url(%s) error(%v)", d.smsSendURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Error("SendSMS hots url(%s) error(%v)", d.smsSendURL+"?"+params.Encode(), res.Code)
		err = fmt.Errorf("SendSMS api response code(%v)", res)
		return
	}
	return
}
