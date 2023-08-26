package history

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go-common/component/metadata/device"
	network2 "go-common/component/metadata/network"
	"go-gateway/app/app-svr/app-interface/interface-legacy/api/history"
	hismdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/history"
)

func TestService_List(t *testing.T) {
	Convey("list", t, WithService(func(s *Service) {
		_, err := s.List(context.TODO(), 27515256, 111, 1, 20, "0", 0)
		So(err, ShouldBeNil)
	}))
}

func TestService_Live(t *testing.T) {
	Convey("live", t, WithService(func(s *Service) {
		_, err := s.Live(context.TODO(), []int64{27515256})
		So(err, ShouldBeNil)
	}))
}

func TestService_LiveList(t *testing.T) {
	Convey("live", t, WithService(func(s *Service) {
		_, err := s.LiveList(context.TODO(), 27515256, 111, 1, 20, "0", 0)
		So(err, ShouldBeNil)
	}))
}

func TestService_Cursor(t *testing.T) {
	Convey("cursor", t, WithService(func(s *Service) {
		testArcsWithPlayUrlParams := &hismdl.ArcsWithPlayUrlParam{
			Qn:        0,
			Fnver:     0,
			Fnval:     0,
			ForceHost: 0,
			Fourk:     0,
			MobiApp:   "iphone",
			Buvid:     "8670",
			NetType:   0,
			TfType:    0,
		}
		_, _, err := s.Cursor(context.TODO(), nil, 8670, 0, testArcsWithPlayUrlParams)
		So(err, ShouldBeNil)
	}))
}

func TestService_CursorGRPC(t *testing.T) {
	Convey("cursorGRPC", t, WithService(func(s *Service) {
		arg := &history.CursorReq{
			Business: "all",
		}
		dev := &device.Device{
			Sid:          "",
			Buvid3:       "",
			Build:        8670,
			Buvid:        "",
			Channel:      "",
			Device:       "",
			RawPlatform:  "aaaa",
			RawMobiApp:   "iphone",
			Model:        "",
			Brand:        "",
			Osver:        "",
			UserAgent:    "",
			Network:      "",
			NetworkType:  0,
			TfISP:        "",
			TfType:       0,
			FawkesEnv:    "",
			FawkesAppKey: "",
		}
		network := &network2.Network{
			Type: 0,
			TF:   0,
		}
		res, err := s.CursorGRPC(context.TODO(), 111004852, *dev, 0, arg, *network)
		fmt.Printf("%+v", res)
		So(err, ShouldBeNil)
	}))
}

func TestService_Search(t *testing.T) {
	Convey("Search", t, WithService(func(s *Service) {
		arg := &history.SearchReq{
			Pn:      1,
			Keyword: "测试",
		}
		dev := device.Device{
			Sid:          "",
			Buvid3:       "",
			Build:        8670,
			Buvid:        "",
			Channel:      "",
			Device:       "",
			RawPlatform:  "aaaa",
			RawMobiApp:   "iphone",
			Model:        "",
			Brand:        "",
			Osver:        "",
			UserAgent:    "",
			Network:      "",
			NetworkType:  0,
			TfISP:        "",
			TfType:       0,
			FawkesEnv:    "",
			FawkesAppKey: "",
		}
		res, err := s.Search(context.TODO(), 14139334, 1, arg, dev)
		fmt.Printf("%d %d", len(res.Items), res.Page.Total)
		So(err, ShouldBeNil)
	}))
}
