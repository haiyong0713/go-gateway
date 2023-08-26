package view

import (
	"context"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
	"go-gateway/app/app-svr/app-car/interface/model/history"
)

func (s *Service) HisReport(c context.Context, mid int64, buvid string, param *history.ReportParam) error {
	const (
		// 历史记录上报业务类型
		_av  = 3
		_pgc = 4
		// 历史记录上报设备
		_dtCar   = 8
		_dtThing = 9
	)
	tp := _pgc
	dt := _dtCar
	if param.Device == model.DeviceThing {
		dt = _dtThing
	}
	if param.Otype == model.GotoAv {
		tp = _av
		param.SeasonType = 0
	}
	serialBusinessType := func() int64 {
		serialType, ok := common.ItemTypeToSerialBusinessType[common.ItemType(param.ItemType)]
		if ok && param.ItemID > 0 {
			return serialType
		}
		return 0
	}()
	if serialBusinessType > 0 && (mid > 0 || buvid != "") { // 上报合集历史
		err := s.serialDao.AddHistory(c, mid, param.Aid, param.Progress, param.ItemID, serialBusinessType, buvid)
		if err != nil { // 忽略合集历史上报错误
			log.Errorc(c, "AddHistory s.serialDao.AddHistory err=%+v", err)
		}
	}
	if err := s.his.Report(c, mid, buvid, tp, dt, param); err != nil {
		return err
	}
	itemToFmInfo(param)
	if param.FmId > 0 {
		s.FmReportToAI(c, mid, buvid, param)
	}
	return nil
}

// FmReportToAI FM播放记录实时上报算法侧
func (s *Service) FmReportToAI(ctx context.Context, mid int64, buvid string, param *history.ReportParam) {
	if param.FmType != fm_v2.AudioVertical && param.FmType != fm_v2.AudioSeason && param.FmType != fm_v2.AudioSeasonUp {
		return
	}
	err := s.fmReport.Do(ctx, func(ctx context.Context) {
		s.FmReportToAIHandler(ctx, mid, buvid, param)
	})
	if err != nil {
		log.Error("FmReportToAI s.fmReport.Do error:%+v", err)
		return
	}
}

func (s *Service) FmReportToAIHandler(ctx context.Context, mid int64, buvid string, param *history.ReportParam) {
	var (
		count int
		err   error
	)
	if param.FmType == fm_v2.AudioSeasonUp || param.FmType == fm_v2.AudioSeason {
		count, err = s.fmDao.GetSeasonOidCount(ctx, fm_v2.SeasonInfoReq{Scene: fm_v2.SceneFm, FmType: param.FmType, SeasonId: param.FmId})
		if err != nil {
			log.Error("FmReportToAI s.fmDao.GetSeasonOidCount fmType:%s, fmId:%d, error:%+v", param.FmType, param.FmId, err)
			count = -1
		}
	}
	ts := time.Unix(param.Timestamp, 0).Format("2006-01-02 15:04:05")
	reportReq := &fm_v2.HistoryReportFm{
		Source:        ToSource(param.FmType),
		Mid:           mid,
		Buvid:         buvid,
		FmID:          param.FmId,
		FmType:        string(param.FmType),
		PlayTime:      ts,
		PlayEvent:     param.PlayEvent,
		Aid:           param.Aid,
		ArchivesCount: count,
	}
	err = s.his.ReportToAI(ctx, reportReq)
	if err != nil {
		log.Error("FmReportToAI s.his.ReportToAI fmType:%s, fmId:%d, error:%+v", param.FmType, param.FmId, err)
		return
	}
}

// 2.3 兼容老版本的FM上报
func itemToFmInfo(param *history.ReportParam) {
	if param.ItemID > 0 && param.ItemType == string(common.ItemTypeFmChannel) {
		param.FmType = fm_v2.AudioVertical
		param.FmId = param.ItemID
	}
	if param.ItemID > 0 && param.ItemType == string(common.ItemTypeFmSerial) {
		param.FmType = fm_v2.AudioSeason
		param.FmId = param.ItemID
	}
}

func ToSource(fmType fm_v2.FmType) string {
	switch fmType {
	case fm_v2.AudioSeason, fm_v2.AudioSeasonUp:
		return "Series"
	case fm_v2.AudioVertical:
		return "Channel"
	default:
		return ""
	}
}
