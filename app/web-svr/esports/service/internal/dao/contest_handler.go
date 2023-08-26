package dao

import (
	"context"

	"go-common/library/log"
	egV2 "go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

// ContestStatusUpdateHandler 该接口内的处理流程需可保证幂等
func (d *dao) contestStatusUpdateHandler(ctx context.Context, contestModel *model.ContestModel, oldContestModel *model.ContestModel) (err error) {
	// 冻结不处理
	if contestModel.Status == model.FreezeTrue {
		return
	}
	if contestModel.ContestStatus == model.ContestStatusIng && canPush(contestModel) {
		err = d.contestBeginHandler(ctx, contestModel, oldContestModel)
		if err != nil {
			log.Errorc(ctx, "[Dao][contestStatusUpdateHandler][contestBeginHandler][Error], error:%+v", err)
			return
		}
	}
	return
}

func (d *dao) contestBeginHandler(ctx context.Context, contestModel *model.ContestModel, oldContestModel *model.ContestModel) (err error) {
	if oldContestModel != nil &&
		oldContestModel.ContestStatus != model.ContestStatusWaiting &&
		oldContestModel.ContestStatus != model.ContestStatusInit {
		// 过滤掉之前非 初始化、未开始的记录
		return
	}
	// 新增或比赛状态为已开始
	if err = d.InitTunnelEvent(ctx, contestModel); err != nil {
		log.Errorc(ctx, "[Service][ContestStatusUpdateHandler][InitTunnelEvent][Error],contestID:%d err:%+v", contestModel.ID, err)
		return
	}
	group := egV2.WithContext(ctx)
	group.Go(func(ctx context.Context) error {
		if contestModel.Special != 0 {
			log.Warnc(ctx, "[Service][ContestStatusUpdateHandler][UpsertTunnelCard][Error], contestID:%d  not default contest", contestModel.ID)
			return nil
		}
		smallErr := d.UpsertTunnelCard(ctx, contestModel)
		if smallErr != nil {
			log.Errorc(ctx, "[Service][ContestStatusUpdateHandler][UpsertTunnelCard][Error], contestID:%d  err:%+v", contestModel.ID, smallErr)
			return smallErr
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		msgErr := d.UpsertTunnelMsgCard(ctx, contestModel)
		if msgErr != nil {
			log.Errorc(ctx, "[Service][ContestStatusUpdateHandler][UpsertTunnelMsgCard][Error], contestID:%d  err:%+v", contestModel.ID, msgErr)
			return msgErr
		}
		return nil
	})
	err = group.Wait()
	return
}

func canPush(contestModel *model.ContestModel) bool {
	// 未填写直播间直接返回
	return contestModel.LiveRoom != 0
}
