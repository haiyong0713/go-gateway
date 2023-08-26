package dynamic

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// BriefDynamics
func TestBriefDynamics(t *testing.T) {
	convey.Convey("BriefDynamics", t, func(convCtx convey.C) {
		var (
			c            = context.Background()
			topicID      = int64(13527)
			frontpageNum = int64(5)
			mid          = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			httpMock("GET", d.briefDynURL).Reply(200).JSON(`{"code":0,"data":{"has_more":1}}`)
			reply, err := d.BriefDynamics(c, topicID, frontpageNum, mid, "", "", 0)
			convCtx.Convey("Then err should be nil.reply should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", reply)
			})
		})
	})
}
