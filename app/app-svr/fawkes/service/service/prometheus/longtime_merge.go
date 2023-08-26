package prometheus

import (
	"context"
	"strconv"
	"time"

	"go-common/library/stat/metric"

	"go-gateway/app/app-svr/fawkes/service/model/gitlab"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

var _metricConnLongProcessMerge = metric.NewGaugeVec(&metric.GaugeVecOpts{
	Namespace: "fawkes",
	Subsystem: "CI",
	Name:      "long_process_merge",
	Help:      "long process merge",
	Labels:    []string{"merge_id", "app_key"},
})

func (s *Service) longProcessMerge() {
	var (
		err            error
		duration       time.Duration
		staticDuration time.Duration
		merging        []*gitlab.GitMerge
		merged         []*gitlab.GitMerge
	)
	if duration, err = time.ParseDuration(s.c.Moni.LongMerge.Duration); err != nil {
		log.Error("parse LongMerge Duration:[%s] error:%v", s.c.Moni.LongMerge.Duration, err)
		return
	}
	if staticDuration, err = time.ParseDuration(s.c.Moni.LongMerge.StatisticalDuration); err != nil {
		log.Error("parse Statistical Duration:[%s] error:%v", s.c.Moni.LongMerge.StatisticalDuration, err)
		return
	}
	if merging, err = s.fkDao.LongMergingProcessSelect(context.Background(), time.Now().Add(-duration), time.Now().Add(-staticDuration)); err != nil {
		log.Error("LongMergingProcessSelect error:%v", err)
		return
	}
	if merged, err = s.fkDao.LongMergedProcessSelect(context.Background(), duration, time.Now().Add(-staticDuration)); err != nil {
		log.Error("LongMergedProcessSelect error:%v", err)
		return
	}
	for _, v := range merging {
		_metricConnLongProcessMerge.Set(float64(time.Since(v.MrStartTime)/time.Minute), strconv.Itoa(v.MergeId), v.AppKey)
	}
	for _, v := range merged {
		_metricConnLongProcessMerge.Set(0, strconv.Itoa(v.MergeId), v.AppKey)
	}
}
