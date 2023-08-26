package manager

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_SpecialCards(t *testing.T) {
	convey.Convey("SpecialCards", t, func(ctx convey.C) {
		var (
			bs []byte
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.SpecialCards([]int64{1, 2, 3})
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
