package steins

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoNodes(t *testing.T) {
	convey.Convey("Nodes", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			keys = []int64{1, 2, 5621, 5622, 5623, 5624, 5625, 5626, 5627}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Nodes(c, keys)
			fmt.Println(res, err)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoNode(t *testing.T) {
	convey.Convey("Node", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			key = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.Node(c, key)
			fmt.Println(res, err)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
