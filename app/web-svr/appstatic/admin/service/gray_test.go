package service

import (
	"fmt"
	"testing"

	"go-gateway/app/web-svr/appstatic/admin/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_Gray(t *testing.T) {
	Convey("TestService_Gray", t, WithService(func(svf *Service) {
		id := int64(1)
		res, err := svf.Gray(id)
		So(err, ShouldBeNil)
		fmt.Printf("%+v", res)
	}))
}

func TestService_AddGray(t *testing.T) {
	Convey("TestService_AddGray", t, WithService(func(svf *Service) {
		v := &model.ResourceGray{
			ResourceId:      1,
			Strategy:        1,
			Salt:            "test",
			BucketStart:     1,
			BucketEnd:       2,
			WhitelistInput:  "1,2,3",
			WhitelistUpload: "test",
		}
		err := svf.AddGray(v)
		So(err, ShouldBeNil)
	}))
}

func TestService_SaveGray(t *testing.T) {
	Convey("TestService_AddGray", t, WithService(func(svf *Service) {
		v := &model.ResourceGray{
			ID:              1,
			ResourceId:      1,
			Strategy:        1,
			Salt:            "test",
			BucketStart:     1,
			BucketEnd:       2,
			WhitelistInput:  "1,2,3",
			WhitelistUpload: "test",
			ManualUpdate:    1,
		}
		err := svf.SaveGray(v)
		So(err, ShouldBeNil)
	}))
}
