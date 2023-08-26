package search

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestService_OgvList(t *testing.T) {
	convey.Convey("SetDarkPub", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			param := &show.SearchOgvLP{
				Pn:    1,
				Ps:    10,
				Stime: 1577808000,
				Etime: 1577808000,
			}
			res, err := s.OgvList(context.Background(), param)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}
