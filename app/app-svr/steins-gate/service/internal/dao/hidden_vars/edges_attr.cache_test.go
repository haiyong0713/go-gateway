package hidden_vars

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoedgeAttrsCache(t *testing.T) {
	convey.Convey("edgeAttrsCache", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			graphID = int64(441)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			attrs, err := d.edgeAttrsCache(c, graphID)
			fmt.Println(attrs, err)
			convCtx.Convey("Then err should be nil.attrs should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(attrs, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoedgeAttrsByGraph(t *testing.T) {
	convey.Convey("edgeAttrsByGraph", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			graphID = int64(441)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			edgeAttrs, err := d.edgeAttrsByGraph(c, graphID)
			estr, _ := json.Marshal(edgeAttrs)
			fmt.Println(string(estr))
			time.Sleep(3 * time.Second)
			convCtx.Convey("Then err should be nil.edgeAttrs should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(edgeAttrs, convey.ShouldNotBeNil)
			})
		})
	})
}
