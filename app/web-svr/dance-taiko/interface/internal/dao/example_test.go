package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaopickExamples(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(4444)
	)
	convey.Convey("PickExamples", t, func(ctx convey.C) {
		p1, err := d.PickExamples(c, aid)
		s, _ := json.Marshal(p1)
		fmt.Println(string(s))
		ctx.Convey("Then err should be nil.p1 should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}
