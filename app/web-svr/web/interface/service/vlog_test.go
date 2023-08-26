package service

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/web/interface/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_Vlog(t *testing.T) {
	Convey("TestService_Vlog should return without err", t, WithService(func(svf *Service) {
		param := &model.VlogParam{
			TID:   21616,
			ChnID: 21616,
			Build: 8550,
			Buvid: "7d5fd2339dbbbd930bd045474bda720b",
			Ps:    20,
			Pn:    1,
		}
		res, err := svf.Vlog(context.Background(), param)
		So(err, ShouldBeNil)
		So(len(res), ShouldBeGreaterThan, 0)
	}))

}

func TestService_VlogRank(t *testing.T) {
	Convey("TestService_VlogRank should return without err", t, WithService(func(svf *Service) {
		param := &model.VlogRankParam{
			TID: 21616,
			Ps:  20,
			Pn:  1,
		}
		res := svf.VlogRank(context.Background(), param)
		So(len(res), ShouldBeGreaterThan, 0)
	}))

}
