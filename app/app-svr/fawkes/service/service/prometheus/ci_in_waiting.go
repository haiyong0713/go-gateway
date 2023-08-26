package prometheus

import (
	"context"
	"go-common/library/stat/metric"

	prometheusmdl "go-gateway/app/app-svr/fawkes/service/model/prometheus"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

var _metricConnCIInWaiting = metric.NewGaugeVec(&metric.GaugeVecOpts{
	Namespace: "fawkes",
	Subsystem: "CI",
	Name:      "ci_in_waiting",
	Help:      "ci build in waiting status count.",
	Labels:    []string{"app_key"},
})

func (s *Service) ciInWaiting() {
	var (
		res *prometheusmdl.CIInWaiting
		err error
	)
	if res, err = s.fkDao.CIInWaiting(context.Background()); err != nil {
		log.Error("CIInWaiting %v", err)
		return
	}
	_metricConnCIInWaiting.Set(res.Count, res.AppKey)
}
