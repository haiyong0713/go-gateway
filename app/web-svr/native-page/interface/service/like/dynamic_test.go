package like

import (
	"context"
	"fmt"
	"testing"

	dymdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_ActIndex(t *testing.T) {
	Convey("ActIndex", t, WithService(func(s *Service) {
		_, err := s.ActIndex(context.Background(), &dymdl.ParamActIndex{PageID: 41}, 15587)
		So(err, ShouldBeNil)
	}))
}

func TestService_LiveDynx(t *testing.T) {
	Convey("LiveDyn", t, WithService(func(s *Service) {
		rly, err := s.LiveDyn(context.Background(), &dymdl.ParamLiveDyn{RoomIDs: []int64{460460}}, 88894834)
		So(err, ShouldBeNil)
		fmt.Printf("%v", rly)
	}))
}
