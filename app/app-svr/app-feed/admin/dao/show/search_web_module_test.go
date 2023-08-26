package show

import (
	"encoding/json"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_valModule(t *testing.T) {
	convey.Convey("ValQuery", t, func(ctx convey.C) {
		var (
			paramModules []*show.SearchWebModuleModule
		)
		if err := json.Unmarshal([]byte("[{\"value\":\"1\"},{\"value\":\"2\"}],{\"value\":\"3\"}],{\"value\":\"4\"}],{\"value\":\"5\"}],{\"value\":\"6\"}]"), &paramModules); err != nil {
			return
		}
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.valModule(paramModules)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_ValQuery(t *testing.T) {
	convey.Convey("ValQuery", t, func(ctx convey.C) {
		var (
			param = []*show.SearchWebModuleQuery{
				{Value: "testaaa1222"},
				//{Value: "test3"},
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.ValQuery(38, param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_SearchSpecialAdd(t *testing.T) {
	convey.Convey("SearchSpecialAdd", t, func(ctx convey.C) {
		var (
			param = &show.SearchWebModuleAP{
				Reason: "test",
				Query:  "[{\"id\":15,\"value\":\"testaaa122\",\"deleted\":0},{\"value\":\"testbbb122\",\"deleted\":0}]",
				Module: "[{\"value\":\"1\"},{\"value\":\"2\"},{\"value\":\"3\"},{\"value\":\"4\"},{\"value\":\"5\"},{\"value\":\"6\"}]",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.WebModuleAdd(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_SearchSpecialUpdate(t *testing.T) {
	convey.Convey("SearchSpecialAdd", t, func(ctx convey.C) {
		var (
			param = &show.SearchWebModuleUP{
				ID:     63,
				Reason: "test",
				Query:  "[{\"id\":706,\"value\":\"testbbbb1\",\"deleted\":0},{\"id\":1,\"value\":\"testbbbb1\",\"deleted\":0},{\"value\":\"test23333dddd2\",\"deleted\":0}]",
				Module: "[{\"value\":\"2\"},{\"value\":\"1\"},{\"value\":\"3\"},{\"value\":\"4\"},{\"value\":\"5\"},{\"value\":\"6\"}]",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			d.WebModuleUpdate(param)
		})
	})
}
