package dao

import (
	"context"
	"time"

	"go-gateway/app/app-svr/kvo/interface/model"

	"go-common/library/log"
)

// SendPlayerAction send action to job.
func (d *Dao) SendTaskAction(c context.Context, k string, act *model.Action) (err error) {
	for i := 0; i < 3; i++ {
		if err = d.taskPub.Send(c, k, act); err != nil {
			log.Error("taskPub.Send(action:%s,data:%s) error(%v)", act.Action, act.Data, err)
			time.Sleep(200 * time.Millisecond)
		} else {
			log.Info("taskPub.Send(action:%s,data:%s) success", act.Action, act.Data)
			break
		}
	}
	return
}

// SendPlayerAction send action to job.
func (d *Dao) SendBuvidTaskAction(c context.Context, k string, act *model.Action) (err error) {
	for i := 0; i < 3; i++ {
		if err = d.buvidTaskPub.Send(c, k, act); err != nil {
			log.Error("buvidTaskPub.Send(action:%s,data:%s) error(%v)", act.Action, act.Data, err)
			time.Sleep(200 * time.Millisecond)
		} else {
			log.Info("buvidTaskPub.Send(action:%s,data:%s) success", act.Action, act.Data)
			break
		}
	}
	return
}
