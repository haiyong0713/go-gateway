package dao

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_Document(t *testing.T) {
	var (
		ctx      = context.TODO()
		checksum = int64(7949982028423363614)
	)
	Convey("document rds ", t, func() {
		doc, err := daoDao.Document(ctx, checksum)
		So(err, ShouldBeNil)
		t.Logf("%v", string(doc))
	})
}

func TestDao_DocumentRds(t *testing.T) {
	var (
		ctx      = context.TODO()
		checksum = int64(6954176954480247924)
	)
	Convey("document rds ", t, func() {
		doc, err := daoDao.Document(ctx, checksum)
		So(err, ShouldBeNil)
		t.Logf("%v", string(doc))
	})
}

func TestDao_UserConfRds(t *testing.T) {
	var (
		ctx = context.TODO()
		mid = int64(1111112645)
	)
	Convey("userconf rds ", t, func() {
		doc, err := daoDao.UserConfRds(ctx, mid, 1)
		So(err, ShouldBeNil)
		t.Logf("%+v", doc)
	})
}

func TestDao_UserDocRds(t *testing.T) {

}

func TestDao_SetUserDocRds(t *testing.T) {

}
