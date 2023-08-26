package http

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-common/library/time"
	"strconv"

	pb "go-gateway/app/app-svr/app-feed/admin/api/search"
	"go-gateway/app/app-svr/app-feed/admin/bvav"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	searchModel "go-gateway/app/app-svr/app-feed/admin/model/search"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/service/search"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

var (
	ctx = context.Background()
)

// Black 黑名单
func blackList(c *bm.Context) {
	var (
		err   error
		pager *searchModel.BlackListPager
	)
	param := &searchModel.BlackListParam{}
	if err = c.Bind(param); err != nil {
		return
	}
	res := map[string]interface{}{}
	if pager, err = searchSvc.BlackList(param); err != nil {
		res["message"] = "获取失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)

}

// interHistory .
func interHistory(c *bm.Context) {
	var (
		err error
	)
	req := &searchModel.InterHisParam{}
	if err = c.Bind(req); err != nil {
		return
	}
	c.JSON(searchSvc.SearchInterHistory(req))
}

// addBlack 添加黑名单
func addBlack(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	param := new(searchModel.Black)
	if err = c.Bind(param); err != nil {
		return
	}
	uid, name := managerInfo(c)
	if err = searchSvc.AddBlack(c, param.Searchword, name, uid); err != nil {
		res["message"] = "获取失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// delBlack 删除黑名单
func delBlack(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	param := new(struct {
		ID int `form:"id" validate:"required"`
	})
	if err = c.Bind(param); err != nil {
		return
	}
	uid, name := managerInfo(c)
	if err = searchSvc.DelBlack(c, param.ID, name, uid); err != nil {
		res["message"] = "获取失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// openAddDarkword 对外 添加黑马词
func openAddDarkword(c *bm.Context) {
	var (
		err  error
		dark searchModel.OpenDark
	)
	res := map[string]interface{}{}
	param := &struct {
		Data string `form:"data" validate:"required"`
	}{}
	if err = c.Bind(param); err != nil {
		return
	}
	if err = json.Unmarshal([]byte(param.Data), &dark); err != nil {
		res["message"] = "参数有误:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = searchSvc.OpenAddDarkword(c, dark); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// openBlacklist 对外 黑名单列表
func openBlacklist(c *bm.Context) {
	var (
		err   error
		black []searchModel.Black
	)
	res := map[string]interface{}{}
	if black, err = searchSvc.BlackAll(); err != nil {
		res["message"] = "获取失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(black, nil)
}

// OpenHotList 对外 黑名单列表
func openHotList(c *bm.Context) {
	var (
		err error
		hot []searchModel.Intervene
	)
	res := map[string]interface{}{}
	if hot, err = searchSvc.OpenHotList(c); err != nil {
		res["message"] = "获取失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(hot, nil)
}

// openDarkword 对外 获取黑马词
func openDarkword(c *bm.Context) {
	var (
		err  error
		dark []searchModel.Dark
	)
	res := map[string]interface{}{}
	if dark, err = searchSvc.GetDarkPub(c); err != nil {
		res["message"] = "获取失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(dark, nil)
}

// openAddHotword 对外 添加搜索热词
func openAddHotword(c *bm.Context) {
	var (
		err error
		hot searchModel.OpenHot
	)
	res := map[string]interface{}{}
	param := &struct {
		Data string `form:"data" validate:"required"`
	}{}
	if err = c.Bind(param); err != nil {
		return
	}
	if err = json.Unmarshal([]byte(param.Data), &hot); err != nil {
		res["message"] = "参数有误:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = searchSvc.OpenAddHotword(c, hot); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// publishHotWord publish hot word
func publishHotWord(c *bm.Context) {
	var (
		err error
		res = map[string]interface{}{}
	)
	uid, name := managerInfo(c)
	if err = searchSvc.SetHotPub(c, name, uid); err != nil {
		res["message"] = "发布失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// publishDark publish dark word
func publishDarkWord(c *bm.Context) {
	var (
		err error
		res = map[string]interface{}{}
	)
	uid, name := managerInfo(c)
	if err = searchSvc.SetDarkPub(c, name, uid); err != nil {
		res["message"] = "发布失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func validateHotInter(t int, v string) (err error) {
	if v != "" && t != 0 {
		var value int64
		if t != common.SeaHotGoToURL {
			value, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return
			}
		}
		switch t {
		case common.SeaHotGoToArch:
			if _, _, err = commonSvc.CardPreview(ctx, common.CardSearchArchive, value); err != nil {
				return
			}
		case common.SeaHotGoToArticle:
			if _, _, err = commonSvc.CardPreview(ctx, common.CardArticle, value); err != nil {
				return
			}
		case common.SeaHotGoToPGC:
			if _, _, err = commonSvc.CardPreview(ctx, common.CardPgcEP, value); err != nil {
				return
			}
		case common.SeaHotGoToURL:
		default:
			return fmt.Errorf("参数错误")
		}
		return
	}
	return
}

// addInter 添加干预
func addInter(c *bm.Context) {
	var (
		err error
		res = map[string]interface{}{}
	)
	param := searchModel.InterveneAdd{}
	if err = c.Bind(&param); err != nil {
		return
	}
	if err = validateImage(param.Type, param.Image); err != nil {
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = validateResource(c, param.Type, param.ResourceId); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if param.GotoType == common.SeaHotGoToArch {
		if param.GotoValue, err = bvav.ToAvStr(param.GotoValue); err != nil {
			res["message"] = "添加失败:" + err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if err = validateHotInter(param.GotoType, param.GotoValue); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	uid, name := managerInfo(c)
	if err = searchSvc.AddInter(c, param, name, uid); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func validateImage(typ int, image string) error {
	if typ == search.SearchInterFire || typ == search.SearchInterNow || typ == search.SearchInterHot || typ == search.SearchInterLive {
		if image == "" {
			return fmt.Errorf("图片不能为空")
		}
	}
	return nil
}

// updateInter 更新干预
func updateInter(c *bm.Context) {
	var (
		err error
		res = map[string]interface{}{}
	)
	param := struct {
		ID         int       `form:"id"`
		Searchword string    `form:"searchword" validate:"required"`
		Rank       int       `form:"position"`
		OldRank    int       `form:"old_position"`
		Tag        string    `form:"tag"`
		Stime      time.Time `form:"stime" validate:"required"`
		Etime      time.Time `form:"etime" validate:"required"`
		Type       int       `form:"type" validate:"required"`
		Image      string    `form:"image"`
		GotoType   int       `form:"goto_type"`
		GotoValue  string    `form:"goto_value"`
		ShowWord   string    `form:"show_word"`
		ResourceId string    `form:"resource_id"`
	}{}
	if err = c.Bind(&param); err != nil {
		return
	}
	if err = validateImage(param.Type, param.Image); err != nil {
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = validateResource(c, param.Type, param.ResourceId); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if param.GotoType == common.SeaHotGoToArch {
		if param.GotoValue, err = bvav.ToAvStr(param.GotoValue); err != nil {
			res["message"] = "更新失败:" + err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if err = validateHotInter(param.GotoType, param.GotoValue); err != nil {
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	inter := searchModel.InterveneAdd{
		ID:         param.ID,
		Searchword: param.Searchword,
		Rank:       param.Rank,
		Tag:        param.Tag,
		Stime:      param.Stime,
		Etime:      param.Etime,
		Type:       param.Type,
		Image:      param.Image,
		GotoType:   param.GotoType,
		GotoValue:  param.GotoValue,
		ShowWord:   param.ShowWord,
		OldRank:    param.OldRank,
		ResourceId: param.ResourceId,
	}
	uid, name := managerInfo(c)
	if err = searchSvc.UpdateInter(c, inter, param.ID, name, uid); err != nil {
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// deleteHot 删除热词
func deleteHot(c *bm.Context) {
	var (
		err error
		res = map[string]interface{}{}
	)
	param := struct {
		ID   int   `form:"id" validate:"required"`
		Type uint8 `form:"type" validate:"required"`
	}{}
	if err = c.Bind(&param); err != nil {
		return
	}
	uid, name := managerInfo(c)
	if err = searchSvc.DeleteHot(c, param.ID, param.Type, name, uid); err != nil {
		res["message"] = "删除失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// deleteDark 删除黑马词
func deleteDark(c *bm.Context) {
	var (
		err error
		res = map[string]interface{}{}
	)
	param := struct {
		ID int `form:"id" validate:"required"`
	}{}
	if err = c.Bind(&param); err != nil {
		return
	}
	uid, name := managerInfo(c)
	if err = searchSvc.DeleteDark(c, param.ID, name, uid); err != nil {
		res["message"] = "删除失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// updateSearch 更新搜索推过来的词
func updateSearch(c *bm.Context) {
	var (
		err error
		res = map[string]interface{}{}
	)
	param := struct {
		ID  int    `form:"id" validate:"required"`
		Tag string `form:"tag"`
	}{}
	if err = c.Bind(&param); err != nil {
		return
	}
	uid, name := managerInfo(c)
	if err = searchSvc.UpdateSearch(c, param.Tag, param.ID, name, uid); err != nil {
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// hotList 搜索热词
func hotList(c *bm.Context) {
	var (
		err    error
		hotout searchModel.HotwordOut
	)
	res := map[string]interface{}{}
	param := struct {
		Date string `form:"date" validate:"required"`
	}{}
	if err = c.Bind(&param); err != nil {
		return
	}
	if hotout, err = searchSvc.HotList(c, param.Date); err != nil {
		res["message"] = "获取热词失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(hotout, nil)
}

// 对热词进行排序
func hotSort(c *bm.Context) {
	var (
		err error
	)
	param := struct {
		ConfigListJson string `form:"config_list" validate:"required"`
	}{}
	if err = c.Bind(&param); err != nil {
		return
	}

	var configList []*searchModel.SortConfigItem
	if err = json.Unmarshal([]byte(param.ConfigListJson), &configList); err != nil {
		c.JSON(nil, err)
		return
	}

	temp := map[string]bool{}
	for _, item := range configList {
		if item.Intervene == 1 {
			if _, ok := temp[item.Searchword]; ok {
				err = ecode.Error(ecode.RequestErr, fmt.Sprintf("排序列表内有重复检索词[%v]", item.Searchword))
				c.JSON(nil, err)
				return
			} else {
				temp[item.Searchword] = true
			}
		}
	}

	uid, name := managerInfo(c)
	if err = searchSvc.HotSort(c, configList, name, uid); err != nil {
		c.JSON(nil, err)
		return
	}

	c.JSON(nil, err)
}

// 获取在线的搜索热词的前20名
func hotTop(c *bm.Context) {
	var (
		err        error
		topListRes searchModel.HotwordOut
	)
	topListRes, err = searchSvc.HotTop(c)

	c.JSON(topListRes, err)
}

// 获取搜索热词的预定池
func hotPending(c *bm.Context) {
	var (
		err            error
		pendingListRes searchModel.HotwordOut
	)

	pendingListRes, err = searchSvc.HotPending(c)

	c.JSON(pendingListRes, err)
}

// darkList 黑马词
func darkList(c *bm.Context) {
	var (
		err     error
		darkout searchModel.DarkwordOut
	)
	res := map[string]interface{}{}
	param := struct {
		Date string `form:"date" validate:"required"`
	}{}
	if err = c.Bind(&param); err != nil {
		return
	}
	if darkout, err = searchSvc.DarkList(c, param.Date); err != nil {
		res["message"] = "获取黑马词失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(darkout, nil)
}

func searchWebCardList(c *bm.Context) {
	var (
		err   error
		pager *show.SearchWebCardPager
	)
	res := map[string]interface{}{}
	req := &show.SearchWebCardLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if pager, err = searchSvc.SearchWebCardList(req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

func addSearchWebCard(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.SearchWebCardAP{}
	if err = c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	req.Person = name
	if err = searchSvc.AddSearchWebCard(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func upSearchWebCard(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.SearchWebCardUP{}
	if err = c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID <= 0 {
		c.JSONMap(map[string]interface{}{"message": "ID 参数不合法"}, ecode.RequestErr)
		c.Abort()
		return
	}
	if err = searchSvc.UpdateSearchWebCard(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func delSearchWebCard(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &struct {
		ID int64 `form:"id" validate:"required"`
	}{}
	if err = c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID <= 0 {
		c.JSONMap(map[string]interface{}{"message": "ID 参数不合法"}, ecode.RequestErr)
		c.Abort()
		return
	}
	if err = searchSvc.DeleteSearchWebCard(req.ID, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func searchWebList(c *bm.Context) {
	var (
		err   error
		pager *show.SearchWebPager
	)
	res := map[string]interface{}{}
	req := &show.SearchWebLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if pager, err = searchSvc.SearchWebList(c, req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

func openSearchWeb(c *bm.Context) {
	var (
		err   error
		pager []*show.OpenSearchWeb
	)
	res := map[string]interface{}{}
	if pager, err = searchSvc.OpenSearchWebList(c); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

func addSearchWeb(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.SearchWebAP{}
	if err = c.BindWith(req, binding.Form); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	req.Person = name

	// 2021M9W1：中期过渡，兼容非视频模块特殊小卡的web卡片类型，后续ipad接入后该字段为必填字段
	if req.PlatVerStr == "" {
		req.PlatVerStr = `[{"platforms":30,"conditions":"","values":""}]`
	}
	if err = json.Unmarshal([]byte(req.PlatVerStr), &req.PlatVer); err != nil {
		res["message"] = "卡片创建失败：版本信息错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}

	if len(req.PlatVer) == 0 {
		res["message"] = "卡片创建失败：版本信息不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = searchSvc.AddSearchWeb(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func upSearchWeb(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.SearchWebUP{}
	if err = c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID <= 0 {
		c.JSONMap(map[string]interface{}{"message": "ID 参数不合法"}, ecode.RequestErr)
		c.Abort()
		return
	}

	// 2021M9W1：中期过渡，兼容非视频模块特殊小卡的web卡片类型，后续ipad接入后该字段为必填字段
	if req.PlatVerStr == "" {
		req.PlatVerStr = `[{"platforms":30,"conditions":"","values":""}]`
	}
	if err = json.Unmarshal([]byte(req.PlatVerStr), &req.PlatVer); err != nil {
		res["message"] = "卡片创建失败：版本信息错误"
		c.JSONMap(res, ecode.RequestErr)
		return
	}

	if len(req.PlatVer) == 0 {
		res["message"] = "卡片创建失败：版本信息不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = searchSvc.UpdateSearchWeb(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func delSearchWeb(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &struct {
		ID int64 `form:"id" validate:"required"`
	}{}
	if err = c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.ID <= 0 {
		c.JSONMap(map[string]interface{}{"message": "ID 参数不合法"}, ecode.RequestErr)
		c.Abort()
		return
	}
	if err = searchSvc.DeleteSearchWeb(req.ID, name, uid); err != nil {
		res["message"] = "卡片删除失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func optSearchWeb(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &struct {
		ID  int64  `form:"id" validate:"required"`
		Opt string `form:"opt" validate:"required"`
	}{}
	if err = c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if err = searchSvc.OptionSearchWeb(req.ID, req.Opt, name, uid); err != nil {
		res["message"] = "修改失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func batchOptSearchWeb(c *bm.Context) {
	var (
		err error
		req = &pb.BatchOptWebReq{}
	)
	if err = c.Bind(req); err != nil {
		return
	}
	req.Uid, req.Uname = util.UserInfo(c)
	c.JSON(searchSvc.BatchOptWeb(c, req))
}

func releaseSearchWeb(c *bm.Context) {
	c.JSON(searchSvc.ReleaseSearchWeb(c))
}

// dySeachList data list
func dySeachList(c *bm.Context) {
	var (
		err   error
		pager *searchModel.DySeaPager
	)
	res := map[string]interface{}{}
	req := &searchModel.DySeachLP{}
	if err = c.Bind(req); err != nil {
		return
	}

	if pager, err = searchSvc.DySeachList(req); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

// addDySeach add data
func addDySeach(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &searchModel.DySeachAP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if uidInter, ok := c.Get("uid"); ok {
		req.Uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		req.Uname = usernameCtx.(string)
	}
	if err = searchSvc.AddDySeach(c, req); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// updateDySeach update data
func updateDySeach(c *bm.Context) {
	var (
		err  error
		name string
		uid  int64
	)
	res := map[string]interface{}{}
	req := &searchModel.DySeachUP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	if err = searchSvc.UpdateDySeach(c, req, name, uid); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// delDySeach del data
func delDySeach(c *bm.Context) {
	var (
		err  error
		name string
		uid  int64
	)
	res := map[string]interface{}{}
	req := &searchModel.DySeachDel{}
	if err = c.Bind(req); err != nil {
		return
	}
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	if err = searchSvc.DeleteDySeach(c, req.ID, name, uid); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// openRecommend .
func openRecommend(c *bm.Context) {
	var (
		err       error
		recommend *searchModel.RecomRes
	)

	res := map[string]interface{}{}
	req := &searchModel.RecomParam{}
	if err = c.Bind(req); err != nil {
		return
	}
	if recommend, err = searchSvc.OpenRecommend2(c, req); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(recommend, nil)
}

func validateResource(ctx context.Context, typ int, resourceId string) (err error) {
	var id int64
	if typ == search.SearchInterLive {
		if resourceId == "" {
			return fmt.Errorf("资源id不能为空")
		}
		if id, err = strconv.ParseInt(resourceId, 10, 64); err != nil {
			return err
		}
		if _, _, err = commonSvc.CardPreview(ctx, common.CardLive, id); err != nil {
			return err
		}
	}
	return
}

func validateShield(t int, v string) (title string, err error) {
	if v != "" && t != 0 {
		var value int64
		value, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return
		}
		switch t {
		case common.SeaShieldAv:
			if title, _, err = commonSvc.CardPreview(ctx, common.CardAv, value); err != nil {
				return
			}
			return
		//	TODO game
		case common.SeaShieldGame:
			return
		case common.SeaShieldPgc:
			if title, _, err = commonSvc.CardPreview(ctx, common.CardPgc, value); err != nil {
				return
			}
			return
		case common.SeaShieldUp:
			if title, _, err = commonSvc.CardPreview(ctx, common.CardUp, value); err != nil {
				return
			}
			return
		case common.SeaShieldLive:
			if title, _, err = commonSvc.CardPreview(ctx, common.CardLive, value); err != nil {
				return
			}
			return
		case common.SeaShieldArt:
			if title, _, err = commonSvc.CardPreview(ctx, common.CardArticle, value); err != nil {
				return
			}
			return
		case common.SeaShieldDync:
			if title, _, err = commonSvc.CardPreview(ctx, common.CardDynamic, value); err != nil {
				return
			}
			return
		case common.SeaShieldGoods:
			// TODO
			return
		case common.SeaShieldShow:
			if title, _, err = commonSvc.CardPreview(ctx, common.CardShow, value); err != nil {
				return
			}
			return
		case common.SeaShieldComic:
			if title, _, err = commonSvc.CardPreview(ctx, common.CardComic, value); err != nil {
				return
			}
			return
		case common.SeaShieldTopic:
			// TODO
			return
		default:
			return "", fmt.Errorf("参数错误")
		}
	}
	return
}

func searchShield(c *bm.Context) {
	var (
		err   error
		pager *show.SearchShieldPager
	)
	res := map[string]interface{}{}
	req := &show.SearchShieldLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if req.IDNew != "" {
		if req.ID, err = bvav.ToAvInt(req.IDNew); err != nil {
			res["message"] = "列表获取失败,bvid转换失败" + err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if pager, err = searchSvc.SearchShieldList(c, req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	for _, v := range pager.Item {
		title, err := validateShield(v.CardType, v.CardValue)
		if err == nil {
			v.Title = title
		}
		if v.CardType == common.SeaShieldAv {
			if v.BvID, err = bvav.AvStrToBvStr(v.CardValue); err != nil {
				v.BvID = err.Error()
				log.Error("searchShield AvStrToBvStr(%s) error(%v)", v.CardValue, err)
				err = nil
			}
		}
	}
	// geteShieldTitle(c, pager.Item)
	c.JSON(pager, nil)
}

func openSearchShield(c *bm.Context) {
	var (
		err   error
		pager []*show.SearchShield
	)
	res := map[string]interface{}{}
	if pager, err = searchSvc.OpenSearchShieldList(c); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

func addSearchShield(c *bm.Context) {
	var (
		err  error
		name string
		uid  int64
	)
	res := map[string]interface{}{}
	req := &show.SearchShieldAP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if req.CardValue, err = bvav.ToAvStr(req.CardValue); err != nil {
		res["message"] = "卡片创建失败,bvid转换失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if _, err = validateShield(req.CardType, req.CardValue); err != nil {
		res["message"] = "添加失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	req.Person = name
	if err = searchSvc.AddSearchShield(c, req, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func upSearchShield(c *bm.Context) {
	var (
		err  error
		name string
		uid  int64
	)
	res := map[string]interface{}{}
	req := &show.SearchShieldUP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if req.CardValue, err = bvav.ToAvStr(req.CardValue); err != nil {
		res["message"] = "卡片创建失败,bvid转换失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if _, err = validateShield(req.CardType, req.CardValue); err != nil {
		res["message"] = "更新失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	req.Person = name
	if err = searchSvc.UpdateSearchShield(c, req, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func optSearchShield(c *bm.Context) {
	var (
		err  error
		name string
		uid  int64
	)
	res := map[string]interface{}{}
	req := &show.SearchShieldOption{}
	if err = c.Bind(req); err != nil {
		return
	}
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	req.Name = name
	req.UID = uid
	if err = searchSvc.OptionSearchShield(req); err != nil {
		res["message"] = "修改失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func batchOptSearchResultSpread(c *bm.Context) {
	var (
		err error
		req = &pb.BatchOptResultSpreadReq{}
	)
	if err = c.Bind(req); err != nil {
		return
	}
	req.Uid, req.Uname = util.UserInfo(c)
	c.JSON(searchSvc.BatchOptResultSpread(c, req))
}

func searchWebModule(c *bm.Context) {
	var (
		err   error
		pager *show.SearchWebModulePager
	)
	res := map[string]interface{}{}
	req := &show.SearchWebModuleLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if pager, err = searchSvc.SearchWebModuleList(c, req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

func addWebModule(c *bm.Context) {
	var (
		err  error
		name string
	)
	res := map[string]interface{}{}
	req := &show.SearchWebModuleAP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	req.UserName = name
	if err = searchSvc.AddSearchWebModule(c, req); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func upSearchWebModule(c *bm.Context) {
	var (
		err  error
		name string
	)
	res := map[string]interface{}{}
	req := &show.SearchWebModuleUP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	req.UserName = name
	if err = searchSvc.UpdateSearchWebModule(c, req); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func optSearchWebModule(c *bm.Context) {
	var (
		err  error
		name string
	)
	res := map[string]interface{}{}
	req := &show.SearchWebModuleOption{}
	if err = c.Bind(req); err != nil {
		return
	}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	req.UserName = name
	if err = searchSvc.OptionSearchWebModule(req); err != nil {
		res["message"] = "修改失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func openSearchModule(c *bm.Context) {
	var (
		err   error
		pager *show.SearchWebModulePager
	)
	res := map[string]interface{}{}
	req := &show.SearchWebModuleLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if pager, err = searchSvc.OpenSearchWebModule(c, req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

// 历史数据
func hotStatistics(c *bm.Context) {
	var (
		err error
		res []*searchModel.StaticticsListItem
		req struct {
			StartTime  string `form:"start_time" validate:"required"`
			EndTime    string `form:"end_time" validate:"required"`
			SearchWord string `form:"search_word" validate:"required"`
		}
	)
	if err = c.Bind(&req); err != nil {
		return
	}
	res, err = searchSvc.GetStatistics(c, req.SearchWord, req.StartTime, req.EndTime)
	c.JSON(res, err)
}

// 实时数据
func hotStatisticsLive(c *bm.Context) {
	var (
		err error
		res []searchModel.StaticticsLiveListItem
		req struct {
			SearchWord string `form:"search_word" validate:"required"`
		}
	)

	if err = c.Bind(&req); err != nil {
		return
	}
	res, err = searchSvc.GetStatisticsLive(c, []string{req.SearchWord})
	c.JSON(res, err)
}

// 给频道服务端用，返回管理后台所有配置过的频道id
func openChannelIds(c *bm.Context) {
	var (
		err error
		res struct {
			Ids  []int64 `json:"ids"`
			Page struct {
				Num   int `json:"num"`
				Size  int `json:"size"`
				Total int `json:"total"`
			} `json:"page"`
		}
		req struct {
			Ps int `form:"ps" default:"20"`
			Pn int `form:"pn" default:"1"`
		}
	)
	if err = c.Bind(&req); err != nil {
		return
	}
	res.Ids, res.Page.Total, err = searchSvc.OpenChannelIdsCache(req.Ps, req.Pn)
	res.Page.Num = req.Pn
	res.Page.Size = req.Ps
	c.JSON(res, err)
}

// UP主别名
func searchAddUpAlias(c *bm.Context) {
	p := new(pb.AddUpAliasReq)
	var err error
	if err = c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}

	_, name := util.UserInfo(c)
	p.Applier = name

	if p.SearchWords == "" {
		err = ecode.Error(-400, "请添加query")
	}

	if p.IsForever == 0 {
		if p.Stime == 0 || p.Etime == 0 || p.Stime >= p.Etime {
			err = ecode.Error(-400, "请填写生效时间")
		}
	}

	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}

	c.JSON(nil, searchSvc.AddUpAlias(c, p))
}

func searchEditUpAlias(c *bm.Context) {
	p := new(pb.EditUpAliasReq)
	var err error
	if err = c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	_, name := util.UserInfo(c)
	p.Applier = name

	if p.Id == 0 {
		err = ecode.Error(-400, "请确认传递参数是否正确")
	}

	if p.SearchWords == "" {
		err = ecode.Error(-400, "请添加query")
	}

	if p.IsForever == 0 {
		if p.Stime == 0 || p.Etime == 0 || p.Stime >= p.Etime {
			err = ecode.Error(-400, "请填写生效时间")
		}
	}

	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}

	c.JSON(nil, searchSvc.EditUpAlias(c, p))
}

func searchToggleUpAlias(c *bm.Context) {
	p := new(pb.ToggleUpAliasReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	if p.Id == 0 || (p.State != 0 && p.State != 1) {
		c.JSON(nil, ecode.Error(-400, "请确认传递参数是否正确"))
		c.Abort()
		return
	}
	c.JSON(nil, searchSvc.ToggleUpAlias(c, p))
}

func searchSearchUpAlias(c *bm.Context) {
	p := new(pb.SearchUpAliasReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := searchSvc.SearchUpAlias(c, p)
	c.JSON(resp, err)
}

func syncSearchUpAlias(c *bm.Context) {
	p := new(pb.SyncUpAliasReq)
	if err := c.BindWith(p, binding.Default(c.Request.Method, c.Request.Header.Get("Content-Type"))); err != nil {
		return
	}
	resp, err := searchSvc.SyncUpAlias(c, p)
	c.JSON(resp, err)
}

func searchExportUpAlias(c *bm.Context) {
	resp, err := searchSvc.ExportUpAlias(c)
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}

	fileContentDisposition := "attachment;filename=result.csv"
	c.Writer.Header().Set("Content-Type", "application/csv")
	c.Writer.Header().Set("Content-Disposition", fileContentDisposition)

	if _, err := c.Writer.Write([]byte(resp)); err != nil {
		c.JSON(nil, err)
		c.Abort()
	}
}
