package hidden_vars

import (
	"context"
	"encoding/json"
	"testing"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaosetHvarCache(t *testing.T) {
	var (
		c      = context.Background()
		mid    = int64(1)
		gid    = int64(2)
		nid    = int64(3)
		cursor = int64(4)
		buvid  = "4444"
		hvars  = map[string]*model.HiddenVar{
			"123": {
				Value: 333,
			},
		}
	)
	convey.Convey("setHvarCache", t, func(ctx convey.C) {
		convey.Println(d.hvarKey(mid, gid, nid, cursor, buvid))
		err := d.setHvarCache(c, mid, gid, nid, cursor, buvid, &model.HiddenVarsRecord{
			Vars: hvars,
		})
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaohvarCache(t *testing.T) {
	var (
		c      = context.Background()
		mid    = int64(1)
		gid    = int64(2)
		nid    = int64(3)
		cursor = int64(4)
		buvid  = "4444"
	)
	convey.Convey("hvarCache", t, func(ctx convey.C) {
		a, err := d.hvarCache(c, mid, gid, nid, cursor, buvid)
		str, _ := json.Marshal(a)
		convey.Println(string(str), err)
		ctx.Convey("Then err should be nil.a should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(a, convey.ShouldNotBeNil)
		})
	})
}
