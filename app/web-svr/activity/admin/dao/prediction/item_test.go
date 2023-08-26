package prediction

import (
	"context"
	"fmt"
	"testing"

	premdl "go-gateway/app/web-svr/activity/admin/model/prediction"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoBatchItem(t *testing.T) {
	convey.Convey("BatchAdd", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			list = make([]*premdl.PredItem, 0, 1)
		)
		list = append(list, &premdl.PredItem{Sid: 10292, State: 1, Desc: "皮卡丘", Image: "", Pid: 3})
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.BatchItem(c, list)
			ctx.Convey("Then err should be nil.rp should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoItemUp(t *testing.T) {
	convey.Convey("BatchAdd", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			list = &premdl.ItemUp{ID: 1, Desc: "魔王one", Image: "", State: 0}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.ItemUp(c, list)
			ctx.Convey("Then err should be nil.rp should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoItemSearch(t *testing.T) {
	convey.Convey("BatchAdd", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			list = &premdl.ItemSearch{ID: 1, Pid: 3, Pn: 1, Ps: 15, State: 1}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.ItemSearch(c, list)
			ctx.Convey("Then err should be nil.rp should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%#v", res)
			})
		})
	})
}
