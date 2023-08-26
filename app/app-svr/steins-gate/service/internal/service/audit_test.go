package service

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/glycerine/goconvey/convey"
)

func TestService_GraphAudit(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("GraphInfo", t, func(ctx convey.C) {
		//err := s.GraphAudit(c, &model.AuditParam{ // 拒绝发消息
		//	ID:           1,
		//	WithNotify:   1,
		//	State:        -20,
		//	RejectReason: "您的稿件不行，主要是【%s】 。。。",
		//	RejectTitle:  "我这个title真的很长吧",
		//})
		err := s.GraphAudit(c, &model.AuditParam{ // 通过
			ID:         1,
			WithNotify: 1,
			State:      1,
		})
		fmt.Println(err)
		ctx.Convey("Then err should be nil.a should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		// time.Sleep(3 * time.Second)
	})
}
