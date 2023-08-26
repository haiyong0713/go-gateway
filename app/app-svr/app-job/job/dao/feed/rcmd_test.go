package feed

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestFeedHots(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("Hots", t, func(ctx convey.C) {
		httpMock("GET", d.hot).Reply(200).
			JSON(`{"note":false,"source_date":"2019-02-11","code":0,"num":500,"list":[{"aid":42888311,"score":177},{"aid":42557612,"score":172}]}`)
		aids, err := d.Hots(c)
		ctx.Convey("Then err should be nil.aids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(aids, convey.ShouldNotBeNil)
		})
	})
}
