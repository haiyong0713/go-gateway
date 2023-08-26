package search

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestNavigationCards(t *testing.T) {
	convey.Convey("TestNavigationCards", t, func(ctx convey.C) {
		var (
			cardId int64 = 1
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			cardMap, err := d.NavigationCards([]int64{cardId})
			fmt.Printf("card(%+v)\n", cardMap[cardId])
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
