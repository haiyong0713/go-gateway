package http

import (
	"context"
	"encoding/json"
	gaiamdl "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/interface"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/interface/conf"
	"go-gateway/app/web-svr/space/interface/model"
)

const (
	_headerBuvid = "Buvid"
	_buvid       = "buvid3"
)

func navNum(c *bm.Context) {
	var (
		vmid, mid int64
		err       error
	)
	midStr := c.Request.Form.Get("mid")
	if vmid, err = strconv.ParseInt(midStr, 10, 64); err != nil || vmid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spcSvc.NavNum(c, mid, vmid), nil)
}

func upStat(c *bm.Context) {
	var (
		mid int64
		err error
	)
	midStr := c.Request.Form.Get("mid")
	if mid, err = strconv.ParseInt(midStr, 10, 64); err != nil || mid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var logMid int64
	if midInter, ok := c.Get("mid"); ok {
		logMid = midInter.(int64)
	}
	// not log in
	if logMid <= 0 {
		c.JSON(struct{}{}, nil)
		return
	}
	c.JSON(spcSvc.UpStat(c, mid))
}

func myInfo(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(spcSvc.MyInfo(c, mid))
}

func notice(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"gt=0"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(spcSvc.Notice(c, v.Mid))
}

func setNotice(c *bm.Context) {
	v := new(struct {
		Notice string `form:"notice"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	notice := strings.Trim(v.Notice, " ")
	if len([]rune(notice)) > conf.Conf.Rule.MaxNoticeLen {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, spcSvc.SetNotice(c, mid, notice))
}

func accTags(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(spcSvc.AccTags(c, v.Mid))
}

func setAccTags(c *bm.Context) {
	v := new(struct {
		Tags []string `form:"tags,split"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	var addTags []string
	if len(v.Tags) != 0 {
		for _, v := range v.Tags {
			if tag := strings.TrimSpace(v); tag != "" {
				addTags = append(addTags, tag)
			}
		}
		if len(addTags) > 0 {
			if err := spcSvc.Filter(c, addTags); err != nil {
				c.JSON(nil, err)
				return
			}
		}
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, spcSvc.SetAccTags(c, mid, v.Tags))
}

// nolint:bilirailguncheck
func accInfo(c *bm.Context) {
	var (
		vmid int64
		err  error
	)
	vmidStr := c.Request.Form.Get("mid")
	if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil || vmid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	v := new(struct {
		Mid   int64  `form:"mid" validate:"gt=0"`
		Token string `form:"token"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	query, _ := json.Marshal(v)
	riskParams := getRiskCommonReq(c, string(query))
	riskParams.Action = "anti_crawler"
	riskParams.Scene = "anti_crawler"
	riskParams.Token = v.Token
	riskParams.VisitRecord = vmid
	accData, err := spcSvc.AccInfo(c, riskParams.Mid, vmid, riskParams)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	if accData != nil {
		if accData.GaiaResType == model.GaiaResponseType_NeedFECheck {
			c.JSON(struct {
				GaData *gaiamdl.RuleCheckReply `json:"ga_data"`
			}{GaData: accData.GaiaData}, ecode.Unauthorized)
		} else if accData.GaiaResType == model.GaiaResponseType_Reject {
			c.JSON(nil, ecode.Error(ecode.AccessDenied, "账号异常,操作失败"))
		} else if accData.GaiaResType == model.GaiaResponseType_TokenInvalid {
			c.JSON(nil, ecode.AccessTokenExpires)
		} else {
			c.JSON(accData, nil)
			// report
			if riskParams.Mid > 0 {
				if err := visitPub.Send(context.Background(), strconv.FormatInt(riskParams.Mid, 10), &model.VisitAct{
					LoginMid: riskParams.Mid,
					Mid:      vmid,
					Referer:  c.Request.Referer(),
					Buvid:    reqBuvid(c),
					Path:     "/x/space/acc/info",
					Ts:       time.Now().Unix(),
				}); err != nil {
					log.Error("%+v", err)
				}
			}
		}
		return
	}
	c.JSON(nil, ecode.RequestErr)
}

func lastPlayGame(c *bm.Context) {
	var (
		mid, vmid int64
		err       error
	)
	vmidStr := c.Request.Form.Get("mid")
	if vmid, err = strconv.ParseInt(vmidStr, 10, 64); err != nil || vmid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(spcSvc.LastPlayGame(c, mid, vmid))
}

func themeList(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(spcSvc.ThemeList(c, mid))
}

func themeActive(c *bm.Context) {
	var (
		themeID int64
		err     error
	)
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	themeIDStr := c.Request.Form.Get("theme_id")
	if themeID, err = strconv.ParseInt(themeIDStr, 10, 64); err != nil || themeID <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, spcSvc.ThemeActive(c, mid, themeID))
}

func relation(c *bm.Context) {
	v := new(struct {
		Vmid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(spcSvc.Relation(c, mid, v.Vmid), nil)
}

func clearCache(c *bm.Context) {
	v := new(struct {
		Msg string `form:"msg" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, spcSvc.ClearCache(c, v.Msg))
}

func clearMsg(c *bm.Context) {
	v := new(struct {
		Type int   `form:"type" validate:"min=1"`
		Mid  int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, spcSvc.ClearMsgCache(c, v.Type, v.Mid))
}

func reqBuvid(c *bm.Context) string {
	buvid := c.Request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := c.Request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	return buvid
}

func clearTopPhotoArc(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, spcSvc.ClearTopPhotoArcByMid(c, v.Mid))
}
