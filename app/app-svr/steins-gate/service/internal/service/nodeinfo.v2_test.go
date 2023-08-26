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

func TestNodeInfoV2(t *testing.T) {
	var (
		c    = context.Background()
		req  = &model.EdgeInfoV2Param{}
		node = model.NodeInfoParam{
			AID:          10114549,
			NodeID:       0,
			GraphVersion: 1815,
			Portal:       0,
			Buvid:        "test123",
		}
		mid int64 = 1065
	)
	req.NodeInfoParam = node
	req.Cursor = -1
	convey.Convey("NodeInfo", t, func(ctx convey.C) {
		graphInfo, err := s.GraphInfo(c, req.AID)
		node, err := s.NodeInfoV2(c, req, mid, graphInfo)
		nn, _ := json.Marshal(node)
		fmt.Printf("%s", nn)
		ctx.Convey("Then err should be nil.node should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(node, convey.ShouldNotBeNil)
		})
		time.Sleep(3 * time.Second)
	})
}
