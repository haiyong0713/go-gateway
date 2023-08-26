package space

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
	spaceclient "go-gateway/app/web-svr/space/interface/api/v1"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestSpacegarbInfo(t *testing.T) {
	var (
		ctx     = context.Background()
		garbMob = &space.Mob{}
		vmid    = int64(27515257)
		mid     = int64(27515257)
	)
	convey.Convey("garbInfo", t, func(c convey.C) {
		equip, _ := s.garbDao.SpaceBGEquip(ctx, vmid)
		// equip.Item.Images[equip.Index] = nil // 异常情况仍然出大会员图
		userAsset, _ := s.garbDao.SpaceBGUserAssetList(ctx, mid, 1, 1)
		s.garbInfo(ctx, garbMob, equip, userAsset, vmid, mid)
		str, _ := json.Marshal(garbMob)
		convey.Println(string(str))
		c.Convey("No return values", func(ctx convey.C) {
		})
	})
}

func TestSpace(t *testing.T) {
	var (
		ctx           = context.Background()
		mid           = int64(0)
		vmid          = int64(27515401)
		plat          = int8(2)
		build         = 9105
		pn            = 1
		ps            = 20
		teenagersMode = 0
		lessonsMode   = 0
		platform      = "ios"
		fromViewAid   = int64(1)
		device        = "pad"
		mobiApp       = "iphone"
		name          = ""
		buvid         = ""
		network       = ""
		adExtra       = ""
		spmid         = ""
		fromSpmid     = ""
		filtered      = ""
	)
	convey.Convey("Space", t, func(c convey.C) {
		res, err := s.Space(ctx, mid, vmid, plat, build, pn, ps, teenagersMode, lessonsMode, fromViewAid, platform, device, mobiApp, name, time.Now(), buvid, network, adExtra, spmid,
			fromSpmid, filtered, false)
		str, _ := json.Marshal(res)
		convey.Println(string(str))
		convey.Println(err)
		c.Convey("No return values", func(ctx convey.C) {
		})
	})
}

func TestUpArcs(t *testing.T) {
	var (
		ctx   = context.Background()
		mid   = int64(27515257)
		vmid  = int64(27515257)
		plat  = int8(2)
		build = 9105
		pn    = 1
		ps    = 20
	)
	convey.Convey("UpArcs", t, func(c convey.C) {
		res, _ := s.UpArcs(ctx, "", "", mid, vmid, pn, ps, build, plat, false,
			"", false, nil)
		str, _ := json.Marshal(res)
		convey.Println(string(str))
		c.Convey("No return values", func(ctx convey.C) {
		})
	})
}

func TestAdjustTab2(t *testing.T) {
	tab := []*space.TabItem{
		{Title: "主页", Param: space.HomeTab},
		{Title: "动态", Param: space.DyanmicTab},
		{Title: "投稿", Param: space.ContributeTab},
		{Title: "专栏", Param: space.ArticleTab},
	}
	activity1 := &spaceclient.UserTabReply{
		TabName:  "活动",
		TabOrder: 1,
	}
	tab1 := adjustTab2(tab, activity1)
	assert.Equal(t, tab1[0].Title, "活动")

	activity2 := &spaceclient.UserTabReply{
		TabName:  "活动",
		TabOrder: 3,
	}
	tab2 := adjustTab2(tab, activity2)
	assert.Equal(t, tab2[2].Title, "活动")

	activity3 := &spaceclient.UserTabReply{
		TabName:  "活动",
		TabOrder: 5,
	}
	tab3 := adjustTab2(tab, activity3)
	assert.Equal(t, tab3[4].Title, "活动")

	activity4 := &spaceclient.UserTabReply{
		TabName:  "活动",
		TabOrder: 10,
	}
	tab4 := adjustTab2(tab, activity4)
	assert.Equal(t, tab4[4].Title, "活动")
}
