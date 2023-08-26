package playurl

import (
	"context"
	"fmt"
	"testing"

	hlsgrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurlhls"

	"github.com/smartystreets/goconvey/convey"
)

func TestHlsScheduler(t *testing.T) {
	var (
		c      = context.Background()
		params = &hlsgrpc.M3U8RequestMsg{
			Platform:  "iphone",
			Cid:       10226772,
			Qn:        16,
			IsSp:      true,
			Mid:       27515255,
			BackupNum: 2,
			Business:  hlsgrpc.Business_UGC,
			ForceHost: 1,
			Type:      hlsgrpc.RequstType_PIP,
			Fnval:     0,
			Fnver:     16,
		}
	)
	convey.Convey("Playurl", t, func(ctx convey.C) {
		p, err := d.HlsScheduler(c, params)
		fmt.Printf("%+v", p)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})

	})
}

func TestMasterScheduler(t *testing.T) {
	var (
		c      = context.Background()
		params = &hlsgrpc.M3U8RequestMsg{
			Platform:  "iphone",
			Cid:       10226772,
			Qn:        16,
			IsSp:      true,
			Mid:       27515255,
			BackupNum: 2,
			Business:  hlsgrpc.Business_UGC,
			ForceHost: 1,
			Type:      hlsgrpc.RequstType_PIP,
			Fnval:     0,
			Fnver:     16,
		}
	)
	convey.Convey("MasterScheduler", t, func(ctx convey.C) {
		p, err := d.MasterScheduler(c, params, nil)
		fmt.Printf("%+v", p)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})

	})
}

func TestM3U8Scheduler(t *testing.T) {
	var (
		c      = context.Background()
		params = &hlsgrpc.M3U8RequestMsg{
			Platform:  "iphone",
			Cid:       10226772,
			Qn:        888, //30216
			IsSp:      true,
			Mid:       27515255,
			BackupNum: 2,
			Business:  hlsgrpc.Business_UGC,
			ForceHost: 1,
			Type:      hlsgrpc.RequstType_PIP,
			Fnval:     0,
			Fnver:     16,
		}
	)
	convey.Convey("Playurl", t, func(ctx convey.C) {
		p, err := d.M3U8Scheduler(c, params)
		fmt.Printf("%+v", p)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})

	})
}
