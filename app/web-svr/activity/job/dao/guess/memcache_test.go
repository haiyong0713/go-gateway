package guess

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoCourseList(t *testing.T) {
	Convey("CourseList", t, func() {
		res, err := d.CourseList(context.Background())
		So(err, ShouldBeNil)
		So(len(res), ShouldBeGreaterThan, 0)
	})
}
