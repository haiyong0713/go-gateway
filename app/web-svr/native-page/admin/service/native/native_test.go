package native

import (
	"context"
	"fmt"
	"testing"
	"time"

	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_DelPage(t *testing.T) {
	Convey("TestService_CheckArc Test", t, WithService(func(s *Service) {
		err := s.DelPage(context.Background(), 1, "liuxuan", "")
		So(err, ShouldBeNil)
	}))
}

func TestService_EditPage(t *testing.T) {
	Convey("TestService_CheckArc Test", t, WithService(func(s *Service) {
		err := s.EditPage(context.Background(), &natmdl.EditParam{}, 0)
		So(err, ShouldBeNil)
	}))
}

func TestService_SearchPage(t *testing.T) {
	Convey("TestService_CheckArc Test", t, WithService(func(s *Service) {
		res, err := s.SearchPage(context.Background(), &natmdl.SearchParam{PageParam: natmdl.PageParam{ID: 1, State: 2}})
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	}))
}

func TestService_Module(t *testing.T) {
	Convey("TestService_CheckArc Test", t, WithService(func(s *Service) {
		_, err := s.Module(context.Background(), &natmdl.ModuleParam{})
		So(err, ShouldBeNil)
	}))
}

func TestService_SearchModule(t *testing.T) {
	Convey("TestService_CheckArc Test", t, WithService(func(s *Service) {
		res, err := s.SearchModule(context.Background(), &natmdl.SearchModule{})
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_PageSkipUrl(t *testing.T) {
	Convey("TestService_CheckArc Test", t, WithService(func(s *Service) {
		err := s.PageSkipUrl(context.Background(), &natmdl.EditParam{}, 1)
		So(err, ShouldBeNil)
	}))
}

func TestService_FindPage(t *testing.T) {
	Convey("TestService_CheckArc Test", t, WithService(func(s *Service) {
		res, err := s.FindPage(context.Background(), "fgo", 0)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_OfflineTab(t *testing.T) {
	Convey("TestService_OfflineTab Test", t, WithService(func(s *Service) {
		var id int32 = 1
		err := s.EditTab(context.Background(), id, 1586310000, 1586310001)
		So(err, ShouldBeNil)
	}))
}

func TestService_SearchTab(t *testing.T) {
	Convey("TestService_SearchTab Test", t, WithService(func(s *Service) {
		location, _ := time.LoadLocation("Local")
		ctimeStart, _ := time.ParseInLocation("2006-01-02 15:04:05", "2020-04-10 15:58:10", location)
		ctimeEnd, _ := time.ParseInLocation("2006-01-02 15:04:05", "2020-04-13 11:58:17", location)
		req := &natmdl.SearchTabReq{
			//ID:         1,
			Title:      "活动底栏-",
			Creator:    "",
			CtimeStart: ctimeStart.Unix(),
			CtimeEnd:   ctimeEnd.Unix(),
			State:      -1,
			Pn:         1,
			Ps:         5,
		}
		list, err := s.SearchTab(context.Background(), req)
		for i, item := range list.List {
			fmt.Printf("%d %s %s\n", i, item.Title, time.Unix(int64(item.Ctime), 0).Format("2006-01-02 15:04:05"))
		}
		So(err, ShouldBeNil)
		So(list, ShouldNotBeNil)
	}))
}

func TestService_GetTabOfPage(t *testing.T) {
	Convey("TestService_GetTabOfPage Test", t, WithService(func(s *Service) {
		var pid int32 = 1
		searchTabItem, err := s.GetTabOfPage(context.Background(), pid)
		fmt.Printf("searchTabItem: %v\n", searchTabItem)
		So(err, ShouldBeNil)
	}))
}
