package service

import (
	"context"
	"fmt"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/tool"
	"time"

	topic "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
)

type cpc100PvResponse struct {
	Code int64 `json:"code"`
	Data int64 `json:"data"`
}

func (s *Service) RefreshCpc100PV(ctx context.Context) (err error) {
	log.Infoc(ctx, "RefreshCpc100PV starting...")
	cli := bm.NewClient(conf.Conf.HTTPClient)
	c := conf.Conf.Cpc100Config
	ticker := time.NewTicker(time.Duration(c.RefreshInterval))
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			res := &cpc100PvResponse{}
			err = cli.Get(ctx, c.PvUrl, "", nil, res)
			if err != nil {
				bs, err1 := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfMarkdown, fmt.Sprintf("RefreshCpc100PV: get pv error %v", err), true)
				if err1 == nil {
					_ = tool.SendCorpWeChatRobotAlarmForVote(bs)
				}
				break
			}
			rideN := float64(1)
			if c.RideN != 0 {
				rideN = c.RideN
			}
			displayPv := int64(float64(res.Data)*rideN) + c.AddN
			err = s.dao.CpcSetPV(ctx, displayPv)
			if err != nil {
				bs, err1 := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfMarkdown, fmt.Sprintf("RefreshCpc100PV: set pv error %v", err), true)
				if err1 == nil {
					_ = tool.SendCorpWeChatRobotAlarmForVote(bs)
				}
			}
		}
	}
}

func (s *Service) RefreshCpc100Topic(ctx context.Context) (err error) {
	log.Infoc(ctx, "RefreshCpc100Topic starting...")
	c := conf.Conf.Cpc100Config
	ticker := time.NewTicker(time.Duration(c.TopicRefreshInterval))
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			res, err := client.TopicClient.BatchGetStats(ctx, &topic.BatchGetStatsReq{TopicIds: c.TopicIds})
			if err != nil {
				bs, err1 := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfMarkdown, fmt.Sprintf("RefreshCpc100Topic: call topic error %v", err), true)
				if err1 == nil {
					_ = tool.SendCorpWeChatRobotAlarmForVote(bs)
				}
				break
			}
			viewCount := int64(0)
			for _, r := range res.Stats {
				viewCount = viewCount + r.View
			}
			err = s.dao.CpcSetTopicView(ctx, viewCount)
			if err != nil {
				bs, err1 := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfMarkdown, fmt.Sprintf("RefreshCpc100Topic: set redis error %v", err), true)
				if err1 == nil {
					_ = tool.SendCorpWeChatRobotAlarmForVote(bs)
				}
			}
		}
	}
}
