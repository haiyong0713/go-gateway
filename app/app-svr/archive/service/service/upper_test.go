package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_UpperCount(t *testing.T) {
	Convey("UpperCount", t, func() {
		_, err := s.UpperCount(context.TODO(), 1)
		So(err, ShouldBeNil)
	})
}

func Test_UppersAidPubTime(t *testing.T) {
	Convey("UppersAidPubTime", t, func() {
		_, err := s.UppersAidPubTime(context.TODO(), []int64{1684013}, 1, 10)
		So(err, ShouldBeNil)
	})
}
