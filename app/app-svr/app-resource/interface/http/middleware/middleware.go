package middleware

import (
	"strconv"
	"time"

	bm "go-common/library/net/http/blademaster"
)

const (
	_headerAppTimestamp = "x-bili-app-ts"
)

func InjectTimestamp() bm.HandlerFunc {
	return func(ctx *bm.Context) {
		now := time.Now()
		ts := strconv.FormatInt(now.Unix(), 10)

		header := ctx.Writer.Header()
		header.Set(_headerAppTimestamp, ts)
	}
}
