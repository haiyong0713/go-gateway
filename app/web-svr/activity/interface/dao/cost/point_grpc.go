package cost

import (
	"context"
	actPlat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/client"
)

// TaskFormulaTotal 任务处公式我的积分
func (d *dao) TaskFormulaTotal(ctx context.Context, mid int64, activityId string) (int64, error) {
	var (
		totalPoint int64 = 0
	)
	countReply, err := client.ActPlatClient.GetFormulaTotal(ctx, &actPlat.GetFormulaTotalReq{
		Activity: activityId,
		Formula:  "total",
		Mid:      mid,
	})
	if err != nil {
		log.Errorc(ctx, "get grpc client.ActPlatClient.GetFormulaTotal() mid(%d) error(%+v)", mid, err)
		return totalPoint, err
	}
	if countReply == nil || countReply.Items == nil {
		log.Warnc(ctx, "get grpc client.ActPlatClient.GetFormulaTotal() mid(%d) historyReply is nil", mid)
		return totalPoint, nil
	}

	totalPoint = countReply.Result
	return totalPoint, nil
}
