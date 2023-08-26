package api

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-common/library/ecode"

	"github.com/smartystreets/goconvey/convey"
)

var client ArchiveClient

func init() {
	var err error
	client, err = NewClient(nil)
	if err != nil {
		panic(err)
	}
}

func TestTypes(t *testing.T) {
	convey.Convey("Types", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.Types(c, &NoArgRequest{})
			ctx.So(err, convey.ShouldBeNil)
			for k, v := range reply.Types {
				ctx.Printf("key:%d id:%d name:%s pid:%d\n", k, v.ID, v.Name, v.Pid)
			}
		})
	})
}

func TestArc(t *testing.T) {
	convey.Convey("TestArc", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.Arc(c, &ArcRequest{Aid: 10113301})
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v\n", reply.Arc)
		})
		ctx.Convey("When error", func(ctx convey.C) {
			reply, err := client.Arc(context.TODO(), &ArcRequest{Aid: 99999999999})
			ctx.So(err, convey.ShouldEqual, ecode.NothingFound)
			ctx.So(reply, convey.ShouldBeNil)
		})
	})
}

func TestArcs(t *testing.T) {
	convey.Convey("TestArcs", t, func(ctx convey.C) {
		var c = context.Background()
		var aids []int64
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.Arcs(c, &ArcsRequest{Aids: aids})
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v\n%+v", reply.Arcs, err)
		})
	})
}

func TestView(t *testing.T) {
	convey.Convey("TestView", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.View(c, &ViewRequest{Aid: 10100696})
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("arc:%+v\n", reply.Arc)
			ctx.Printf("pages:%+v\n", reply.Pages)
		})
		ctx.Convey("When empty", func(ctx convey.C) {
			reply, err := client.View(c, &ViewRequest{Aid: 99999999999})
			ctx.So(err, convey.ShouldNotBeNil)
			ctx.So(reply, convey.ShouldBeNil)
		})
	})
}

func TestViews(t *testing.T) {
	convey.Convey("TestViews", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.Views(c, &ViewsRequest{Aids: []int64{10100696}})
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v\n", reply.Views)
		})
		ctx.Convey("When empty", func(ctx convey.C) {
			arcs, err := client.Views(c, &ViewsRequest{Aids: []int64{99999999999}})
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(arcs.Views, convey.ShouldBeNil)
		})
		ctx.Convey("When len(aid)>50", func(ctx convey.C) {
			var aids []int64
			for i := 0; i < 5; i++ {
				aids = append(aids, -11111)
			}
			_, err := client.Views(c, &ViewsRequest{Aids: aids})
			ctx.So(ecode.Cause(err).Code(), convey.ShouldEqual, ecode.RequestErr)
		})
	})
}

func TestVideoShot(t *testing.T) {
	convey.Convey("VideoShot", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			aid = int64(10114084)
			cid = int64(10164715)
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.VideoShot(c, &VideoShotRequest{Aid: aid, Cid: cid})
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v\n", reply.Vs)
		})
	})
}

func TestUpCount(t *testing.T) {
	convey.Convey("UpCount", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(27515615)
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.UpCount(c, &UpCountRequest{Mid: mid})
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v\n", reply.Count)
		})
	})
}

func TestUpsPassed(t *testing.T) {
	convey.Convey("UpsPassed", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			mids = []int64{27515615}
			pn   = int32(1)
			ps   = int32(10)
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.UpsPassed(c, &UpsPassedRequest{Mids: mids, Pn: pn, Ps: ps})
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v\n", reply.UpsPassed)
		})
	})
}

func TestUpArcs(t *testing.T) {
	convey.Convey("UpArcs", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			mid = int64(27515245)
			pn  = int32(1)
			ps  = int32(20)
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.UpArcs(c, &UpArcsRequest{Mid: mid, Pn: pn, Ps: ps})
			ctx.So(err, convey.ShouldBeNil)
			for _, v := range reply.Arcs {
				ctx.Printf("%d %d \n", v.Aid, v.AttributeV2)
			}
		})
	})
}

func TestArcsWithPlayurl(t *testing.T) {
	convey.Convey("ArcsWithPlayurl", t, func(ctx convey.C) {
		var (
			c    = context.Background()
			aids = []int64{10114084}
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.ArcsWithPlayurl(c, &ArcsWithPlayurlRequest{Aids: aids, Platform: "ios", Ip: "123.111.111.111"})
			ctx.So(err, convey.ShouldBeNil)
			for _, v := range reply.ArcWithPlayurl {
				ctx.Printf("%+v\n", v)
			}
		})
	})
}

func TestCreators(t *testing.T) {
	convey.Convey("TestCreators", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.Creators(c, &CreatorsRequest{Aids: []int64{10318363}})
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v\n", reply.Info)
		})
	})
}

func TestSimpleArc(t *testing.T) {
	convey.Convey("TestSimpleArc", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.SimpleArc(c, &SimpleArcRequest{Aid: 10318363})
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%+v\n", reply.Arc)
		})
	})
}

func TestSimpleArcs(t *testing.T) {
	convey.Convey("TestSimpleArcs", t, func(ctx convey.C) {
		var c = context.Background()
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.SimpleArcs(c, &SimpleArcsRequest{Aids: []int64{10318363, 10100696}})
			ss, _ := json.Marshal(reply)
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%s", ss)
		})
	})
}

func TestArcPlayer(t *testing.T) {
	convey.Convey("ArcsPlayer", t, func(ctx convey.C) {
		video1 := PlayAv{}
		video1.Aid = 880072935
		video1.PlayVideos = append(video1.PlayVideos, &PlayVideo{
			Cid: 10282899,
		}, &PlayVideo{
			Cid: 10282898,
		})
		video2 := PlayAv{}
		video2.Aid = 400103150
		video2.PlayVideos = append(video2.PlayVideos, &PlayVideo{
			Cid: 10223279,
		}, &PlayVideo{})
		req := ArcsPlayerRequest{}
		req = ArcsPlayerRequest{
			BatchPlayArg: &BatchPlayArg{
				MobiApp: "iphone",
				Ip:      "127.0.0.1",
			},
		}
		req.PlayAvs = append(req.PlayAvs, &video1, &video2)
		a, e := client.ArcsPlayer(context.TODO(), &req)
		fmt.Println(e)
		fmt.Printf("%v", a)
	})
}
