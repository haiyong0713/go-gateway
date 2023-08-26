package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/web/interface/model"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_TagAids(t *testing.T) {
	Convey("should return without err", t, WithService(func(svf *Service) {
		total, res, err := svf.TagAids(context.Background(), 1, 1, 1)
		So(err, ShouldBeNil)
		So(total, ShouldBeGreaterThan, 0)
		So(len(res), ShouldBeGreaterThan, 0)
	}))

}

func TestService_TagArchives(t *testing.T) {
	Convey("TestService_TagArchives", t, WithService(func(svf *Service) {
		req := &model.TagArcsReq{
			TagID:  600,
			Source: 0,
			Offset: "",
			PS:     20,
		}
		reply, err := svf.TagArchives(context.Background(), req)
		data, _ := json.MarshalIndent(reply, "", "\t")
		fmt.Printf("TestService_TagArchives\n%s", data)
		So(err, ShouldBeNil)
	}))
}
