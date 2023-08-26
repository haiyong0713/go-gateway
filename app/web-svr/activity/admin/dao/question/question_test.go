package question

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/admin/model/question"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaBatchAddDetail(t *testing.T) {
	convey.Convey("BatchAddDetail", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			list = make([]*question.AddDetailArg, 0, 2)
		)
		list = append(list, &question.AddDetailArg{BaseID: 10292, RightAnswer: "answer", Name: "性格", WrongAnswer: "worn", Attribute: 3})
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.BatchAddDetail(c, list)
			ctx.Convey("Then err should be nil.rp should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaSaveBase(t *testing.T) {
	convey.Convey("SaveBase", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			arg = &question.SaveBaseArg{ID: 1}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SaveBase(c, arg)
			ctx.Convey("Then err should be nil.rp should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaSaveDetail(t *testing.T) {
	convey.Convey("SaveDetail", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			arg = &question.SaveDetailArg{ID: 1}
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SaveDetail(c, arg)
			ctx.Convey("Then err should be nil.rp should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
