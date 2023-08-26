package archive

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestArchiveArc(t *testing.T) {
	var (
		c   = context.Background()
		aid = int64(10112118)
	)
	convey.Convey("Arc", t, func(ctx convey.C) {
		a, err := d.Arc(c, aid)
		ctx.Convey("Then err should be nil.a should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(a, convey.ShouldNotBeNil)
		})
	})
}

func TestArchiveArcs(t *testing.T) {
	var (
		c    = context.Background()
		aids = []int64{10112118}
	)
	convey.Convey("Arcs", t, func(ctx convey.C) {
		res, err := d.Arcs(c, aids)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestArchiveArcsWithPage(t *testing.T) {
	var (
		c    = context.Background()
		aids = []int64{10112118, 10112119, 10112129, 10112120, 10112121}
	)
	convey.Convey("Arcs", t, func(ctx convey.C) {
		res, err := d.ArcsWithPage(c, aids)
		fmt.Println(res)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestArchiveFlowJudge(t *testing.T) {
	var (
		c        = context.Background()
		aids     = []int64{10112102}
		flowCtrl = &conf.FlowCtrl{
			Secret:    "a25eef460c047075ffa0f6713c53a4bac79442eb",
			OidLength: 30,
			Source:    "hot_weekly_selected",
		}
	)
	convey.Convey("FlowJudge", t, func(ctx convey.C) {
		noHotAids, hotDownAids, err := d.FlowJudge(c, aids, flowCtrl)
		ctx.Convey("Then err should be nil.forbidAids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(noHotAids), convey.ShouldBeLessThan, len(aids))
			ctx.So(len(hotDownAids), convey.ShouldBeLessThan, len(aids))
		})
	})
}

func TestDao_ArchiveSearchBan(t *testing.T) {
	var (
		c         = context.Background()
		aid int64 = 640001692
	)
	convey.Convey("ArchiveSearchBan", t, func(ctx convey.C) {
		err := d.ArchiveSearchBan(c, aid)
		ctx.Convey("Then err should be nil.ArchiveSearchBan should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_ArchiveSearchBan2(t *testing.T) {
	var (
		c = context.Background()
		// xiongjianjun：640001692稿件比较老，grpc接口在uat环境可能会出现未禁搜的情况，预发和线上不会出现
		// uat使用以下id：
		// 960041359 480017502 760065168
		aid int64 = 760065168
	)
	convey.Convey("ArchiveSearchBan2", t, func(ctx convey.C) {
		err := d.ArchiveSearchBan2(c, aid)
		ctx.Convey("Then err should be nil.ArchiveSearchBan should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestDao_ArchiveAudit(t *testing.T) {
	var (
		c         = context.Background()
		aid int64 = 10100680
	)
	convey.Convey("ArchiveAudit", t, func(ctx convey.C) {
		res, err := d.ArchiveAudit(c, aid)
		ctx.Convey("Then err should be nil.ArchiveSearchBan should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	})
}
