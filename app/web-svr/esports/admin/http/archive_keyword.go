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

func keywordInfo(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.KeywordInfo(c, v.ID))
}

func keywordList(c *bm.Context) {
	var (
		list []*model.EsArchiveKeyword
		cnt  int64
		err  error
	)
	v := new(struct {
		Pn      int64  `form:"pn" validate:"min=0" default:"1"`
		Ps      int64  `form:"ps" validate:"min=0,max=30" default:"20"`
		Keyword string `form:"keyword"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if list, cnt, err = esSvc.KeywordList(c, v.Pn, v.Ps, v.Keyword); err != nil {
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

func addKeyword(c *bm.Context) {
	v := new(model.EsArchiveKeyword)
	if err := c.Bind(v); err != nil {
		return
	}
	v.UserInfo = userInfo(c)
	if err := esSvc.AddKeyword(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "关键词创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func editKeyword(c *bm.Context) {
	v := new(model.EsArchiveKeyword)
	if err := c.Bind(v); err != nil {
		return
	}
	v.UserInfo = userInfo(c)
	if err := esSvc.EditKeyword(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "关键词修改失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func delKeyword(c *bm.Context) {
	v := new(struct {
		IDs []int64 `form:"ids,split" validate:"gt=0,dive,gt=0"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.DelKeyword(c, v.IDs))
}

func importKeyword(c *bm.Context) {
	var (
		err  error
		data []byte
	)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("importKeyword upload err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Error("importKeyword ioutil.ReadAll err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("importKeyword r.ReadAll() err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if l := len(records); l > conf.Conf.Rule.MaxAutoRows || l <= 1 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	// ImportParam.
	var keywords []*model.KeywordImportParam
	keyMap := make(map[string]string, len(records))
	for i, v := range records {
		if i == 0 {
			continue
		}
		keyword := new(model.KeywordImportParam)
		// strKey
		strKey := v[0]
		if _, ok := keyMap[strKey]; ok {
			continue
		}
		keyword.Keyword = strKey
		keyMap[strKey] = strKey
		// gids
		keyword.Gids = v[1]
		// match ids
		keyword.MatchIDs = v[2]
		keywords = append(keywords, keyword)
	}
	if len(keywords) == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, esSvc.KeywordImport(c, keywords))
}
