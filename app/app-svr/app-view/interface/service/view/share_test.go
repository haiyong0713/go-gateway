package view

import (
	"context"
	"fmt"
	"testing"

	"github.com/glycerine/goconvey/convey"
	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/app-view/interface/model/share"
)

func TestShareInfo(t *testing.T) {
	params := &share.InfoParam{
		Bvid: "240042029",
		Mid:  27515316,
	}
	Convey("ShareInfo", t, func(ctx convey.C) {
		res, err := s.ShareInfo(context.Background(), params)
		fmt.Printf("=====%#v======", res)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
