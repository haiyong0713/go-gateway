package service

import (
	"encoding/json"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"
)

func (s *Service) liveFollowproc() {
	defer s.waiter.Done()
	if s.liveFollowSub == nil {
		return
	}
	for {
		msg, ok := <-s.liveFollowSub.Messages()
		if !ok {
			log.Info("liveFollowproc databus exit!")
			return
		}
		msg.Commit()
		m := new(like.LiveMsg)
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("liveFollowproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		log.Info("liveFollowproc key:%s partition:%d offset:%d value:%b", msg.Key, msg.Partition, msg.Offset, msg.Value)
	}
}
