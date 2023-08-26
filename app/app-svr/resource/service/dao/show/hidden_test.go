package show

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

// TestHiddens test
func TestHiddens(t *testing.T) {
	convey.Convey("TestHiddens", t, func(ctx convey.C) {
		ctx.Convey("When everyting is correct", func(ctx convey.C) {
			rly, limits, err := d.Hiddens(context.Background(), time.Now())
			ctx.Convey("Error should be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				s, _ := json.Marshal(rly)
				l, _ := json.Marshal(limits)
				ctx.Printf("s:%s l:%s", s, l)
			})
		})
	})
}
