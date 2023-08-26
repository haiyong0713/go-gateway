package view

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_ViewInfoc(t *testing.T) {
	Convey("ViewInfoc", t, func() {
		s.ViewInfoc(0, 0, "test", "0", "", "", "", "", "", time.Now(), errors.New("test"), 1, "", "", "mobile", "")
	})
}

func Test_RelateInfoc(t *testing.T) {
	Convey("RelateInfoc", t, func() {
		s.RelateInfoc(0, 0, 0, "", "", "", "", "", "", "", "", nil, time.Now(),
			0, 1, 0, "", "", "", "", nil, nil, "", nil, 0)
	})
}

func Test_infocAd(t *testing.T) {
	Convey("RelateInfoc", t, func() {
		s.infocAd(context.Background(),
			"m", 123, "network", 123, "dd", 321, []int32{3, 4, 5}, "ios", "", "")
	})
}
