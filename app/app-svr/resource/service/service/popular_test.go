package service

import (
	"context"
	"testing"

	pb "go-gateway/app/app-svr/resource/service/api/v1"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_GetS10PopEntranceAids(t *testing.T) {
	Convey("GetS10PopEntranceAids", t, WithService(func(s *Service) {
		res, err := s.GetS10PopEntranceAids(context.Background(), &pb.GetS10PopEntranceAidsReq{})
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeEmpty)
	}))
}
