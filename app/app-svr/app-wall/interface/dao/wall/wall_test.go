package wall

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-wall/interface/conf"

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

func TestWallAll(t *testing.T) {
	Convey("WallAll", t, func() {
		res, err := d.WallAll(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
