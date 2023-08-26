package show

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestDao_SearchOgv(t *testing.T) {
	convey.Convey("Relate", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			res, err := d.SearchOgv(context.Background())
			ctx.Convey("Error should be nil, res should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}
