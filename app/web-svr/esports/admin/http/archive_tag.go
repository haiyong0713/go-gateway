package http

import (
	"encoding/csv"
	"io/ioutil"
	"strings"

	"go-common/library/ecode"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/admin/conf"
	"go-gateway/app/web-svr/esports/admin/model"
)

func autotagInfo(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.AutotagInfo(c, v.ID))
}

func autotagList(c *bm.Context) {
	var (
		list []*model.EsArchiveTag
		cnt  int64
		err  error
	)
	v := new(struct {
		Pn  int64  `form:"pn" validate:"min=0" default:"1"`
		Ps  int64  `form:"ps" validate:"min=0,max=30" default:"20"`
		Tag string `form:"tag"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if list, cnt, err = esSvc.AutotagList(c, v.Pn, v.Ps, v.Tag); err != nil {
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

func addAutotag(c *bm.Context) {
	v := new(model.EsArchiveTag)
	if err := c.Bind(v); err != nil {
		return
	}
	v.UserInfo = userInfo(c)
	if err := esSvc.AddAutotag(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "tag创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func editAutotag(c *bm.Context) {
	v := new(model.EsArchiveTag)
	if err := c.Bind(v); err != nil {
		return
	}
	v.UserInfo = userInfo(c)
	if err := esSvc.EditAutotag(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "tag修改失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func delAutotag(c *bm.Context) {
	v := new(struct {
		IDs []int64 `form:"ids,split" validate:"gt=0,dive,gt=0"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.DelAutotag(c, v.IDs))
}

func importAutotag(c *bm.Context) {
	var (
		err  error
		data []byte
	)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("importTag upload err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Error("importTag ioutil.ReadAll err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("importTag r.ReadAll() err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if l := len(records); l > conf.Conf.Rule.MaxAutoRows || l <= 1 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	// ImportParam.
	var tags []*model.TagImportParam
	midMap := make(map[string]string, len(records))
	for i, v := range records {
		if i == 0 {
			continue
		}
		tag := new(model.TagImportParam)
		// tag
		strTag := v[0]
		if _, ok := midMap[strTag]; ok {
			continue
		}
		tag.Tag = strTag
		midMap[strTag] = strTag
		// gids
		tag.Gids = v[1]
		// match ids
		tag.MatchIDs = v[2]
		tags = append(tags, tag)
	}
	if len(tags) == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, esSvc.AutotagImport(c, tags))
}
