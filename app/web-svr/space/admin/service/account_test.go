package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_Fans(t *testing.T) {
	Convey("Official", t, WithService(func(s *Service) {
		res, err := s.Fans(context.Background(), []int64{27515257, 27515258})
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
