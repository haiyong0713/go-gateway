package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/question"
	"go-gateway/app/web-svr/activity/interface/service"
	"go-gateway/ecode"
)

func questionStart(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.QuestionStart(c, v.Sid, mid))
}

func questionAnswer(c *bm.Context) {
	v := new(struct {
		Sid        int64  `form:"sid" validate:"min=1"`
		QuestionID int64  `form:"id" validate:"min=1"`
		PoolID     int64  `form:"pool_id" validate:"min=1"`
		Index      int64  `form:"index" validate:"min=1"`
		Answer     string `form:"answer"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	data, err := service.LikeSvc.Answer(c, v.Sid, v.PoolID, mid, v.QuestionID, v.Index, v.Answer)
	if err == nil && data.Finish == 1 {
		if func() bool {
			for _, sid := range conf.Conf.Cpc100.QuestionSid {
				if sid == v.Sid {
					return true
				}
			}
			return false
		}() {
			err = service.TaskSvr.ActSend(c, mid, conf.Conf.Cpc100.QuestionBusiness, conf.Conf.Cpc100.Activity, data.AnswerTime)
		}
	}
	c.JSON(data, err)
}

func questionNext(c *bm.Context) {
	v := new(struct {
		Sid    int64 `form:"sid" validate:"min=1"`
		PoolID int64 `form:"pool_id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.Next(c, v.Sid, v.PoolID, mid))
}

func questionMyRecords(c *bm.Context) {
	v := new(struct {
		Sids  []int64 `form:"sids,split" validate:"required"`
		State []int64 `form:"state,split" validate:"required"`
		Pn    int64   `form:"pn" default:"1"`
		Ps    int64   `form:"ps" default:"10"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(service.LikeSvc.MyQuestionRecords(c, mid, v.Sids, v.State, v.Pn, v.Ps))
}

// questionAndAnswer 返回问题和答案
func questionAndAnswer(c *bm.Context) {
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.QuestionAnswer(c, v.Sid))
}

func gaokaoQuestion(c *bm.Context) {
	v := new(question.GKQuestReq)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.GKQuestion(c, v))
}

func gaokaoRank(c *bm.Context) {
	v := new(question.GKRankReq)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(service.LikeSvc.GKRank(c, v))
}

func gaokaoReport(c *bm.Context) {
	v := new(question.GKRankReq)
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	if mid < 0 {
		c.JSON(nil, ecode.ReqParamErr)
		return
	}
	c.JSON(service.LikeSvc.GKReportScore(c, mid, v))
}
