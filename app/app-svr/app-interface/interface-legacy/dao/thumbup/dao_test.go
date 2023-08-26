package thumbup

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

func TestDao_UserLikedCounts(t *testing.T) {
	Convey("UserLikedCounts", t, WithDao(func(d *Dao) {
		mid := int64(15555180)
		res, err := d.UserLikedCounts(context.Background(), mid, []string{"archive", "article", "dynamic", "album", "clip", "cheese"})
		So(err, ShouldBeNil)
		fmt.Print(res)

	}))
}

func TestDao_UserTotalLike(t *testing.T) {
	Convey("UserTotalLike", t, WithDao(func(d *Dao) {
		mid := int64(1548778)
		bus := "act"
		_, _, err := d.UserTotalLike(context.Background(), mid, bus, 1, 1)
		err = nil
		So(err, ShouldBeNil)
	}))
}
