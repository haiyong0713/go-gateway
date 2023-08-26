package http

import (
	"strconv"
	"strings"
	"time"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"
	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"
	taGrpcModel "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/bvav"
	"go-gateway/app/app-svr/app-feed/admin/model/article"
	"go-gateway/app/app-svr/app-feed/admin/model/comic"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/game"
	"go-gateway/app/app-svr/app-feed/admin/model/live"
	"go-gateway/app/app-svr/app-feed/admin/util"
	"go-gateway/pkg/idsafe/bvid"
)

func managerInfo(c *bm.Context) (uid int64, username string) {
	if nameInter, ok := c.Get("username"); ok {
		username = nameInter.(string)
	}
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if username == "" {
		cookie, err := c.Request.Cookie("username")
		if err != nil {
			log.Error("managerInfo get cookie error (%v)", err)
			return
		}
		username = cookie.Value
		c, err := c.Request.Cookie("uid")
		if err != nil {
			log.Error("managerInfo get cookie error (%v)", err)
			return
		}
		uidInt, _ := strconv.Atoi(c.Value)
		uid = int64(uidInt)
	}
	return
}

func cardPreview(c *bm.Context) {
	var (
		err   error
		title string
		res   = map[string]interface{}{}
		raw   interface{}
		id    int64
	)
	type Card struct {
		Type string `form:"type" validate:"required"`
		ID   string `form:"id" validate:"required"`
	}
	card := &Card{}
	if err = c.Bind(card); err != nil {
		return
	}
	// todo: 为什么？？？
	if strings.HasPrefix(card.ID, "bv") {
		res["message"] = "bv id 需要全部大写"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if strings.HasPrefix(card.ID, "BV") {
		if id, err = bvid.BvToAv(card.ID); err != nil {
			res["message"] = err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	} else {
		if id, err = strconv.ParseInt(card.ID, 10, 64); err != nil {
			res["message"] = err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	if title, raw, err = commonSvc.CardPreview(c, card.Type, id); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	titleReturn := common.CardPreview{
		Title: title,
		Raw:   raw,
	}
	c.JSON(titleReturn, nil)
}

func cardPreviewBatch(c *bm.Context) {
	var (
		err error
		ids []string
		res = map[string]interface{}{}
	)
	type Req struct {
		Type string `form:"type" validate:"required"`
		Ids  string `form:"ids" validate:"required"`
	}
	req := &Req{}
	if err = c.Bind(req); err != nil {
		return
	}
	if ids = strings.Split(req.Ids, ","); len(ids) == 0 {
		res["message"] = "无效id列表"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	for _, id := range ids {
		if strings.HasPrefix(id, "bv") {
			res["message"] = "bv id 需要全部大写"
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	c.JSON(commonSvc.CardPreviewBatch(c, req.Type, ids))
}

func actionLog(c *bm.Context) {
	var (
		res = map[string]interface{}{}
	)
	param := &common.Log{}
	if err := c.Bind(param); err != nil {
		return
	}
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	if param.Starttime == "" {
		param.Starttime = firstOfMonth.Format("2006-01-02")
	}
	if param.Endtime == "" {
		param.Endtime = lastOfMonth.Format("2006-01-02")
	}
	param.Starttime = param.Starttime + " 00:00:00"
	param.Endtime = param.Endtime + " 23:59:59"
	searchRes, err := commonSvc.LogAction(c, param)
	if err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	res["data"] = searchRes.Item
	res["pager"] = searchRes.Page
	c.JSONMap(res, nil)
}

// actionAddLog add action log
func actionAddLog(c *bm.Context) {
	var (
		err error
	)
	type Log struct {
		Type   int    `form:"type" validate:"required"`
		Uame   string `form:"uname"`
		UID    int64  `form:"uid"`
		Action string `form:"action"`
		Param  string `form:"param"`
	}
	l := &Log{}
	if err = c.Bind(l); err != nil {
		return
	}
	if err = util.AddLogs(l.Type, l.Uame, l.UID, 0, l.Action, l.Param); err != nil {
		log.Error("actionAddLog error(%v)", err)
		return
	}
	c.JSON(nil, nil)
}

func cardType(c *bm.Context) {
	var (
		res = map[string]interface{}{}
	)
	res["data"] = commonSvc.CardType()
	c.JSONMap(res, nil)
}

func archiveType(c *bm.Context) {
	var (
		data map[int32]*arcmdl.Tp
		err  error
		res  = map[string]interface{}{}
	)
	type Archive struct {
		IDs []int32 `form:"ids,split" validate:"required,dive,gt=0"`
	}
	arc := &Archive{}
	if err = c.Bind(arc); err != nil {
		return
	}
	if data, err = commonSvc.ArchivesType(arc.IDs); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	res["data"] = data
	c.JSONMap(res, nil)
}

func tagType(c *bm.Context) {
	var (
		data map[int64]*taGrpcModel.Tag
		err  error
		res  = map[string]interface{}{}
	)
	type Tag struct {
		IDs []int64 `form:"ids,split" validate:"required,dive,gt=0"`
	}
	tag := &Tag{}
	if err = c.Bind(tag); err != nil {
		return
	}
	if data, err = commonSvc.TagGrpc(tag.IDs); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	res["data"] = data
	c.JSONMap(res, nil)
}

func archives(c *bm.Context) {
	var (
		err  error
		res  = map[string]interface{}{}
		arcs *arcgrpc.ArcsReply
		ids  []int64
	)
	type Arc struct {
		IDs []string `form:"ids,split" validate:"required,dive,gt=0"`
	}
	arc := &Arc{}
	if err = c.Bind(arc); err != nil {
		return
	}
	if ids, err = bvav.AvsStrToAvsIntSlice(arc.IDs); err != nil {
		res["message"] = "BVID 转换失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if arcs, err = commonSvc.Archives(ids); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	resMap := map[string]interface{}{}
	for _, v := range arcs.Arcs {
		resMap[strconv.Itoa(int(v.Aid))] = v
		curBvid, _ := bvid.AvToBv(v.Aid)
		resMap[curBvid] = v
	}
	res["data"] = resMap
	c.JSONMap(res, nil)
}

func notify(c *bm.Context) {
	var (
		res        = map[string]interface{}{}
		err        error
		archives   *arcgrpc.ArcsReply
		msg, title string
		mid        int64
		room       map[int64]*live.Room
		article    *article.Article
	)
	type Param struct {
		Busness int64 `form:"busness" validate:"required"`
		Type    int64 `form:"type" validate:"required"`
		ID      int64 `form:"id" validate:"required"`
	}
	p := &Param{}
	if err = c.Bind(p); err != nil {
		return
	}
	//天马的业务通知
	if p.Busness == common.NotifyBusnessTianma {
		//稿件
		if p.Type == common.NotifyTypArchive {
			if archives, err = commonSvc.Archives([]int64{p.ID}); err != nil {
				res["message"] = err.Error()
				c.JSONMap(res, ecode.RequestErr)
				return
			}
			if _, ok := archives.Arcs[p.ID]; !ok {
				res["message"] = "无效稿件ID"
				c.JSONMap(res, ecode.RequestErr)
				return
			}
			title = common.NotifyTitleArchive
			msg = archives.Arcs[p.ID].Title
			mid = archives.Arcs[p.ID].Author.Mid
		} else if p.Type == common.NotifyTypLive {
			if room, err = commonSvc.LiveRooms(c, []int64{p.ID}); err != nil {
				res["message"] = err.Error()
				c.JSONMap(res, ecode.RequestErr)
				return
			}
			if _, ok := room[p.ID]; !ok {
				res["message"] = "无效直播ID"
				c.JSONMap(res, ecode.RequestErr)
				return
			}
			title = common.NotifyTitleLive
			msg = room[p.ID].Title
			mid = room[p.ID].UID
		} else if p.Type == common.NotifyTypArticle {
			if article, err = commonSvc.Article(c, []int64{p.ID}); err != nil {
				res["message"] = err.Error()
				c.JSONMap(res, ecode.RequestErr)
				return
			}
			title = common.NotifyTitleArticle
			msg = article.Data.Title
			mid = article.Data.Author.Mid
		} else {
			log.Error("notify param(%+v)error", p)
			return
		}
		if err = commonSvc.Notify([]int64{mid}, p.Busness, title, msg); err != nil {
			res["message"] = err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
	}
	c.JSON(nil, nil)
}

func upInfo(c *bm.Context) {
	var (
		err     error
		res     = map[string]interface{}{}
		accCard *accgrpc.Card
	)
	type Arc struct {
		ID int64 `form:"id" validate:"required,gt=0"`
	}
	arc := &Arc{}
	if err = c.Bind(arc); err != nil {
		return
	}
	if accCard, err = commonSvc.UpInfo(c, arc.ID); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	res["data"] = accCard
	c.JSONMap(res, nil)
}

func comicInfo(c *bm.Context) {
	var (
		err  error
		res  = map[string]interface{}{}
		data []*comic.ComicInfo
	)
	type Card struct {
		ID int64 `form:"id" validate:"required"`
	}
	card := &Card{}
	if err = c.Bind(card); err != nil {
		return
	}
	if data, err = commonSvc.Comic(c, card.ID); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(data, nil)
}

// gameInfo .
func gameInfo(c *bm.Context) {
	var (
		err  error
		res  = map[string]interface{}{}
		data *game.Info
	)
	type Card struct {
		ID int64 `form:"id" validate:"required"`
	}
	card := &Card{}
	if err = c.Bind(card); err != nil {
		return
	}
	if data, err = commonSvc.AppGameInfo(c, card.ID); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(data, nil)
}
