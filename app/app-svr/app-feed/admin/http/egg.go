package http

import (
	"encoding/json"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	eggModel "go-gateway/app/app-svr/app-feed/admin/model/egg"

	"github.com/jinzhu/gorm"
)

func addEgg(c *bm.Context) {
	var (
		err error
		p   []eggModel.Plat
		pic eggModel.EggPic
	)
	res := map[string]interface{}{}
	param := new(eggModel.Obj)
	if err = c.Bind(param); err != nil {
		return
	}
	uid, name := managerInfo(c)
	e := &eggModel.Egg{
		Stime:            param.Stime,
		Etime:            param.Etime,
		ShowCount:        param.ShowCount,
		Type:             param.Type,
		ReType:           param.ReType,
		ReValue:          param.ReValue,
		UID:              uid,
		Publish:          eggModel.NotPublish,
		Person:           name,
		Delete:           eggModel.NotDelete,
		PreTime:          param.PreTime,
		Mids:             param.Mids,
		MaskTransparency: param.MaskTransparency,
		MaskColor:        param.MaskColor,
	}
	if err = json.Unmarshal([]byte(param.Plat), &p); err != nil {
		res["message"] = "参数有误:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = checkEgg(param.Query, e); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if param.Type == eggModel.EggTypePic {
		if err = json.Unmarshal([]byte(param.Pic), &pic); err != nil {
			res["message"] = "图片彩蛋: 图片参数有误:" + err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if err = eggSvc.AddEgg(e, p, param.Query, pic); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// delEgg delete egg
func delEgg(c *bm.Context) {
	var (
		err error
		egg *eggModel.Egg
	)
	res := map[string]interface{}{}
	param := &struct {
		ID uint `form:"id" validate:"required"`
	}{}
	if err = c.Bind(param); err != nil {
		return
	}
	if egg, err = eggSvc.EggWithID(param.ID); err != nil {
		if err == gorm.ErrRecordNotFound {
			res["message"] = "找不到数据:" + err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		res["message"] = "删除失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if egg.Publish == eggModel.Publish {
		res["message"] = "已发布彩蛋不能删除"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	uid, name := managerInfo(c)
	if err = eggSvc.DelEgg(param.ID, name, uid); err != nil {
		res["message"] = "删除失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// pubEgg publish egg
func pubEgg(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	param := &struct {
		ID      uint  `form:"id" validate:"required"`
		Publish uint8 `form:"publish"`
	}{}
	if err = c.Bind(param); err != nil {
		return
	}
	uid, name := managerInfo(c)
	if err = eggSvc.PubEgg(param.ID, param.Publish, name, uid); err != nil {
		res["message"] = "发布失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// updateEgg update egg
func updateEgg(c *bm.Context) {
	var (
		err error
		p   []eggModel.Plat
		pic eggModel.EggPic
	)
	res := map[string]interface{}{}
	param := new(eggModel.ObjUpdate)
	if err = c.Bind(param); err != nil {
		return
	}
	uid, name := managerInfo(c)
	e := &eggModel.Egg{
		ID:               param.ID,
		Stime:            param.Stime,
		Etime:            param.Etime,
		ShowCount:        param.ShowCount,
		Type:             param.Type,
		ReType:           param.ReType,
		ReValue:          param.ReValue,
		UID:              uid,
		Publish:          param.Publish,
		Person:           name,
		Delete:           eggModel.NotDelete,
		PreTime:          param.PreTime,
		Mids:             param.Mids,
		MaskTransparency: param.MaskTransparency,
		MaskColor:        param.MaskColor,
	}
	if err = json.Unmarshal([]byte(param.Plat), &p); err != nil {
		res["message"] = "json解析失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = checkEgg(param.Query, e); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if param.Type == eggModel.EggTypePic {
		if err = json.Unmarshal([]byte(param.Pic), &pic); err != nil {
			res["message"] = "图片彩蛋: 图片参数有误:" + err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if err = eggSvc.UpdateEgg(e, p, param.Query, pic); err != nil {
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// checkEgg check egg error
func checkEgg(w []string, e *eggModel.Egg) (err error) {
	var (
		flag       bool
		eW         string
		queryLimit = 100
	)
	if len(w) > queryLimit {
		err = fmt.Errorf("搜索词不能大于%d个", queryLimit)
		return
	}
	if flag, eW, err = eggSvc.IsWdExist(w, e.Stime, e.Etime, e.ID); err != nil {
		log.Error("eggSrv.checkEgg IsWdExist error(%v)", err)
		return
	}
	if flag {
		err = fmt.Errorf("搜索词 (%v) 已有彩蛋，请勿重复添加", eW)
		return
	}
	return
}

// indexEgg get egg list
func indexEgg(c *bm.Context) {
	var (
		err  error
		eggs *eggModel.IndexPager
	)
	res := map[string]interface{}{}
	param := &eggModel.IndexParam{}
	if err = c.Bind(param); err != nil {
		return
	}
	if eggs, err = eggSvc.IndexEgg(param); err != nil {
		res["message"] = "查询失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(eggs, nil)
}

// searchEgg search api for search
func searchEgg(c *bm.Context) {
	var (
		err  error
		eggs []eggModel.SearchEgg
	)
	res := map[string]interface{}{}
	if eggs, err = eggSvc.SearchEgg(); err != nil {
		res["message"] = "搜索查询失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(eggs, nil)
}

// SearchEggWeb search api for web
func SearchEggWeb(c *bm.Context) {
	var (
		err  error
		eggs []eggModel.SearchEggWeb
	)
	res := map[string]interface{}{}
	if eggs, err = eggSvc.SearchEggWeb(); err != nil {
		res["message"] = "Web搜索查询失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(eggs, nil)
}
