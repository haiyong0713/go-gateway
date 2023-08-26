package channel

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/model/channel"
	"path/filepath"
	"testing"
	"time"

	. "github.com/glycerine/goconvey/convey"
)

var (
	svf *Service
	mid int64 = 27515255
	c         = context.Background()
)

func init() {
	dir, _ := filepath.Abs("../../cmd/web-interface-test.toml")
	_ = flag.Set("conf", dir)
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Init(conf.Conf.Log)
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

func printDetail(detail interface{}) {
	data, _ := json.MarshalIndent(detail, "", "\t")
	fmt.Printf("%+v\n", string(data))
}

func TestService_Red(t *testing.T) {
	Convey("TestService_Red", t, WithService(func(s *Service) {
		res, err := s.Red(c, mid)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_Subscribe(t *testing.T) {
	Convey("TestService_Subscribe", t, WithService(func(s *Service) {
		var (
			channelID int64 = 34
		)
		err := s.Subscribe(c, mid, &channel.SubscribeReq{ID: channelID})
		subList, err := s.SubscribedList(c, mid)

		var chanIDs []int64
		for _, channel := range subList.StickChannels {
			chanIDs = append(chanIDs, channel.ID)
		}
		for _, channel := range subList.NormalChannels {
			chanIDs = append(chanIDs, channel.ID)
		}
		So(err, ShouldBeNil)
		So(chanIDs, ShouldContain, channelID)
	}))
}

func TestService_Unsubscribe(t *testing.T) {
	Convey("TestService_Unsubscribe", t, WithService(func(s *Service) {
		var (
			channelID int64 = 34
		)
		err := s.Unsubscribe(c, mid, &channel.UnsubscribeReq{ID: channelID})
		subList, err := s.SubscribedList(c, mid)

		var chanIDs []int64
		for _, channel := range subList.StickChannels {
			chanIDs = append(chanIDs, channel.ID)
		}
		for _, channel := range subList.NormalChannels {
			chanIDs = append(chanIDs, channel.ID)
		}
		So(err, ShouldBeNil)
		So(chanIDs, ShouldNotContain, channelID)
	}))
}

func TestService_SubscribedList(t *testing.T) {
	Convey("TestService_SubscribedList", t, WithService(func(s *Service) {
		res, err := s.SubscribedList(c, mid)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_Stick(t *testing.T) {
	Convey("TestService_Stick", t, WithService(func(s *Service) {
		var (
			req = &channel.StickReq{
				StickList:  "22692,600",
				NormalList: "",
			}
		)
		err := s.Stick(c, mid, req)
		So(err, ShouldBeNil)
		res, err := s.SubscribedList(c, mid)
		if res != nil {
			printDetail(res)
		}
	}))
}

func TestService_HotList(t *testing.T) {
	Convey("TestService_HotList", t, WithService(func(s *Service) {
		var (
			req = &channel.HotListReq{
				Offset:   "",
				NeedArc:  true,
				PageSize: 2,
			}
		)
		res, err := s.HotList(c, mid, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_ViewList(t *testing.T) {
	Convey("TestService_ViewList", t, WithService(func(s *Service) {
		res, err := s.ViewList(c, mid)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_Detail(t *testing.T) {
	Convey("TestService_Detail", t, WithService(func(s *Service) {
		var (
			channelID int64 = 600
		)
		res, err := s.Detail(c, mid, &channel.WebDetailReq{ID: channelID})
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_FeaturedList(t *testing.T) {
	Convey("TestService_FeaturedList", t, WithService(func(s *Service) {
		var (
			req = &channel.FeaturedListReq{
				ChannelID:  600,
				Offset:     "",
				FilterType: 2020,
				PageSize:   5,
			}
		)
		res, err := s.FeaturedList(c, mid, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_MultipleList(t *testing.T) {
	Convey("TestService_MultipleList", t, WithService(func(s *Service) {
		var (
			req = &channel.MultipleListReq{
				ChannelID: 600,
				Offset:    "",
				SortType:  channel.MultiHot,
				PageSize:  5,
			}
		)
		res, err := s.MultipleList(c, mid, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_Search(t *testing.T) {
	Convey("TestService_Search", t, WithService(func(s *Service) {
		var (
			req = &channel.SearchReq{
				Keyword:  "吃饭",
				Page:     1,
				PageSize: 5,
			}
		)
		res, err := s.Search(c, mid, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
