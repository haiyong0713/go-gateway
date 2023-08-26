package relation

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	"github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-interface")
		flag.Set("conf_token", "1mWvdEwZHmCYGoXJCVIdszBOPVdtpXb3")
		flag.Set("tree_id", "2688")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-interface-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}

func TestStat(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(27515256)
	)
	convey.Convey("Stat", t, func(ctx convey.C) {
		rly, err := d.Stat(c, mid)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			fmt.Printf("%v", rly)
		})
	})
}

func TestFollowersUnread(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(27515257)
	)
	convey.Convey("FollowersUnread", t, func(ctx convey.C) {
		rly, err := d.FollowersUnread(c, mid)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			fmt.Printf("%v", rly)
		})
	})
}

func TestFollowings(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(27515257)
	)
	convey.Convey("Followings", t, func(ctx convey.C) {
		rly, err := d.Followings(c, mid)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			fmt.Printf("%v", rly)
		})
	})
}

func TestRelations(t *testing.T) {
	var (
		c    = context.Background()
		mid  = int64(27515257)
		fids = []int64{111005049}
	)
	convey.Convey("Relations", t, func(ctx convey.C) {
		rly, err := d.Relations(c, mid, fids)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			fmt.Printf("%v", rly)
		})
	})
}

func TestTag(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(27515257)
		tid = int64(-10)
	)
	convey.Convey("Tag", t, func(ctx convey.C) {
		rly, err := d.Tag(c, mid, tid)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			fmt.Printf("%v", rly)
		})
	})
}

func TestFollowersUnreadCount(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(27515257)
	)
	convey.Convey("FollowersUnreadCount", t, func(ctx convey.C) {
		rly, err := d.FollowersUnreadCount(c, mid)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			fmt.Printf("%v", rly)
		})
	})
}

func TestDao_StatsGRPC(t *testing.T) {
	var (
		c    = context.Background()
		mids = []int64{1, 2, 3}
	)
	convey.Convey("StatsGRPC Test", t, func(ctx convey.C) {
		_, err := d.StatsGRPC(c, mids)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_Relation(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("StatsGRPC Test", t, func(ctx convey.C) {
		_, err := d.Relation(c, 15555180, 27515414)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_SpecialEffect(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("StatsGRPC Test", t, func(ctx convey.C) {
		_, err := d.SpecialEffect(c, 15555180, 27515414, "")
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_Interrelations(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("Interrelations", t, func(ctx convey.C) {
		res, err := d.Interrelations(c, 111005889, []int64{111005050})
		ss, _ := json.Marshal(res)
		fmt.Printf("%s", ss)
		ctx.Convey("Then err should be nil", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
