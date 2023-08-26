package service

import (
	"context"
	"encoding/json"

	"go-gateway/app/app-svr/archive/service/api"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"
)

const _thumbUp = 1

func (s *Service) statThumbupproc() {
	defer s.waiter.Done()
	c := context.Background()
	if s.thumbupSub == nil {
		return
	}
	for {
		msg, ok := <-s.thumbupSub.Messages()
		if !ok {
			log.Info("statThumbupproc databus exit!")
			return
		}
		msg.Commit()
		m := new(like.StatLikeMsg)
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Info("statThumbupproc databus exit!")
			continue
		}
		if m.Type == _typeArc && m.Action == _thumbUp {
			if _, ok := s.restartArc[m.ID]; ok {
				s.singleDoTask(c, m.Mid, s.c.Restart2020.LikeTaskID)
			}
			//if _, ok := s.yelAndGreenArc[m.ID]; ok {
			//	s.singleDoTask(c, m.Mid, s.c.YelAndGreen.LikeTaskID)
			//}
			if _, ok := s.mobileGameArc[m.ID]; ok {
				s.singleDoTask(c, m.Mid, s.c.MobileGame.LikeTaskID)
			}
			if _, ok := s.stupidArc[m.ID]; ok {
				s.singleDoTask(c, m.Mid, s.c.Stupid.LikeTaskID)
			}
			if _, ok := s.staffArc[m.ID]; ok {
				reply, err := s.arcClient.Arc(c, &api.ArcRequest{Aid: m.ID})
				if err != nil {
					log.Error("statThumbupproc s.arcClient.Arc(%d) error(%v)", m.ID, err)
					continue
				}
				if reply.Arc.AttrVal(api.AttrBitIsCooperation) == api.AttrYes {
					s.singleDoTask(c, m.UpMid, s.c.Staff.LikeTaskID)
					for _, v := range reply.Arc.StaffInfo {
						if v == nil || v.Mid == 0 || v.Mid == m.UpMid {
							continue
						}
						s.singleDoTask(c, v.Mid, s.c.Staff.LikeTaskID)
					}
				}
			}
		}
		log.Info("statThumbupproc key:%s partition:%d offset:%d value:%s", msg.Key, msg.Partition, msg.Offset, string(msg.Value))
	}
}
