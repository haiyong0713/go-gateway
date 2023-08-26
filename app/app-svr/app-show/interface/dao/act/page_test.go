package act

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	natpagemdl "go-gateway/app/web-svr/native-page/interface/api"

	"git.bilibili.co/bapis/bapis-go/activity/service"
	"github.com/smartystreets/goconvey/convey"
)

func TestActNatConfig(t *testing.T) {
	convey.Convey("NatConfig", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			arg = &natpagemdl.NatConfigReq{Pid: 102}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1, err := d.NatConfig(c, arg)
			convCtx.Convey("Then err should be nil.p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestActActLikes(t *testing.T) {
	convey.Convey("ActLikes", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			arg = &api.ActLikesReq{Sid: 10507, Mid: 155809, SortType: 1, Ps: 1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			reply, err := d.ActLikes(c, arg)
			convCtx.Convey("Then err should be nil.reply should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(reply)
			})
		})
	})
}

func TestActActLiked(t *testing.T) {
	convey.Convey("ActLiked", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			arg = &api.ActLikedReq{Sid: 10436, Mid: 155809, Lid: 1, Score: 1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.ActLiked(c, arg)
			convCtx.Convey("Then err should be nil.p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestActModuleConfig(t *testing.T) {
	convey.Convey("ModuleConfig", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			arg = &natpagemdl.ModuleConfigReq{ModuleID: 5218}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.ModuleConfig(c, arg)
			convCtx.Convey("Then err should be nil.req should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// ModuleMixExt
func TestModuleMixExt(t *testing.T) {
	convey.Convey("ModuleMixExt", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			arg = &natpagemdl.ModuleMixExtReq{ModuleID: 5218, Ps: 1, Offset: 0, MType: 1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.ModuleMixExt(c, arg)
			convCtx.Convey("Then err should be nil.req should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestModuleMixExts(t *testing.T) {
	convey.Convey("ModuleMixExts", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			arg = &natpagemdl.ModuleMixExtsReq{ModuleID: 15295, Ps: 1, Offset: 0}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.ModuleMixExts(c, arg)
			convCtx.Convey("Then err should be nil.req should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				fmt.Printf("%v", res)
			})
		})
	})
}

// ReserveFollowings
func TestReserveFollowings(t *testing.T) {
	convey.Convey("ReserveFollowings", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.ReserveFollowings(c, 15555181, []int64{10635, 10636, 10637})
			convCtx.Convey("Then err should be nil.req should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				str, _ := json.Marshal(res)
				fmt.Printf("%s", str)
			})
		})
	})
}

func TestAddReserve(t *testing.T) {
	convey.Convey("AddReserve", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddReserve(c, 10636, 15555181)
			convCtx.Convey("Then err should be nil.req should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// DelReserve
func TestDelReserve(t *testing.T) {
	convey.Convey("DelReserve", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelReserve(c, 10636, 15555181)
			convCtx.Convey("Then err should be nil.req should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestNativePages(t *testing.T) {
	convey.Convey("NativePages", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			pids = []int64{6091, 6083, 6082, 6079, 6061, 6060}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1, err := d.NativePages(c, pids)
			convCtx.Convey("Then err should be nil.p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(p1, convey.ShouldNotBeNil)
				fmt.Printf("%v", p1)
			})
		})
	})
}
