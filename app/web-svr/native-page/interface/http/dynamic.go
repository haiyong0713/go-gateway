package http

import (
	bm "go-common/library/net/http/blademaster"

	dynmdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"
)

func actIndex(c *bm.Context) {
	arg := &dynmdl.ParamActIndex{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ua := c.Request.Header.Get("user-agent")
	arg.MobiApp = dynmdl.ParseUserAgent2MobiApp(ua)
	if res, err := c.Request.Cookie("Buvid"); err == nil {
		arg.Buvid = res.Value
	} else {
		if res, err := c.Request.Cookie("buvid3"); err == nil {
			arg.Buvid = res.Value
		}
	}
	c.JSON(likeSvc.ActIndex(c, arg, mid))
}

func menuTab(c *bm.Context) {
	arg := &dynmdl.ParamMenuTab{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ua := c.Request.Header.Get("user-agent")
	arg.MobiApp = dynmdl.ParseUserAgent2MobiApp(ua)
	if res, err := c.Request.Cookie("Buvid"); err == nil {
		arg.Buvid = res.Value
	} else {
		if res, err := c.Request.Cookie("buvid3"); err == nil {
			arg.Buvid = res.Value
		}
	}
	c.JSON(likeSvc.MenuTab(c, arg, mid))
}

func inlineTab(c *bm.Context) {
	arg := &dynmdl.ParamActInline{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	ua := c.Request.Header.Get("user-agent")
	arg.MobiApp = dynmdl.ParseUserAgent2MobiApp(ua)
	if res, err := c.Request.Cookie("Buvid"); err == nil {
		arg.Buvid = res.Value
	} else {
		if res, err := c.Request.Cookie("buvid3"); err == nil {
			arg.Buvid = res.Value
		}
	}
	c.JSON(likeSvc.InlineTab(c, arg, mid))
}

func natPages(c *bm.Context) {
	arg := &dynmdl.ParamNatPages{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(likeSvc.NatPages(c, arg.PageIDs))
}

func actDynamic(c *bm.Context) {
	arg := &dynmdl.ParamActDynamic{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.ActDynamic(c, arg, mid))
}

func resourceDyn(c *bm.Context) {
	arg := &dynmdl.ParamResourceDyn{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.ResourceDyn(c, arg, mid))
}

func resourceRole(c *bm.Context) {
	arg := &dynmdl.ParamResourceRole{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(likeSvc.ResourceRole(c, arg))
}

func natModule(c *bm.Context) {
	arg := &dynmdl.ParamNatModule{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(likeSvc.NatModule(c, arg))
}

func seasonIDs(c *bm.Context) {
	arg := &dynmdl.ParamSeasonIDs{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.SeasonIDs(c, arg, mid))
}

func seasonSource(c *bm.Context) {
	arg := &dynmdl.ParamSeasonSource{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.SeasonSource(c, arg, mid))
}

func resourceAid(c *bm.Context) {
	arg := &dynmdl.ParamResourceAid{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(likeSvc.ResourceAid(c, arg))
}

func timelineSource(c *bm.Context) {
	arg := &dynmdl.ParamTimelineSource{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(likeSvc.TimelineSource(c, arg.FID, 0, arg.Offset, arg.Ps))
}

func newVideoAid(c *bm.Context) {
	arg := &dynmdl.ParamAid{}
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(likeSvc.NewVideoAid(c, arg))
}

func newVideoDyn(c *bm.Context) {
	arg := &dynmdl.ParamVideoDyn{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.NewVideoDyn(c, arg, mid))
}

func liveDyn(c *bm.Context) {
	arg := &dynmdl.ParamLiveDyn{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.LiveDyn(c, arg, mid))
}

func minePages(c *bm.Context) {
	arg := &dynmdl.ParamMinePages{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.MinePages(c, mid, arg.Offset, arg.Ps))
}

func upActPages(c *bm.Context) {
	arg := &dynmdl.ParamMinePages{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.UpActPages(c, mid, arg.Offset, arg.Ps))
}

func tsPage(c *bm.Context) {
	arg := &dynmdl.ParamTsPage{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.TsPageResource(c, mid, arg.PID))
}

func minePageAdd(c *bm.Context) {
	c.JSON(nil, likeSvc.MinePageAdd(c))
}

func minePageSave(c *bm.Context) {
	arg := &dynmdl.ParamMinePageSave{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.MinePageSave(c, mid, arg))
}

func tsRemark(c *bm.Context) {
	var arg struct {
		Remark string `form:"remark" validate:"required"`
		Type   int    `form:"type"` //1:title 2:remark
	}
	if err := c.Bind(&arg); err != nil {
		return
	}
	var err error
	if arg.Type == 1 {
		_, err = likeSvc.TitleCheck(c, arg.Remark)
	} else {
		err = likeSvc.RemarkCheck(c, arg.Remark)
	}
	c.JSON(nil, err)
}

func tsWhiteSave(c *bm.Context) {
	c.JSON(likeSvc.TsWhiteSave(c))
}

func tsWhite(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.TsWhite(c, mid))
}

func inlineTsWhite(c *bm.Context) {
	v := new(struct {
		Uid int64 `form:"uid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(likeSvc.InlineTsWhite(c, v.Uid))
}

func myArchiveList(c *bm.Context) {
	var args struct {
		Pn int64 `form:"pn"`
		Ps int64 `form:"ps"`
	}
	if err := c.Bind(&args); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	if args.Pn < 1 {
		args.Pn = 1
	}
	if args.Ps < 1 || args.Ps > 20 {
		args.Ps = 20
	}
	c.JSON(likeSvc.MyArchiveList(c, mid, args.Pn, args.Ps))
}

func actArchiveList(c *bm.Context) {
	args := &dynmdl.ActArchiveListReq{}
	if err := c.Bind(args); err != nil {
		return
	}
	if args.Ps < 1 || args.Ps > 20 {
		args.Ps = 20
	}
	c.JSON(likeSvc.ActArchiveList(c, args))
}

func resourceOrigin(c *bm.Context) {
	arg := &dynmdl.ParamResourceOrigin{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.ResourceOrigin(c, arg, mid))
}

func progress(c *bm.Context) {
	args := &dynmdl.ProgressReq{}
	if err := c.Bind(args); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.Progress(c, args, mid))
}

func tsSetting(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.TsSetting(c, mid))
}

func tsSpace(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.TsSpace(c, mid))
}

func tsSpaceSave(c *bm.Context) {
	req := &dynmdl.TsSpaceSaveReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	req.From = dynmdl.SpaceSaveFromUser
	c.JSON(nil, likeSvc.TsSpaceSave(c, req, mid))
}
func editorOrigin(c *bm.Context) {
	arg := &dynmdl.ParamEditorOrigin{}
	if err := c.Bind(arg); err != nil {
		return
	}
	var buvid string
	if res, err := c.Request.Cookie("Buvid"); err == nil {
		buvid = res.Value
	} else {
		if res, err := c.Request.Cookie("buvid3"); err == nil {
			buvid = res.Value
		}
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.EditorOrigin(c, arg, mid, buvid))
}

func edViewedArcs(c *bm.Context) {
	req := &dynmdl.EdViewedArcsReq{}
	if err := c.Bind(req); err != nil {
		return
	}
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.EdViewedArcs(c, req, mid))
}

func partition(c *bm.Context) {
	c.JSON(likeSvc.Partition(c))
}

func partitionV2(c *bm.Context) {
	var mid int64
	if midInter, ok := c.Get("mid"); ok {
		mid = midInter.(int64)
	}
	c.JSON(likeSvc.PartitionV2(c, mid))
}
