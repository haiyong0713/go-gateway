package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"

	mdl "go-gateway/app/app-svr/fawkes/service/model/app"
)

// FeedbackList show feedbacks
func FeedbackList(c *bm.Context) {
	p := new(mdl.FeedbackReq)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(s.FeedbackSvr.FeedbackList(c, p))
}

// FeedbackAdd add a feedback
func FeedbackAdd(c *bm.Context) {
	p := new(mdl.FeedbackReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.FeedbackSvr.FeedbackAdd(c, p))
}

// FeedbackUpdate update a feedback
func FeedbackUpdate(c *bm.Context) {
	p := &mdl.FeedbackReq{}
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	c.JSON(s.FeedbackSvr.FeedbackUpdate(c, p))
}

// FeedbackInfo show a feedback info by id
func FeedbackInfo(c *bm.Context) {
	p := new(mdl.FeedbackReq)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(s.FeedbackSvr.FeedbackInfo(c, p))
}

// FeedbackDel delete a feedback
func FeedbackDel(c *bm.Context) {
	p := new(mdl.FeedbackReq)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(s.FeedbackSvr.FeedbackDel(c, p))
}

// FeedBackTapdBugCreate
func FeedBackTapdBugCreate(c *bm.Context) {
	var (
		err error
		res = map[string]interface{}{}
	)
	p := &mdl.FeedbackTapdBug{}
	if err := c.Bind(p); err != nil {
		return
	}
	bugID, err := s.FeedbackSvr.FBTapdBugCreate(c, p)
	if err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, err)
		return
	}
	c.JSON(bugID, nil)
}
