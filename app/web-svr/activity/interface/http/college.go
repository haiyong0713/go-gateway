package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/ecode"
)

// collegeBind 绑定学校
func collegeBind(c *bm.Context) {
	v := new(struct {
		CollegeID int64 `form:"college_id" validate:"required,min=1"`
		Year      int   `form:"year"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, ecode.ActivityOverEnd)
	// midStr, _ := c.Get("mid")
	// mid := midStr.(int64)
	// params := headers(c)
	// c.JSON(collegeSvc.Bind(c, mid, params.Buvid, v.CollegeID, v.Year))

}

// collegeProvinceRank 省排行
func collegeProvinceRank(c *bm.Context) {
	v := new(struct {
		ProvinceID int64 `form:"province_id" validate:"required,min=1"`
		Pn         int   `form:"pn" validate:"min=1" default:"1"`
		Ps         int   `form:"ps" validate:"min=1" default:"10"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, ecode.ActivityOverEnd)
	// c.JSON(collegeSvc.ProvinceRank(c, v.ProvinceID, v.Ps, v.Pn))

}

// collegeNationwideRank 全国排行
func collegeNationwideRank(c *bm.Context) {
	v := new(struct {
		Pn int `form:"pn" validate:"min=1" default:"1"`
		Ps int `form:"ps" validate:"min=1" default:"10"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, ecode.ActivityOverEnd)
	// c.JSON(collegeSvc.NationwideRank(c, v.Ps, v.Pn))

}

// collegeList 全量学校列表
func collegeList(c *bm.Context) {
	v := new(struct {
		Key string `form:"key" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, ecode.ActivityOverEnd)
	// c.JSON(collegeSvc.GetAllCollege(c, v.Key))
}

// collegePeopleRank 获取校内用户排行
func collegePeopleRank(c *bm.Context) {
	v := new(struct {
		CollegeID int64 `form:"college_id" validate:"min=1"`
		Pn        int   `form:"pn" validate:"min=1" default:"1"`
		Ps        int   `form:"ps" validate:"min=1" default:"10"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, ecode.ActivityOverEnd)

	// c.JSON(collegeSvc.CollegePeopleRank(c, v.CollegeID, v.Ps, v.Pn))
}

// collegePersonal 用户信息
func collegePersonal(c *bm.Context) {
	c.JSON(nil, ecode.ActivityOverEnd)
}

// collegeDetail 学校详情
func collegeDetail(c *bm.Context) {
	v := new(struct {
		CollegeID int64 `form:"college_id" validate:"required,min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, ecode.ActivityOverEnd)
	// c.JSON(collegeSvc.Detail(c, v.CollegeID))
}

// collegeTabArchive tab下的稿件信息
func collegeTabArchive(c *bm.Context) {
	v := new(struct {
		TabType   int   `form:"tab_type" validate:"required,min=1" default:"1"`
		CollegeID int64 `form:"college_id" validate:"required,min=1"`
		Pn        int   `form:"pn" validate:"min=1" default:"1"`
		Ps        int   `form:"ps" validate:"min=1" default:"10"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, ecode.ActivityOverEnd)
	// midStr, ok := c.Get("mid")
	// var mid int64
	// if ok {
	// 	mid = midStr.(int64)
	// }
	// c.JSON(collegeSvc.ArchiveList(c, mid, v.CollegeID, v.TabType, v.Ps, v.Pn))
}

// collegeArchiveIsActivity tab下的稿件信息
func collegeArchiveIsActivity(c *bm.Context) {
	v := new(struct {
		Aid int64 `form:"aid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, ecode.ActivityOverEnd)
	// midStr, ok := c.Get("mid")
	// var mid int64
	// if ok {
	// 	mid = midStr.(int64)
	// }
	// c.JSON(collegeSvc.AidIsCollege(c, mid, v.Aid))
}

func collegeUploadVersion(c *bm.Context) {
	c.JSON(nil, ecode.ActivityOverEnd)

	// c.JSON(collegeSvc.UpdateVersion())
}

func collegeTask(c *bm.Context) {
	c.JSON(nil, ecode.ActivityOverEnd)
	// 	midStr, ok := c.Get("mid")
	// 	var mid int64
	// 	if ok {
	// 		mid = midStr.(int64)
	// 	}
	// 	c.JSON(collegeSvc.Task(c, mid))
}

func collegeInviter(c *bm.Context) {
	c.JSON(nil, ecode.ActivityOverEnd)
	// midStr, ok := c.Get("mid")
	// var mid int64
	// if ok {
	// 	mid = midStr.(int64)
	// }
	// c.JSON(collegeSvc.InviterCollege(c, mid))

}

func collegeFollow(c *bm.Context) {
	c.JSON(nil, ecode.ActivityOverEnd)
	// midStr, ok := c.Get("mid")
	// var mid int64
	// if ok {
	// 	mid = midStr.(int64)
	// }
	// c.JSON(collegeSvc.Follow(c, mid))

}
