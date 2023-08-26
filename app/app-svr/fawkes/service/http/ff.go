package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/utils/collection"
	"go-common/library/xstr"

	ffmdl "go-gateway/app/app-svr/fawkes/service/model/ff"
)

func appFFWhithlist(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.FFSvr.AppFFWhithlist(c, appKey, env))
}

func appFFWhithlistAdd(c *bm.Context) {
	var (
		params                = c.Request.Form
		res                   = map[string]interface{}{}
		appKey, env, userName string
		mids                  []int64
		err                   error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if midStr := params.Get("mid"); midStr == "" {
		res["message"] = "mid不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	} else if mids, err = xstr.SplitInts(midStr); err != nil {
		res["message"] = "mid值非法"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.FFSvr.AppFFWhithlistAdd(c, appKey, env, userName, mids))
}

func appFFWhithlistDel(c *bm.Context) {
	var (
		params                = c.Request.Form
		res                   = map[string]interface{}{}
		appKey, env, userName string
		mid                   int64
		err                   error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if mid, err = strconv.ParseInt(params.Get("mid"), 10, 64); err != nil {
		res["message"] = "mid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.FFSvr.AppFFWhithlistDel(c, appKey, env, userName, mid))
}

func appFFConfigSet(c *bm.Context) {
	var (
		bs                                                                           []byte
		res                                                                          = map[string]interface{}{}
		appKey, env, key, desc, salt, status, bucket, version, unVersion, romVersion string
		brand, unBrand, network, isp, channel, whith, blackList, userName, blackMid  string
		bucketCount                                                                  int64
		err                                                                          error
	)
	if bs, err = ioutil.ReadAll(c.Request.Body); err != nil {
		res["message"] = fmt.Sprintf("ioutil.ReadAll() error(%v)", err)
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.Request.Body.Close()
	// params
	var ffs *ffmdl.FF
	if err = json.Unmarshal(bs, &ffs); err != nil {
		res["message"] = fmt.Sprintf("http submit() json.Unmarshal(%s) error(%v)", string(bs), err)
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if appKey = ffs.AppKey; appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = ffs.Env; env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if key = ffs.Key; env == "" {
		res["message"] = "key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	key = strings.ToLower(key)
	if desc = ffs.Desc; desc == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if status = ffs.Status; status == "" {
		res["message"] = "status异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if status != "um" && status != "dm" && status != "u" {
		res["message"] = "status值非法"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	salt = ffs.Salt
	if bucket = ffs.Bucket; bucket == "" {
		res["message"] = "bucket异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	version = ffs.Version
	if version != "" {
		var vj *ffmdl.Version
		if err = json.Unmarshal([]byte(version), &vj); err != nil {
			res["message"] = "version值非法A"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if vj.Min == 0 && vj.Max == 0 {
			res["message"] = "version值非法B"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	unVersion = ffs.UnVersion
	romVersion = ffs.RomVersion
	brand = ffs.Brand
	unBrand = ffs.UnBrand
	network = ffs.Network
	isp = ffs.ISP
	channel = ffs.Channel
	if channel != "" {
		if _, err = xstr.SplitInts(channel); err != nil {
			res["message"] = "channel值非法"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	whith = ffs.Whith
	if whith != "" {
		if _, err = xstr.SplitInts(whith); err != nil {
			res["message"] = "whith值非法"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	bucketCount = ffs.BucketCount
	blackList = ffs.BlackList
	if blackList != "" {
		var bls []*ffmdl.FF
		if err = json.Unmarshal([]byte(blackList), &bls); err != nil {
			res["message"] = "black_list值非法"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	blackMid = ffs.BlackMid
	if blackMid != "" {
		if _, err = collection.ToSliceInt(strings.Split(blackMid, ",")); err != nil {
			res["message"] = "black_mid值非法"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.FFSvr.AppFFConfigSet(c, appKey, env, userName, key, desc, status, salt, bucket, version, unVersion,
		romVersion, brand, unBrand, network, isp, channel, whith, blackList, blackMid, bucketCount))
}

func appFFList(c *bm.Context) {
	var (
		params                           = c.Request.Form
		res                              = map[string]interface{}{}
		appKey, env, userName, filterKey string
		pn, ps                           int
		err                              error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn <= 0 {
		res["message"] = "pn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		res["message"] = "ps异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	filterKey = params.Get("filter_key")
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(s.FFSvr.AppFFList(c, appKey, env, userName, filterKey, pn, ps))
}

func appFFPublish(c *bm.Context) {
	var (
		params                      = c.Request.Form
		res                         = map[string]interface{}{}
		appKey, env, userName, desc string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if desc = params.Get("description"); desc == "" {
		res["message"] = "description异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(nil, s.FFSvr.AppFFPublish(c, appKey, env, userName, desc))
}

func appFFHistory(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		pn, ps      int
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pn, err = strconv.Atoi(params.Get("pn")); err != nil || pn <= 0 {
		res["message"] = "pn异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps, err = strconv.Atoi(params.Get("ps")); err != nil {
		res["message"] = "ps异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ps < 1 || ps > 20 {
		ps = 20
	}
	c.JSON(s.FFSvr.AppFFHistory(c, appKey, env, pn, ps))
}

func appFFHistoryByID(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
		ffid   int64
		err    error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if ffid, err = strconv.ParseInt(params.Get("id"), 10, 64); err != nil {
		res["message"] = "id异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.FFSvr.AppFFHistoryByID(c, appKey, ffid))
}

func appFFDiff(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
		fvid        int64
		err         error
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if fvid, err = strconv.ParseInt(params.Get("fvid"), 10, 64); err != nil {
		res["message"] = "fvid异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.FFSvr.AppFFDiff(c, appKey, env, fvid))
}

func appFFPublishDiff(c *bm.Context) {
	var (
		params      = c.Request.Form
		res         = map[string]interface{}{}
		appKey, env string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.FFSvr.AppFFPublishDiff(c, appKey, env))
}

func appFFConfig(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, env, key string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if key = params.Get("key"); key == "" {
		res["message"] = "key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.FFSvr.AppFFConfig(c, appKey, env, key))
}

func appFFConfigDel(c *bm.Context) {
	var (
		params           = c.Request.Form
		res              = map[string]interface{}{}
		appKey, env, key string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if env = params.Get("env"); env == "" {
		res["message"] = "env异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if key = params.Get("key"); key == "" {
		res["message"] = "key异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, s.FFSvr.AppFFConfigDel(c, appKey, env, key))
}

func appFFModifyCount(c *bm.Context) {
	var (
		params = c.Request.Form
		res    = map[string]interface{}{}
		appKey string
	)
	if appKey = params.Get("app_key"); appKey == "" {
		res["message"] = "appkey异常"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(s.FFSvr.AppFFModifyCount(c, appKey))
}
