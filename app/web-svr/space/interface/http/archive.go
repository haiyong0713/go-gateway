package http

import (
	"encoding/json"
	gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/ecode"
	"go-gateway/app/web-svr/space/interface/conf"
	"go-gateway/app/web-svr/space/interface/model"
	"strconv"
	"strings"
)

func topArc(c *bm.Context) {
	var (
		mid, vmid int64
		err       error
	)
	vmidStr := c.Request.Form.Get("vmid")
	if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil || vmid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spcSvc.TopArc(c, mid, vmid))
}

func setTopArc(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid    int64  `form:"aid"`
		Bvid   string `form:"bvid"`
		Reason string `form:"reason"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	reason := strings.TrimSpace(v.Reason)
	if len([]rune(reason)) > conf.Conf.Rule.MaxTopReasonLen {
		c.JSON(nil, ecode.TopReasonLong)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, spcSvc.SetTopArc(c, mid, aid, reason))
}

func cancelTopArc(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, spcSvc.DelTopArc(c, mid))
}

func masterpiece(c *bm.Context) {
	var (
		mid, vmid int64
		err       error
	)
	vmidStr := c.Request.Form.Get("vmid")
	if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil || vmid <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spcSvc.Masterpiece(c, mid, vmid))
}

func addMasterpiece(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid    int64  `form:"aid"`
		Bvid   string `form:"bvid"`
		Reason string `form:"reason"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	reason := strings.TrimSpace(v.Reason)
	if len([]rune(reason)) > conf.Conf.Rule.MaxMpReasonLen {
		c.JSON(nil, ecode.TopReasonLong)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, spcSvc.AddMasterpiece(c, mid, aid, reason))
}

func editMasterpiece(c *bm.Context) {
	var (
		aid, preAid int64
		err         error
	)
	v := new(struct {
		Aid     int64  `form:"aid"`
		Bvid    string `form:"bvid"`
		PreAid  int64  `form:"pre_aid"`
		PreBvid string `form:"pre_bvid"`
		Reason  string `form:"reason"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	if preAid, err = bvArgCheck(v.PreAid, v.PreBvid); err != nil {
		c.JSON(nil, err)
		return
	}
	reason := strings.TrimSpace(v.Reason)
	if len([]rune(reason)) > conf.Conf.Rule.MaxMpReasonLen {
		c.JSON(nil, ecode.TopReasonLong)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, spcSvc.EditMasterpiece(c, mid, preAid, aid, reason))
}

func cancelMasterpiece(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, spcSvc.CancelMasterpiece(c, mid, aid))
}

func arcSearch(c *bm.Context) {
	var (
		v = new(model.SearchArg)
	)
	if err := c.Bind(v); err != nil {
		return
	}
	if v.CheckType != "" {
		if _, ok := model.ArcCheckType[v.CheckType]; !ok {
			c.JSON(nil, xecode.RequestErr)
			return
		}
		if v.CheckID <= 0 {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	query, _ := json.Marshal(v)
	riskParams := getRiskCommonReq(c, string(query))
	riskParams.Action = "anti_crawler"
	riskParams.Scene = "anti_crawler"
	riskParams.Token = v.Token
	riskParams.VisitRecord = v.Mid
	res, err := spcSvc.ArcSearch(c, riskParams.Mid, v, riskParams)
	if err != nil {
		if !xecode.EqualError(xecode.NothingFound, err) {
			err = xecode.Degrade
		}
		c.JSON(nil, err)
		return
	}
	if res != nil {
		if res.GaiaResType == model.GaiaResponseType_NeedFECheck {
			c.JSON(struct {
				GaData *gaiamdl.RuleCheckReply `json:"ga_data"`
			}{GaData: res.GaiaData}, xecode.Unauthorized)
		} else if res.GaiaResType == model.GaiaResponseType_Reject {
			c.JSON(nil, xecode.Error(xecode.AccessDenied, "账号异常,操作失败"))
		} else if res.GaiaResType == model.GaiaResponseType_TokenInvalid {
			c.JSON(nil, xecode.AccessTokenExpires)
		} else {
			c.JSON(res, nil)
		}
		return
	}
	c.JSON(nil, xecode.RequestErr)
}

func arcList(c *bm.Context) {
	var (
		rs  *model.UpArc
		err error
	)
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
		Pn  int32 `form:"pn" default:"1" validate:"min=1"`
		Ps  int32 `form:"ps" default:"20" validate:"min=1,max=50"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if rs, err = spcSvc.UpArcs(c, v.Mid, v.Pn, v.Ps); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"pn":    int64(v.Pn),
		"ps":    int64(v.Ps),
		"count": rs.Count,
	}
	data["page"] = page
	data["archives"] = rs.List
	c.JSON(data, nil)
}
