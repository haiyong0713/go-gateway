package dao

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_SetUserDocRds(t *testing.T) {
	var (
		ctx   = context.TODO()
		mid   = int64(1111112645)
		buvid = "Y04C74038475378A45019D205CCF8259B329"
	)
	Convey("set mid empty cache", t, func() {
		err := testDao.SetUserDocRds(ctx, mid, buvid, 1, []byte{})
		So(err, ShouldBeNil)
	})
	Convey("set buvid empty cache", t, func() {
		err := testDao.SetUserDocRds(ctx, 0, buvid, 1, []byte{})
		So(err, ShouldBeNil)
	})
}

func TestDao_UserDocRds(t *testing.T) {
	var (
		ctx   = context.TODO()
		mid   = int64(11111126456)
		buvid = "Y04C74038475378A45019D205CCF8259B329z"
	)
	Convey("mid empty cache", t, func() {
		rm, err := testDao.UserDocRds(ctx, mid, buvid, 1)
		So(rm, ShouldBeNil)
		So(err, ShouldBeNil)
	})
	Convey("buvid empty cache", t, func() {
		rm, err := testDao.UserDocRds(ctx, 0, buvid, 1)
		So(rm, ShouldBeNil)
		So(err, ShouldBeNil)
	})
}

func TestDao_UserDocRds1(t *testing.T) {
	var (
		ctx   = context.TODO()
		mid   = int64(1111112645)
		buvid = "Y04C74038475378A45019D205CCF8259B329"
	)
	Convey("mid empty cache", t, func() {
		rm, err := testDao.UserDocRds(ctx, mid, buvid, 1)
		So(rm, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
	Convey("buvid empty cache", t, func() {
		rm, err := testDao.UserDocRds(ctx, 0, buvid, 1)
		So(rm, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}
