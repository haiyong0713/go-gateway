package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_RegAllArc(t *testing.T) {
	Convey("TestService_RegAllArc", t, WithService(func(s *Service) {
		var (
			ctx = context.Background()
			rid = 127
			ps  = 10
			pn  = 1
		)
		arcs, count, err := s.RegAllArc(ctx, int64(rid), ps, pn)
		So(len(arcs), ShouldBeGreaterThan, 0)
		So(count, ShouldBeGreaterThan, 0)
		So(err, ShouldBeNil)
	}))
}

func TestService_RegOriginArc(t *testing.T) {
	Convey("TestService_RegOriginArc", t, WithService(func(s *Service) {
		var (
			ctx = context.Background()
			rid = 127
			ps  = 10
			pn  = 1
		)
		arcs, count, err := s.RegOriginArc(ctx, int64(rid), ps, pn)
		So(len(arcs), ShouldBeGreaterThan, 0)
		So(count, ShouldBeGreaterThan, 0)
		So(err, ShouldBeNil)
	}))
}

func TestService_RegionCnt(t *testing.T) {
	Convey("TestService_RegOriginArc", t, WithService(func(s *Service) {
		var (
			ctx = context.Background()
			rid = []int32{127}
		)
		arcs, err := s.RegionCnt(ctx, rid)
		So(len(arcs), ShouldBeGreaterThan, 0)
		So(err, ShouldBeNil)
	}))
}
