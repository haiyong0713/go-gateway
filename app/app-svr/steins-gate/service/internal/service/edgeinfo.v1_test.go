package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestEdgeInfo(t *testing.T) {
	var (
		c   = context.Background()
		req = &model.NodeInfoParam{
			AID:          10114549,
			EdgeID:       16603,
			GraphVersion: 1815,
			Portal:       0,
			Buvid:        "test123",
		}
		mid int64 = 104
	)
	req.Cursor = -1
	convey.Convey("EdgeInfo", t, func(ctx convey.C) {
		graphInfo, err := s.GraphInfo(c, req.AID)
		node, err := s.EdgeInfo(c, req, mid, graphInfo)
		nn, _ := json.Marshal(node)
		fmt.Printf("%s", nn)
		ctx.Convey("Then err should be nil.node should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(node, convey.ShouldNotBeNil)
		})
		time.Sleep(3 * time.Second)
	})
}

func TestEdgeInfoPreview(t *testing.T) {
	var (
		c   = context.Background()
		req = &model.NodeinfoPreReq{
			AID:    10114549,
			EdgeID: 16677,
			Portal: 0,
		}
		mid int64 = 105
	)
	req.Cursor = -1
	graphInfo, _ := s.dao.GraphInfoPreview(c, req.AID)
	convey.Convey("EdgeInfoPreview", t, func(ctx convey.C) {
		node, err := s.EdgeInfoPreview(c, req, mid, graphInfo)
		nn, _ := json.Marshal(node)
		fmt.Printf("%s", nn)
		ctx.Convey("Then err should be nil.node should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(node, convey.ShouldNotBeNil)
		})
	})
	time.Sleep(3 * time.Second)
}
