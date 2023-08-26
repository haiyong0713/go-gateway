package common

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/common"
)

func TestService_commonItems(t *testing.T) {
	convey.Convey("TestService_commonItems", t, WithService(func(s *Service) {
		var (
			c      = context.Background()
			device = model.DeviceInfo{
				AccessKey: "e2cf401a17df4803eac905dee4b1ad61",
				MobiApp:   "bilibili_things",
				Channel:   "byd",
			}
			// 合集测试
			reqSerial = &commonItemsReq{
				Mid:              14139334,
				Buvid:            "XY317C63F6B78666114B90BB3167CB1C5A3F2",
				FmSerialIds:      []int64{333, 444},
				FmSerialIdAidMap: map[int64]int64{333: 440117562, 444: 840121797},
			}
		)
		commonItems, err := s.commonItems(c, device, reqSerial)
		convey.So(err, convey.ShouldBeNil)
		bytes, _ := json.Marshal(commonItems)
		_, _ = convey.Printf("res:%+v", string(bytes))
	}))

}

func TestService_VideoTabs(t *testing.T) {
	convey.Convey("TestService_VideoTabs", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = &common.VideoTabsReq{DeviceInfo: model.DeviceInfo{Channel: "byd"}}
		)
		tabs, err := s.VideoTabs(c, req)
		convey.So(err, convey.ShouldBeNil)
		bytes, _ := json.Marshal(tabs)
		_, _ = convey.Printf("res:%+v", string(bytes))
	}))
}
