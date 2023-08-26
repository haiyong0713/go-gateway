package http

import (
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"

	gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
)

func coins(c *bm.Context) {
	var (
		mid, aid int64
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
	c.JSON(webSvc.Coins(c, mid, aid))
}

func addCoin(c *bm.Context) {
	var (
		err     error
		actLike = "cointolike"
	)
	v := new(struct {
		Aid        int64  `form:"aid"`
		Bvid       string `form:"bvid"`
		Multiply   int64  `form:"multiply" validate:"min=1"`
		Avtype     int64  `form:"avtype"`
		Business   string `form:"business"`
		Upid       int64  `form:"upid"`
		SelectLike int    `form:"select_like"`
		Gaia       int    `form:"ga"`
		Token      string `form:"tk"`
		Source     string `form:"source"`
		EabX       int8   `form:"eab_x" default:"0"`
		Ramval     int64  `form:"ramval" default:"0"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Avtype == 0 {
		v.Avtype = model.CoinAddArcType
	}
	if v.Avtype != model.CoinAddArcType && v.Avtype != model.CoinAddArtType {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if v.Avtype == model.CoinAddArtType && v.Upid == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if v.Avtype == model.CoinAddArcType {
		if v.Aid, err = bvArgCheck(v.Aid, v.Bvid); err != nil {
			c.JSON(nil, err)
			return
		}
	} else if v.Aid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if v.Business != "" {
		if v.Business == model.CoinArcBusiness {
			v.Avtype = model.CoinAddArcType
		} else if v.Business == model.CoinArtBusiness {
			v.Avtype = model.CoinAddArtType
		} else {
			c.JSON(nil, ecode.RequestErr)
			return
		}
	} else {
		switch v.Avtype {
		case model.CoinAddArcType:
			v.Business = model.CoinArcBusiness
		case model.CoinAddArtType:
			v.Business = model.CoinArtBusiness
		default:
			c.JSON(nil, ecode.RequestErr)
			return
		}
	}
	query, _ := json.Marshal(v)
	riskParams := getRiskCommonReq(c, string(query))
	riskParams.Token = v.Token
	riskParams.Source = v.Source
	riskParams.EabX = v.EabX
	riskParams.Ramval = v.Ramval
	riskParams.Gaia = v.Gaia
	riskParams.Avid = v.Aid
	riskParams.UpMid = v.Upid
	riskParams.CoinNum = v.Multiply
	res, err := webSvc.AddCoin(c, v.Aid, riskParams.Mid, v.Upid, v.Multiply, v.Avtype, v.Business, v.SelectLike, riskParams)
	likeRes := false
	isLikeRisk := false
	if res != nil {
		likeRes = res.Like
		isLikeRisk = res.IsRisk
	}
	if err == nil || isLikeRisk {
		itemType := "av"
		if v.Avtype == model.CoinAddArtType {
			itemType = "article"
		}
		isRiskInt := "0"
		if isLikeRisk {
			isRiskInt = "1"
		}
		infocData := model.UserActInfoc{
			Buvid:    riskParams.Buvid,
			Client:   "web",
			Ip:       metadata.String(c, metadata.RemoteIP),
			Uid:      v.Upid,
			Aid:      v.Aid,
			Mid:      riskParams.Mid,
			Sid:      reqSid(c),
			Refer:    riskParams.Referer,
			Url:      riskParams.Api,
			ItemID:   strconv.FormatInt(v.Aid, 10),
			ItemType: itemType,
			Action:   "coin",
			Ua:       riskParams.UserAgent,
			Ts:       strconv.FormatInt(time.Now().Unix(), 10),
			IsRisk:   isRiskInt,
		}
		webSvc.InfocV2(infocData)
		if likeRes || isLikeRisk {
			infocData.Action = actLike
			webSvc.InfocV2(infocData)
		}
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
			c.JSON(struct {
				Like bool `json:"like"`
			}{Like: likeRes}, err)
		}
		return
	}
	c.JSON(nil, ecode.RequestErr)
}

func coinExp(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(webSvc.CoinExp(c, mid))
}

func coinList(c *bm.Context) {
	var (
		ls    []*model.CoinArc
		count int
		err   error
	)
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
		Pn  int   `form:"pn" default:"1" validate:"min=1"`
		Ps  int   `form:"ps" default:"20" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if ls, count, err = webSvc.CoinList(c, v.Mid, v.Pn, v.Ps); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSONMap(map[string]interface{}{
		"count": count,
		"data":  ls,
	}, err)
}
