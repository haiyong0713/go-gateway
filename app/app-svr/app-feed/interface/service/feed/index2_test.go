package feed

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_Index2(t *testing.T) {
	Convey(t.Name(), t, func() {
		gotIs, gotConfig, gotInfoc, _, _, err := s.Index2(context.Background(), "", 0, 0, nil, 1, "", "", time.Now())
		So(gotIs, ShouldNotBeNil)
		So(gotConfig, ShouldNotBeNil)
		So(gotInfoc, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestService_indexBanner2(t *testing.T) {
	Convey(t.Name(), t, func() {
		gotBanners, gotVersion, err := s.indexBanner2(context.Background(), 0, "", 0, nil, nil)
		So(gotBanners, ShouldNotBeNil)
		So(gotVersion, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestService_mergeItem2(t *testing.T) {
	Convey(t.Name(), t, func() {
		gotIs, gotAdInfom := s.mergeItem2(context.Background(), 0, 0, nil, nil, nil, nil, "", nil, false, nil, nil)
		So(gotIs, ShouldNotBeNil)
		So(gotAdInfom, ShouldNotBeNil)
	})
}

func TestService_dealAdLoc(t *testing.T) {
	Convey(t.Name(), t, func() {
		s.dealAdLoc(context.Background(), nil, nil, nil, time.Now(), "", 0, nil)
	})
}

func TestService_dealItem2(t *testing.T) {
	Convey(t.Name(), t, func() {
		gotIs, gotIsAI := s.dealItem2(context.Background(), 0, "", 0, nil, nil, false, false, false, time.Now(), nil, nil)
		So(gotIs, ShouldNotBeNil)
		So(gotIsAI, ShouldNotBeNil)
	})
}

func TestService_Converge(t *testing.T) {
	Convey(t.Name(), t, func() {
		gotIs, gotConverge, _, _, _, err := s.Converge(context.Background(), 0, 0, nil, "", time.Now())
		So(gotIs, ShouldNotBeNil)
		So(gotConverge, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
