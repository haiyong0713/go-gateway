package common

import (
	"context"
	"github.com/smartystreets/goconvey/convey"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
	"testing"
)

func Test_TabItemsStrategy(t *testing.T) {
	convey.Convey("Test_TabItemsStrategy_audio_season", t, WithService(func(s *Service) {
		var (
			ctx = context.Background()
			req = &fm_v2.HandleTabItemsReq{
				DeviceInfo: model.DeviceInfo{
					AccessKey: "",
					MobiApp:   "android_bilithings",
					Channel:   "byd",
					Build:     2020001,
				},
				Mid:    20321343,
				FmType: fm_v2.AudioSeason,
				FmId:   333,
			}
		)
		resp, err := TabItemsStrategy(ctx, req)
		convey.ShouldBeNil(err)
		convey.ShouldNotBeNil(resp)
		_, _ = convey.Println(toJson(resp))
	}))
}

func Test_seasonTabAB(t *testing.T) {
	convey.Convey("Test_seasonTabAB", t, WithService(func(s *Service) {
		var (
			ctx   = context.Background()
			param = &fm_v2.FmShowParam{
				DeviceInfo: model.DeviceInfo{Build: 2010001},
				Mid:        23432942,
				Buvid:      "XY123456",
			}
			tabs = []*fm_v2.TabItem{
				{
					FmType:        fm_v2.AudioSeason,
					FmId:          333,
					Title:         "原始标题",
					FirstArcTitle: "实验组标题",
				},
			}
		)
		tabAB := s.seasonTabAB(ctx, param, tabs)
		convey.ShouldEqual(tabAB[0].Title, "实验组标题")
		_, _ = convey.Printf("tabAB:%s\n", toJson(tabAB))
	}))
}
