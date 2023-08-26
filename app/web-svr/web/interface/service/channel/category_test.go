package channel

import (
	chanmdl "go-gateway/app/web-svr/web/interface/model/channel"
	"testing"

	. "github.com/glycerine/goconvey/convey"
)

func TestService_CategoryList(t *testing.T) {
	Convey("TestService_CategoryList", t, WithService(func(s *Service) {
		res, err := s.CategoryList(c)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_ChannelArcList(t *testing.T) {
	Convey("TestService_ChannelArcList", t, WithService(func(s *Service) {
		req := &chanmdl.ChannelArcListReq{
			ID:     28,
			Offset: "",
		}
		res, err := s.ChannelArcList(c, mid, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_ChannelList(t *testing.T) {
	Convey("TestService_ChannelList", t, WithService(func(s *Service) {
		req := &chanmdl.ChannelListReq{
			ID:       28,
			Offset:   "",
			PageSize: 5,
		}
		res, err := s.ChannelList(c, mid, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
