package show

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestShowRawGoodHisRes(t *testing.T) {
	convey.Convey("RawGoodHisRes", t, func(convCtx convey.C) {
		var (
			ctx = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := dao.RawGoodHisRes(ctx)
			str, _ := json.Marshal(res)
			fmt.Println(string(str), err)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
