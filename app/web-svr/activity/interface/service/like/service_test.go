package like

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"fmt"

	. "github.com/smartystreets/goconvey/convey"
)

var svf *Service

func init() {
	dir, _ := filepath.Abs("../../cmd/activity-test.toml")
	flag.Set("conf", dir)
	if err := conf.Init(); err != nil {
		panic(err)
	}
	if svf == nil {
		svf = New(conf.Conf)
	}
	time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		f(svf)
	}
}

func TestService_Subject(t *testing.T) {
	Convey("should return without err", t, WithService(func(svf *Service) {
		res, err := svf.Subject(context.Background(), 1)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	}))
}

func TestService_LikeAct(t *testing.T) {
	Convey("should return without err", t, WithService(func(svf *Service) {
		res, err := svf.LikeAct(context.Background(), &like.ParamAddLikeAct{Sid: 10296, Lid: 13513, Score: 1}, 27515615)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	}))
}

func TestService_LikeActList(t *testing.T) {
	Convey("should return without err", t, WithService(func(svf *Service) {
		res, err := svf.LikeActList(context.Background(), 10296, 2089809, []int64{13510, 13511, 13514, 13513})
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
		fmt.Printf("res %v", res)
	}))
}

func TestService_StoryKingAct(t *testing.T) {
	Convey("should return without err", t, WithService(func(svf *Service) {
		res, err := svf.StoryKingAct(context.Background(), &like.ParamStoryKingAct{Sid: 10365, Lid: 2357, Score: 1}, 27515615)
		So(err, ShouldBeNil)
		fmt.Printf("%v", res)
	}))
}

func TestService_StoryKingLeftTime(t *testing.T) {
	Convey("should return without err", t, WithService(func(svf *Service) {
		res, err := svf.StoryKingLeftTime(context.Background(), 10296, 55555)
		So(err, ShouldBeNil)
		fmt.Printf("%d", res)
	}))
}

func TestService_storyEachUsed(t *testing.T) {
	Convey("should return without err", t, WithService(func(svf *Service) {
		res, err := svf.storyEachUsed(context.Background(), 10296, 216761, 13538)
		So(err, ShouldBeNil)
		fmt.Printf("%d", res)
	}))
}

func TestService_StoryKingList(t *testing.T) {
	Convey("should return without err", t, WithService(func(svf *Service) {
		res, err := svf.StoryKingList(context.Background(), &like.ParamList{Sid: 1, Pn: 1, Ps: 100, Type: "random"}, 27515257)
		So(err, ShouldBeNil)
		fmt.Printf("%v", res)
	}))
}

func TestService_UpList(t *testing.T) {
	Convey("should return without err", t, WithService(func(svf *Service) {
		res, err := svf.UpList(context.Background(), &like.ParamList{Sid: 10259, Pn: 1, Ps: 100, Type: "random"}, 27515257)
		So(err, ShouldBeNil)
		if res != nil {
			if len(res.List) > 0 {
				for _, v := range res.List {
					fmt.Printf("%v %v", v.Item, v.Object)
				}
			}
		}
	}))
}

func TestService_LikeMyList(t *testing.T) {
	Convey("LikeMyList", t, WithService(func(svf *Service) {
		sid := int64(10429)
		mid := int64(27515249)
		res, err := svf.LikeMyList(context.Background(), sid, mid, 15, 1)
		So(err, ShouldBeNil)
		if res != nil {
			if len(res.List) > 0 {
				for _, v := range res.List {
					fmt.Printf("%v %v", v.Item, v.Object)
				}
			}
		}
	}))
}

func TestService_ActLikes(t *testing.T) {
	Convey("ActLikes", t, WithService(func(svf *Service) {
		arg := &like.ArgActLikes{Sid: 10436, Ps: 15, Pn: 1, SortType: 3, Mid: 15555180}
		_, err := svf.ActLikes(context.Background(), arg)
		So(err, ShouldBeNil)
	}))
}
