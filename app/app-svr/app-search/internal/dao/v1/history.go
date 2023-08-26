package v1

import (
	"context"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
)

func (d *dao) GetHistoryFrequent(ctx context.Context, req *hisApi.HistoryFrequentReq) (*hisApi.HistoryFrequentReply, error) {
	return d.hisClient.HistoryFrequent(ctx, req)
}
