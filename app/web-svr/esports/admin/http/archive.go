package http

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/bvav"
	"go-gateway/app/web-svr/esports/admin/conf"
	"go-gateway/app/web-svr/esports/admin/model"
)

func arcList(c *bm.Context) {
	var (
		list []*model.ArcResult
		cnt  int
		err  error
	)
	res := make(map[string]interface{})
	v := new(model.ArcListParam)
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Aid, err = bvav.ToAvStr(v.Aid); err != nil {
		res["message"] = "参数错误:" + err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	if list, cnt, err = esSvc.ArcList(c, v); err != nil {
		c.JSON(nil, err)
		return
	}
	data := make(map[string]interface{}, 2)
	page := map[string]int{
		"num":   v.Pn,
		"size":  v.Ps,
		"total": cnt,
	}
	data["page"] = page
	data["list"] = list
	c.JSON(data, nil)
}

// uniqueAid .
func uniqueAid(values []int64) (res []int64, err error) {
	mapAids := make(map[int64]bool)
	for _, v := range values {
		if mapAids[v] {
			return nil, fmt.Errorf("稿件ID(%d)重复", v)
		}
		res = append(res, v)
		mapAids[v] = true
	}
	return res, nil
}

func batchAddArc(c *bm.Context) {
	var (
		err error
	)
	res := make(map[string]interface{})
	v := new(model.ArcAddParam)
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Aids, err = bvav.AvsStrToAvsIntSlice(v.AidsStr); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	if v.Aids, err = uniqueAid(v.Aids); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	if len(v.Aids) > conf.Conf.Rule.MaxBatchArcLimit {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, esSvc.BatchAddArc(c, v))
}

func batchEditArc(c *bm.Context) {
	var (
		err error
	)
	res := make(map[string]interface{})
	v := new(model.ArcAddParam)
	if err = c.Bind(v); err != nil {
		return
	}
	if v.Aids, err = bvav.AvsStrToAvsIntSlice(v.AidsStr); err != nil {
		res["message"] = "bvid转换失败:" + err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	if v.Aids, err = uniqueAid(v.Aids); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	if len(v.Aids) > conf.Conf.Rule.MaxBatchArcLimit {
		res["message"] = "稿件数量超过最大限制"
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	if err = esSvc.BatchEditArc(c, v); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func editArc(c *bm.Context) {
	v := new(model.ArcImportParam)
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.EditArc(c, v))
}

func arcImportCSV(c *bm.Context) {
	var (
		err  error
		data []byte
	)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Error("arcImportCSV upload err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Error("arcImportCSV ioutil.ReadAll err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("r.ReadAll() err(%v)", err)
		c.JSON(nil, xecode.RequestErr)
		return
	}
	if l := len(records); l > conf.Conf.Rule.MaxCSVRows || l <= 1 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	var arcs []*model.ArcImportParam
	aidMap := make(map[int64]int64, len(arcs))
	for _, v := range records {
		arc := new(model.ArcImportParam)
		var (
			aidStr string
		)
		// aid
		if aidStr, err = bvav.ToAvStr(v[0]); err != nil {
			log.Warn("arcImportCSV ToAvStr(%s) error(%v)", v[0], err)
			continue
		}
		if aid, err := strconv.ParseInt(aidStr, 10, 64); err != nil || aid <= 0 {
			log.Warn("arcImportCSV strconv.ParseInt(%s) error(%v)", v[0], err)
			continue
		} else {
			if _, ok := aidMap[aid]; ok {
				continue
			}
			arc.Aid = aid
			aidMap[aid] = aid
		}
		// gids
		if gids, err := xstr.SplitInts(v[1]); err != nil {
			log.Warn("arcImportCSV gids xstr.SplitInts(%s) aid(%d) error(%v)", v[1], arc.Aid, err)
		} else {
			for _, id := range gids {
				if id > 0 {
					arc.Gids = append(arc.Gids, id)
				}
			}
		}
		// match ids
		if matchIDs, err := xstr.SplitInts(v[2]); err != nil {
			log.Warn("arcImportCSV match xstr.SplitInts(%s) aid(%d) error(%v)", v[2], arc.Aid, err)
		} else {
			for _, id := range matchIDs {
				if id > 0 {
					arc.MatchIDs = append(arc.MatchIDs, id)
				}
			}
		}
		// team ids
		if teamIDs, err := xstr.SplitInts(v[3]); err != nil {
			log.Warn("arcImportCSV team xstr.SplitInts(%s) aid(%d) error(%v)", v[3], arc.Aid, err)
		} else {
			for _, id := range teamIDs {
				if id > 0 {
					arc.TeamIDs = append(arc.TeamIDs, id)
				}
			}
		}
		// tag ids
		if tagIDs, err := xstr.SplitInts(v[4]); err != nil {
			log.Warn("arcImportCSV tag xstr.SplitInts(%s) aid(%d) error(%v)", v[4], arc.Aid, err)
		} else {
			for _, id := range tagIDs {
				if id > 0 {
					arc.TagIDs = append(arc.TagIDs, id)
				}
			}
		}
		// years
		if years, err := xstr.SplitInts(v[5]); err != nil {
			log.Warn("arcImportCSV year xstr.SplitInts(%s) aid(%d) error(%v)", v[5], arc.Aid, err)
		} else {
			for _, id := range years {
				if id > 0 {
					arc.Years = append(arc.Years, id)
				}
			}
		}
		arcs = append(arcs, arc)
	}
	if len(arcs) == 0 {
		c.JSON(nil, xecode.RequestErr)
		return
	}
	c.JSON(nil, esSvc.ArcImportCSV(c, arcs))
}

func batchDelArc(c *bm.Context) {
	v := new(struct {
		Aids []int64 `form:"aids,split" validate:"dive,gt=1,required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, esSvc.BatchDelArc(c, v.Aids))
}

func batchPassArc(c *bm.Context) {
	var err error
	res := make(map[string]interface{})
	v := new(struct {
		IDs []int64 `form:"ids,split" validate:"gt=0,dive,gt=0"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if err = esSvc.BatchPassArc(c, v.IDs); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func batchNopassArc(c *bm.Context) {
	var err error
	res := make(map[string]interface{})
	v := new(struct {
		IDs []int64 `form:"ids,split" validate:"gt=0,dive,gt=0"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	if err = esSvc.BatchNopassArc(c, v.IDs); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, xecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}
