package statistics

import (
	"context"

	statisticsmdl "go-gateway/app/app-svr/fawkes/service/model/statistics"
)

func (s *Service) StatisticsLine(c context.Context, matchOptions *statisticsmdl.FawkesMatchOption) (res []*statisticsmdl.FawkesMoni, err error) {
	res, err = s.fkDao.FawkesMoniLine(c, matchOptions)
	return
}

func (s *Service) StatisticsPie(c context.Context, matchOptions *statisticsmdl.FawkesMatchOption) (res []*statisticsmdl.FawkesMoni, err error) {
	res, err = s.fkDao.FawkesMoniPie(c, matchOptions)
	return
}

func (s *Service) CommonInfoList(c context.Context, matchOptions *statisticsmdl.FawkesMatchOption) (res []interface{}, err error) {
	switch matchOptions.EventID {
	case statisticsmdl.CIBUILD:
		res, err = s.fkDao.CIInfoList(c, matchOptions)
	case statisticsmdl.CIJOB:
		res, err = s.fkDao.CIJobList(c, matchOptions)
	case statisticsmdl.LASER:
		res, err = s.fkDao.SttLaserList(c, matchOptions)
	case statisticsmdl.CICOMPILE:
		res, err = s.fkDao.SttCICompileList(c, matchOptions)
	case statisticsmdl.TECHNOLOGYSTORAGE, statisticsmdl.TECHNOLOGYSTORAGE_UNREGISTERED:
		res, err = s.fkDao.TechnologyInfoList(c, matchOptions)
	case statisticsmdl.TECHNOLOGYQUANTITY:
		res, err = s.fkDao.TechnologyQuantityInfoList(c, matchOptions)
	}
	return
}
