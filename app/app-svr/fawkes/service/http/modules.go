package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/xstr"

	mdlmdl "go-gateway/app/app-svr/fawkes/service/model/modules"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func sizeRecord(c *bm.Context) {
	var (
		bs  []byte
		err error
	)
	req := c.Request
	res := map[string]interface{}{}
	bs, err = ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("ioutil.ReadAll() error(%v)", err)
		res["message"] = fmt.Sprintf("%v", err)
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	req.Body.Close()
	// params
	var m = &mdlmdl.ModuleSizeReq{}
	if err = json.Unmarshal(bs, m); err != nil {
		log.Error("http sendmail() json.Unmarshal(%s) error(%v)", string(bs), err)
		res["message"] = fmt.Sprintf("%v", err)
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MdlSvr.AddModuleSize(c, m))
}

func groupAdd(c *bm.Context) {
	var (
		params              = c.Request.Form
		appKey, name, cname string
		res                 = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if name = params.Get("name"); name == "" {
		res["message"] = "group name 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	cname = params.Get("cname")
	c.JSON(nil, s.MdlSvr.AddGroup(c, appKey, name, cname))
}

func groupChange(c *bm.Context) {
	var (
		params               = c.Request.Form
		appKey, mName, gName string
		res                  = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if mName = params.Get("m_name"); mName == "" {
		res["message"] = "m_name 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gName = params.Get("g_name"); gName == "" {
		res["message"] = "g_name 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.MdlSvr.ChangeGroup(c, appKey, mName, gName, userName))
}

func groupEdit(c *bm.Context) {
	var (
		params              = c.Request.Form
		appKey, name, cName string
		gID                 int64
		err                 error
		res                 = map[string]interface{}{}
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if hasAuth := s.MngSvr.AuthAdminRole(c, appKey, userName); !hasAuth {
		res["message"] = "无权限操作"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	name = params.Get("name")
	cName = params.Get("cname")
	if name == "" && cName == "" {
		res["message"] = "name 和 cname 都为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MdlSvr.EditGroup(c, gID, name, cName))
}

func groupDel(c *bm.Context) {
	var (
		params = c.Request.Form
		appKey string
		gID    int64
		err    error
		res    = map[string]interface{}{}
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if hasAuth := s.MngSvr.AuthAdminRole(c, appKey, userName); !hasAuth {
		res["message"] = "无权限操作"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gID, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MdlSvr.DeleteGroup(c, gID))
}

func moduleGroupList(c *bm.Context) {
	var (
		params = c.Request.Form
		appKey string
		res    = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.MdlSvr.ListModuleGroup(c, appKey))
}

func listGroups(c *bm.Context) {
	var (
		params = c.Request.Form
		appKey string
		res    = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.MdlSvr.ListAllGroups(c, appKey))
}

func listSizeType(c *bm.Context) {
	var (
		params = c.Request.Form
		appKey string
		limit  int
		err    error
		res    = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if limit, err = strconv.Atoi(params.Get("limit")); err != nil {
		limit = 20
	}
	c.JSON(s.MdlSvr.ListSizeTypes(c, appKey, limit))
}

func sizeModule(c *bm.Context) {
	var (
		params                  = c.Request.Form
		appKey, mName, sizeType string
		res                     = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if mName = params.Get("m_name"); mName == "" {
		res["message"] = "m_name 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	sizeType = params.Get("size_type")
	c.JSON(s.MdlSvr.ListModuleSize(c, appKey, mName, sizeType))
}

func sizeGroup(c *bm.Context) {
	var (
		params                             = c.Request.Form
		appKey, gName, sizeType            string
		resRatio, codeRatio, xcassetsRatio float64
		limit                              int
		err                                error
		res                                = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gName = params.Get("g_name"); gName == "" {
		res["message"] = "g_name 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if limit, err = strconv.Atoi(params.Get("limit")); err != nil {
		limit = 20
	}
	sizeType = params.Get("size_type")
	if sizeType == "" {
		if resRatio, err = strconv.ParseFloat(params.Get("res_ratio"), 64); err != nil {
			res["message"] = "res_ratio 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if codeRatio, err = strconv.ParseFloat(params.Get("code_ratio"), 64); err != nil {
			res["message"] = "code_ratio 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		xcassetsRatioStr := params.Get("xcassets_ratio")
		if xcassetsRatioStr == "" {
			xcassetsRatio = 0
		} else {
			if xcassetsRatio, err = strconv.ParseFloat(xcassetsRatioStr, 64); err != nil {
				res["message"] = "xcassets_ratio 参数错误"
				c.JSONMap(res, ecode.RequestErr)
				return
			}
		}
	}
	c.JSON(s.MdlSvr.ListGroupSize(c, appKey, gName, sizeType, resRatio, codeRatio, xcassetsRatio, limit))
}

func groupsSizeInBuild(c *bm.Context) {
	var (
		params                             = c.Request.Form
		appKey, sizeType                   string
		buildID                            int64
		resRatio, codeRatio, xcassetsRatio float64
		err                                error
		res                                = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	sizeType = params.Get("size_type")
	if sizeType == "" {
		if resRatio, err = strconv.ParseFloat(params.Get("res_ratio"), 64); err != nil {
			res["message"] = "res_ratio 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if codeRatio, err = strconv.ParseFloat(params.Get("code_ratio"), 64); err != nil {
			res["message"] = "code_ratio 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		xcassetsRatioStr := params.Get("xcassets_ratio")
		if xcassetsRatioStr == "" {
			xcassetsRatio = 0
		} else {
			if xcassetsRatio, err = strconv.ParseFloat(xcassetsRatioStr, 64); err != nil {
				res["message"] = "xcassets_ratio 参数错误"
				c.JSONMap(res, ecode.RequestErr)
				return
			}
		}
	}
	c.JSON(s.MdlSvr.ListGroupSizeInBuild(c, appKey, sizeType, buildID, resRatio, codeRatio, xcassetsRatio))
}

func modulesSizeInGroupVersion(c *bm.Context) {
	var (
		params                             = c.Request.Form
		appKey, gName, sizeType            string
		buildID                            int64
		resRatio, codeRatio, xcassetsRatio float64
		err                                error
		res                                = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if gName = params.Get("g_name"); gName == "" {
		res["message"] = "g_name 为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if buildID, err = strconv.ParseInt(params.Get("build_id"), 10, 64); err != nil {
		res["message"] = "build_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	sizeType = params.Get("size_type")
	if sizeType == "" {
		if resRatio, err = strconv.ParseFloat(params.Get("res_ratio"), 64); err != nil {
			res["message"] = "res_ratio 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if codeRatio, err = strconv.ParseFloat(params.Get("code_ratio"), 64); err != nil {
			res["message"] = "code_ratio 参数错误"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		xcassetsRatioStr := params.Get("xcassets_ratio")
		if xcassetsRatioStr == "" {
			xcassetsRatio = 0
		} else {
			if xcassetsRatio, err = strconv.ParseFloat(xcassetsRatioStr, 64); err != nil {
				res["message"] = "xcassets_ratio 参数错误"
				c.JSONMap(res, ecode.RequestErr)
				return
			}
		}
	}
	c.JSON(s.MdlSvr.ListModuleSizeInGroup(c, appKey, gName, sizeType, buildID, resRatio, codeRatio, xcassetsRatio))
}

func modulesConfTotalSizeSet(c *bm.Context) {
	var (
		params            = c.Request.Form
		appKey, version   string
		totalSize         int64
		moduleGroupIDList []int64
		err               error
		res               = map[string]interface{}{}
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if version = params.Get("version"); version == "" {
		res["message"] = "version 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if moduleGroupIDListstr := params.Get("module_group_id"); moduleGroupIDListstr == "" {
		res["message"] = "module_group_id 不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	} else if moduleGroupIDList, err = xstr.SplitInts(moduleGroupIDListstr); err != nil {
		res["message"] = "module_group_id 值非法"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if totalSize, err = strconv.ParseInt(params.Get("total_size"), 10, 64); err != nil {
		res["message"] = "total_size 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.MdlSvr.ModulesConfTotalSizeSet(c, appKey, version, userName, moduleGroupIDList, totalSize))
}

func modulesConfSet(c *bm.Context) {
	var (
		params                                                                             = c.Request.Form
		appKey, version, description                                                       string
		percentage                                                                         float64
		moduleGroupID, totalSize, fixedSize, applyNormalSize, applyForceSize, externalSize int64
		err                                                                                error
		res                                                                                = map[string]interface{}{}
	)
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if version = params.Get("version"); version == "" {
		res["message"] = "version 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if percentage, err = strconv.ParseFloat(params.Get("percentage"), 64); err != nil {
		res["message"] = "percentage 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if moduleGroupID, err = strconv.ParseInt(params.Get("module_group_id"), 10, 64); err != nil {
		res["message"] = "module_group_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if totalSize, err = strconv.ParseInt(params.Get("total_size"), 10, 64); err != nil {
		res["message"] = "total_size 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if fixedSize, err = strconv.ParseInt(params.Get("fixed_size"), 10, 64); err != nil {
		res["message"] = "fixed_size 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if applyNormalSize, err = strconv.ParseInt(params.Get("apply_normal_size"), 10, 64); err != nil {
		res["message"] = "module_group_id 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if applyForceSize, err = strconv.ParseInt(params.Get("apply_force_size"), 10, 64); err != nil {
		res["message"] = "apply_force 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if externalSize, err = strconv.ParseInt(params.Get("external_size"), 10, 64); err != nil {
		res["message"] = "external_size 参数错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	description = params.Get("description")
	c.JSON(nil, s.MdlSvr.ModulesConfSet(c, appKey, version, description, userName, percentage, moduleGroupID, totalSize, fixedSize, applyNormalSize, applyForceSize, externalSize))
}

func modulesConfGet(c *bm.Context) {
	var (
		params          = c.Request.Form
		appKey, version string
		getNewest       bool
		err             error
		res             = map[string]interface{}{}
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "app_key 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if version = params.Get("version"); version == "" {
		res["message"] = "version 异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if getNewest, err = strconv.ParseBool(c.Request.FormValue("get_newest")); err != nil {
		getNewest = false
	}
	c.JSON(s.MdlSvr.ModulesConfGet(c, appKey, version, getNewest))
}
