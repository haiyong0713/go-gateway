package garb

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGarbSpaceBG(t *testing.T) {
	var (
		c      = context.Background()
		garbID = int64(1648)
	)
	convey.Convey("SpaceBG", t, func(ctx convey.C) {
		result, err := d.SpaceBG(c, garbID)
		fmt.Println(result, err)
		ctx.Convey("Then err should be nil.result should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldNotBeNil)
		})
	})
}

func TestGarbUserFanInfo(t *testing.T) {
	var (
		c      = context.Background()
		mid    = int64(27515403)
		garbID = int64(211)
	)
	convey.Convey("UserFanInfo", t, func(ctx convey.C) {
		number, err := d.UserFanInfo(c, mid, garbID)
		fmt.Println("fffff", number, err)
		ctx.Convey("Then err should be nil.number should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(number, convey.ShouldNotBeNil)
		})
	})
}

func TestGarbUserFanInfos(t *testing.T) {
	var (
		c           = context.Background()
		mid         = int64(27515403)
		suitItemIDs = []int64{211, 1135, 248}
	)
	convey.Convey("UserFanInfos", t, func(ctx convey.C) {
		result, err := d.UserFanInfos(c, mid, suitItemIDs)
		str, _ := json.Marshal(result)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.result should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldNotBeNil)
		})
	})
}

func TestGarbSpaceBGEquip(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(27515257)
		// mid = int64(110000259)
	)
	convey.Convey("SpaceBGEquip", t, func(ctx convey.C) {
		reply, err := d.SpaceBGEquip(c, mid)
		str, _ := json.Marshal(reply)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.reply should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(reply, convey.ShouldNotBeNil)
		})
	})
}

func TestGarbSpaceBGUserAssetList(t *testing.T) {
	var (
		c   = context.Background()
		mid = int64(27515403)
		pn  = int64(1)
		ps  = int64(20)
	)
	convey.Convey("SpaceBGUserAssetList", t, func(ctx convey.C) {
		res, err := d.SpaceBGUserAssetList(c, mid, pn, ps)
		str, _ := json.Marshal(res)
		fmt.Println(string(str))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestGarbLoad(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("SpaceBGUserAssetList", t, func(ctx convey.C) {
		err := d.SpaceBGLoad(c, 27515257, 1649, 2)
		fmt.Println(err)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUserAsset(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("UserAsset", t, func(ctx convey.C) {
		_, err := d.UserAsset(c, 1810, 27515255)
		fmt.Println(err)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
