package common

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
)

func TestService_FmLike(t *testing.T) {
	convey.Convey("TestService_FmLike", t, WithService(func(s *Service) {
		var (
			c     = metadata.NewContext(context.Background(), metadata.MD{"remote_ip": "165.53.43.21"})
			param = &fm_v2.FmLikeParam{
				DeviceInfo: model.DeviceInfo{
					MobiApp:  "android_bilithings",
					Device:   "Android",
					Platform: "android",
					Channel:  "byd",
					Build:    2000007,
				},
				Mid:    27515254,
				UpMid:  2074573507,
				Oid:    880121692,
				Action: _actionLike,
				Path:   "/x/v2/car/fm/like",
				UA:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.64 Safari/537.36",
			}
		)

		err := s.FmLike(c, param)
		convey.So(err, convey.ShouldBeNil)
	}))

}
