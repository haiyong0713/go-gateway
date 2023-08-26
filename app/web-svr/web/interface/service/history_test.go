package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_WxHistoryCursor(t *testing.T) {
	Convey("test nav Nav", t, WithService(func(s *Service) {
		var (
			mid int64 = 111006313
		)
		res, err := s.WxHistoryCursor(context.Background(), mid, 0, 0, "", 10)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		Printf("%+v", res)
	}))
}
