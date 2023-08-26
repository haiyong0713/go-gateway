package http

import (
	"encoding/json"
	gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	"go-common/library/ecode"
	"strconv"
	"time"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"
)

func like(c *bm.Context) {
	var (
		aid int64
		err error
	)
	v := new(struct {
		Aid    int64  `form:"aid"`
		Bvid   string `form:"bvid"`
		Like   int8   `form:"like" validate:"min=1,max=4,required"`
		EabX   int8   `form:"eab_x" default:"0"`
		Ramval int64  `form:"ramval" default:"0"`
		Token  string `form:"tk"`
		Source string `form:"source"`
		Gaia   int    `form:"ga"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
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
	res, err := webSvc.Like(c, aid, riskParams.Mid, v.Like, riskParams)
	if err == nil || res != nil && res.IsRisk {
		isRiskInt := "0"
		if res.IsRisk {
			isRiskInt = "1"
		}
		webSvc.InfocV2(model.UserActInfoc{
			Buvid:    riskParams.Buvid,
			Client:   "web",
			Ip:       metadata.String(c, metadata.RemoteIP),
			Uid:      res.UpID,
			Aid:      aid,
			Mid:      riskParams.Mid,
			Sid:      reqSid(c),
			Refer:    riskParams.Referer,
			Url:      riskParams.Api,
			ItemID:   strconv.FormatInt(aid, 10),
			ItemType: "av",
			Action:   model.LikeType[v.Like],
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
			c.JSON(nil, err)
		}
		return
	}
	c.JSON(nil, ecode.RequestErr)
}

func likeTriple(c *bm.Context) {
	var (
		aid       int64
		actTriple = "triplelike"
		err       error
	)
	v := new(struct {
		Aid    int64  `form:"aid"`
		Bvid   string `form:"bvid"`
		EabX   int8   `form:"eab_x" default:"0"`
		Ramval int64  `form:"ramval" default:"0"`
		Token  string `form:"tk"`
		Source string `form:"source"`
		Gaia   int    `form:"ga"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
		c.JSON(nil, err)
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
	res, err := webSvc.LikeTriple(c, aid, riskParams.Mid, riskParams)
	if err != nil {
		return
	}
	if res != nil && res.Anticheat {
		isRiskInt := "0"
		if res.IsRisk {
			isRiskInt = "1"
		}
		webSvc.InfocV2(model.UserActInfoc{
			Buvid:    riskParams.Buvid,
			Client:   "web",
			Ip:       metadata.String(c, metadata.RemoteIP),
			Uid:      res.UpID,
			Aid:      aid,
			Mid:      riskParams.Mid,
			Sid:      reqSid(c),
			Refer:    riskParams.Referer,
			Url:      riskParams.Api,
			ItemID:   strconv.FormatInt(aid, 10),
			ItemType: "av",
			Action:   actTriple,
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
			c.JSON(res, err)
		}
		return
	}
	c.JSON(nil, ecode.RequestErr)
}

func hasLike(c *bm.Context) {
	var (
		aid, mid int64
		err      error
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
	mid = midStr.(int64)
	c.JSON(webSvc.HasLike(c, aid, mid))
}

func upLikeImg(ctx *bm.Context) {
	v := new(struct {
		Vmid      int64 `form:"vmid" validate:"min=1"`
		Following bool  `form:"following"`
		Aid       int64 `form:"aid"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var mid int64
	midStr, ok := ctx.Get("mid")
	if ok {
		mid = midStr.(int64)
	}
	ctx.JSON(webSvc.UpLikeImg(ctx, mid, v.Vmid, v.Aid, v.Following))
}
