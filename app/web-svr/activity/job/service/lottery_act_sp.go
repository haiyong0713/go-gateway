package service

import (
	"context"
	"encoding/json"
	"go-common/library/log"

	lmdl "go-gateway/app/web-svr/activity/job/model/like"

	videoupapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

const (
	counterVideo    = "videoup"
	counterVideoCut = "videoup_bcut"
)

func (s *Service) actPlatHistoryMsgSpToCh() {
	defer s.waiter.Done()
	c := context.Background()
	if s.actPlatHistorySpSub == nil {
		return
	}
	for {
		msg, ok := <-s.actPlatHistorySpSub.Messages()
		if !ok {
			s.lotteryConsumewaiter.Done()
			log.Info("actPlatHistoryMsgSpToCh databus:actPlatHistorySub ActPlatHistory-T exit!")
			return
		}
		msg.Commit()
		m := &lmdl.ActPlatHistoryVideoMsg{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("actPlatHistoryMsgSpToCh json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		log.Info("actPlatHistoryMsgSpToCh databus:actPlatHistorySub ActPlatHistory mid(%d)", m.MID)

		if m.MID <= 0 || m.Diff <= 0 {
			continue
		}
		if m.Counter != counterVideo || m.Activity != s.c.SpringFestival2021.Activity {
			continue
		}
		newM := &lmdl.ActPlatHistoryMsg{}
		// 验证是否是必剪
		if len(m.Raw) > 0 {
			for _, v := range m.Raw {
				if v.New != nil {
					arg := &videoupapi.ArcMaterialsReq{
						AID: v.New.Aid,
						MTp: -1,
					}
					resRly, err := s.videoupClient.ArcMaterials(c, arg)
					if err != nil {
						log.Errorc(c, "bcutCount s.videoupClient.ArcMaterials aid(%d) error(%+v)", arg.AID, err)
						continue
					}
					if isBcut(resRly.UpFrom) {
						newM.Diff++
					}
				}

			}
		}
		if newM.Diff == 0 {
			continue
		}
		newM.Counter = counterVideoCut
		newM.Activity = s.c.SpringFestival2021.Activity
		newM.MID = m.MID
		newM.TimeStamp = m.TimeStamp
		s.actPlatHistoryCh <- newM
		log.Info("actPlatHistoryMsgSpToCh success mid(%d) diff(%d) counter(%s) activity(%s) ", newM.MID, newM.Diff, newM.Counter, newM.Activity)
	}
}
