package service

import (
	"encoding/json"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"
)

func (s *Service) statCoinproc() {
	defer s.waiter.Done()
	if s.coinSub == nil {
		return
	}
	for {
		msg, ok := <-s.coinSub.Messages()
		if !ok {
			log.Info("statCoinproc databus exit!")
			return
		}
		msg.Commit()
		m := new(like.StatCoinMsg)
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("statCoinproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		switch m.Type {
		case _typeArc:
		}
		log.Info("statCoinproc key:%s partition:%d offset:%d value:%s", msg.Key, msg.Partition, msg.Offset, string(msg.Value))
	}
}
