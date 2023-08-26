package dao

import (
	"context"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go-common/library/conf/paladin"
)

var (
	daoDao Dao
)

func TestMain(m *testing.M) {
	var err error
	if paladin.DefaultClient, err = paladin.NewFile("../../cmd/cmd/"); err != nil {
		panic(err)
	}
	redis, err := NewRedis()
	if err != nil {
		panic(err)
	}
	db, err := NewDB()
	if err != nil {
		panic(err)
	}
	daoDao, _ = New(redis, db)
	os.Exit(m.Run())
}

func TestDao_UserConf(t *testing.T) {
	var (
		ctx = context.TODO()
		mid = int64(1111112645)
	)
	Convey("userconf rds ", t, func() {
		doc, err := daoDao.UserConf(ctx, mid, 1)
		So(err, ShouldBeNil)
		t.Logf("%+v", doc)
	})
}
