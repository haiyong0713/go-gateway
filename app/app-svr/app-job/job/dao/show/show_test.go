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
		res, err := d.Heads(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestItems(t *testing.T) {
	Convey("Items", t, func() {
		res, err := d.Items(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestTempHeads(t *testing.T) {
	Convey("TempHeads", t, func() {
		res, err := d.TempHeads(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestTempItems(t *testing.T) {
	Convey("TempItems", t, func() {
		res, err := d.TempItems(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
