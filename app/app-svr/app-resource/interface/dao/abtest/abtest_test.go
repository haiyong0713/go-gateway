package abtest

import (
	"context"
	"flag"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	dao *Dao
)

func ctx() context.Context {
	return context.Background()
}

// TestMain dao ut.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-resource")
		flag.Set("conf_token", "z8JNX5MFIyDxyBsqwQyF6pnjWQ5YOA14")
		flag.Set("tree_id", "2722")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "uat-config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-resource-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	dao = New(conf.Conf)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}

// TestExperimentLimit dao ut.
func TestExperimentLimit(t *testing.T) {
	Convey("get Experiment", t, func() {
		res, err := dao.ExperimentLimit(ctx())
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestExperimentByIDs dao ut.
func TestExperimentByIDs(t *testing.T) {
	Convey("get Experiment By IDs", t, func() {
		_, err := dao.ExperimentByIDs(ctx(), []int64{121})
		So(err, ShouldBeNil)
	})
}

// TestAbServer dao ut.
func TestAbServer(t *testing.T) {
	Convey("get AbServer", t, func() {
		res, err := dao.AbServer(ctx(), "E9B3F095-AA9E-4557-A1FB-758D85EE853C140082infoc", "phone", "iphone", "1", "iphone XS", "", "", 8400, 142231)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}
