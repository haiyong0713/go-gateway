package sms

import (
	"context"
	"strconv"

	smsgrpc "git.bilibili.co/bapis/bapis-go/community/service/sms"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

type Dao struct {
	client smsgrpc.SmsClient
}

func NewDao(cfg *conf.Config) *Dao {
	client, err := smsgrpc.NewClient(cfg.SmsClient)
	if err != nil {
		panic(err)
	}
	return &Dao{client: client}
}

func (d *Dao) SendSms(c context.Context, mobile int64, tcode, tparam string) error {
	req := &smsgrpc.SendReq{
		Mobile:  strconv.FormatInt(mobile, 10),
		Country: "86",
		Tcode:   tcode,
		Tparam:  tparam,
	}
	if _, err := d.client.Send(c, req); err != nil {
		log.Errorc(c, "Fail to send sms, req=%+v error=%+v", req, err)
		return err
	}
	return nil
}
