package service

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/archive/service/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_GoRegionTotal(t *testing.T) {
	Convey("RegionTotal BigData", t, WithService(func(s *Service) {
		for _, v := range s.regionTotal {
			So(v, ShouldBeGreaterThan, 0)
		}
	}))
	Convey("RegionTotal Live", t, WithService(func(s *Service) {
		So(s.live, ShouldBeGreaterThan, 0)
	}))
}

func TestService_GoRegionArcs(t *testing.T) {
	var (
		c                   = context.Background()
		rid           int32 = 168
		pn, ps, count int   = 1, 10, 0
		err           error
		rs            []*api.Arc
	)

	Convey("Region Arcives", t, WithService(func(s *Service) {
		rs, count, err = s.GoRegionArcs3(c, rid, pn, ps, false)
		So(err, ShouldBeNil)
		So(count, ShouldBeGreaterThan, 0)
		Printf("%+v", rs)
	}))

}

func TestService_GoRegionsArcs(t *testing.T) {
	var (
		c     = context.Background()
		rids  = []int32{1, 3, 4, 5, 13, 36, 129, 119, 23, 11, 155, 160, 165, 168}
		count = 15
		err   error
		rs    map[int32][]*api.Arc
	)
	Convey("Regions Archives", t, WithService(func(s *Service) {
		rs, err = s.GoRegionsArcs3(c, rids, count)
		So(err, ShouldBeNil)
		Printf("%+v", rs)
	}))

}

func TestService_GoRegionTagArcs(t *testing.T) {
	var (
		c                   = context.Background()
		tid           int64 = 123456
		rid           int32 = 168
		pn, ps, count int   = 1, 5, 0
		err           error
		rs            []*api.Arc
	)
	Convey("Region Tag Archives", t, WithService(func(s *Service) {
		rs, count, err = s.GoRegionTagArcs3(c, rid, tid, pn, ps)
		So(err, ShouldBeNil)
		Printf("%+v", rs)
		Printf("%d", count)
	}))
}
