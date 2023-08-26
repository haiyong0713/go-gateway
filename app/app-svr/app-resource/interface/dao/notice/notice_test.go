package notice

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-resource/interface/component"
	"go-gateway/app/app-svr/app-resource/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

// TestMain dao ut.
func TestMain(m *testing.M) {
	flag.Set("conf", "../../cmd/app-resource-test.toml")
	if err := conf.Init(); err != nil {
		panic(err)
	}
	if err := component.InitByCfg(conf.Conf.MySQL.Show); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	m.Run()
	os.Exit(0)
}

func TestAll(t *testing.T) {
	Convey("get All all", t, func() {
		res, err := d.All(ctx(), time.Now())
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestDao_PackagePushList(t *testing.T) {
	list, err := d.PackagePushList(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, list)
	for _, v := range list {
		assert.NotNil(t, v)
		t.Log(v)
	}
}
