package college

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/college"
)

// SendPoint ...
func (d *dao) SendPoint(c context.Context, mid int64, data *college.ActPlatActivityPoints) (err error) {
	midStr := fmt.Sprintf("%d", mid)
	if err = d.actPlatPub.Send(c, midStr, data); err != nil {
		log.Errorc(c, "SendPoint: d.actPlatPub.Send(%v) error(%v)", *data, err)
		return
	}
	log.Infoc(c, "SendPoint: d.actPlatPub.Send(%v)", *data)
	return
}
