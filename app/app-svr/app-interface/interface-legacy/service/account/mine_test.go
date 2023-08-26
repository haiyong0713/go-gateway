package account

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"

	"go-common/library/conf/paladin.v2"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

var s *Service

func init() {
	dir, _ := filepath.Abs("../../cmd/app-interface-test.toml")
	flag.Set("conf", dir)
	c := conf.Conf
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	if err := paladin.Get("app-interface-test.toml").UnmarshalTOML(c); err != nil {
		panic(err)
	}
	s = New(c)
	time.Sleep(3 * time.Second)
}

func Test_Mine(t *testing.T) {
	Convey("Test_Mine", t, func() {
		mine, err := s.Mine(context.Background(), 27515255, "ios", "hans", "", "", "", 0, 1, 0, 0, "phone", "", "", "", 0)
		So(err, ShouldBeNil)
		Printf("%+v", mine.Answer)
		for _, s := range mine.Sections {
			Printf("%+v\n", s.Title)
			for _, item := range s.Items {
				Printf("%+v\n", item.Title)
			}
		}
		for _, s := range mine.SectionsV2 {
			Printf("%+v\n", s.Title)
			for _, item := range s.Items {
				Printf("%+v\n", item.Title)
			}
		}
	})
}

func Test_MineIpad(t *testing.T) {
	Convey("Test_Mine", t, func() {
		mine, err := s.MineIpad(context.Background(), 16840123, "ios", "hans", "", "", "", "", "", 0, 0, 0, 0, "")
		So(err, ShouldBeNil)
		for _, item := range mine.IpadSections {
			Printf("%+v\n", item.Title)
		}
	})
}

func Test_MyInfo(t *testing.T) {
	Convey("Test_MyInfo", t, func() {
		myinfo, err := s.Myinfo(context.Background(), 27515251, "iphone")
		So(err, ShouldBeNil)
		Printf("%+v", myinfo)
	})
}

func Test_GameTipsDmpFilter(t *testing.T) {
	in := []*space.GameTip{
		{
			ID:         1,
			IsDirected: 0,
		},
		{
			ID:         2,
			IsDirected: 0,
		}, //没有人群包限制
		{
			ID:         3,
			IsDirected: 1,
			DmpId:      0,
		}, //命中的人群包
		{
			ID:         4,
			IsDirected: 1,
			DmpId:      1,
		}, //不存在的人群包
		{
			ID:         5,
			IsDirected: 1,
			DmpId:      101,
		}, //没有命中的人群包
	}
	result := map[int64]struct{}{
		1: {},
		2: {},
		3: {},
	}
	gameTips := s.gameTipsDmpFilter(context.Background(), in, 2553322)
	for _, v := range gameTips {
		_, ok := result[v.ID]
		assert.Equal(t, true, ok)
	}
}

func TestBirthday(t *testing.T) {
	now := time.Now()
	timeTest := []struct {
		time time.Time
		res  bool
	}{
		{
			time: now.AddDate(-1, 0, 0),
			res:  true,
		},
		{
			time: now.AddDate(-9, 0, 0),
			res:  true,
		},
		{
			time: now.AddDate(0, -2, 0),
			res:  false,
		},
		{
			time: now.AddDate(-2, 0, -2),
			res:  false,
		},
	}
	for _, v := range timeTest {
		assert.Equal(t, v.res, isBirthday(v.time, now))
	}
}
