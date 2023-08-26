package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestServiceGraphInfo(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(3333)
	)
	convey.Convey("GraphInfo", t, func(ctx convey.C) {
		a, err := s.GraphInfo(c, aid)
		fmt.Println(a, err)
		ctx.Convey("Then err should be nil.a should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(a, convey.ShouldNotBeNil)
		})
	})
}

func TestServiceView(t *testing.T) {
	var (
		c   = context.Background()
		req = &api.ViewReq{
			Aid: 10200126,
			Mid: 0,
		}
	)
	convey.Convey("View", t, func(ctx convey.C) {
		resp, err := s.View(c, req)
		fmt.Println(resp, err)
		ctx.Convey("Then err should be nil.resp should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(resp, convey.ShouldNotBeNil)
		})
	})
}

func TestServiceViews(t *testing.T) {
	var (
		c   = context.Background()
		req = &api.ViewsReq{
			Aids:            []int64{10113631, 10113518},
			AidsWithHistory: []int64{10113631, 10113518, 10113690, 10113691},
			Mid:             111006313,
			Buvid:           "gaga",
		}
	)
	/**
		GraphIDs: []int64{659, 631, 622},
	AIDs:     []int64{10113518, 10113631, 10113690},
	MID:      111006313,
	*/
	convey.Convey("Views", t, func(ctx convey.C) {
		resp, err := s.Views(c, req)
		fmt.Println(err)
		str, _ := json.Marshal(resp)
		fmt.Println(string(str))
		time.Sleep(2 * time.Second)
		ctx.Convey("Then err should be nil.resp should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(resp, convey.ShouldNotBeNil)
		})
	})
}

func TestServiceGraphView(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(10113690)
	)
	convey.Convey("GraphView", t, func(ctx convey.C) {
		page, graph, eval, err := s.GraphView(c, aid)
		fmt.Println(page, graph, eval, err)
		ctx.Convey("Then err should be nil.page and graph should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(page, convey.ShouldNotBeNil)
			ctx.So(graph, convey.ShouldNotBeNil)
		})
	})
}

func TestServiceMarkEvaluations(t *testing.T) {
	var (
		c    = context.Background()
		mid  = int64(111005921)
		aids []int64
	)
	aids = append(aids, 10200126)
	aids = append(aids, 10113593)
	convey.Convey("MarkEvaluations", t, func(ctx convey.C) {
		res, err := s.MarkEvaluations(c, mid, aids)
		fmt.Println(res, err)
		ctx.Convey("Then err should be nil. res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
