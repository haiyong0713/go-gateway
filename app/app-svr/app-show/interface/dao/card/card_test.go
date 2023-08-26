package card

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-show/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-show")
		flag.Set("conf_token", "Pae4IDOeht4cHXCdOkay7sKeQwHxKOLA")
		flag.Set("tree_id", "2687")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-show-test.toml")

	}
	flag.Parse()
	cfg, err := confInit()
	if err != nil {
		panic(err)
	}
	d = New(cfg)
	os.Exit(m.Run())
}

func confInit() (*conf.Config, error) {
	err := paladin.Init()
	if err != nil {
		return nil, err
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("app-show.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func TestColumns(t *testing.T) {
	Convey("Columns", t, func() {
		res, err := d.Columns(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestPosRecs(t *testing.T) {
	Convey("PosRecs", t, func() {
		res, err := d.PosRecs(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestRecContents(t *testing.T) {
	Convey("RecContents", t, func() {
		res, _, err := d.RecContents(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestNperContents(t *testing.T) {
	Convey("NperContents", t, func() {
		res, _, err := d.NperContents(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestColumnNpers(t *testing.T) {
	Convey("ColumnNpers", t, func() {
		res, err := d.ColumnNpers(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestColumnList(t *testing.T) {
	Convey("ColumnList", t, func() {
		res, err := d.ColumnList(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestCard(t *testing.T) {
	Convey("Card", t, func() {
		res, err := d.Card(ctx(), time.Now())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestCardPlat(t *testing.T) {
	Convey("CardPlat", t, func() {
		res, err := d.CardPlat(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestCardSet(t *testing.T) {
	Convey("CardSet", t, func() {
		res, err := d.CardSet(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestEventTopic(t *testing.T) {
	Convey("EventTopic", t, func() {
		res, err := d.EventTopic(ctx())
		So(res, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestSeries(t *testing.T) {
	Convey("Series", t, func() {
		res, err := d.Series(ctx(), "weekly_selected")
		Convey("Then err should be nil.res should not be nil.", func() {
			So(res, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}

func TestSerieConfig(t *testing.T) {
	Convey("SerieConfig", t, func() {
		res, err := d.SerieConfig(ctx(), "weekly_selected", 1)
		Convey("Then err should be nil.res should not be nil.", func() {
			So(res, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}

func TestSelectedRes(t *testing.T) {
	Convey("SelectedRes", t, func() {
		res, err := d.SelectedRes(ctx(), 1)
		Convey("Then err should be nil.res should not be nil.", func() {
			So(res, ShouldNotBeEmpty)
			So(err, ShouldBeNil)
		})
	})
}
