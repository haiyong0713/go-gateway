package like

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/web-svr/activity/interface/model/prediction"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_PreJudge(t *testing.T) {
	Convey("PreJudge should return without err", t, WithService(func(svf *Service) {
		arg := &prediction.PreParams{Sid: 10292, Point: 55, NickName: "layang"}
		lis, err := svf.PreJudge(context.Background(), arg)
		So(err, ShouldBeNil)
		fmt.Printf("%v", lis)
	}))
}

func TestService_PreListSet(t *testing.T) {
	Convey("PreListSet should return without err", t, WithService(func(svf *Service) {
		err := svf.PreListSet(context.Background(), 10292)
		So(err, ShouldBeNil)
	}))
}

func TestService_ItemListSet(t *testing.T) {
	Convey("ItemListSet should return without err", t, WithService(func(svf *Service) {
		err := svf.ItemListSet(context.Background(), 17)
		So(err, ShouldBeNil)
	}))
}

func TestService_PreItemUp(t *testing.T) {
	Convey("ItemListSet should return without err", t, WithService(func(svf *Service) {
		err := svf.PreItemUp(context.Background(), 368)
		So(err, ShouldBeNil)
	}))
}
