package splash_screen

import (
	. "github.com/glycerine/goconvey/convey"
	splashModel "go-gateway/app/app-svr/app-feed/admin/model/splash_screen"
	"testing"
)

func Test_GetSelectConfigListAll(t *testing.T) {
	Convey("GetSelectConfigListAll", t, func() {
		res, count, err := testD.GetSelectConfigListByPage(1, 10)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		So(count, ShouldBeGreaterThanOrEqualTo, 0)
	})
}

func Test_SaveCategories(t *testing.T) {
	Convey("SaveCategories", t, func() {
		categories := []*splashModel.Category{
			{
				ID:   1,
				Name: "日常",
			},
			{
				ID:   2,
				Name: "测试1",
			},
			{
				ID:   8,
				Name: "测试2",
			},
		}
		res, err := testD.SaveCategories(categories, "test")
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	})
}
