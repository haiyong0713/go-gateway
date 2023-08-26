package show

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

// TestIcons test
func TestIcons(t *testing.T) {
	convey.Convey("TestIcons", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			rly, err := d.Icons(context.Background(), time.Now(), time.Now())
			ctx.Convey("Error should be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				s, _ := json.Marshal(rly)
				ctx.Printf("s:%s", s)
			})
		})
	})
}
