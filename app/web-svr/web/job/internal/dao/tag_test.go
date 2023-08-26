package dao

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_TagHots(t *testing.T) {
	var (
		rid int64 = 13
	)
	Convey("TagHots", t, func(convCtx C) {
		Convey("When everything goes positive", func(convCtx C) {
			data, err := d.TagHots(ctx, rid)
			So(err, ShouldBeNil)
			So(len(data), ShouldBeGreaterThan, 0)
		})
	})
}
