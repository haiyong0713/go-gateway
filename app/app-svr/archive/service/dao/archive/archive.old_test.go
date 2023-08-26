package archive

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestVideosByAidCids(t *testing.T) {
	var (
		c       = context.TODO()
		aidCids = make(map[int64][]int64)
	)

	aidCids[320006469] = []int64{10342767, 10342768, 10342776}
	aidCids[320006468] = []int64{}

	convey.Convey("VideosByAidCids", t, func(ctx convey.C) {
		_, err := d.VideosByAidCids(c, aidCids)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
