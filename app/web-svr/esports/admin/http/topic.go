package http

import (
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/admin/model"
)

func addTopicVideoList(c *bm.Context) {
	v := new(model.VideoList)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.AddTopicVideoList(c, v))
}

func editTopicVideoList(c *bm.Context) {
	v := new(model.VideoList)
	if err := c.Bind(v); err != nil {
		return
	}
	if v.ID <= 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(esSvc.EditTopicVideoList(c, v))
}

func forbidTopicVideoList(c *bm.Context) {
	v := new(struct {
		ID    int64 `form:"id" validate:"min=1"`
		State int   `form:"state" validate:"min=0,max=1" default:"1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.ForbidTopicVideoList(c, v.ID, v.State))
}

func infoTopicVideoList(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.VideoListInfo(c, v.ID))
}

func topicVideoLists(c *bm.Context) {
	var (
		list []*model.VideoList
		cnt  int64
		err  error
	)
	v := new(struct {
		Pn    int64  `form:"pn" validate:"min=0"`
		Ps    int64  `form:"ps" validate:"min=0,max=30"`
		Title string `form:"title"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Pn == 0 {
		v.Pn = 1
	}
	if v.Ps == 0 {
		v.Ps = 20
	}
	if list, cnt, err = esSvc.TopicVideoLists(c, v.Pn, v.Ps, v.Title); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"count": cnt,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func checkArchive(ctx *bm.Context) {
	v := new(struct {
		ArchiveIDs []string `form:"archive_ids,split" validate:"gt=0,dive,gt=0"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(esSvc.TopicCheckArchives(ctx, v.ArchiveIDs))
}

func videoFilter(c *bm.Context) {
	p := new(model.ParamVideoFilter)
	if err := c.Bind(p); err != nil {
		return
	}
	c.JSON(esSvc.TopicVideoFilter(c, p))
}
