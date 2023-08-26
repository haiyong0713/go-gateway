package http

import (
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/pkg/idsafe/bvid"
)

func filterMatch(c *bm.Context) {
	p := new(model.ParamFilter)
	if err := c.Bind(p); err != nil {
		return
	}
	if p.Stime != "" {
		if _, err := time.Parse("2006-01-02", p.Stime); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	c.JSON(eSvc.FilterMatch(c, p))
}

func listContest(c *bm.Context) {
	var (
		mid   int64
		err   error
		total int
		list  []*model.Contest
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	p := new(model.ParamContest)
	if err = c.Bind(p); err != nil {
		return
	}
	if p.Stime != "" {
		if _, err = time.Parse("2006-01-02", p.Stime); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if p.Etime != "" {
		if _, err = time.Parse("2006-01-02", p.Etime); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if list, total, err = eSvc.ListContest(c, mid, p); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func appContest(c *bm.Context) {
	var (
		mid   int64
		err   error
		total int
		list  []*model.Contest
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	p := new(model.ParamContest)
	if err = c.Bind(p); err != nil {
		return
	}
	if p.Stime != "" {
		if _, err = time.Parse("2006-01-02", p.Stime); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
		p.Etime = p.Stime
	}
	if list, total, err = eSvc.ListContest(c, mid, p); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func calendar(c *bm.Context) {
	var err error
	p := new(model.ParamFilter)
	if err = c.Bind(p); err != nil {
		return
	}
	if _, err = time.Parse("2006-01-02", p.Stime); err != nil {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if _, err = time.Parse("2006-01-02", p.Etime); err != nil {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(eSvc.Calendar(c, p))
}

func filterVideo(c *bm.Context) {
	p := new(model.ParamFilter)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.FilterVideo(c, p))
}

func listVideo(c *bm.Context) {
	var (
		err       error
		total     int
		list      []*arcmdl.Arc
		videoList []*model.Video
	)
	p := new(model.ParamVideo)
	if err = c.Bind(p); err != nil {
		return
	}
	if list, total, err = eSvc.ListVideo(c, p); err != nil {
		c.JSON(nil, err)
		return
	}
	for _, v := range list {
		if v == nil {
			continue
		}
		video := &model.Video{Arc: v}
		if video.Bvid, err = bvid.AvToBv(v.Aid); err != nil {
			log.Error("listVideo AvToBv(%v)error (%v)", v.Aid, err)
			err = nil
			continue
		}
		videoList = append(videoList, video)
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = videoList
	c.JSON(data, nil)
}

func actVideos(c *bm.Context) {
	var (
		err error
	)
	param := new(struct {
		MmID int64 `form:"mm_id"  validate:"gt=0"`
	})
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(eSvc.ActModules(c, param.MmID))
}

func active(c *bm.Context) {
	var (
		err error
	)
	param := new(struct {
		Aid int64 `form:"aid"  validate:"gt=0"`
		Tp  int64 `form:"tp"`
	})
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(eSvc.ActPage(c, param.Aid, param.Tp))
}

func actPoints(c *bm.Context) {
	var (
		mid   int64
		err   error
		total int
		list  []*model.Contest
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	p := new(model.ParamActPoint)
	if err = c.Bind(p); err != nil {
		return
	}
	if list, total, err = eSvc.ActPoints(c, mid, p); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func actKnockout(c *bm.Context) {
	var (
		err error
	)
	param := new(struct {
		MdID int64 `form:"md_id"  validate:"gt=0"`
	})
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(eSvc.ActKnockout(c, param.MdID))
}

func actTop(c *bm.Context) {
	var (
		mid   int64
		err   error
		total int
		list  []*model.Contest
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	p := new(model.ParamActTop)
	if err = c.Bind(p); err != nil {
		return
	}
	if p.Stime != "" {
		if _, err = time.Parse("2006-01-02", p.Stime); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if p.Etime != "" {
		if _, err = time.Parse("2006-01-02", p.Etime); err != nil {
			c.JSON(nil, xecode.RequestErr)
			return
		}
	}
	if list, total, err = eSvc.ActTop(c, mid, p); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func search(c *bm.Context) {
	var (
		mid   int64
		buvid string
		err   error
	)
	p := new(model.ParamSearch)
	if err = c.Bind(p); err != nil {
		return
	}
	if ck, err := c.Request.Cookie("buvid3"); err == nil {
		buvid = ck.Value
	}
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(eSvc.Search(c, mid, p, buvid))
}
func season(c *bm.Context) {
	var (
		err   error
		total int
		list  []*model.Season
	)
	p := new(model.ParamSeason)
	if err = c.Bind(p); err != nil {
		return
	}
	if list, total, err = eSvc.Season(c, p); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func appSeason(c *bm.Context) {
	var (
		err   error
		total int
		list  []*model.Season
	)
	p := new(model.ParamSeason)
	if err = c.Bind(p); err != nil {
		return
	}
	if list, total, err = eSvc.AppSeason(c, p); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func gameRank(c *bm.Context) {
	c.JSON(eSvc.GameRank(c))
}

func gameSeason(c *bm.Context) {
	param := new(struct {
		Gid int64 `form:"gid"  validate:"min=0"`
	})
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(eSvc.GameSeason(c, param.Gid))
}

func contest(c *bm.Context) {
	var (
		mid int64
		err error
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	param := new(struct {
		Cid      int64 `form:"cid"  validate:"gt=0"`
		Platform int64 `form:"platform"`
	})
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(eSvc.Contest(c, mid, param.Cid, param.Platform))
}

func contestV2(c *bm.Context) {
	var (
		mid int64
		err error
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	param := new(struct {
		Cid      int64 `form:"cid"  validate:"gt=0"`
		Platform int64 `form:"platform"`
	})
	if err = c.Bind(param); err != nil {
		return
	}
	c.JSON(eSvc.ContestWithMatchRecord(c, mid, param.Cid, param.Platform))
}

func recent(c *bm.Context) {
	var (
		mid  int64
		err  error
		list []*model.Contest
	)
	if midStr, ok := c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	param := &model.ParamCDRecent{}
	if err = c.Bind(param); err != nil {
		return
	}
	if list, err = eSvc.Recent(c, mid, param); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(list, nil)
}

func intervene(c *bm.Context) {
	var (
		err   error
		total int64
		list  []*model.SearchRes
	)
	p := new(struct {
		Ps int64 `form:"ps" default:"10" validate:"min=1"`
		Pn int64 `form:"pn" default:"1" validate:"min=1"`
	})
	if err = c.Bind(p); err != nil {
		return
	}
	if list, total, err = eSvc.Intervene(c, p.Pn, p.Ps); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   p.Pn,
		"size":  p.Ps,
		"total": total,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func s10Tab(c *bm.Context) {
	var mid int64

	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}

	p := new(struct {
		AvCID int64 `form:"av_cid" default:"0"`
	})
	if err := c.Bind(p); err != nil {
		return
	}

	m := make(map[string]interface{}, 0)
	{
		m["contest_area"] = eSvc.S10Tab4Contest(c, mid, p.AvCID)
		m["tasks"], _ = eSvc.TasksAndPoints(c, mid)
		m["season_id"] = conf.LoadSeasonContestWatch().SeasonID
	}
	c.JSON(m, nil)
}

func s10TabOfContest(c *bm.Context) {
	var mid int64

	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}

	p := new(struct {
		AvCID int64 `form:"av_cid" default:"0"`
	})
	if err := c.Bind(p); err != nil {
		return
	}

	data := eSvc.S10Tab4Contest(c, mid, p.AvCID)
	c.JSON(data, nil)
}

func s10LiveContestSeries(c *bm.Context) {
	var mid int64

	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}

	data := eSvc.S10LiveContestSeries(c, mid)
	c.JSON(data, nil)
}

func s10LiveMoreContest(c *bm.Context) {
	var mid int64

	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}

	data := eSvc.S10LiveMoreContest(c, mid)
	c.JSON(data, nil)
}

func s10Poster4Activity(ctx *bm.Context) {
	var mid int64

	if d, ok := ctx.Get("mid"); ok {
		mid = d.(int64)
	}

	data := eSvc.S10Poster4Activity(ctx, mid)
	ctx.JSON(data, nil)
}

func s10ScoreAnalysis(ctx *bm.Context) {
	param := new(model.ScoreAnalysisRequest)
	if err := ctx.Bind(param); err != nil {
		return
	}

	data := eSvc.S10ScoreAnalysis(ctx, param)
	ctx.JSON(data, nil)
}

func s10CurrentSeries(c *bm.Context) {
	data := eSvc.S10CurrentSeries(c)
	c.JSON(data, nil)
}

func s10MoreContest(c *bm.Context) {
	var mid int64

	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}

	data := eSvc.S10MoreContest(c, mid)
	c.JSON(data, nil)
}

func s10Tasks(c *bm.Context) {
	mstr, _ := c.Get("mid")
	mid, _ := mstr.(int64)
	c.JSON(eSvc.TasksAndPoints(c, mid))
}

func s10RankingData(c *bm.Context) {
	param := new(struct {
		RoundID      string `form:"round_id"`
		NeedPrevious bool   `form:"need_previous"`
		From         string `form:"from" default:"ugc-tab"`
		Eliminate    int    `form:"eliminate"`
	})
	if err := c.Bind(param); err != nil {
		return
	}
	c.JSON(eSvc.S10RankingData(c, param.RoundID, param.NeedPrevious, param.From, param.Eliminate))
}

func matchSeasons(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamMatchSeasons)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.MatchSeasonsInfo(c, mid, p))
}

func seasonTeams(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamSeasonTeams)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.SeasonTeamsInfo(c, mid, p))
}

func seasonsInfo(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	p := new(model.ParamSeasonsInfo)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(eSvc.BatchSeasonsInfo(c, mid, p))
}

func webReplyWall(c *bm.Context) {
	var mid int64
	if d, ok := c.Get("mid"); ok {
		mid = d.(int64)
	}
	c.JSON(eSvc.WebReplyWall(c, mid))
}
