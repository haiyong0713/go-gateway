package http

import (
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	statisticsmdl "go-gateway/app/app-svr/fawkes/service/model/statistics"
)

func statisticsLine(c *bm.Context) {
	var (
		res         = map[string]interface{}{}
		matchOption *statisticsmdl.FawkesMatchOption
		err         error
	)
	matchOption = new(statisticsmdl.FawkesMatchOption)
	if err = c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.ClassType == "" {
		res["message"] = "class_type 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.StatisticsSvr.StatisticsLine(c, matchOption))
}

func statisticsPie(c *bm.Context) {
	var (
		res         = map[string]interface{}{}
		matchOption *statisticsmdl.FawkesMatchOption
		err         error
	)
	matchOption = new(statisticsmdl.FawkesMatchOption)
	if err = c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.ClassType == "" {
		res["message"] = "class_type 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if matchOption.ClassType == "" {
		res["message"] = "column 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.StatisticsSvr.StatisticsPie(c, matchOption))
}

func commonInfoList(c *bm.Context) {
	var (
		res         = map[string]interface{}{}
		matchOption *statisticsmdl.FawkesMatchOption
		err         error
	)
	matchOption = new(statisticsmdl.FawkesMatchOption)
	if err := c.Bind(matchOption); err != nil {
		res["message"] = "参数解析异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = matchOption.Check(); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.StatisticsSvr.CommonInfoList(c, matchOption))
}
