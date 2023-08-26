package service

import (
	"context"
	"encoding/json"
	"testing"

	"go-gateway/app/web-svr/space/interface/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestService_DynamicList(t *testing.T) {
	convey.Convey("test dynamic list", t, WithService(func(s *Service) {
		arg := &model.DyListArg{
			Vmid: 908085,
			Pn:   1,
			Qn:   16,
		}
		list, err := s.DynamicList(context.Background(), arg)
		convey.So(err, convey.ShouldBeNil)
		bs, _ := json.Marshal(list)
		convey.Println(string(bs))
	}))
}

func TestService_BehaviorList(t *testing.T) {
	convey.Convey("test dynamic list", t, WithService(func(s *Service) {
		var (
			vmid     int64 = 2089809
			lastTime int64 = 1545289186
			ps             = 20
		)
		list := s.BehaviorList(context.Background(), 0, vmid, lastTime, ps)
		convey.Println(len(list))
		bs, _ := json.Marshal(list)
		convey.Println(string(bs))
	}))
}
