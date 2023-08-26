package http

import (
	"encoding/json"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	rankmdl "go-gateway/app/web-svr/activity/admin/model/rank"
)

func rankCreate(c *bm.Context) {
	arg := new(rankmdl.CreateReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(rankSrv.Create(c, arg.SID, arg.SIDSource, arg.Stime, arg.Etime, userName))
}

func rankExport(c *bm.Context) {
	arg := new(rankmdl.ExportReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankSrv.RankResultExport(c, arg.ID, arg.AttributeType, userName))
}

func rankDetail(c *bm.Context) {
	arg := new(rankmdl.DetailReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(rankSrv.Detail(c, arg.SID, arg.SIDSource))
}

func rankUpdate(c *bm.Context) {
	arg := new(rankmdl.Rank)
	if err := c.Bind(arg); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, rankSrv.Update(c, arg.ID, arg, userName))
}

func rankOffline(c *bm.Context) {
	arg := new(rankmdl.OfflineReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, rankSrv.Offline(c, arg.ID))
}

func rankIntervention(c *bm.Context) {
	arg := new(rankmdl.GetInterventionReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(rankSrv.GetBlackOrWhite(c, arg.ID, arg.ObjectType, arg.InterventionType, arg.Pn, arg.Ps))
}

func updateIntervention(c *bm.Context) {
	arg := new(rankmdl.UpdateInterventionReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, rankSrv.UpdateBlackOrWhite(c, arg.ID, arg.List))
}

func rankResult(c *bm.Context) {
	arg := new(rankmdl.ResultReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(rankSrv.RankResult(c, arg.ID, arg.AttributeType, arg.Pn, arg.Ps))
}

func rankResultUpdate(c *bm.Context) {
	arg := new(rankmdl.ResultEditReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	err := json.Unmarshal([]byte(arg.Result), &arg.ResultStruct)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, rankSrv.UpdateRankResult(c, arg.ID, arg.ResultStruct))
}

func rankPublish(c *bm.Context) {
	arg := new(rankmdl.PublishReq)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, rankSrv.Publish(c, arg.ID, arg.AttributeType, arg.Batch))
}
