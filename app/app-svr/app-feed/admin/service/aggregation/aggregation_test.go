package aggregation

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/aggregation"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_AddAggregation(t *testing.T) {
	Convey("TestService_AddAggregation", t, WithService(func(s *Service) {
		var (
			c     = context.Background()
			param = aggregation.AggPub{
				HotTitle: "lol",
				Title:    "lol",
				SubTitle: "lol",
				Image:    "www.bilibili.com",
			}
			tagIDs = []int64{9222}
		)
		err := s.AddAggregation(c, param, tagIDs, "", 0)
		So(err, ShouldBeNil)
	}))
}

func TestService_UpdateAggregation(t *testing.T) {
	Convey("TestService_UpdateAggregation", t, WithService(func(s *Service) {
		var (
			c     = context.Background()
			param = aggregation.AggPub{
				HotTitle: "lol",
				Title:    "lol",
				SubTitle: "lol",
				Image:    "www.bilibili.com",
			}
			tagIDs = []int64{9222}
		)
		err := s.UpdateAggregation(c, param, tagIDs, "", 0)
		So(err, ShouldBeNil)
	}))
}

func TestService_AggOperate(t *testing.T) {
	Convey("TestService_AggOperate", t, WithService(func(s *Service) {
		var c = context.Background()
		err := s.AggOperate(c, 1, 0, 0, "", "")
		So(err, ShouldBeNil)
	}))
}

func TestService_AggregationList(t *testing.T) {
	Convey("TestService_AggregationList", t, WithService(func(s *Service) {
		var (
			c     = context.Background()
			param = &aggregation.AggListReq{
				ID: 1,
			}
		)
		res, err := s.AggregationList(c, param)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_Tag(t *testing.T) {
	Convey("TestService_Tag", t, WithService(func(s *Service) {
		var c = context.Background()
		res, err := s.Tag(c, "lol")
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_AggView(t *testing.T) {
	Convey("TestService_AggView", t, WithService(func(s *Service) {
		var c = context.Background()
		res, err := s.AggView(c, 199, false)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	}))
}

func TestService_AggTagAdd(t *testing.T) {
	Convey("TestService_AggTagAdd", t, WithService(func(s *Service) {
		var c = context.Background()
		err := s.AggTagAdd(c, 189, []int64{600, 14530})
		So(err, ShouldBeNil)
	}))
}

func TestService_AggTagDel(t *testing.T) {
	Convey("TestService_AggTagDel", t, WithService(func(s *Service) {
		var c = context.Background()
		err := s.AggTagDel(c, 189, 600)
		So(err, ShouldBeNil)
	}))
}
