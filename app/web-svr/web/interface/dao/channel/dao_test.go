package channel

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/conf"
	chanmdl "go-gateway/app/web-svr/web/interface/model/channel"

	changrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	. "github.com/glycerine/goconvey/convey"
)

var (
	dao          *Dao
	mid          int64 = 27515255
	categoryType int32 = 100
	channelID    int64 = 600
	c                  = context.Background()
)

func init() {
	dir, _ := filepath.Abs("../../cmd/web-interface-test.toml")
	_ = flag.Set("conf", dir)
	if err := conf.Init(); err != nil {
		panic(err)
	}
	log.Init(conf.Conf.Log)
	if dao == nil {
		dao = New(conf.Conf)
	}
	time.Sleep(time.Second)
}

func WithDao(f func(d *Dao)) func() {
	return func() {
		f(dao)
	}
}

func printDetail(detail interface{}) {
	data, _ := json.MarshalIndent(detail, "", "\t")
	fmt.Printf("%+v\n", string(data))
}

func TestDao_SubscribedChannel(t *testing.T) {
	Convey("TestDao_SubscribedChannel", t, WithDao(func(d *Dao) {
		res, err := d.SubscribedChannel(c, mid)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_NewNotify(t *testing.T) {
	Convey("TestDao_NewNotify", t, WithDao(func(d *Dao) {
		res, err := d.NewNotify(c, mid)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_Category(t *testing.T) {
	Convey("TestDao_Category", t, WithDao(func(d *Dao) {
		res, err := d.Category(c)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_ViewChannel(t *testing.T) {
	Convey("TestDao_ViewChannel", t, WithDao(func(d *Dao) {
		res, err := d.ViewChannel(c, mid)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_ChannelList(t *testing.T) {
	Convey("TestDao_ChannelList", t, WithDao(func(d *Dao) {
		req := &changrpc.ChannelListReq{
			Mid:          mid,
			CategoryType: categoryType,
			Offset:       "",
			Ps:           5,
		}
		res, err := d.ChannelList(c, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_ChannelResourceList(t *testing.T) {
	Convey("TestDao_ChannelResourceList", t, WithDao(func(d *Dao) {
		req := &changrpc.ChannelResourceListReq{
			Mid:          mid,
			Offset:       "",
			CategoryType: categoryType,
			Typ:          1,
			Ps:           5,
			Count:        5,
		}
		res, err := d.ChannelResourceList(c, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_UpdateSubscribe(t *testing.T) {
	Convey("TestDao_UpdateSubscribe", t, WithDao(func(d *Dao) {
		tops := ""
		cids := ""
		err := d.UpdateSubscribe(c, mid, tops, cids)
		So(err, ShouldBeNil)
	}))
}

func TestDao_ChannelDetail(t *testing.T) {
	Convey("TestDao_ChannelDetail", t, WithDao(func(d *Dao) {
		res, err := d.ChannelDetail(c, mid, channelID)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_ResourceList(t *testing.T) {
	Convey("TestDao_ResourceList", t, WithDao(func(d *Dao) {
		req := &changrpc.ResourceListReq{
			ChannelId: channelID,
			TabType:   changrpc.TabType_TAB_TYPE_TOTAL,
			SortType:  changrpc.TotalSortType_SORT_BY_HOT,
			Offset:    "",
			PageSize:  5,
			Mid:       mid,
			Typ:       1,
		}
		res, err := d.ResourceList(c, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_SearchChannel(t *testing.T) {
	Convey("TestDao_SearchChannel", t, WithDao(func(d *Dao) {
		res, err := d.SearchChannel(c, mid, []int64{channelID})
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_SearchChannelsInfo(t *testing.T) {
	Convey("TestDao_SearchChannelsInfo", t, WithDao(func(d *Dao) {
		req := &changrpc.SearchChannelsInfoReq{
			Mid:   mid,
			Cids:  []int64{channelID},
			Count: 5,
		}
		res, err := d.SearchChannelsInfo(c, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_SearchEs(t *testing.T) {
	Convey("TestDao_SearchEs", t, WithDao(func(d *Dao) {
		res, _, err := d.SearchEs(c, "测试", 5, 5, chanmdl.EsStateOK)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_HotChannel(t *testing.T) {
	Convey("TestDao_HotChannel", t, WithDao(func(d *Dao) {
		req := &changrpc.HotChannelReq{
			Mid:    mid,
			Offset: "",
			Ps:     5,
			Count:  5,
			Typ:    1,
		}
		res, err := d.HotChannel(c, mid, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_RelativeChannel(t *testing.T) {
	Convey("TestDao_RelativeChannel", t, WithDao(func(d *Dao) {
		req := &changrpc.RelativeChannelReq{
			Mid:  mid,
			Cids: []int64{channelID},
		}
		res, err := d.RelativeChannel(c, req)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
