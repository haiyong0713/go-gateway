package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/web/interface/model"
	"go-gateway/app/web-svr/web/interface/model/rcmd"
)

func ranking(c *bm.Context) {
	var (
		rid                    int64
		rankType, day, arcType int
		err                    error
	)
	params := c.Request.Form
	ridStr := params.Get("rid")
	rankTypeStr := params.Get("type")
	dayStr := params.Get("day")
	arcTypeStr := params.Get("arc_type")
	if rid, err = strconv.ParseInt(ridStr, 10, 64); err != nil || rid < 0 {
		rid = 0
	}
	if rankType, err = strconv.Atoi(rankTypeStr); err != nil {
		rankType = 1
	}
	if day, err = strconv.Atoi(dayStr); err != nil {
		day = 3
	}
	if arcType, err = strconv.Atoi(arcTypeStr); err != nil {
		arcType = 0
	}
	if err = checkType(rid, rankType, day, arcType); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(webSvc.Ranking(c, int16(rid), rankType, model.DayType[day], model.ArcType[arcType]))
}

func rankingV2(ctx *bm.Context) {
	v := new(struct {
		Type string `form:"type" default:"all"`
		Rid  int64  `form:"rid" validate:"min=0"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	typ, ok := model.RankV2Types[v.Type]
	if !ok {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	// 传了分区id时重置类型
	if v.Rid > 0 {
		typ = 0
	}
	ctx.JSON(webSvc.RankingV2(ctx, typ, v.Rid))
}

func checkType(rid int64, rankType, day, arcType int) (err error) {
	if _, ok := model.RankType[rankType]; !ok {
		err = ecode.RequestErr
		return
	}
	if _, ok := model.DayType[day]; !ok {
		err = ecode.RequestErr
		return
	}
	if _, ok := model.ArcType[arcType]; !ok {
		err = ecode.RequestErr
	}
	// bangumi and rookie not have recent contribution
	if (rid == 33 || rankType == 3) && arcType == 1 {
		err = ecode.RequestErr
	}
	return
}

func rankingIndex(c *bm.Context) {
	var (
		day int
		err error
	)
	params := c.Request.Form
	dayStr := params.Get("day")
	if day, err = strconv.Atoi(dayStr); err != nil {
		day = 3
	}
	if err = checkIndexDay(day); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(webSvc.RankingIndex(c, day))
}

func rankingRegion(c *bm.Context) {
	var (
		day, original int
		rid           int64
		err           error
	)
	params := c.Request.Form
	ridStr := params.Get("rid")
	dayStr := params.Get("day")
	originalStr := params.Get("original")
	if rid, err = strconv.ParseInt(ridStr, 10, 64); err != nil || rid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if day, err = strconv.Atoi(dayStr); err != nil {
		day = 3
	}
	if _, ok := model.RegionDayAll[day]; !ok {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if original, err = strconv.Atoi(originalStr); err != nil {
		original = 0
	} else if original != 1 && original != 0 {
		original = 0
	}
	c.JSON(webSvc.RankingRegion(c, rid, day, original))
}

func rankingRecommend(c *bm.Context) {
	var (
		rid int64
		err error
	)
	params := c.Request.Form
	ridStr := params.Get("rid")
	if rid, err = strconv.ParseInt(ridStr, 10, 64); err != nil || rid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(webSvc.RankingRecommend(c, rid))
}

func lpRankingRecommend(ctx *bm.Context) {
	v := struct {
		Business string `form:"business" validate:"required"`
	}{}
	if err := ctx.Bind(&v); err != nil {
		return
	}
	ctx.JSON(webSvc.LpRankingRecommend(ctx, v.Business))
}

func rankingTag(c *bm.Context) {
	var (
		rid, tagID int64
		err        error
	)
	params := c.Request.Form
	ridStr := params.Get("rid")
	tagIDStr := params.Get("tag_id")
	if rid, err = strconv.ParseInt(ridStr, 10, 64); err != nil || rid <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if tagID, err = strconv.ParseInt(tagIDStr, 10, 64); err != nil || tagID <= 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(webSvc.RankingTag(c, int16(rid), tagID))
}

func checkIndexDay(day int) (err error) {
	err = ecode.RequestErr
	for _, dayItem := range model.IndexDayType {
		if day == dayItem {
			err = nil
			return
		}
	}
	return
}

func webTop(c *bm.Context) {
	c.JSON(webSvc.WebTop(c))
}

func webTopRcmd(ctx *bm.Context) {
	param := struct {
		FreshType  int   `form:"fresh_type"`
		Version    int   `form:"version" validate:"min=0,max=1"`
		Ps         int   `form:"ps" validate:"min=0,max=14"`
		FreshIdx   int64 `form:"fresh_idx"`
		FreshIdx1h int64 `form:"fresh_idx_1h"`
	}{}
	if err := ctx.Bind(&param); err != nil {
		return
	}
	var buvid string
	if ck, err := ctx.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	var mid int64
	if midInter, ok := ctx.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	ctx.JSON(webSvc.WebTopRcmd(ctx, buvid, mid, param.FreshType, param.Version, param.Ps, ip, param.FreshIdx, param.FreshIdx1h))
}

// func webTopFeedRcmd(ctx *bm.Context) {
// 	param := struct {
// 		FreshType   int64  `form:"fresh_type"`
// 		Ps          int64  `form:"ps" validate:"min=0,max=30"`
// 		FreshIdx    int64  `form:"fresh_idx"`
// 		FreshIdx1h  int64  `form:"fresh_idx_1h"`
// 		FeedVersion string `form:"feed_version" default:"V0"`
// 		YNum        int64  `form:"y_num"`
// 	}{}
// 	if err := ctx.Bind(&param); err != nil {
// 		return
// 	}
// 	if param.Ps < 1 {
// 		param.Ps = 30
// 	}
// 	var buvid string
// 	if ck, err := ctx.Request.Cookie("buvid3"); err == nil {
// 		buvid = ck.Value
// 	}
// 	var mid int64
// 	if midInter, ok := ctx.Get("mid"); ok {
// 		mid = midInter.(int64)
// 	}
// 	ip := metadata.String(ctx, metadata.RemoteIP)
// 	ctx.JSON(webSvc.WebTopFeedRcmd(ctx, mid, param.FreshType, param.Ps, param.FreshIdx, param.FreshIdx1h, param.YNum, param.FeedVersion, ip, buvid))
// }

func webTopFeedRcmdV2(ctx *bm.Context) {
	param := &rcmd.TopRcmdReq{}
	if err := ctx.Bind(param); err != nil {
		return
	}
	if param.Ps < 1 {
		param.Ps = 30
	}
	if ck, err := ctx.Request.Cookie("buvid3"); err == nil {
		param.Buvid = ck.Value
	}
	if midInter, ok := ctx.Get("mid"); ok {
		param.Mid = midInter.(int64)
	}
	if sidCookie, err := ctx.Request.Cookie("sid"); err == nil {
		param.Sid = sidCookie.Value
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	param.Ip = ip
	param.IsFeed = 1
	param.UserAgent = ctx.Request.UserAgent()
	ctx.JSON(webSvc.WebTopFeedRcmdV2(ctx, param))
}
