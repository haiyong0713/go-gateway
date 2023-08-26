package service

import (
	"context"
	"fmt"
	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/tool"
	"time"

	"go-common/library/sync/errgroup.v2"
)

var notifyShouldMentionUser = env.DeployEnv == env.DeployEnvProd

func (s *Service) VoteRefreshDSItemNotEnd(ctx context.Context) (err error) {
	ticker := time.NewTicker(time.Duration(conf.Conf.Vote.DSItemsRefreshDurForNotEnd))
	for {
		select {
		case <-ticker.C:
			log.Errorc(ctx, "VoteRefreshDSItemNotEnd started")
			s.innerRefreshDSItem(ctx, api.ListVoteActivityForRefreshReqType_ListVoteActivityForRefreshReqTypeNotEnded)
			log.Errorc(ctx, "VoteRefreshDSItemNotEnd finished")
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) VoteRefreshDSItemEndWithin90(ctx context.Context) (err error) {
	ticker := time.NewTicker(time.Duration(conf.Conf.Vote.DSItemsRefreshDurForEndWithin90))
	for {
		select {
		case <-ticker.C:
			log.Errorc(ctx, "VoteRefreshDSItemEndWithin90 started")
			s.innerRefreshDSItem(ctx, api.ListVoteActivityForRefreshReqType_ListVoteActivityForRefreshReqTypeEndWithin90)
			log.Errorc(ctx, "VoteRefreshDSItemEndWithin90 finished")
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) innerRefreshDSItem(ctx context.Context, typ api.ListVoteActivityForRefreshReqType) {
	res, err := client.ActivityClient.ListVoteActivityForRefresh(ctx, &api.ListVoteActivityForRefreshReq{
		Type: typ,
	})
	if err != nil {
		log.Errorc(ctx, "ListVoteActivityForRefresh error: %v", err)
		return
	}
	if len(res.Activitys) == 0 {
		return
	}
	var eg errgroup.Group
	for _, activity := range res.Activitys {
		id := activity.Id
		name := activity.Name
		eg.Go(func(ctx context.Context) error {
			_, err := client.ActivityClient.RefreshVoteActivityDSItems(ctx, &api.RefreshVoteActivityDSItemsReq{ActivityId: id})
			if err != nil {
				cause := ecode.Cause(err)
				msg := fmt.Sprintf("ENV: %v. VoteRefreshDSItemNotEnd for activityId: %v, Name: %v,  error: %v, cause: %v", env.DeployEnv, id, name, err, cause.Message())
				log.Errorc(ctx, "%s", msg)
				if bs, err1 := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfText, msg, notifyShouldMentionUser); err1 == nil {
					_ = tool.SendCorpWeChatRobotAlarmForVote(bs)
				}
			} else {
				log.Errorc(ctx, "VoteRefreshDSItemNotEnd for activityId: %v success", id)
			}
			return err
		})
	}

	err = eg.Wait()
	if err != nil {
		log.Errorc(ctx, "VoteRefreshDSItemNotEnd fail: %v", err)
	}
}

func (s *Service) VoteRefreshRankNotEnd(ctx context.Context) (err error) {
	ticker := time.NewTicker(time.Duration(conf.Conf.Vote.RealTimeRankRefreshDurForNotEnd))
	for {
		select {
		case <-ticker.C:
			log.Errorc(ctx, "VoteRefreshRankNotEnd started")
			s.innerRefreshExternalRank(ctx, api.ListVoteActivityForRefreshReqType_ListVoteActivityForRefreshReqTypeNotEnded)
			s.innerRefreshInternalRank(ctx, api.ListVoteActivityForRefreshReqType_ListVoteActivityForRefreshReqTypeNotEnded)
			log.Errorc(ctx, "VoteRefreshRankNotEnd finished")
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) VoteRefreshRankEndWithin90(ctx context.Context) (err error) {
	ticker := time.NewTicker(time.Duration(conf.Conf.Vote.RealTimeRankRefreshDurForEndWithin90))
	for {
		select {
		case <-ticker.C:
			log.Errorc(ctx, "VoteRefreshRankEndWithin90 started")
			s.innerRefreshExternalRank(ctx, api.ListVoteActivityForRefreshReqType_ListVoteActivityForRefreshReqTypeEndWithin90)
			s.innerRefreshInternalRank(ctx, api.ListVoteActivityForRefreshReqType_ListVoteActivityForRefreshReqTypeEndWithin90)
			log.Errorc(ctx, "VoteRefreshRankEndWithin90 finished")
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) innerRefreshExternalRank(ctx context.Context, typ api.ListVoteActivityForRefreshReqType) {
	res, err := client.ActivityClient.ListVoteActivityForRefresh(ctx, &api.ListVoteActivityForRefreshReq{
		Type: typ,
	})
	if err != nil {
		log.Errorc(ctx, "ListVoteActivityForRefresh error: %v", err)
		return
	}
	if len(res.Activitys) == 0 {
		return
	}
	var eg errgroup.Group
	for _, activity := range res.Activitys {
		id := activity.Id
		name := activity.Name
		switch activity.Rule.VoteUpdateRule {
		case int64(api.VoteCountUpdateRule_VoteCountUpdateRuleRealTime):
			//票数刷新规则: 实时
			//票数zset和DB中的total_votes在投票时已经刷新,此时直接刷新排行榜
			eg.Go(func(ctx context.Context) error {
				_, err = client.ActivityClient.RefreshVoteActivityRankExternal(ctx, &api.RefreshVoteActivityRankExternalReq{ActivityId: id})
				if err != nil {
					cause := ecode.Cause(err)
					msg := fmt.Sprintf("ENV: %v. innerRefreshExternalRank for activityId: %v RefreshVoteActivityRankExternal error: %v, cause: %v", env.DeployEnv, id, err, cause.Message())
					log.Errorc(ctx, "%s", msg)
					if bs, err1 := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfText, msg, notifyShouldMentionUser); err1 == nil {
						_ = tool.SendCorpWeChatRobotAlarmForVote(bs)
					}
				} else {
					log.Errorc(ctx, "innerRefreshExternalRank for activityId: %v success", id)
				}
				return err
			})
		case int64(api.VoteCountUpdateRule_VoteCountUpdateRuleOnTime):
			//票数刷新规则: 定时
			//需要刷新DB中的total_votes和缓存中的zset,并刷新排行榜
			now := time.Now()
			y, m, d := now.Date()
			h := now.Hour()
			todayNotRefreshed := activity.LastRankRefreshTime < time.Date(y, m, d, 0, 0, 0, 0, now.Location()).Unix()
			shouldRefreshNow := todayNotRefreshed && int64(h) >= activity.Rule.VoteUpdateCron
			if shouldRefreshNow {
				eg.Go(func(ctx context.Context) error {
					_, err := client.ActivityClient.RefreshVoteActivityRankZset(ctx, &api.RefreshVoteActivityRankZsetReq{ActivityId: id})
					if err != nil {
						cause := ecode.Cause(err)
						msg := fmt.Sprintf("ENV: %v. innerRefreshExternalRank for activity Id: %v, Name: %v, RefreshVoteActivityRankZset error: %v, cause: %v", env.DeployEnv, id, name, err, cause.Message())
						log.Errorc(ctx, "%s", msg)
						if bs, err1 := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfText, msg, notifyShouldMentionUser); err1 == nil {
							_ = tool.SendCorpWeChatRobotAlarmForVote(bs)
						}
						return err
					} else {
						log.Errorc(ctx, "innerRefreshExternalRank for activityId: %v RefreshVoteActivityRankZset success", id)
					}
					_, err = client.ActivityClient.RefreshVoteActivityRankExternal(ctx, &api.RefreshVoteActivityRankExternalReq{ActivityId: id})
					if err != nil {
						cause := ecode.Cause(err)
						msg := fmt.Sprintf("ENV: %v. innerRefreshExternalRank for activity Id: %v, Name: %v,RefreshVoteActivityRankExternal error: %v, cause: %v", env.DeployEnv, id, name, err, cause.Message())
						log.Errorc(ctx, "%s", msg)
						if bs, err1 := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfText, msg, notifyShouldMentionUser); err1 == nil {
							_ = tool.SendCorpWeChatRobotAlarmForVote(bs)
						}
					} else {
						log.Errorc(ctx, "innerRefreshExternalRank for activityId: %v RefreshVoteActivityRankExternal success", id)
					}
					return err
				})
			}
		}
	}
	err = eg.Wait()
	if err != nil {
		log.Errorc(ctx, "innerRefreshExternalRank fail: %v", err)
	}
}

func (s *Service) innerRefreshInternalRank(ctx context.Context, typ api.ListVoteActivityForRefreshReqType) {
	res, err := client.ActivityClient.ListVoteActivityForRefresh(ctx, &api.ListVoteActivityForRefreshReq{
		Type: typ,
	})
	if err != nil {
		log.Errorc(ctx, "ListVoteActivityForRefresh error: %v", err)
		return
	}
	if len(res.Activitys) == 0 {
		return
	}
	var eg errgroup.Group
	for _, activity := range res.Activitys {
		id := activity.Id
		name := activity.Name
		eg.Go(func(ctx context.Context) error {
			_, err := client.ActivityClient.RefreshVoteActivityRankInternal(ctx, &api.RefreshVoteActivityRankInternalReq{ActivityId: id})
			if err != nil {
				cause := ecode.Cause(err)
				msg := fmt.Sprintf("ENV: %v. innerRefreshInternalRank for activityId: %v, name: %v,error: %v, cause: %v", env.DeployEnv, id, name, err, cause.Message())
				log.Errorc(ctx, "%s", msg)
				if bs, err1 := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfText, msg, notifyShouldMentionUser); err1 == nil {
					_ = tool.SendCorpWeChatRobotAlarmForVote(bs)
				}
			} else {
				log.Errorc(ctx, "innerRefreshInternalRank for activityId: %v success", id)
			}
			return err
		})
	}
	err = eg.Wait()
	if err != nil {
		log.Errorc(ctx, "innerRefreshInternalRank fail: %v", err)
	}
}
