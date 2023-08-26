package service

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_ArcsAllow(t *testing.T) {
	Convey("TestService_ArcsAllow", t, WithService(func(s *Service) {
		res, err := s.ArcsAllow(context.Background(), []int64{10318407, 10318755, 10318753})
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
