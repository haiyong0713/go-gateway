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

func TestEdgeInfoV2(t *testing.T) {
	var (
		c         = context.Background()
		req       = &model.EdgeInfoV2Param{}
		nodeParam = model.NodeInfoParam{
			AID:          10114549,
			EdgeID:       0,
			GraphVersion: 2091,
			Portal:       0,
			Buvid:        "test123",
		}
		mid int64 = 1874
	)
	req.NodeInfoParam = nodeParam
	req.Cursor = -1
	convey.Convey("EdgeInfo", t, func(ctx convey.C) {
		graphInfo, err := s.GraphInfo(c, req.AID)
		node, err := s.EdgeInfoV2(c, req, mid, graphInfo)
		nn, _ := json.Marshal(node)
		fmt.Printf("%s", nn)
		ctx.Convey("Then err should be nil.node should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(node, convey.ShouldNotBeNil)
		})
		time.Sleep(3 * time.Second)
	})
}

func TestEdgeInfoV2Pre(t *testing.T) {
	var (
		c         = context.Background()
		req       = &model.EdgeInfoV2PreReq{}
		mid int64 = 1874
	)
	req.AID = 10114549
	req.EdgeID = 0
	req.Cursor = -1
	convey.Convey("EdgeInfoV2Preview", t, func(ctx convey.C) {
		graphInfo, err := s.GraphInfo(c, req.AID)
		edge, err := s.EdgeInfoV2Preview(c, req, mid, graphInfo)
		nn, _ := json.Marshal(edge)
		fmt.Printf("%s", nn)
		ctx.Convey("Then err should be nil.node should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(edge, convey.ShouldNotBeNil)
		})
		time.Sleep(3 * time.Second)
	})
}
