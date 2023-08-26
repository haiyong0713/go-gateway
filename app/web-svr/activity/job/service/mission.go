package service

import (
	"context"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/conf"
	lmdl "go-gateway/app/web-svr/activity/job/model/like"
	"strconv"
	"strings"
	"time"
)

func (s *Service) refreshValidActivity(ctx context.Context) (err error) {
	duration := time.Duration(_defaultRefreshTicker) * time.Second
	if conf.Conf.MissionConfig != nil && conf.Conf.MissionConfig.RefreshTickerSecond != 0 {
		duration = time.Duration(conf.Conf.MissionConfig.RefreshTickerSecond) * time.Second
	}
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			err = s.doRefreshValidActivity(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) doRefreshValidActivity(ctx context.Context) (err error) {
	resp, err := s.actGRPC.GetValidMissionActivityIds(ctx, &api.NoReply{})
	if err != nil {
		log.Errorc(ctx, "[refreshValidActivity][GetValidMissionActivityIds][Error], err:%+v")
		return
	}
	for _, actId := range resp.ActIds {
		_, err = s.actGRPC.RefreshValidMissionActivityCache(ctx, &api.RefreshValidMissionActivityCacheReq{
			ActId: actId,
		})
		if err != nil {
			log.Errorc(ctx, "[refreshValidActivity][RefreshValidMissionActivityCache][Error], err:%+v")
		}
	}
	return
}

func (s *Service) missionGroupsConsumer(ctx context.Context, item interface{}) railgun.MsgPolicy {
	notifyMsg, ok := item.(*lmdl.ThresholdNotifyMsg)
	if !ok || notifyMsg.MID == 0 || notifyMsg.Activity == "" || notifyMsg.Counter == "" || notifyMsg.Diff == 0 {
		return railgun.MsgPolicyIgnore
	}
	_, groupId, err := s.groupInfoGet(ctx, notifyMsg.Counter)
	if err != nil {
		return railgun.MsgPolicyIgnore
	}
	_, err = s.actGRPC.GroupConsumerForTaskComplete(ctx, &api.GroupConsumerForTaskCompleteReq{
		GroupId:   groupId,
		Mid:       notifyMsg.MID,
		Total:     notifyMsg.Total,
		Timestamp: notifyMsg.TimeStamp,
	})
	if err != nil {
		return railgun.MsgPolicyAttempts
	}
	return railgun.MsgPolicyNormal
}

func (s *Service) groupInfoGet(ctx context.Context, counter string) (actId, groupId int64, err error) {
	list := strings.Split(counter, "_ss_")
	if len(list) != 3 {
		err = xecode.Errorf(xecode.RequestErr, "非合法节点组")
		return
	}
	actId, err = strconv.ParseInt(list[0], 10, 64)
	if err != nil {
		log.Errorc(ctx, "[groupInfoGet][ActId][Error], err:%+v", err)
		return
	}
	groupId, err = strconv.ParseInt(list[1], 10, 64)
	if err != nil {
		log.Errorc(ctx, "[groupInfoGet][GroupId][Error], err:%+v", err)
		return
	}
	return
}

func (s *Service) makeUpReceiveRecords(ctx context.Context) (err error) {
	duration := time.Duration(_defaultRefreshTicker) * time.Second
	if conf.Conf.MissionConfig != nil && conf.Conf.MissionConfig.MakeUpReceiveRecordSecond != 0 {
		duration = time.Duration(conf.Conf.MissionConfig.MakeUpReceiveRecordSecond) * time.Second
	}
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			err = s.doMakeUp(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) doMakeUp(ctx context.Context) (err error) {
	beginTime := time.Now().Unix()
	log.Infoc(ctx, "[doMakeUp][UpdateBegin], beginTime:%d", beginTime)

	resp, err := s.actGRPC.GetValidMissionActivityIds(ctx, &api.NoReply{})
	if err != nil {
		log.Errorc(ctx, "[doMakeUp][GetValidMissionActivityIds][Error], err:%+v")
		return
	}
	if resp == nil || resp.ActIds == nil || len(resp.ActIds) == 0 {
		return
	}
	for _, actId := range resp.ActIds {
		for tableIndex := 0; tableIndex < 100; tableIndex++ {
			_ = s.oneTableDoMakeUp(ctx, actId, int32(tableIndex))
		}
	}
	endTime := time.Now().Unix()
	log.Infoc(ctx, "[doMakeUp][UpdateFinish], endTime:%d, interval:%d", endTime, endTime-beginTime)
	return
}

func (s *Service) oneTableDoMakeUp(ctx context.Context, actId int64, tableIndex int32) (err error) {
	count := 0
	for {
		resp, errR := s.actGRPC.GetMissionReceivingRecords(ctx, &api.GetMissionReceivingRecordsReq{
			ActId:      actId,
			TableIndex: tableIndex,
		})
		if errR != nil {
			log.Errorc(ctx, "[oneTableDoMakeUp][GetMissionReceivingRecords], err:%+v", errR)
			err = errR
			return
		}
		if resp == nil || resp.List == nil || len(resp.List) == 0 {
			return
		}
		for _, record := range resp.List {
			_, errG := s.actGRPC.RetryMissionReceiveRecord(ctx, &api.RetryMissionReceiveRecordReq{
				ReceiveId: record.ReceiveId,
				Mid:       record.Mid,
				ActId:     record.ActId,
			})
			if errG != nil {
				log.Errorc(ctx, "[oneTableDoMakeUp][GetMissionReceivingRecords], err:%+v", errG)
				continue
			}
			count++
		}

		if len(resp.List) < 100 {
			log.Infoc(ctx, "[oneTableDoMakeUp][UpdateSuccess], actId:%d, tableIndex:%d, updateCount:%d", actId, tableIndex, count)
			return
		}
	}
}
