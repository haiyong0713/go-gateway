package prediction

import (
	"context"
	"fmt"
	"testing"

	premdl "go-gateway/app/web-svr/activity/admin/model/prediction"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoBatchAdd(t *testing.T) {
	convey.Convey("BatchAdd", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			list = make([]*premdl.Prediction, 0, 1)
		)
		list = append(list, &premdl.Prediction{Sid: 10292, State: 1, Name: "性格", Min: 0, Max: 0, Pid: 0, Type: 1})
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.BatchAdd(c, list)
			ctx.Convey("Then err should be nil.rp should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSearch(t *testing.T) {
	convey.Convey("Search", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ser = &premdl.PredSearch{ID: 1, Sid: 10292, Pid: 0, Type: 0, State: 1, Pn: 1, Ps: 15}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			list, err := d.Search(c, ser)
			ctx.Convey("Then err should be nil.rp should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Print(list)
			})
		})
	})
}

// PresUp
func TestDaoPresUp(t *testing.T) {
	convey.Convey("Search", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ser = &premdl.PresUp{ID: 1, Name: "性格", State: 1}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.PresUp(c, ser)
			ctx.Convey("Then err should be nil.rp should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
