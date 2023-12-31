package location

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-view/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func init() {
	dir, _ := filepath.Abs("../../cmd/app-view-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestAuthPIDs(t *testing.T) {
	Convey("get AuthPIDs", t, func() {
		_, err := d.AuthPIDs(ctx(), "417,1521", "127.0.0.0")
		So(err, ShouldBeNil)
	})
}

func TestArchive(t *testing.T) {
	Convey("get Archive", t, func() {
		_, err := d.Archive(ctx(), 16816128, 16816128, "127.0.0.0", "127.0.0.0")
		So(err, ShouldBeNil)
	})
}
