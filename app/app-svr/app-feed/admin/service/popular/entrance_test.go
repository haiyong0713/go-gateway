package popular

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPopEntranceSave(t *testing.T) {
	Convey("TestPopEntranceSave", t, WithService(func(s *Service) {
		param := &show.EntranceSave{
			EntranceCore: show.EntranceCore{
				Title:       "t123",
				Icon:        "t1",
				RedirectUri: "www.bilibili.com",
				Rank:        0,
				ModuleID:    "t1",
				Grey:        20,
				RedDot:      0,
				RedDotText:  "有更新",
				BuildLimit:  `[{"plat":0,"condition_start":"gt","build_start":10007,"condition_end":"lt","build_end":11111},{"plat":1,"condition_start":"gt","build_start":8509,"condition_end":"lt","build_end":8517}]`,
				WhiteList:   "1,2,3",
				BlackList:   "4,5,6",
				State:       1,
			},
			Version: 1,
		}
		err := s.PopEntranceSave(context.Background(), param, 0, "")
		So(err, ShouldBeNil)
	}))
}

func TestPopularEntrance(t *testing.T) {
	Convey("TestPopularEntrance", t, WithService(func(s *Service) {
		res, err := s.PopularEntrance(context.Background(), 3, 1, 30)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestPopEntranceOperate(t *testing.T) {
	Convey("TestPopEntranceOperate", t, WithService(func(s *Service) {
		err := s.PopEntranceOperate(context.Background(), 12, 0, 1, "", "t1")
		So(err, ShouldBeNil)
	}))
}

func TestRedDotUpdate(t *testing.T) {
	Convey("TestRedDotUpdate", t, WithService(func(s *Service) {
		err := s.RedDotUpdate(context.Background(), "lx", "good-history")
		So(err, ShouldBeNil)
	}))
}

func TestPopularView(t *testing.T) {
	Convey("PopularView", t, WithService(func(s *Service) {
		res, err := s.PopularView(context.Background(), 280)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestPopularViewSave(t *testing.T) {
	Convey("PopularViewSave", t, WithService(func(s *Service) {
		err := s.PopularViewSave(context.Background(), 280, "http://i0.hdslb.com/bfs/app/b077f3c38759da7a5d9eb5007bcc07221b53f0d2.png")
		So(err, ShouldBeNil)
	}))
}

func TestPopularViewAdd(t *testing.T) {
	var ids []int64
	ids = append(ids, 123)
	ids = append(ids, 124)
	Convey("PopularViewAdd", t, WithService(func(s *Service) {
		err := s.PopularViewAdd(context.Background(), 280, ids)
		So(err, ShouldBeNil)
	}))
}

func TestPopularViewOperate(t *testing.T) {
	Convey("PopularViewOperate", t, WithService(func(s *Service) {
		err := s.PopularViewOperate(context.Background(), 280, 10098256, 0, 2)
		So(err, ShouldBeNil)
	}))
}

func TestPopularTagAdd(t *testing.T) {
	var ids []int64
	ids = append(ids, 125)
	ids = append(ids, 126)
	Convey("PopularTagAdd", t, WithService(func(s *Service) {
		err := s.PopularTagAdd(context.Background(), 280, ids)
		So(err, ShouldBeNil)
	}))
}

func TestPopularTagDel(t *testing.T) {
	Convey("PopularTagDel", t, WithService(func(s *Service) {
		err := s.PopularTagDel(context.Background(), 280, 13863)
		So(err, ShouldBeNil)
	}))
}

func TestPopularMiddleSave(t *testing.T) {
	Convey("PopularMiddleSave", t, WithService(func(s *Service) {
		//err := s.PopularMiddleSave(context.Background(), 1, 1, "")
		//So(err, ShouldBeNil)
	}))
}

func TestPopularMiddleList(t *testing.T) {
	Convey("PopularMiddleList", t, WithService(func(s *Service) {
		res, err := s.PopularMiddleList(context.Background(), 1, 10)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
