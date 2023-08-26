package question

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQuestiondetailKey(t *testing.T) {
	Convey("detailKey", t, func() {
		var (
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := detailKey(id)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestQuestionlastLogKey(t *testing.T) {
	Convey("lastLogKey", t, func() {
		var (
			mid    = int64(0)
			baseID = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := lastLogKey(mid, baseID)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}
