package http

import (
	bm "go-common/library/net/http/blademaster"
	xtime "go-common/library/time"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/service"
)

func upLaunchCheck(c *bm.Context) {
	midInter, _ := c.Get("mid")
	loginMid := midInter.(int64)
	c.JSON(service.LikeSvc.UpLaunchCheck(c, loginMid))
}

func upLaunch(c *bm.Context) {
	v := new(struct {
		Title     string     `form:"title" validate:"min=1,max=20"`
		Stime     xtime.Time `form:"stime" validate:"min=1"`
		Etime     xtime.Time `form:"etime" validate:"min=1"`
		Statement string     `form:"statement" validate:"min=1,max=200"`
		Aid       int64      `form:"aid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	loginMid := midInter.(int64)
	c.JSON(nil, service.LikeSvc.UpLaunch(c, v.Title, v.Statement, v.Stime, v.Etime, v.Aid, loginMid))
}

func upCheck(c *bm.Context) {
	v := new(struct {
		Uid int64 `form:"uid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	res, err := service.LikeSvc.UpCheck(c, v.Uid)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(map[string]interface{}{"status": res}, nil)
}

func canCreateUpReserve(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	res, err := service.LikeSvc.CanUpCreateActReserve(c, &pb.CanUpCreateActReserveReq{Mid: v.Mid})
	if err != nil {
		c.JSON(nil, err)
		return
	}
	var state int
	if _, ok := res.List[int64(pb.UpActReserveRelationType_Archive)]; ok {
		state = int(pb.CanUpCreateActReservePermissionType_Allow)
	}
	c.JSON(struct {
		State int
	}{state}, nil)
}

func upArchiveList(c *bm.Context) {
	midInter, _ := c.Get("mid")
	loginMid := midInter.(int64)
	c.JSON(service.LikeSvc.UpArchiveList(c, loginMid))
}

func upActPage(c *bm.Context) {
	var loginMid int64
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.UpActPage(c, loginMid, v.Sid))
}

func upDo(c *bm.Context) {
	v := new(struct {
		Sid            int64   `form:"sid" validate:"min=1"`
		TotalTime      int64   `form:"total_time"`
		MatchedPercent float32 `form:"matched_percent"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midInter, _ := c.Get("mid")
	loginMid := midInter.(int64)
	c.JSON(service.LikeSvc.UpActDo(c, v.Sid, loginMid, v.TotalTime, v.MatchedPercent))
}

func upActRank(c *bm.Context) {
	var loginMid int64
	v := new(struct {
		Sid int64 `form:"sid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midInter, ok := c.Get("mid"); ok {
		loginMid = midInter.(int64)
	}
	c.JSON(service.LikeSvc.UpActRank(c, v.Sid, loginMid))
}
