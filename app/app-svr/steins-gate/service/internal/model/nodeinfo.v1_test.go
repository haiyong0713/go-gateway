package model

import (
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestModelFromNode(t *testing.T) {
	var (
		nodeinfo = &api.GraphNode{
			ShowTime: 0,
		}
		choices = []*Choice{
			{
				CID:       123,
				IsDefault: 0,
			},
			{
				CID:       345,
				IsDefault: 0,
			},
			{
				CID:       567,
				IsDefault: 0,
			},
			{
				CID:       789,
				IsDefault: 0,
			},
			{
				CID:       567,
				IsDefault: 1,
			},
		}
		aid = int64(9999)
	)
	convey.Convey("FromNode", t, func(ctx convey.C) {
		pre := new(Preload)
		pre.FromNode(nodeinfo, choices, aid)
		str, _ := json.Marshal(pre)
		fmt.Println(string(str))
		ctx.Convey("No return values", func(ctx convey.C) {
		})
	})
}
