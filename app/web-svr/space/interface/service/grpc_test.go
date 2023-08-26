package service

import (
	"context"
	v1 "go-gateway/app/web-svr/space/interface/api/v1"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_Official(t *testing.T) {
	Convey("test notice", t, WithService(func(s *Service) {
		data, err := s.Official(context.Background(), &v1.OfficialRequest{Mid: 1})
		So(err, ShouldBeNil)
		Printf("%v", data)
	}))
}
