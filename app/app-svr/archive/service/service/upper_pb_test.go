package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_UpperPassed3(t *testing.T) {
	Convey("UpperPassed3", t, func() {
		_, err := s.UpperPassed3(context.TODO(), 27515615, 1, 20)
		So(err, ShouldNotBeNil)
	})
}
