package pack

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"

	pkmdl "go-gateway/app/app-svr/fawkes/job/internal/model/pack"
	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"
)

type statistics struct {
	RateStr      string  `json:"success"`
	Rate         float64 `json:"-"`
	FailList     []int64 `json:"failed"`
	BatchSum     int64   `json:"batchSum"`
	DeleteFailed int64   `json:"deleteFailed"`
	UpdateFail   int64   `json:"updateFail"`
}

// initCleanRailgun 注册定时任务
func (s *Service) initCleanRailgun() {
	r := railgun.NewRailGun("NAS盘清理", nil,
		railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: s.jobCfg.Corn}),
		railgun.NewCronProcessor(nil, s.CleanRailgun))
	s.packRailgun = r
	r.Start()
}

func (s *Service) CleanRailgun(c context.Context) (msg railgun.MsgPolicy) {
	emptyDate := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	var startTime, endTime time.Time
	if emptyDate == s.jobCfg.Start && emptyDate == s.jobCfg.End {
		// 全空 时间为六个月前
		nowTime := time.Now()
		startTime = nowTime.AddDate(0, -6, -2)
		endTime = nowTime.AddDate(0, -6, -1)
	} else {
		startTime = s.jobCfg.Start
		endTime = s.jobCfg.End
	}
	re, err := s.Clean(c, startTime, endTime, s.jobCfg.PackType, s.jobCfg.AppKey)
	if err != nil {
		log.Error("CleanRailgun occur error: %+v", err)
	}
	log.Info("CleanRailgun statistics clean startTime[%s] endTime[%s], statistics[%+v]", startTime.String(), endTime.String(), re)
	return railgun.MsgPolicyNormal
}

func (s *Service) Clean(c context.Context, tStart, tEnd time.Time, pkgTypes []int64, appKey string) (re interface{}, err error) {
	var (
		list []*cimdl.BuildPack
		keys []*pkmdl.BuildKey
		resp *pkmdl.DeleteResp
		sum  int64 // 需要删除的总条数
	)
	if list, err = s.dao.QueryPackList(c, s.out, tStart.Unix(), tEnd.Unix(), pkgTypes, appKey); err != nil {
		log.Error("query pack error: %+v", err)
		return
	}
	for _, v := range list {
		keys = append(keys, &pkmdl.BuildKey{BuildId: v.BuildID, AppKey: v.AppKey})
	}
	if sum = int64(len(list)); sum == 0 {
		re = &statistics{
			BatchSum: 0,
		}
		return
	}
	if resp, err = s.dao.DeleteExpiredPack(c, s.out, keys); err != nil {
		log.Error("delete expired pack error: %+v", err)
		return
	}
	if resp == nil {
		log.Warn("delete expired pack response is nil")
		return
	}
	r, _ := strconv.ParseFloat(fmt.Sprintf("%.4f", float64(sum-int64(len(resp.BuildIdFail)))/float64(sum)), 64)
	return &statistics{
		BatchSum:     sum,
		RateStr:      fmt.Sprintf("%.2f", r*100) + "%",
		Rate:         r,
		FailList:     resp.BuildIdFail,
		DeleteFailed: int64(len(resp.BuildIdFail)),
		UpdateFail:   sum - int64(len(resp.BuildIdFail)) - resp.AffectedRows,
	}, err
}
