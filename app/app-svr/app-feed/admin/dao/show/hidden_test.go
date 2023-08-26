package show

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/hidden"

	"github.com/smartystreets/goconvey/convey"
)

func TestHiddenSave(t *testing.T) {
	var (
		hiddenArg = &hidden.Hidden{
			//ID:      1,
			SID:     1,
			RID:     3,
			Channel: "oppo",
			Stime:   time.Time(1574666727),
			Etime:   time.Time(1574666740),
		}
		limit = &hidden.HiddenLimit{
			Plat:       0,
			Build:      9999,
			Conditions: "gt",
		}
		limits = make(map[int8]*hidden.HiddenLimit)
	)
	limits[0] = limit
	limits[1] = &hidden.HiddenLimit{
		Plat:       1,
		Build:      888,
		Conditions: "lt",
	}
	convey.Convey("HiddenSave", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			id, err := d.HiddenSave(context.Background(), hiddenArg, limits)
			fmt.Println(id)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestHiddenLimits(t *testing.T) {
	convey.Convey("HiddenLimits", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			res, err := d.HiddenLimits(context.Background(), []int64{1, 2})
			str, _ := json.Marshal(res)
			fmt.Println(string(str))
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestUpdateHiddenState(t *testing.T) {
	convey.Convey("UpdateHiddenState", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			_, err := d.UpdateHiddenState(context.Background(), 1, 0, 1)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestHidden(t *testing.T) {
	convey.Convey("Hidden", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			res, err := d.Hidden(context.Background(), 2)
			str, _ := json.Marshal(res)
			fmt.Println(string(str))
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestHiddens(t *testing.T) {
	convey.Convey("Hiddens", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			res, total, err := d.Hiddens(context.Background(), 1, 15)
			str, _ := json.Marshal(res)
			fmt.Println(string(str))
			fmt.Println(total)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestRegion(t *testing.T) {
	convey.Convey("Region", t, func(ctx convey.C) {
		ctx.Convey("When everything go positive", func(ctx convey.C) {
			res, err := d.Region(context.Background(), 1)
			str, _ := json.Marshal(res)
			fmt.Println(string(str))
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
