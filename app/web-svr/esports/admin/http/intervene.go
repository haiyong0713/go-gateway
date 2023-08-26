package http

import (
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/admin/model"
)

func infoIntervene(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.InterveneInfo(c, v.ID))
}

func listIntervene(c *bm.Context) {
	var (
		list []*model.SearchInfo
		cnt  int64
		err  error
	)
	v := new(struct {
		Pn    int64  `form:"pn" validate:"min=0" default:"1"`
		Ps    int64  `form:"ps" validate:"min=0,max=20" default:"20"`
		Sort  int64  `form:"sort" default:"1"`
		Title string `form:"title"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if list, cnt, err = esSvc.InterveneList(c, v.Pn, v.Ps, v.Sort, v.Title); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int64{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": cnt,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

func addIntervene(c *bm.Context) {
	v := new(model.EsSearchCard)
	if err := c.Bind(v); err != nil {
		return
	}
	v.QueryName = strings.ToLower(v.QueryName)
	if err := esSvc.AddIntervene(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "搜索干预创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func editIntervene(c *bm.Context) {
	v := new(model.EsSearchCard)
	if err := c.Bind(v); err != nil {
		return
	}
	v.QueryName = strings.ToLower(v.QueryName)
	if err := esSvc.EditIntervene(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "搜索干预修改失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func forbidIntervene(c *bm.Context) {
	v := new(struct {
		ID    int64 `form:"id" validate:"min=1"`
		State int   `form:"state" validate:"min=0,max=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.ForbidIntervene(c, v.ID, v.State))
}
