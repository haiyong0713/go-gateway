package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_TopPhotoArcs(t *testing.T) {
	Convey("Official", t, WithService(func(s *Service) {
		res, err := s.TopPhotoArcs(context.Background(), []int64{14139334})
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
