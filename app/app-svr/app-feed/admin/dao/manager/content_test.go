package manager

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_ContentCards(t *testing.T) {
	convey.Convey("Contents", t, func(ctx convey.C) {
		var (
			bs []byte
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.ContentCards([]int64{20, 21, 22})
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			bs, err = json.Marshal(res)
			if err != nil {
				return
			}
			fmt.Println(string(bs))
		})

	})
}
