package dao

import (
	"context"
	"go-common/library/log"
	actApi "go-gateway/app/web-svr/activity/interface/api"
)

const (
	_contestIDsBulkSize         = 100
	_esportsBusiness            = 1
	_callActivityGuessDefaultPn = 1
)

// GetGuessDetail 该接口后续通过activity模块的优化进行
// 竞猜更多通过赛程id列表 或 单一 赛程 id获取竞猜记录
func (d *dao) GetGuessDetail(ctx context.Context, contestIds []int64, mid int64) (guessMap map[int64]bool, err error) {
	guessMap = make(map[int64]bool)
	idsCount := len(contestIds)
	if idsCount == 0 {
		return
	}
	for i := 0; i < idsCount; i += _contestIDsBulkSize {
		var partIDs []int64
		if i+_contestIDsBulkSize > idsCount {
			partIDs = contestIds[i:]
		} else {
			partIDs = contestIds[i : i+_contestIDsBulkSize]
		}
		req := new(actApi.UserGuessMatchsReq)
		{
			req.Mid = mid
			req.Business = _esportsBusiness
			req.Oids = partIDs
			req.Pn = _callActivityGuessDefaultPn
			req.Ps = int64(len(partIDs))
		}
		tmpResp, tmpErr := d.activityClient.UserGuessMatchs(ctx, req)
		if tmpErr != nil {
			log.Errorc(ctx, "contest component componentContestGuessMap  mid(%d) contestIDList(%+v) error(%+v)", mid, contestIds, tmpErr)
			return
		}
		for _, userGuess := range tmpResp.UserGroup {
			guessMap[userGuess.Oid] = true
		}
	}
	return
}
