package http

import (
	"fmt"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-feed/admin/bvav"
	"go-gateway/app/app-svr/app-feed/admin/middle"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
	xecode "go-gateway/app/app-svr/app-feed/ecode"

	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"

	taGrpcModel "git.bilibili.co/bapis/bapis-go/community/interface/tag"
)

// webRcmdAuth .
func webRcmdAuth(c *bm.Context) {
	var (
		permsStr   []string
		name       string
		roleValues *common.Role
		err        error
		perms      interface{}
		ok         bool
	)
	res := map[string]interface{}{}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	if perms, ok = c.Get(middle.CtxPermissions); ok {
		if permsStr, ok = perms.([]string); !ok {
			c.JSON(nil, nil)
			return
		}
	}
	if roleValues, err = webSvc.WebRcmdRole(name, common.AuthWebRcmdAdmin, permsStr); err != nil {
		res["message"] = "获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(roleValues, nil)
}

// webRcmdAuth .
func appRcmdAuth(c *bm.Context) {
	var (
		permsStr   []string
		name       string
		roleValues *common.Role
		err        error
		perms      interface{}
		ok         bool
	)
	res := map[string]interface{}{}
	if usernameCtx, ok := c.Get("username"); ok {
		name = usernameCtx.(string)
	}
	if perms, ok = c.Get(middle.CtxPermissions); ok {
		if permsStr, ok = perms.([]string); !ok {
			c.JSON(permsStr, nil)
			return
		}
	}
	if roleValues, err = webSvc.WebRcmdRole(name, common.AuthAppRcmdAdmin, permsStr); err != nil {
		res["message"] = "获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(roleValues, nil)
}

func webRcmdCardList(c *bm.Context) {
	var (
		err   error
		pager *show.WebRcmdCardPager
	)
	res := map[string]interface{}{}
	req := &show.WebRcmdCardLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if pager, err = webSvc.WebRcmdCardList(req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

func addWebRcmdCard(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.WebRcmdCardAP{}
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
	if err = webSvc.AddWebRcmdCard(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func upWebRcmdCard(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.WebRcmdCardUP{}
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
	if err = webSvc.UpdateWebRcmdCard(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func delWebRcmdCard(c *bm.Context) {
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
	if err = webSvc.DeleteWebRcmdCard(req.ID, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func webRcmdList(c *bm.Context) {
	var (
		err   error
		pager *show.WebRcmdPager
	)
	res := map[string]interface{}{}
	req := &show.WebRcmdLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if req.ID, err = bvav.ToAvStr(req.ID); err != nil {
		res["message"] = "BVID 转换失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if pager, err = webSvc.WebRcmdList(c, req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	for k, v := range pager.Item {
		if v.Partition != "" {
			var typestr string
			typestr, err = commonSvc.ArcTypeString(v.Partition)
			if err != nil {
				pager.Item[k].PartitionNames = "列表获取失败 " + err.Error()
			} else {
				pager.Item[k].PartitionNames = typestr
			}
		}
		if v.Tag != "" {
			var tags map[int64]string
			tags, err = commonSvc.ArcTagString(v.Tag)
			if err != nil {
				pager.Item[k].TagNames = "列表获取失败 " + err.Error()
			} else {
				pager.Item[k].TagNames = tags
			}
		}
		if v.Avid != "" {
			if v.BvidRelate, err = bvav.ToBvStr(v.Avid); err != nil {
				log.Error("webRcmdList ToAvBvStr(%s) Error(%v)", v.Avid, err)
				err = nil
			}
		}
	}
	c.JSON(pager, nil)
}

func validate(p *Param) (err error) {
	var (
		id   int64
		card string
	)
	if id, err = strconv.ParseInt(p.CardValue, 10, 64); err != nil {
		return
	}
	if p.CardType == common.WebRcmdSpecial {
		card = common.CardWebRcmdSpecial
	} else if p.CardType == common.WebRcmdAV {
		card = common.CardAv
	} else if p.CardType == common.WebRcmdGame {
		card = common.CardWebRcmdGame
	}
	if _, _, err = commonSvc.CardPreview(ctx, card, id); err != nil {
		return
	}
	if len(p.Avids) > 0 {
		var (
			am *arcgrpc.ArcsReply
		)
		if am, err = commonSvc.Archives(p.Avids); err != nil {
			return
		}
		for _, aid := range p.Avids {
			if _, ok := am.Arcs[aid]; !ok {
				err = fmt.Errorf("错误类型（无效ID %v）", aid)
				return
			}
		}
	}
	if len(p.Tags) > 0 {
		var (
			tags map[int64]*taGrpcModel.Tag
		)
		if tags, err = commonSvc.TagGrpc(p.Tags); err != nil {
			if ecode.EqualError(xecode.TagNotExist, err) {
				err = fmt.Errorf("找不到tag数据")
			}
			return
		}
		for _, tid := range p.Tags {
			if _, ok := tags[tid]; !ok {
				err = fmt.Errorf("ID 为%d的tag信息找不到", tid)
				return
			}
		}
	}
	if len(p.Partitions) > 0 {
		partitions := commonSvc.ArcType
		for _, pid := range p.Partitions {
			if _, ok := partitions[pid]; !ok {
				err = fmt.Errorf("ID 为%d的分区信息找不到", pid)
				return
			}
		}
	}
	return
}

// Param validate param
type Param struct {
	Partitions  []int32 `form:"partition,split" validate:"required,dive,gt=0"`
	Tags        []int64 `form:"tag,split" validate:"required,dive,gt=0"`
	Avids       []int64
	CardType    int      `json:"card_type" form:"card_type" validate:"required"`
	CardValue   string   `json:"card_value" form:"card_value" validate:"required"`
	AvidsString []string `form:"avid,split" validate:"required,dive,gt=0"`
}

func addWebRcmd(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.WebRcmdAP{}
	if err = c.Bind(req); err != nil {
		return
	}
	p := &Param{}
	if err = c.Bind(p); err != nil {
		return
	}
	if req.CardValue, err = bvav.ToAvStr(req.CardValue); err != nil {
		res["message"] = "卡片创建失败,bvid转换失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	p.CardValue = req.CardValue
	if p.Avids, err = bvav.AvsStrToAvsIntSlice(p.AvidsString); err != nil {
		res["message"] = "卡片创建失败,bvid转换失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if req.Avid, err = bvav.ToAvsStr(req.Avid); err != nil {
		res["message"] = "卡片创建失败,bvid转换失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = validate(p); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	req.Person = name
	if err = webSvc.AddWebRcmd(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func upWebRcmd(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.WebRcmdUP{}
	if err = c.Bind(req); err != nil {
		return
	}
	p := &Param{}
	if err = c.Bind(p); err != nil {
		return
	}
	if req.CardValue, err = bvav.ToAvStr(req.CardValue); err != nil {
		res["message"] = "卡片创建失败,bvid转换失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	p.CardValue = req.CardValue
	if p.Avids, err = bvav.AvsStrToAvsIntSlice(p.AvidsString); err != nil {
		res["message"] = "卡片创建失败,bvid转换失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if req.Avid, err = bvav.ToAvsStr(req.Avid); err != nil {
		res["message"] = "卡片创建失败,bvid转换失败" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = validate(p); err != nil {
		res["message"] = "卡片创建失败: " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
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
	if err = webSvc.UpdateWebRcmd(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func delWebRcmd(c *bm.Context) {
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
	if err = webSvc.DeleteWebRcmd(req.ID, name, uid); err != nil {
		res["message"] = "卡片删除失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func optWebRcmd(c *bm.Context) {
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
	if err = webSvc.OptionWebRcmd(req.ID, req.Opt, name, uid, 0); err != nil {
		res["message"] = "修改失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func batchOptWebRcmd(c *bm.Context) {
	var (
		req = &struct {
			IDs []int64 `form:"ids,split" validate:"required,dive,gt=0"`
			Opt string  `form:"opt" validate:"required"`
		}{}
	)
	if err := c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	c.JSON(webSvc.BatchOptionWebRcmd(req.IDs, req.Opt, name, uid))
}
