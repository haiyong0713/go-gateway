package api

import (
	"context"
	"encoding/json"
	"testing"

	"go-common/library/ecode"

	"github.com/smartystreets/goconvey/convey"
)

var client ArchiveHonorClient

func init() {
	var err error
	client, err = NewClient(nil)
	if err != nil {
		panic(err)
	}
}

func TestHonor(t *testing.T) {
	convey.Convey("Honor", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			req = &HonorRequest{Aid: 520053499}
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.Honor(c, req)
			ctx.So(err, convey.ShouldBeNil)
			for _, v := range reply.Honor {
				ctx.Printf("%+v", v)
			}
		})
	})
}

func TestHonors(t *testing.T) {
	convey.Convey("Honors", t, func(ctx convey.C) {
		var (
			c   = context.Background()
			req = &HonorsRequest{Aids: []int64{1, 2}}
		)
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			reply, err := client.Honors(c, req)
			ctx.So(err, convey.ShouldBeNil)
			vv, _ := json.Marshal(reply)
			ctx.Printf("%s", vv)
		})
		aids := []int64{}
		i := 0
		for i = 0; i < 51; i++ {
			aids = append(aids, int64(1))
		}
		req = &HonorsRequest{Aids: aids}
		ctx.Convey("When len(aids) is more than 50", func(ctx convey.C) {
			_, err := client.Honors(c, req)
			ctx.So(ecode.Cause(err).Code(), convey.ShouldEqual, ecode.RequestErr)
		})
	})
}
