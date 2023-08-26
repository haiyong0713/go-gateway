package dao

import (
	"context"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_HgetAllUserDoc(t *testing.T) {
	var (
		ctx = context.TODO()
		mid = int64(1111112645)
	)
	Convey("add player config ", t, func() {
		res, err := testDao.HgetAllUserDoc(ctx, mid, "", 1)
		t.Logf("%+v", res)
		So(err, ShouldBeNil)
	})
}

func TestDao_HMsetUserDoc(t *testing.T) {
	var (
		ctx = context.TODO()
		mid = int64(1111112645)
		dst = map[string]string{
			"switch": strconv.FormatBool(true),
		}
	)
	Convey("add player config ", t, func() {
		err := testDao.HMsetUserDoc(ctx, mid, "", 1, dst)
		So(err, ShouldBeNil)
	})
}

func TestDao_DelUserDoc(t *testing.T) {
	var (
		ctx = context.TODO()
		mid = int64(1111112645)
	)
	Convey("add player config ", t, func() {
		ok, err := testDao.DelUserDoc(ctx, mid, "", 1)
		So(ok, ShouldBeTrue)
		So(err, ShouldBeNil)
	})
}
