package dao

import (
	"context"
	"strconv"

	smsgrpc "git.bilibili.co/bapis/bapis-go/community/service/sms"
	"go-common/library/log"
)

func (d *Dao) SendSms(c context.Context, mobile int64, tcode, tparam string) error {
	req := &smsgrpc.SendReq{
		Mobile:  strconv.FormatInt(mobile, 10),
		Country: "86",
		Tcode:   tcode,
		Tparam:  tparam,
	}
	if _, err := d.smsClient.Send(c, req); err != nil {
		log.Errorc(c, "Fail to send sms, req=%+v error=%+v", req, err)
		return err
	}
	return nil
}
