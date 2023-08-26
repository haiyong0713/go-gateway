package like

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/smartystreets/goconvey/convey"
)

func TestItemAndContent(t *testing.T) {
	convey.Convey("ItemAndContent", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			item = &like.Item{Wid: 123, Mid: 1587485, Sid: 10301, Type: 1, State: 0}
			cont = &like.LikeContent{Message: "meeage", IPv6: []byte{}}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ItemAndContent(c, item, cont)
			ctx.Convey("Then err should be nil.likes should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%d", res)
			})
		})
	})
}
