package assist

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

func init() {
	dir, _ := filepath.Abs("../../cmd/app-view-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func ctx() context.Context {
	return context.Background()
}

func TestAssist(t *testing.T) {
	Convey("get Assist all", t, func() {
		res, err := d.MultiAssist(ctx(), []int64{1})
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}
