package live

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-wall/job/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func init() {
	dir, _ := filepath.Abs("../../cmd/app-wall-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestPack(t *testing.T) {
	Convey("Pack", t, func() {
		err := d.Pack(ctx(), 1, 1)
		So(err, ShouldBeNil)
	})
}

func TestAddVip(t *testing.T) {
	Convey("AddVip", t, func() {
		_, err := d.AddVip(ctx(), 1, 1)
		So(err, ShouldBeNil)
	})
}
