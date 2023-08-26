package service

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/web-svr/native-page/admin/model"
)

var (
	uid      int64 = 1
	username       = "test"
)

func TestService_AddWhiteList(t *testing.T) {
	Convey("TestService_AddWhiteList", t, WithService(func(s *Service) {
		err := s.AddWhiteList(context.Background(), &model.AddWhiteListReq{Mid: 1}, uid, username, "")
		So(err, ShouldBeNil)
	}))
}

func TestService_DeleteWhiteList(t *testing.T) {
	Convey("TestService_DeleteWhiteList", t, WithService(func(s *Service) {
		err := s.DeleteWhiteList(context.Background(), &model.DeleteWhiteListReq{ID: 1}, uid, username)
		So(err, ShouldBeNil)
	}))
}

func TestService_WhiteList(t *testing.T) {
	Convey("TestService_WhiteList", t, WithService(func(s *Service) {
		res, err := s.WhiteList(context.Background(), &model.GetWhiteListReq{Mid: 1, Pn: 1, Ps: 20})
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		if len(res.List) == 0 {
			return
		}
		for _, v := range res.List {
			fmt.Printf("%+v\n", v)
		}
	}))
}
