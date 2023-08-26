package search

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/admin/model/search"

	"github.com/smartystreets/goconvey/convey"
)

func TestService_Recommend(t *testing.T) {
	convey.Convey("SetDarkPub", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			param := &search.RecomParam{
				Ts:       time.Now().Unix(),
				Pn:       1,
				Ps:       10,
				CardType: []int{15},
			}
			res, err := s.OpenRecommend(context.Background(), param)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}
