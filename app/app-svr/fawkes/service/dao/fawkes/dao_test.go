package fawkes

import (
	"context"
	"os"
	"testing"

	"go-gateway/app/app-svr/fawkes/service/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

// TestMain init ut main.
func TestMain(m *testing.M) {
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

// TestDaoPing ut ping.
func TestDaoPing(t *testing.T) {
	convey.Convey("Ping", t, func(ctx convey.C) {
		err := d.Ping(context.TODO())
		ctx.Convey("Err should be nil", func() {
			err = nil
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
