package record

import (
	"context"

	"go-common/library/log"
	"go-common/library/stat/prom"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

const (
	_portalAdvance = 0
)

func recStateProm(recState int) {
	switch recState {
	case model.RecWithoutCursor:
		prom.BusinessInfoCount.Incr("Rec_Without_Cursor")
	case model.RecWithCursorInProgress:
		prom.BusinessInfoCount.Incr("Rec_With_Cursor_In_Progress")
	case model.RecWithCursorPerfect:
		prom.BusinessInfoCount.Incr("Rec_With_Cursor_Perfect")
	}
}

func (d *Dao) GetLastRecord(c context.Context, params *model.NodeInfoParam, requestID, mid int64, pullHandler model.PullHandler) (lastRec *api.GameRecords, err error) {
	if lastRec, err = d.Record(c, mid, params.GraphVersion, params.Buvid); err != nil {
		log.Error("s.dao.Record err(%v) mid(%d) graph(%d)", err, mid, params.GraphVersion)
		return
	}
	if params == nil || lastRec == nil || params.Portal != _portalAdvance { // 只处理有存档，并且是向前走的情况下的兼容
		return
	}
	if currentID := pullHandler(lastRec); currentID == requestID { // 如果Portal=0，并且请求id是上一个id，认为是回溯
		params.Portal = 1
		prom.BusinessInfoCount.Incr("Portal_Zero_Retry")
		return
	}
	prom.BusinessInfoCount.Incr("Portal_Zero_Normal")
	return
}

// WriteRecord get record info
func (d *Dao) WriteRecord(c context.Context, params *model.NodeInfoParam, mid, requestID, rootID int64, newRecHandler model.NewRecordHandler,
	pressHandler model.PressHandler, fromRecHandler model.FromRecHandler, lastRec *api.GameRecords, cacheOnly bool) (inChoices, inCursorChoices string, fromID, currentCursor int64, hvarReq *model.HvarReq, err error) {
	var recState int
	// 根据之前的存档记录，当前ID和portal来生成新的存档记录
	if inChoices, inCursorChoices, fromID, currentCursor, recState, err = newRecHandler(lastRec, requestID, rootID, params.Cursor, params.Portal); err != nil {
		return
	}
	recStateProm(recState)
	newRec := &api.GameRecords{GraphId: params.GraphVersion, Aid: params.AID, Mid: mid, Choices: inChoices, Buvid: params.Buvid, CursorChoice: inCursorChoices, CurrentCursor: currentCursor}
	pressHandler(newRec, requestID) // 根据graph类型去写入不同的字段
	if !cacheOnly {
		if err = d.AddRecord(c, newRec, false); err != nil {
			log.Error("d.AddRecord(%v) error(%v)", newRec, err)
			return
		}
	}
	hvarReq = fromRecHandler(lastRec, params.Portal, rootID, requestID, currentCursor, inChoices, inCursorChoices)
	if fanoutErr := d.cache.Do(c, func(ctx context.Context) {
		//nolint:errcheck
		d.AddCacheRecord(ctx, newRec, params)
	}); fanoutErr != nil {
		log.Error("fanout do mid:%d, graphID:%d, err(%+v) ", mid, params.GraphVersion, fanoutErr)
	}
	return
}

// WriteRecordPreview get record info
func (d *Dao) WriteRecordPreview(c context.Context, mid, requestID, graphID, aid, rootID, cursor int64, portal int32, newRecHandler model.NewRecordHandler,
	pressHandler model.PressHandler, fromRecHandler model.FromRecHandler) (inChoices, inCursorChoices string, currentCursor int64, hvarReq *model.HvarReq, err error) {
	var (
		lastRec  *api.GameRecords
		recState int
	)
	// 预览无重复请求问题
	if lastRec, err = d.RawRecord(c, mid, graphID, true); err != nil {
		log.Error("s.dao.Record err(%v) mid(%d) graph(%d)", err, mid, graphID)
		return
	}
	if inChoices, inCursorChoices, _, currentCursor, recState, err = newRecHandler(lastRec, requestID, rootID, cursor, portal); err != nil {
		return
	}
	recStateProm(recState)
	newRec := &api.GameRecords{GraphId: graphID, Aid: aid, Mid: mid, Choices: inChoices, HiddenVars: "", GlobalVars: "", CurrentCursor: currentCursor, CursorChoice: inCursorChoices}
	pressHandler(newRec, requestID) // 根据graph类型去写入不同的字段
	if err = d.AddRecord(c, newRec, true); err != nil {
		log.Error("d.AddRecord(%v) error(%v)", newRec, err)
		return
	}
	hvarReq = fromRecHandler(lastRec, portal, rootID, requestID, currentCursor, inChoices, inCursorChoices)
	return

}
