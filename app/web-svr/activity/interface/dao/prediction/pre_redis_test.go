package prediction

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestPreListKey(t *testing.T) {
	convey.Convey("preListKey", t, func(ctx convey.C) {
		var (
			sid = int64(7)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := preListKey(sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestItemListKey(t *testing.T) {
	convey.Convey("itemListKey", t, func(ctx convey.C) {
		var (
			pid = int64(7)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			p1 := itemListKey(pid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestAddPreSet(t *testing.T) {
	convey.Convey("AddPreSet", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{7}
			sid = int64(10292)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddPreSet(c, ids, sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestPreList(t *testing.T) {
	convey.Convey("PreList", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			sid = int64(10292)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			list, err := d.PreList(c, sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", list)
				fmt.Print(err)
			})
		})
	})
}

func TestDelPreSet(t *testing.T) {
	convey.Convey("DelPreSet", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{7}
			sid = int64(10292)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.DelPreSet(c, ids, sid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestAddItemPreSet(t *testing.T) {
	convey.Convey("AddItemPreSet", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{7, 10, 5, 8, 9, 20, 150}
			pid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.AddItemPreSet(c, ids, pid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestItemRandMember(t *testing.T) {
	convey.Convey("ItemRandMember", t, func(ctx convey.C) {
		var (
			c     = context.Background()
			pid   = int64(1)
			count = 2
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			list, err := d.ItemRandMember(c, pid, count)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", list)
			})
		})
	})
}

func TestDelItemPreSet(t *testing.T) {
	convey.Convey("DelItemPreSet", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{7}
			pid = int64(1)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.DelItemPreSet(c, ids, pid)
			ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
