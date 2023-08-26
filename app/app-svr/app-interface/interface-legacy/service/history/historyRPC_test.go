package history

import (
	"context"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	api "go-gateway/app/app-svr/app-interface/interface-legacy/api/history"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_HistoryTab(t *testing.T) {
	Convey("HistoryTab", t, WithService(func(s *Service) {
		dev := device.Device{
			Buvid: "test",
		}
		req := &api.HistoryTabReq{
			Source: api.HistorySource_history,
		}
		_, err := s.HistoryTabRPC(context.TODO(), 27515256, dev, req)
		So(err, ShouldBeNil)
	}))
}

func TestService_CursorV2(t *testing.T) {
	Convey("HistoryTab", t, WithService(func(s *Service) {
		dev := device.Device{
			Buvid: "test",
		}
		req := &api.CursorV2Req{
			Business: "goods",
		}
		network := network.Network{
			Type: _androidBroadcast,
		}

		_, err := s.CursorV2GRPC(context.TODO(), 27515256, dev, 111, req, network)
		So(err, ShouldBeNil)
	}))
}
