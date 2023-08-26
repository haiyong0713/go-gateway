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

func TestNodeInfo(t *testing.T) {
	var (
		c   = context.Background()
		req = &model.NodeInfoParam{
			AID:          10114549,
			NodeID:       17054,
			GraphVersion: 1815,
			Portal:       1,
			Buvid:        "test123",
		}
		mid int64 = 1001
	)
	req.Cursor = 3
	convey.Convey("NodeInfo", t, func(ctx convey.C) {
		graphInfo, err := s.GraphInfo(c, req.AID)
		node, err := s.NodeInfo(c, req, mid, graphInfo)
		nn, _ := json.Marshal(node)
		fmt.Printf("%s", nn)
		ctx.Convey("Then err should be nil.node should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(node, convey.ShouldNotBeNil)
		})
		time.Sleep(3 * time.Second)
	})
}
