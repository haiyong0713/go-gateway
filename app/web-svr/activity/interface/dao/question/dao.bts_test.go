package question

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQuestionDetail(t *testing.T) {
	Convey("Detail", t, func() {
		var (
			c  = context.Background()
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.Detail(c, id)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestQuestionDetails(t *testing.T) {
	Convey("Details", t, func() {
		var (
			c   = context.Background()
			ids = []int64{}
		)
		Convey("When everything goes positive", func() {
			res, err := d.Details(c, ids)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestQuestionLastQuesLog(t *testing.T) {
	Convey("LastQuesLog", t, func() {
		var (
			c      = context.Background()
			mid    = int64(0)
			baseID = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.LastQuesLog(c, mid, baseID)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}
