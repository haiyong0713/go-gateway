package http

import (
	"encoding/json"
	gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	"net/http"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"
	chmdl "go-gateway/app/web-svr/web/interface/model/channel"
)

const (
	_headerBuvid = "Buvid"
	_buvid       = "buvid3"
)

func view(c *bm.Context) {
	var (
		mid, cid int64
		err      error
		rs       *model.View
	)
	v := new(struct {
		Aid        int64  `form:"aid"`
		Bvid       string `form:"bvid"`
		OutReferer string `form:"out_referer"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	// get mid
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	cidStr := c.Request.Form.Get("cid")
	if cidStr != "" {
		if cid, err = strconv.ParseInt(cidStr, 10, 64); err != nil || cid < 0 {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	cdnIP := c.Request.Header.Get("X-Cache-Server-Addr")
	if rs, _, err = webSvc.View(c, v.Aid, cid, mid, cdnIP, v.OutReferer, ""); err != nil {
		if webSvc.SLBRetry(err) {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Error("%+v", err)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func archiveStat(c *bm.Context) {
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
	c.JSON(webSvc.ArchiveStat(c, aid))
}

func addShare(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid          int64  `form:"aid"`
		Bvid         string `form:"bvid"`
		RoomID       int64  `form:"room_id"`
		UpID         int64  `form:"up_id"`
		ParentAreaID int64  `form:"parent_area_id"`
		AreaID       int64  `form:"area_id"`
		EabX         int8   `form:"eab_x" default:"0"`
		Ramval       int64  `form:"ramval" default:"0"`
		Token        string `form:"tk"`
		Source       string `form:"source"`
		Gaia         int    `form:"ga"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	aid, err = bvArgCheck(v.Aid, v.Bvid)
	if err != nil && v.RoomID <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	query, _ := json.Marshal(v)
	riskParams := getRiskCommonReq(c, string(query))
	riskParams.Token = v.Token
	riskParams.Source = v.Source
	riskParams.EabX = v.EabX
	riskParams.Ramval = v.Ramval
	riskParams.Gaia = v.Gaia
	riskParams.Avid = v.Aid
	res, err := webSvc.AddShare(c, aid, riskParams.Mid, v.RoomID, v.UpID, v.ParentAreaID, v.AreaID, riskParams)
	// add Anti cheat
	if err == nil || res != nil && res.IsRisk {
		isRiskInt := "0"
		if res.IsRisk {
			isRiskInt = "1"
		}
		itemType := "av"
		if v.RoomID > 0 {
			itemType = "live"
		}
		webSvc.InfocV2(model.UserActInfoc{
			Buvid:    riskParams.Buvid,
			Client:   "web",
			Ip:       metadata.String(c, metadata.RemoteIP),
			Uid:      v.UpID,
			Aid:      aid,
			Mid:      riskParams.Mid,
			Sid:      reqSid(c),
			Refer:    riskParams.Referer,
			Url:      riskParams.Api,
			ItemID:   strconv.FormatInt(aid, 10),
			ItemType: itemType,
			Action:   "share",
			Ua:       riskParams.UserAgent,
			Ts:       strconv.FormatInt(time.Now().Unix(), 10),
			IsRisk:   isRiskInt,
		})
	}
	if res != nil {
		if res.GaiaResType == model.GaiaResponseType_NeedFECheck {
			c.JSON(struct {
				GaData *gaiamdl.RuleCheckReply `json:"ga_data"`
			}{GaData: res.GaiaData}, ecode.Unauthorized)
		} else if res.GaiaResType == model.GaiaResponseType_Reject {
			c.JSON(nil, ecode.Error(ecode.AccessDenied, "账号异常,操作失败"))
		} else if res.GaiaResType == model.GaiaResponseType_TokenInvalid {
			c.JSON(nil, ecode.AccessTokenExpires)
		} else {
			if err != nil {
				c.JSON(nil, err)
				return
			}
			c.JSON(res.Shares, nil)
		}
		return
	}
	c.JSON(nil, ecode.RequestErr)
}

func description(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
		Page int64  `form:"page" validate:"min=0"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(webSvc.Description(c, aid, v.Page))
}

func desc2(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
		Page int64  `form:"page" validate:"min=0"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(webSvc.Desc2(c, aid, v.Page))
}

func arcReport(c *bm.Context) {
	var (
		aid, mid int64
		err      error
	)
	v := new(struct {
		Aid    int64  `form:"aid"`
		Bvid   string `form:"bvid"`
		Type   int64  `form:"type"`
		Reason string `form:"reason"`
		Pics   string `form:"pics"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	c.JSON(nil, webSvc.ArcReport(c, mid, aid, v.Type, v.Reason, v.Pics))
}

func appealTags(c *bm.Context) {
	c.JSON(webSvc.AppealTags(c))
}

func arcAppeal(c *bm.Context) {
	var (
		mid, aid int64
		err      error
	)
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
		Tid  int64  `form:"tid" validate:"min=1"`
		Desc string `form:"desc" validate:"required"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	params := c.Request.Form
	midStr, _ := c.Get("mid")
	mid = midStr.(int64)
	data := make(map[string]string)
	data["oid"] = strconv.FormatInt(aid, 10)
	for name := range params {
		switch name {
		case "tid":
			data["tid"] = strconv.FormatInt(v.Tid, 10)
		case "desc":
			data["description"] = v.Desc
		default:
			data[name] = params.Get(name)
		}
	}
	c.JSON(nil, webSvc.ArcAppeal(c, mid, data))
}

func authorRecommend(c *bm.Context) {
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
	c.JSON(webSvc.AuthorRecommend(c, aid))
}

func relatedArcs(c *bm.Context) {
	var (
		aid, mid int64
		err      error
	)
	v := new(struct {
		Aid               int64  `form:"aid"`
		Bvid              string `form:"bvid"`
		NeedOperationCard int    `form:"need_operation_card"`
		WebRmRepeat       int    `form:"web_rm_repeat"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	list, card, _, err := webSvc.RelatedArcs(c, aid, mid, reqBuvid(c), false, v.NeedOperationCard == 1, v.WebRmRepeat == 1, false, nil)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{})
	data["data"] = list
	if card != nil {
		data["spec"] = card
	}
	c.JSONMap(data, nil)
}

func detail(c *bm.Context) {
	var (
		mid int64
		err error
		rs  *model.Detail
	)
	v := new(struct {
		Aid               int64  `form:"aid"`
		Bvid              string `form:"bvid"`
		NeedOperationCard int    `form:"need_operation_card"`
		WebRmRepeat       int    `form:"web_rm_repeat"`
		NeedHootShare     int    `form:"need_hot_share"`
		NeedElec          int    `form:"need_elec"`
		OutReferer        string `form:"out_referer"`
		RecommendType     string `form:"recommend_type"`
		NeedRcmdReason    int    `form:"need_rcmd_reason"`
		Platform          string `form:"platform"`
		PageNo            int    `form:"page_no" default:"1" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	cdnIP := c.Request.Header.Get("X-Cache-Server-Addr")
	if rs, err = webSvc.Detail(c, v.Aid, mid, cdnIP, v.OutReferer, reqBuvid(c), v.RecommendType, v.Platform, v.NeedRcmdReason == 1, v.NeedOperationCard == 1, v.WebRmRepeat == 1, v.NeedHootShare == 1, v.NeedElec == 1, v.PageNo); err != nil {
		if webSvc.SLBRetry(err) {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Error("%+v", err)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func detailTag(c *bm.Context) {
	var (
		mid int64
		err error
		rs  []*chmdl.VideoTag
	)
	v := new(struct {
		Aid             int64  `form:"aid"`
		Bvid            string `form:"bvid"`
		Cid             int64  `form:"cid"`
		IsH5Subdivision int    `form:"is_h5_subdivision"` // H5唤端实验，需要up主的tag和相关视频的tag
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	if rs, err = webSvc.DetailTag(c, v.Aid, mid, v.Cid, nil, v.IsH5Subdivision == 1); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(rs, nil)
}

func arcUGCPay(c *bm.Context) {
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
	c.JSON(webSvc.ArcUGCPay(c, mid, aid))
}

func arcRelation(c *bm.Context) {
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
	data, err := webSvc.ArcRelation(c, mid, aid)
	if err != nil {
		if webSvc.SLBRetry(err) {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Error("%+v", err)
			return
		}
		c.JSON(nil, err)
		return
	}
	c.JSON(data, nil)
}

func arcSpecRcmd(c *bm.Context) {
	c.JSON(nil, ecode.NothingFound)
}

func avConfig(c *bm.Context) {
	res := map[string]interface{}{
		"show_bv": true,
	}
	c.JSON(res, nil)
}

func arcCustomConfig(ctx *bm.Context) {
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	aid, err := bvArgCheck(v.Aid, v.Bvid)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(webSvc.ArcCustomConfig(ctx, aid))
}

func arcPremiere(c *bm.Context) {
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
	})
	var err error
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	res, err := webSvc.Premiere(c, v.Aid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(res, nil)
}

func arcPremiereInfo(c *bm.Context) {
	v := new(struct {
		Aid  int64  `form:"aid"`
		Bvid string `form:"bvid"`
	})
	var err error
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
		return
	}
	res, err := webSvc.PremiereInfo(c, v.Aid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(res, nil)
}
