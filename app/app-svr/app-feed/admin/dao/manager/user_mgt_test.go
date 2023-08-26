package manager

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_UserAdmin(t *testing.T) {
	convey.Convey("SpecialCards", t, func(ctx convey.C) {
		var (
			bs []byte
		)
		name := "quguolin"
		level := 2
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.UserRole(name, level)
			ctx.So(err, convey.ShouldBeNil)
			bs, _ = json.Marshal(res)
			fmt.Println(string(bs))
		})

	})
}

func TestDao_UserGroup(t *testing.T) {
	convey.Convey("SpecialCards", t, func(ctx convey.C) {
		var (
			bs []byte
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.UserGroupByPids([]int{})
			ctx.So(err, convey.ShouldBeNil)
			bs, _ = json.Marshal(res)
			fmt.Println(string(bs))
		})

	})
}

func TestDao_UserGroupByName(t *testing.T) {
	convey.Convey("SpecialCards", t, func(ctx convey.C) {
		var (
			bs []byte
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.UserGroupByName("test")
			ctx.So(err, convey.ShouldBeNil)
			bs, _ = json.Marshal(res)
			fmt.Println(string(bs))
		})

	})
}
