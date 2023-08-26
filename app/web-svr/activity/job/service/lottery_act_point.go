package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"strings"

	"go-common/library/railgun"
	l "go-gateway/app/web-svr/activity/job/model/like"
	lmdl "go-gateway/app/web-svr/activity/job/model/like"
)

// LotteryRailgun ...
type LotteryRailgun struct {
	Cfg     *railgun.Config
	Databus *railgun.DatabusV1Config
	Batch   *railgun.BatchConfig
}

func (s *Service) actPointMsgConsume(msg railgun.Message) (*railgun.SingleUnpackMsg, error) {
	historyMsg := &lmdl.ThresholdNotifyMsg{}
	if err := json.Unmarshal(msg.Payload(), historyMsg); err != nil {
		return nil, err
	}
	return &railgun.SingleUnpackMsg{
		Group: historyMsg.MID,
		Item:  historyMsg,
	}, nil
}

func (s *Service) buildLotteryCounter(c context.Context, sid int64, groupID int64) string {
	return fmt.Sprintf("%d_%d", sid, groupID)
}

func (s *Service) buildMessageLotteryCounter(c context.Context, counter string) string {
	list := strings.Split(counter, "_ss_")
	if len(list) != 3 {
		return ""
	}
	return fmt.Sprintf("%s_%s", list[0], list[1])
}

func (s *Service) actPointAddLotteryTimes(c context.Context, item interface{}) railgun.MsgPolicy {
	notifyMsg, ok := item.(*lmdl.ThresholdNotifyMsg)
	if !ok || notifyMsg.MID == 0 || notifyMsg.Activity == "" || notifyMsg.Counter == "" || notifyMsg.Diff == 0 {
		return railgun.MsgPolicyIgnore
	}
	addTimes, ok1 := s.lotteryTypeAddTimes[lmdl.LotteryActPoint]
	if !ok1 {
		return railgun.MsgPolicyIgnore
	}
	mapAddTimes := make(map[string]*l.Lottery)
	for _, v := range addTimes {
		info := &lmdl.LotteryActPointInfo{}
		err := json.Unmarshal([]byte(v.Info), info)
		if err != nil {
			continue
		}
		mapAddTimes[s.buildLotteryCounter(c, info.SID, info.GroupID)] = v
	}
	counter := s.buildMessageLotteryCounter(c, notifyMsg.Counter)
	if counter == "" {
		return railgun.MsgPolicyIgnore
	}
	if add, ok2 := mapAddTimes[counter]; ok2 {
		if notifyMsg.Diff > 0 {
			for i := 0; i < int(notifyMsg.Diff); i++ {
				index := i
				orderNo := fmt.Sprintf("%d_%d_%d_%d", notifyMsg.MID, add.ID, notifyMsg.TimeStamp, index)

				err := s.dao.LotteryAddTimesPub(c, notifyMsg.MID, &lmdl.LotteryAddTimesMsg{
					MID:        notifyMsg.MID,
					SID:        add.Sid,
					CID:        add.ID,
					ActionType: lmdl.LotteryActPoint,
					OrderNo:    orderNo,
				})
				if err != nil {
					log.Errorc(c, "actPointAddLotteryTimes err s.dao.LotteryAddTimesPub mid(%d) sid(%s) cid(%d) orderNo(%s) err(%v)", notifyMsg.MID, add.Sid, add.ID, orderNo, err)
				} else {
					log.Infoc(c, "actPointAddLotteryTimes success s.dao.LotteryAddTimesPub mid(%d) sid(%s) cid(%d) orderNo(%s)", notifyMsg.MID, add.Sid, add.ID, orderNo)
				}
			}
		}

	}
	return railgun.MsgPolicyNormal
}
