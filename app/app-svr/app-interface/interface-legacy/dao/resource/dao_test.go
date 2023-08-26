package resource

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func init() {
	dir, _ := filepath.Abs("../../cmd/app-interface-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestResBanner(t *testing.T) {
	Convey("Banner", t, func() {
		res, err := d.Banner(ctx(), "", "", "", "", "", "", "", 1, 1, 1)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestEntrancesIsHidden(t *testing.T) {
	Convey("EntrancesIsHidden", t, func() {
		res, err := d.EntrancesIsHidden(ctx(), []int64{258, 243, 244}, 9999999, 1, "xiaomi")
		ress, _ := json.Marshal(res)
		fmt.Printf("%s", ress)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestMngIcon(t *testing.T) {
	Convey("MngIcon", t, func() {
		res, err := d.MngIcon(ctx(), []int64{245, 242, 201, 257}, 1, 1)
		ress, _ := json.Marshal(res)
		fmt.Printf("%s", ress)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestMineSections(t *testing.T) {
	Convey("MngIcon", t, func() {
		res, err := d.MineSections(ctx(), 1, 1, 99999, "", "hans")
		ress, _ := json.Marshal(res)
		fmt.Printf("%s", ress)
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}
