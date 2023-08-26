package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Stat3(t *testing.T) {
	Convey("Stat3", t, func() {
		_, err := s.Stat3(context.TODO(), 14761597)
		So(err, ShouldBeNil)
	})
}

func Test_Stats3(t *testing.T) {
	Convey("Stats3", t, func() {
		_, err := s.Stats3(context.TODO(), []int64{14761597})
		So(err, ShouldBeNil)
	})
}
