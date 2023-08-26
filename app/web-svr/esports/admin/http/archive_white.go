package http

import (
	"encoding/csv"
	"io/ioutil"
	"strconv"
	"strings"

	"go-common/library/ecode"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/esports/admin/conf"
	"go-gateway/app/web-svr/esports/admin/model"
)

func whiteInfo(c *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(esSvc.WhiteInfo(c, v.ID))
}

func whiteList(c *bm.Context) {
	var (
		list []*model.EsArchiveWhite
		cnt  int64
		err  error
	)
	v := new(struct {
		Mid string `form:"mid"`
		Pn  int64  `form:"pn" validate:"min=0" default:"1"`
		Ps  int64  `form:"ps" validate:"min=0,max=30" default:"20"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if list, cnt, err = esSvc.WhiteList(c, v.Pn, v.Ps, v.Mid); err != nil {
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

func addWhite(c *bm.Context) {
	v := new(model.EsArchiveWhite)
	if err := c.Bind(v); err != nil {
		return
	}
	v.UserInfo = userInfo(c)
	if err := esSvc.AddWhite(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "白名单创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func editWhite(c *bm.Context) {
	v := new(model.EsArchiveWhite)
	if err := c.Bind(v); err != nil {
		return
	}
	v.UserInfo = userInfo(c)
	if err := esSvc.EditWhite(c, v); err != nil {
		res := map[string]interface{}{}
		res["message"] = "白名单修改失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func delWhite(c *bm.Context) {
	v := new(struct {
		IDs []int64 `form:"ids,split" validate:"gt=0,dive,gt=0"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.DelWhite(c, v.IDs))
}

func importWhite(c *bm.Context) {
	var (
		err  error
		data []byte
	)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("importWhite upload err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Error("importWhite ioutil.ReadAll err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("importWhite r.ReadAll() err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if l := len(records); l > conf.Conf.Rule.MaxAutoRows || l <= 1 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	// ImportParam.
	var whites []*model.WhiteImportParam
	midMap := make(map[int64]int64, len(records))
	for i, v := range records {
		if i == 0 {
			continue
		}
		white := new(model.WhiteImportParam)
		// mid
		if mid, err := strconv.ParseInt(v[0], 10, 64); err != nil || mid <= 0 {
			log.Warn("importWhite strconv.ParseInt(%s) error(%v)", v[0], err)
			continue
		} else {
			if _, ok := midMap[mid]; ok {
				continue
			}
			white.Mid = mid
			midMap[mid] = mid
		}
		// gids
		white.Gids = v[1]
		// match ids
		white.MatchIDs = v[2]
		whites = append(whites, white)
	}
	if len(whites) == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, esSvc.WhiteImport(c, whites))
}
