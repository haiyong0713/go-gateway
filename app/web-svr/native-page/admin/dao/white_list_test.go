package dao

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go-gateway/app/web-svr/native-page/admin/model"
)

var record = &model.WhiteListRecord{
	ID:          1,
	Mid:         1,
	Creator:     "test",
	CreatorUID:  2,
	Modifier:    "test",
	ModifierUID: 3,
	State:       StateValid,
}

func TestDao_AddWhiteList(t *testing.T) {
	Convey("TestDao_AddWhiteList", t, func() {
		c := context.Background()
		Convey("When everything goes positive", func() {
			res, err := d.AddWhiteList(c, record)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})
	})
}

func TestDao_UpdateWhiteList(t *testing.T) {
	Convey("TestDao_UpdateWhiteList", t, func() {
		c := context.Background()
		Convey("When everything goes positive", func() {
			attrs := map[string]interface{}{
				"modifier":     "test_2",
				"modifier_uid": 3,
				"state":        StateInvalid,
			}
			err := d.UpdateWhiteList(c, record.ID, attrs)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDao_WhiteList(t *testing.T) {
	Convey("TestDao_WhiteList", t, func() {
		c := context.Background()
		Convey("When everything goes positive", func() {
			res, _, err := d.WhiteList(c, record.Mid, 1, 20)
			for _, v := range res {
				fmt.Printf("%+v\n", v)
			}
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})
	})
}
