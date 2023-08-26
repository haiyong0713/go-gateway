package search

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	model "go-gateway/app/app-svr/app-feed/admin/model/search"

	"github.com/smartystreets/goconvey/convey"
)

func TestService_BrandBlacklistList(t *testing.T) {
	convey.Convey("SetDarkPub", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			param := &model.BrandBlacklistListReq{
				Username: "litongyu",
				Uid:      1,
				Pn:       1,
				Ps:       20,
			}
			res, err := s.BrandBlacklistList(context.Background(), param)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}
