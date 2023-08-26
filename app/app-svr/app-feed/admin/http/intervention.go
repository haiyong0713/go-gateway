package http

import (
	"strconv"

	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/model/intervention"
	"go-gateway/app/app-svr/app-feed/admin/util"
	"go-gateway/pkg/idsafe/bvid"
)

func createIntervention(c *bm.Context) {
	var (
		err     error
		request = &intervention.Detail{}
	)
	if err = c.Bind(request); err != nil {
		return
	}
	_, username := util.UserInfo(c)

	detail, err := interventionSrv.CreateIntervention(request, username)
	c.JSON(detail, err)
	//nolint:gosimple
	return
}

func editIntervention(c *bm.Context) {
	var (
		err     error
		request = &intervention.Detail{}
	)
	if err = c.Bind(request); err != nil {
		return
	}
	_, username := util.UserInfo(c)

	detail, err := interventionSrv.EditIntervention(request, username)
	c.JSON(detail, err)
	//nolint:gosimple
	return
}

func changeIntervention(c *bm.Context) {
	var (
		err     error
		request = &intervention.Detail{}
	)
	if err = c.Bind(request); err != nil {
		return
	}
	_, username := util.UserInfo(c)

	detail, err := interventionSrv.ChangeIntervention(request, username)
	c.JSON(detail, err)
	//nolint:gosimple
	return
}

func searchIntervention(c *bm.Context) {
	var (
		err   error
		param = new(struct {
			Id        string `form:"id"`
			Avid      int64  `form:"avid"`
			Title     string `form:"title"`
			StartTime int64  `form:"start_time"`
			EndTime   int64  `form:"end_time"`
			CreatedBy string `form:"created_by"`
			Pn        int    `form:"pn" default:"1"`
		})
	)
	if err = c.Bind(param); err != nil {
		return
	}

	conditions := intervention.Detail{
		Avid:      param.Avid,
		Bvid:      param.Id,
		Title:     param.Title,
		StartTime: param.StartTime,
		EndTime:   param.EndTime,
		CreatedBy: param.CreatedBy,
	}

	result, err := interventionSrv.DetailList(&conditions, param.Pn)
	c.JSON(result, err)
	//nolint:gosimple
	return
}

func createOptLog(c *bm.Context) {
	var (
		err     error
		request = &intervention.OptLogDetail{}
	)
	if err = c.Bind(request); err != nil {
		return
	}

	err = interventionSrv.CreateOptLog(request)
	c.JSON(request, err)
	//nolint:gosimple
	return
}

func searchInterventionLogs(c *bm.Context) {
	var (
		err   error
		param = new(struct {
			InterventionId uint   `form:"intervention_id"`
			Id             string `form:"id"`
			Avid           int64  `form:"avid"`
			OpUser         string `form:"op_user"`
			Pn             int    `form:"pn" default:"1"`
		})
		avid int64
	)

	if err = c.Bind(param); err != nil {
		c.JSON(nil, err)
		return
	}

	if param.Avid != 0 {
		avid = param.Avid
	} else if param.Id != "" {
		avid, err = strconv.ParseInt(param.Id, 10, 64)
		if err != nil {
			// 不是avid，就当做bvid
			avid, err = bvid.BvToAv(param.Id)
			if err != nil {
				c.JSON(nil, err)
				return
			}
		}
	}

	conditions := intervention.OptLogDetail{
		InterventionId: param.InterventionId,
		Avid:           avid,
		OpUser:         param.OpUser,
	}

	result, err := interventionSrv.OpLogList(&conditions, param.Pn)
	c.JSON(result, err)
	//nolint:gosimple
	return
}
