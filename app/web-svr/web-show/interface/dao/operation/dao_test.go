package operation

import (
	"context"
	"flag"
	"path/filepath"
	"testing"

	"go-gateway/app/web-svr/web-show/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var d *Dao

func WithDao(f func(d *Dao)) func() {
	return func() {
		dir, _ := filepath.Abs("../cmd/web-show-test.toml")
		flag.Set("conf", dir)
		conf.Init()
		if d == nil {
			d = New(conf.Conf)
		}
		f(d)
	}
}

func TestDao_Operation(t *testing.T) {
	Convey("test operation", t, WithDao(func(d *Dao) {
		data, err := d.Operation(context.TODO())
		So(err, ShouldBeNil)
		Printf("%+v", data)
	}))
}
