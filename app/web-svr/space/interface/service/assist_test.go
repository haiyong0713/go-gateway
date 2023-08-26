package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_RiderList(t *testing.T) {
	Convey("article list test", t, WithService(func(s *Service) {
		mid := int64(27515256)
		pn := 1
		ps := 10
		res, err := s.RiderList(context.Background(), mid, int(pn), int(ps))
		So(err, ShouldBeNil)
		So(res, ShouldBeNil)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
