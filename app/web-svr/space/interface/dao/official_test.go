package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	pb "go-gateway/app/web-svr/space/interface/api/v1"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_Official(t *testing.T) {
	convey.Convey("privacyHit", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			data, err := d.RawOfficial(context.Background(), &pb.OfficialRequest{Mid: 1})
			ctx.Convey("Then err should be nil.data should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(data, convey.ShouldNotBeNil)
				bs, _ := json.Marshal(data)
				fmt.Println(string(bs))
			})
		})
	})
}
