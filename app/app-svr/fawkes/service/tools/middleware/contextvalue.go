package middleware

import (
	"context"

	bm "go-common/library/net/http/blademaster"

	toolmdl "go-gateway/app/app-svr/fawkes/service/model/tool"
)

func ContextValues() bm.HandlerFunc {
	return func(c *bm.Context) {
		v := &toolmdl.ContextValues{}
		file, fileHeader, err := c.Request.FormFile("file")
		if err == nil {
			v.File = file
			v.FileHeader = fileHeader
		}
		if name, ok := c.Get("username"); ok {
			v.Username = name.(string)
		}
		c.Context = context.WithValue(c.Context, toolmdl.ContentKey, v)
		c.Next()
	}
}
