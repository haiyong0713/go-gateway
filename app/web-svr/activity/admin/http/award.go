package http

import (
	"encoding/json"
	"fmt"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/admin/model"
)

func awardDetail(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(actSrv.AwardDetail(c, v.ID))
}

func awardSubList(c *bm.Context) {
	v := new(struct {
		Keyword string `form:"keyword"`
		State   string `form:"state"`
		Pn      int64  `form:"pn" default:"1" validate:"min=1"`
		Ps      int64  `form:"ps" default:"10" validate:"min=1,max=20"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	list, count, err := actSrv.AwardList(c, v.Keyword, v.State, v.Pn, v.Ps)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": count,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func awardAdd(c *bm.Context) {
	v := new(model.AddAwardArg)
	if err := c.Bind(v); err != nil {
		return
	}
	if usernameCtx, ok := c.Get("username"); ok {
		v.Author = usernameCtx.(string)
	}
	if v.SidType == 1 {
		if _, err := xstr.SplitInts(v.OtherSids); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	c.JSON(nil, actSrv.AwardAdd(c, v))
}

func awardSave(c *bm.Context) {
	v := new(model.SaveAwardArg)
	if err := c.Bind(v); err != nil {
		return
	}
	if usernameCtx, ok := c.Get("username"); ok {
		v.Author = usernameCtx.(string)
	}
	if v.SidType == 1 {
		if _, err := xstr.SplitInts(v.OtherSids); err != nil {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	c.JSON(nil, actSrv.AwardSave(c, v))
}

func awardSubLog(c *bm.Context) {
	v := new(struct {
		Oid int64 `form:"oid" validate:"min=1"`
		Mid int64 `form:"mid"`
		Pn  int64 `form:"pn" default:"1" validate:"min=1"`
		Ps  int64 `form:"ps" default:"10" validate:"min=1,max=20"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	list, count, err := actSrv.AwardLog(c, v.Oid, v.Mid, v.Pn, v.Ps)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": count,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func awardSubLogExport(c *bm.Context) {
	v := new(struct {
		Oid int64 `form:"oid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	result, _, err := actSrv.AwardLogExport(c, v.Oid)
	if err != nil {
		c.Bytes(-500, "application/csv", nil)
		return
	}
	csvBytes, _ := json.Marshal(result)
	c.Writer.Header().Set("Content-Type", "application/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%d.csv", "award_log", v.Oid))
	c.Writer.Write([]byte("\xEF\xBB\xBF"))
	c.Writer.Write(csvBytes)
}
