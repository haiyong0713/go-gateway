package service

import (
	"encoding/json"

	"go-common/library/log"
	lmdl "go-gateway/app/web-svr/activity/job/model/like"
)

const (
	_activityReply = "activity"
)

func (s *Service) replyproc() {
	defer s.waiter.Done()
	if s.replySub == nil {
		return
	}
	var (
		err error
	)
	for {
		msg, ok := <-s.replySub.Messages()
		if !ok {
			log.Info("databus: activity-job binlog replyproc exit!")
			return
		}
		msg.Commit()
		m := &lmdl.Reply{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("replyproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		switch m.Type {
		case _activityReply:

		}
		log.Info("replyproc  success key:%s partition:%d offset:%d value:%s ", msg.Key, msg.Partition, msg.Offset, msg.Value)
	}
}
