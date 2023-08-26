package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_S9Result(t *testing.T) {
	Convey("test service Search", t, WithService(func(s *Service) {
		mid := int64(10000)
		sid := int64(1)
		res, err := s.S9Result(context.Background(), mid, sid)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}

func TestService_S9Record(t *testing.T) {
	Convey("S9Record", t, WithService(func(s *Service) {
		mid := int64(10000)
		sid := int64(1)
		res, err := s.S9Record(context.Background(), mid, sid, 1, 10)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
