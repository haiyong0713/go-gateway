package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/render"
	"go-gateway/app/api-gateway/api-manager/internal/model"
	"net/http"
	"strconv"
)

func codeGenerate(ctx *bm.Context) {
	form := new(model.CodeGeneratorReq)
	if err := ctx.Bind(form); err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	reply, err := svc.GenerateCode(ctx, form.ApiID)
	bcode := 0
	message := ""
	if err != nil {
		log.Errorc(ctx, "generate code return err %v", err)
		bcode = ecode.ServerErr.Code()
		message = err.Error()
	}

	jsonRender(ctx, bcode, message, reply)
}

func jsonRender(ctx *bm.Context, bcode int, message string, data interface{}) {
	header := ctx.Writer.Header()
	header.Set("bili-status-code", strconv.FormatInt(int64(bcode), 10))
	ctx.Render(http.StatusOK, render.JSON{
		Code:    bcode,
		Message: message,
		Data:    data,
	})
}
