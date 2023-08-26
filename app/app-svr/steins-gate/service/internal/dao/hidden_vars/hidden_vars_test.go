package hidden_vars

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoHiddenVars(t *testing.T) {
	var (
		c         = context.Background()
		mid       = int64(0)
		graphInfo = &api.GraphInfo{}
		req       = &model.HvarReq{}
	)
	convey.Convey("HiddenVars", t, func(ctx convey.C) {
		res, err := d.HiddenVars(c, mid, graphInfo, req, "")
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestDaorawHiddenVarsRec(t *testing.T) {
	convey.Convey("rawHiddenVarsRec", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			mid       = int64(0)
			graphid   = int64(0)
			currentid = int64(0)
			cursor    = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.rawHiddenVarsRec(c, mid, graphid, currentid, cursor)
			fmt.Println(res, err)
			convCtx.Convey("Then err should be nil. res should be greater than 0", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestDao_addHiddenVarsRec(t *testing.T) {
	convey.Convey("addHiddenVarsRec", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			rec = &model.HiddenVarRec{
				MID:       int64(0),
				GraphID:   int64(0),
				CurrentID: int64(0),
				CursorID:  int64(0),
				Value:     "",
			}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.addHiddenVarsRec(c, rec)
			convCtx.Convey("Then err should be nil", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
