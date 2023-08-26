package dao

import (
	"github.com/glycerine/goconvey/convey"
	"testing"
)

func Test_GetOpenIDByMID(t *testing.T) {
	convey.Convey("GetOpenIDByMID", t, func() {
		mid := int64(2082679085)
		res, err := testD.GetOpenIDByMID(mid, "f9a9yne6hp1x4mtj")
		convey.ShouldBeNil(err)
		convey.ShouldNotBeNil(res)
	})
}

func Test_GetMIDByUID(t *testing.T) {
	convey.Convey("GetMIDByUID", t, func() {
		uid := "2082679085"
		res, err := testD.GetMIDByUID(uid, "f9a9yne6hp1x4mtj")
		convey.ShouldBeNil(err)
		convey.ShouldNotBeNil(res)
	})
}

func Test_GetMIDByOpenID(t *testing.T) {
	convey.Convey("GetMIDByOpenID", t, func() {
		openID := "ba2d8e15511c4b1ab9f9958ffb9c537e"
		res, err := testD.GetMIDByOpenID(openID, "f9a9yne6hp1x4mtj")
		convey.ShouldBeNil(err)
		convey.ShouldNotBeNil(res)
	})
}
