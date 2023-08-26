package service

import (
	"context"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/job/internal/model"

	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestServicearcHandle(t *testing.T) {
	var (
		arc = &model.Archive{
			Aid:       10113518,
			State:     0,
			Attribute: 537936115,
		}
	)
	convey.Convey("arcHandle", t, func(ctx convey.C) {
		s.arcHandle(arc)
		ctx.Convey("No return values", func(ctx convey.C) {
		})
	})
}

func TestServicereturnGraph(t *testing.T) {
	var (
		c   = context.Background()
		req = &model.ReqReturnGraph{
			Arc: &api.Arc{
				Title: "123",
				Aid:   10113448,
				Author: api.Author{
					Mid: 1234,
				},
			},
			GraphID: 4,
		}
	)
	convey.Convey("returnGraph", t, func(ctx convey.C) {
		s.returnGraph(c, req)
		ctx.Convey("No return values", func(ctx convey.C) {
		})
	})
}
