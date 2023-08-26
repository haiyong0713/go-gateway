package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/web-svr/esports/interface/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_Search(t *testing.T) {
	Convey("test service Search", t, WithService(func(s *Service) {
		arg := &model.ParamSearch{
			Keyword: "赛",
			Pn:      1,
			Ps:      30,
		}
		res, err := s.Search(context.Background(), 0, arg, "")
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
