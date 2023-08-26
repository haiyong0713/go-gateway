package http

import (
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
)

func indexIcon(c *bm.Context) {
	c.JSON(resSvc.IndexIcon(c), nil)
}

func playerIcon(c *bm.Context) {
	var (
		params = c.Request.Form
		aid    int64
		typeId int32
		tagIds []int64
		mid    int64
	)
	aid, _ = strconv.ParseInt(params.Get("aid"), 10, 64)
	mid, _ = strconv.ParseInt(params.Get("mid"), 10, 64)
	//nolint:gosec
	tid, _ := strconv.Atoi(params.Get("type_id"))
	showPlayicon, _ := strconv.ParseBool(params.Get("show_playicon"))
	typeId = int32(tid)
	tagIds, _ = xstr.SplitInts(params.Get("tag_ids"))
	c.JSON(resSvc.PlayerIcon(c, aid, tagIds, typeId, mid, showPlayicon, false))
}

func playerPgcIcon(c *bm.Context) {
	var (
		err error
	)
	req := &struct {
		SID          int64 `form:"sid" validate:"required"`
		Mid          int64 `form:"mid"`
		ShowPlayicon bool  `form:"show_playicon"`
	}{}
	if err = c.Bind(req); err != nil {
		return
	}
	c.JSON(resSvc.PlayerPgcIcon(c, req.SID, req.Mid, req.ShowPlayicon), nil)
}

func cmtbox(c *bm.Context) {
	var (
		params = c.Request.Form
		id     int64
		err    error
	)
	if id, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(resSvc.Cmtbox(c, id))
}

func regionCard(c *bm.Context) {
	var (
		params = c.Request.Form
		err    error
	)
	mobiApp := params.Get("mobi_app")
	buildStr := params.Get("build")
	// check params
	build, err := strconv.Atoi(buildStr)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	device := params.Get("device")
	plat := model.Plat(mobiApp, device)
	c.JSON(resSvc.RegionCard(c, plat, build))
}

func audit(c *bm.Context) {
	c.JSON(resSvc.Audit(c), nil)
}
func dySearch(c *bm.Context) {
	c.JSON(resSvc.DySearch(), nil)
}

func customConfig(ctx *bm.Context) {
	req := &pb.CustomConfigRequest{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(resSvc.CustomConfig(ctx, req))
}

func isUploader(ctx *bm.Context) {
	req := &model.WhiteCheckForm{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(resSvc.IsUploaderWhiteCheck(ctx, req))
}

func isNotUploader(ctx *bm.Context) {
	req := &model.WhiteCheckForm{}
	if err := ctx.Bind(req); err != nil {
		return
	}
	ctx.JSON(resSvc.IsNotUploaderWhiteCheck(ctx, req))
}
