package act

import (
	"context"
	"time"

	"go-common/library/log"
	actmdl "go-gateway/app/web-svr/activity/interface/model/actplat"
)

// ClickTask 点击任务
func (s *Service) ClickTask(ctx context.Context, mid int64, activity, business string) (err error) {
	// 给邀请人完成任务
	timeStamp := time.Now().Unix()
	return s.actSend(ctx, mid, mid, activity, business, timeStamp)
}

func (s *Service) SupplymentActSend(ctx context.Context, mid int64, source int64, activity, business string, timeStamp int64) (err error) {
	return s.actSend(ctx, mid, source, activity, business, timeStamp)
}

// actSend
func (s *Service) actSend(ctx context.Context, mid int64, source int64, activity, business string, timeStamp int64) (err error) {
	data := &actmdl.ActivityPoints{
		Timestamp: timeStamp,
		Mid:       mid,
		Source:    source,
		Activity:  activity,
		Business:  business,
	}
	err = s.actDao.Send(ctx, mid, data)
	if err != nil {
		log.Errorc(ctx, "s.actDao.Send data(%+v) err(%v)", data, err)
	}
	return err
}
