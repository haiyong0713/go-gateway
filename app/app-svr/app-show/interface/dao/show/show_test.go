package show

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func ctx() context.Context {
	return context.Background()
}

func TestHeads(t *testing.T) {
	Convey("Heads", t, func() {
		res, err := dao.Heads(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestItems(t *testing.T) {
	Convey("Items", t, func() {
		res, err := dao.Items(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestTempHeads(t *testing.T) {
	Convey("TempHeads", t, func() {
		res, err := dao.TempHeads(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestTempItems(t *testing.T) {
	Convey("TempItems", t, func() {
		res, err := dao.TempItems(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
