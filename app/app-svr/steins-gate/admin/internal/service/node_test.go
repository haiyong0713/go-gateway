package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestService_NodeInfoAudit(t *testing.T) {
	var (
		c       = context.Background()
		graphid = int64(3333)
		nodeid  = int64(1)
	)
	convey.Convey("NodeInfoAudit", t, func(ctx convey.C) {

		nodeShow, err := s.NodeInfoAudit(c, graphid, nodeid)
		fmt.Println(nodeShow, err)
		ctx.Convey("Then err should be nil.nodeShow should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(nodeShow, convey.ShouldNotBeNil)
		})
	})
}

func TestService_EdgeInfoV2Audit(t *testing.T) {
	var (
		c       = context.Background()
		graphid = int64(1414)
		edgeid  = int64(0)
	)
	convey.Convey("EdgeInfoV2Audit", t, func(ctx convey.C) {
		edgeShow, err := s.EdgeInfoV2Audit(c, graphid, edgeid)
		nn, _ := json.Marshal(edgeShow)
		fmt.Printf("%s", nn)
		ctx.Convey("Then err should be nil.nodeShow should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(edgeShow, convey.ShouldNotBeNil)
		})
	})
}
