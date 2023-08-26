package http

import (
	"encoding/json"
	"net/http"

	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/management-job/api"

	"github.com/pkg/errors"
)

func taskDo(ctx *bm.Context) {
	req := &api.TaskDoReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	if err := parseParams(req, ctx.Request); err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(svc.TaskDo(ctx, req))
}

func parseParams(dst *api.TaskDoReq, req *http.Request) error {
	params := req.Form.Get("params")
	if params == "" {
		return nil
	}
	dst.Params = &api.Params{}
	if err := json.Unmarshal([]byte(params), dst.Params); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func rawConfig(ctx *bm.Context) {
	req := &api.RawConfigReq{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(svc.RawConfig(ctx, req))
}
