package aids

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-admin/admin/conf"
	"go-gateway/app/app-svr/app-admin/admin/model/aids"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func init() {
	dir, _ := filepath.Abs("../../cmd/app-admin-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestInsert(t *testing.T) {
	Convey("insert aids", t, func() {
		a := &aids.Param{
			Aid: 654,
		}
		err := d.Insert(ctx(), a)
		So(err, ShouldBeNil)
	})
}
