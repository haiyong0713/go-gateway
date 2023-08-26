package service

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_Archives(t *testing.T) {
	Convey("service test", t, WithService(func(s *Service) {
		res, err := s.Archives(context.Background(), []int64{10110582, 10110581})
		So(err, ShouldBeNil)
		for _, v := range res {
			fmt.Printf("%+v", v)
		}
	}))
}
