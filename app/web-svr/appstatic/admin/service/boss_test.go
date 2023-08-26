package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_CheckCdnStatus(t *testing.T) {
	Convey("TestService_CheckCdnStatus", t, WithService(func(svf *Service) {
		err := svf.BossCdnPublishCheck(context.Background(), 2)
		So(err, ShouldBeNil)
	}))
}
