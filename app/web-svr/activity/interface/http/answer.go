package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/service"
)

func answerUserInfo(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.AnswerUserInfo(c, mid))
}

func answerQuestion(c *bm.Context) {
	p := new(like.ParamQuestion)
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.AnswerQuestion(c, mid, p))
}

func weekTop(c *bm.Context) {
	c.JSON(service.LikeSvc.WeekTop(c))
}

func answerResult(c *bm.Context) {
	p := new(like.ParamResult)
	if err := c.Bind(p); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	request := c.Request
	p.UA = request.UserAgent()
	p.Referer = request.Referer()
	p.IP = metadata.String(c, metadata.RemoteIP)
	buvid := request.Header.Get(_headerBuvid)
	if buvid == "" {
		cookie, _ := request.Cookie(_buvid)
		if cookie != nil {
			buvid = cookie.Value
		}
	}
	p.Buvid = buvid
	p.Origin = request.Header.Get("Origin")
	c.JSON(service.LikeSvc.AnswerResult(c, mid, p))
}

func answerRank(c *bm.Context) {
	c.JSON(service.LikeSvc.AnswerRank(c))
}

func answerPendant(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.AnswerPendant(c, mid))
}

func knowRule(c *bm.Context) {
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.KnowRule(c, mid))
}

func shareAddHP(c *bm.Context) {
	v := new(struct {
		CurrentRound int64 `form:"current_round" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, service.LikeSvc.ShareAddHP(c, mid, v.CurrentRound))
}
