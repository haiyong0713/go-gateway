package show

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-job/job/conf"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	flag.Set("conf", "../../cmd/app-job-test.toml")
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	d.client.SetTransport(gock.DefaultTransport)
	m.Run()
	os.Exit(0)
}

func Test_BeginTran(t *testing.T) {
	Convey("BeginTran", t, func() {
		tx, err := d.BeginTran(context.TODO())
		So(err, ShouldBeNil)
		err = tx.Commit()
		So(err, ShouldBeNil)
	})
}

func Test_PTime(t *testing.T) {
	Convey("PTime", t, func() {
		_, err := d.PTime(context.TODO(), time.Now())
		So(err, ShouldBeNil)
	})
}

func Test_Pub(t *testing.T) {
	Convey("Pub", t, func() {
		tx, err := d.BeginTran(context.TODO())
		So(err, ShouldBeNil)
		err = d.Pub(tx, 127)
		So(err, ShouldBeNil)
		err = tx.Commit()
		So(err, ShouldBeNil)
	})
}

func Test_PingDB(t *testing.T) {
	Convey("PingDB", t, func() {
		err := d.PingDB(context.TODO())
		So(err, ShouldBeNil)
	})
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	d.client.SetTransport(gock.DefaultTransport)
	return r
}
