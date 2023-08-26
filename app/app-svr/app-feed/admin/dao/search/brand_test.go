package search

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	model "go-gateway/app/app-svr/app-feed/admin/model/search"

	"github.com/smartystreets/goconvey/convey"
)

func TestBrandBlacklistList(t *testing.T) {
	convey.Convey("SetSearchAuditStat", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			total, list, err := d.BrandBlacklistList(c, &model.BrandBlacklistListReq{
				Pn:      1,
				Ps:      20,
				State:   1,
				Keyword: "1",
				Order:   2,
			})
			l, _ := json.Marshal(list)
			fmt.Printf("total: (%v), list: (%v)\n", total, string(l))
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
