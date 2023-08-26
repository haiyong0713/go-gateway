package utils

import (
	"context"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/net/trace"

	toolmdl "go-gateway/app/app-svr/fawkes/service/model/tool"
)

// CopyTrx 返回一个background context 但是会保留原本的trace信息
func CopyTrx(ctx context.Context) context.Context {
	c := metadata.WithContext(ctx)
	tr, ok := trace.FromContext(ctx)
	if ok {
		c = trace.NewContext(c, tr)
	}
	return c
}

// GetUsername 从bm.Context 或者 context.Context中获取用户名
func GetUsername(ctx interface{}) (username string) {
	switch c := ctx.(type) {
	case *bm.Context:
		username = getUserFromBMContext(c)
	case bm.Context:
		username = getUserFromBMContext(&c)
	case context.Context:
		username = getUserFromContext(c)
	}

	return
}

func getUserFromBMContext(ctx *bm.Context) (username string) {
	if v := ctx.Value(toolmdl.ContentKey); v != nil {
		ctxValue := v.(*toolmdl.ContextValues)
		if ctxValue.FromOpenAPI {
			return ctxValue.OpenAPIUser
		} else {
			return ctxValue.Username
		}
	} else {
		if user, isVerify := ctx.Get("username"); isVerify {
			return user.(string)
		}
	}
	return
}

func getUserFromContext(ctx context.Context) (username string) {
	if v := ctx.Value(toolmdl.ContentKey); v != nil {
		ctxValue := v.(*toolmdl.ContextValues)
		if ctxValue.FromOpenAPI {
			return ctxValue.OpenAPIUser
		}
		return ctxValue.Username
	}
	return ctx.Value("username").(string)
}
