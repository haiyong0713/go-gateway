package http

import (
	"encoding/json"
	"io/ioutil"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func createWF(ctx *bm.Context) {
	req := &struct {
		ID         int64  `json:"id"`
		ApiName    string `json:"api_name"`
		OnlyDeploy bool   `json:"only_deploy"`
	}{}
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	defer ctx.Request.Body.Close()
	if err := json.Unmarshal(data, req); err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(svc.CreateWF(ctx, req.ID, req.ApiName, req.OnlyDeploy))
}

func wfStatus(ctx *bm.Context) {
	form := new(struct {
		ApiName string `form:"api_name"`
	})
	if err := ctx.Bind(form); err != nil {
		return
	}
	ctx.JSON(svc.GetWFStatus(ctx, form.ApiName))
}

func resumeWF(ctx *bm.Context) {
	req := &struct {
		ID      int64  `json:"id"`
		ApiName string `json:"api_name"`
	}{}
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	defer ctx.Request.Body.Close()
	if err := json.Unmarshal(data, req); err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, svc.ResumeWF(ctx, req.ID, req.ApiName))
}

func stopWF(ctx *bm.Context) {
	req := &struct {
		ID      int64  `json:"id"`
		ApiName string `json:"api_name"`
	}{}
	data, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	defer ctx.Request.Body.Close()
	if err := json.Unmarshal(data, req); err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	ctx.JSON(nil, svc.StopWF(ctx, req.ID, req.ApiName))
}
