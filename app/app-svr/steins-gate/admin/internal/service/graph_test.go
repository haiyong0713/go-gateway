package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestService_GraphShow(t *testing.T) {
	var (
		c       = context.Background()
		graphid = int64(3333)
	)
	convey.Convey("GraphShow", t, func(ctx convey.C) {

		nodeShow, err := s.GraphShow(c, graphid)
		fmt.Println(nodeShow, err)
		ctx.Convey("Then err should be nil.nodeShow should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(nodeShow, convey.ShouldNotBeNil)
		})
	})
}
