package sidebar

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var d *Dao

func init() {
	dir, _ := filepath.Abs("../cmd/app-interface-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
}

func WithDao(f func(d *Dao)) func() {
	return func() {
		Reset(func() {})
		f(d)
	}
}

func TestDao_Sidebar(t *testing.T) {
	Convey("ArchiveInfo", t, WithDao(func(d *Dao) {
		res, err := d.Sidebars(context.TODO())
		fmt.Printf("%+v", res)
		So(err, ShouldBeNil)
	}))
}
