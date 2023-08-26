package fawkes

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_TopResource(t *testing.T) {
	Convey("TestDao_TopResource", t, func() {
		resource, err := d.BenderTopResource(context.Background())
		So(err, ShouldBeNil)
		So(resource, ShouldNotBeEmpty)
	})
}
