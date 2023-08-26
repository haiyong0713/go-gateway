package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/net/trace"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/job/model/match"
	"go-gateway/app/web-svr/activity/tools/lib/function"
	"strconv"
	"time"
)

// archiveActionCanal .
func (s *Service) archiveActionCanal() {
	defer s.waiter.Done()
	if s.archiveBinLogSub == nil {
		return
	}
	var err error
	c := context.Background()
	for {
		msg, ok := <-s.archiveBinLogSub.Messages()
		if !ok {
			log.Errorc(c, "databus: activity-job binlog archiveBinLogSub archive-T exit!")
			return
		}
		msg.Commit()
		m := &match.Message{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Errorc(c, "archiveBinLogSub json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		switch m.Table {
		case _archiveTable:
			// archiveBinLogSub data update
			if m.Action == match.ActUpdate {
				newArc := &like.Archive{}
				oldArc := &like.Archive{}
				if err = json.Unmarshal(m.New, newArc); err != nil {
					log.Errorc(c, "archiveBinLogSub json.Unmarshal(%s) error(%+v)", m.New, err)
					continue
				}
				if err = json.Unmarshal(m.Old, oldArc); err != nil {
					log.Errorc(c, "archiveBinLogSub json.Unmarshal(%s) error(%+v)", m.Old, err)
					continue
				}
				// 定时发布稿件如果同步发布了预约的话 检测稿件审核通过or审核驳回状态
				upActReserveCronCtx := trace.SimpleServerTrace(context.Background(), "up_act_reserve_cron")
				if function.InInt64Slice(int64(newArc.State), []int64{like.StateForbidUserDelay, like.StateForbidRecycle, like.StateForbidPolice, like.StateForbidLock, like.StateForbidUpDelete}) {
					log.Infoc(upActReserveCronCtx, "s.ReserveAuditHandleByArcCronPubState old(%+v) new(%+v)", oldArc, newArc)
					if err = s.ReserveAuditHandleByArcCronPubState(upActReserveCronCtx, oldArc, newArc); err != nil {
						log.Errorc(upActReserveCronCtx, like.UpActReserveArcCronLogPrefix+err.Error())
					}
				}
			}
		}
		log.Infoc(c, "archiveBinLogSub success key:%s partition:%d offset:%d value:%s", msg.Key, msg.Partition, msg.Offset, msg.Value)
	}
}

func (s *Service) ReserveRelationChannelAuditNotify(ctx context.Context, oldObj *like.UpActReserveRelation, newObj *like.UpActReserveRelation) (err error) {
	message := new(like.UpActReserveRelationChannelAudit)
	// 直播预约抽奖
	if newObj.LotteryType > 0 && newObj.DynamicID != "" && newObj.LotteryID != "" {

		message.ReserveID = strconv.FormatInt(newObj.Sid, 10)
		message.ReserveAudit = convertAudit(newObj.Audit)
		message.DynamicID = newObj.DynamicID
		message.DynamicAudit = newObj.DynamicAudit
		message.LotteryID = newObj.LotteryID
		message.LotteryAudit = newObj.LotteryAudit

		// 老状态是审核中 变为 审核拒绝
		if oldObj.Audit == like.UpActReserveAudit && newObj.Audit == like.UpActReserveReject {
			message.ReserveAuditFirst = true
		}

		data, _ := json.Marshal(message)
		if err = retry.WithAttempts(ctx, "UpActReserveRelationChannelAudit", 10, netutil.DefaultBackoffConfig, func(c context.Context) error {
			return component.UpActReserveRelationChannelAudit.Send(ctx, strconv.FormatInt(newObj.Sid, 10)+strconv.FormatInt(newObj.Mid, 10), data)
		}); err != nil {
			err = errors.Wrap(err, "component.UpActReserveRelationChannelAudit.Send err")
			return
		}
	}
	return
}

func convertAudit(input int64) (output int64) {
	switch input {
	case like.UpActReserveReject:
		output = like.UpActReserveChannelReject
	case like.UpActReserveAudit:
		output = like.UpActReserveChannelAudit
	case like.UpActReservePassDelayAudit:
		output = like.UpActReserveChannelPass
	case like.UpActReservePass:
		output = like.UpActReserveChannelPass
	}
	return
}

func (s *Service) PubUpActReserveRelationLotteryReserve(ctx context.Context, newObj *like.ActReserveField) (err error) {
	// 预约抽奖活动推送数据
	isLotteryUpActReserve, err := s.IsUpActReserveLottery(ctx, newObj.Sid)
	if err != nil {
		err = errors.Wrap(err, "s.IsUpActReserveLottery err")
		return
	}
	if isLotteryUpActReserve {
		message := &like.ActivityReservePub{
			Sid:         newObj.Sid,
			Mid:         newObj.Mid,
			State:       newObj.State,
			TimeVersion: time.Now().UnixNano() / 1000,
		}
		data, _ := json.Marshal(message)
		if err = retry.WithAttempts(ctx, "UpActReserveLotteryUserReserveState", 10, netutil.DefaultBackoffConfig, func(c context.Context) error {
			return component.UpActReserveLotteryUserReserveState.Send(ctx, strconv.FormatInt(newObj.Sid, 10)+strconv.FormatInt(newObj.Mid, 10), data)
		}); err != nil {
			err = errors.Wrap(err, "component.UpActReserveLotteryUserReserveState.Send err")
			return
		}
	}

	return
}

func (s *Service) IsUpActReserveLottery(ctx context.Context, sid int64) (isLotteryUpActReserve bool, err error) {
	if v, ok := s.isUpActReserveLottery.Load(sid); ok {
		if v == true {
			isLotteryUpActReserve = true
			return
		}
		return
	} else {
		if isLotteryUpActReserve, err = s.OriginalIsUpActReserveLottery(ctx, sid); err != nil {
			err = errors.Wrap(err, "s.OriginalIsUpActReserve err")
			return
		}
		s.isUpActReserveLottery.Store(sid, isLotteryUpActReserve)
	}
	return
}

func (s *Service) OriginalIsUpActReserveLottery(ctx context.Context, sid int64) (isLotteryUpActReserve bool, err error) {
	var relation *like.UpActReserveRelation
	if err = retry.WithAttempts(ctx, "act_subject", 10, netutil.DefaultBackoffConfig, func(c context.Context) error {
		var e error
		if relation, e = s.dao.GetUpActReserveRelationInfoBySid(ctx, sid); e != nil {
			e = errors.Wrap(e, "s.dao.GetUpActReserveRelationInfoBySid err")
			return e
		}
		return nil
	}); err != nil {
		err = errors.Wrap(err, "retry.WithAttempts err")
		return
	}

	// 抽奖预约
	if relation.Sid > 0 && relation.LotteryType > 0 {
		isLotteryUpActReserve = true
	}

	return
}

func (s *Service) CacheData(typ int64) (res interface{}) {
	if typ == 1 {
		data := make(map[int64]bool)
		s.isUpActReserveLottery.Range(func(key, value interface{}) bool {
			sid := key.(int64)
			data[sid] = value.(bool)
			return true
		})
		res = data
	}
	return
}

// 创建预约抽奖私信卡
func (s *Service) UpActReserveRelationLotteryNotifyCard(ctx context.Context, oldObj, newObj *like.UpActReserveRelation) (err error) {
	// 草稿态=>激活态
	if oldObj.State == int64(api.UpActReserveRelationState_UpReserveEdit) && newObj.State == int64(api.UpActReserveRelationState_UpReserveRelated) {
		log.Infoc(ctx, "UpActReserveRelationLotteryNotifyCard oldObj(%+v) newObj(%+v)", oldObj, newObj)
		// 预约抽奖活动推送数据
		var isLotteryUpActReserve bool
		isLotteryUpActReserve, err = s.IsUpActReserveLottery(ctx, newObj.Sid)
		if err != nil {
			err = errors.Wrap(err, "s.IsUpActReserveLottery err")
			return
		}
		if isLotteryUpActReserve {
			var relations *api.UpActReserveInfoReply
			if err = retry.WithAttempts(ctx, "UpActReserveInfo", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
				relations, err = client.ActivityClient.UpActReserveInfo(ctx, &api.UpActReserveInfoReq{Sids: []int64{newObj.Sid}})
				log.Infoc(ctx, "client.ActivityClient.UpActReserveInfo sid(%+v) reply(%+v)", newObj.Sid, relations)
				return err
			}); err != nil {
				err = errors.Wrap(err, "retry.WithAttempts client.ActivityClient.UpActReserveInfo err")
				return
			}

			if _, ok := relations.List[newObj.Sid]; !ok {
				err = fmt.Errorf("relations.List none relations(%+v) sid(%+v)", relations.List, newObj.Sid)
				return
			}

			if err = s.CreateUpActReserveRelationLotteryCard(ctx, relations.List[newObj.Sid]); err != nil {
				err = errors.Wrapf(err, "CreateUpActReserveRelationLotteryCard err")
				return
			}
			return
		}
	}
	return
}

// 预约抽奖通知
func (s *Service) UpActReserveRelationLotteryNotify(ctx context.Context, newObj *like.Reserve) (err error) {
	// 预约抽奖活动推送数据
	isLotteryUpActReserve, err := s.IsUpActReserveLottery(ctx, newObj.Sid)
	if err != nil {
		err = errors.Wrap(err, "s.IsUpActReserveLottery err")
		return
	}
	if isLotteryUpActReserve {
		log.Infoc(ctx, "UpActReserveRelationLotteryNotify newObj(%+v)", newObj)
		// 创建私信
		if err = s.CreateUpActReserveRelationLotteryNotify(ctx, newObj); err != nil {
			err = errors.Wrapf(err, "CreateUpActReserveRelationLotteryNotify err")
			return
		}
	}
	return
}
