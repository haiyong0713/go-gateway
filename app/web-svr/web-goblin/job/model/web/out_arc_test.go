package web

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPushArc_ForbidArc(t *testing.T) {
	Convey("ForbidArc", t, func() {
		arc1 := &PushArc{Title: "aa啊aa"}
		forbid1 := arc1.ForbidArc()
		So(forbid1, ShouldBeTrue)
		arc2 := &PushArc{Title: "aaa啊aa"}
		forbid2 := arc2.ForbidArc()
		So(forbid2, ShouldBeFalse)
	})
}
