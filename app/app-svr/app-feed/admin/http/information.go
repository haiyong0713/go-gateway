package http

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// recommendCardList
func recommendCardList(c *bm.Context) {
	var (
		err  error
		list *show.RecommendCardList
		avid int64
	)
	res := map[string]interface{}{}
	req := &show.RecommendCardListReq{}
	if err = c.Bind(req); err != nil {
		return
	}
	// 时间段参数验证
	if req.Stime > 0 && req.Etime > 0 && req.Stime >= req.Etime {
		res["message"] = "展示起止时间校验失败"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	// 转avid
	if avid, err = req.AvIDVal(); err != nil {
		res["message"] = "卡片ID校验失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	} else {
		req.AvID = avid
	}

	if list, err = infoSvc.RecommendCardList(c, req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	for _, item := range list.List {
		var (
			id    int64
			title string
		)
		id, err = strconv.ParseInt(item.CardID, 10, 64)
		if err != nil {
			res["message"] = err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		item.AvID = id
		if item.CardType == show.CardTypeAv {
			if item.BvID, err = common.GetBvID(item.AvID); err != nil {
				err = nil
				continue
			}
		}
		if title, _, err = commonSvc.CardPreview(c, show.CardPreviewType[item.CardType], id); err != nil {
			item.CardTitle = fmt.Sprintf("id为%d错误:", id) + err.Error()
			continue
		}
		item.CardTitle = title

	}
	c.JSON(list, nil)
}

// addRecommendCard
func addRecommendCard(c *bm.Context) {
	var (
		err     error
		cardId  string
		avid    int64
		overlap bool
	)
	res := map[string]interface{}{}
	req := &show.RecommendCardAddReq{}
	if err = c.Bind(req); err != nil {
		return
	}
	// 操作者信息
	req.Uid, req.Uname = util.UserInfo(c)
	if req.Uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	// 展示起止时间验证
	if req.Etime <= req.Stime {
		c.JSONMap(map[string]interface{}{"message": "卡片展示结束时间需大于开始时间"}, ecode.RequestErr)
		c.Abort()
		return
	}
	if req.Stime.Time().Unix() < time.Now().Unix() {
		c.JSONMap(map[string]interface{}{"message": "卡片展示生效时间需大于当前时间"}, ecode.RequestErr)
		c.Abort()
		return
	}

	// 卡片id验证
	if cardId, avid, err = validateCardId(c, req.CardType, req.CardID); err != nil {
		c.JSONMap(map[string]interface{}{"message": "卡片ID 校验失败：" + err.Error()}, ecode.RequestErr)
		c.Abort()
		return
	}
	req.CardID = cardId
	req.AvID = avid
	req.ApplyReason = util.TrimStrSpace(req.ApplyReason)

	// 视频段overlap验证
	checkParams := &show.RecommendCardIntervalCheckReq{
		ID:       0,
		CardPos:  req.CardPos,
		PosIndex: req.PosIndex,
		Stime:    req.Stime,
		Etime:    req.Etime,
	}
	if overlap, err = infoSvc.IntervalCheckRecommendCard(c, checkParams); err != nil {
		return
	}
	if overlap {
		c.JSONMap(map[string]interface{}{"message": "该位置已有运营卡片，请检查后重新提交"}, ecode.RequestErr)
		c.Abort()
		return
	}

	if err = infoSvc.AddRecommendCard(c, req); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// modifyRecommendCard
func modifyRecommendCard(c *bm.Context) {
	var (
		err     error
		cardId  string
		avid    int64
		overlap bool
	)
	res := map[string]interface{}{}
	req := &show.RecommendCardModifyReq{}
	if err = c.Bind(req); err != nil {
		return
	}
	// 操作者信息
	req.Uid, req.Uname = util.UserInfo(c)
	if req.Uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	// 展示起止时间验证
	if req.Etime <= req.Stime {
		c.JSONMap(map[string]interface{}{"message": "卡片展示结束时间需大于开始时间"}, ecode.RequestErr)
		c.Abort()
		return
	}
	if req.Stime.Time().Unix() < time.Now().Unix() {
		c.JSONMap(map[string]interface{}{"message": "卡片展示生效时间需大于当前时间"}, ecode.RequestErr)
		c.Abort()
		return
	}
	// 卡片id验证
	if cardId, avid, err = validateCardId(c, req.CardType, req.CardID); err != nil {
		c.JSONMap(map[string]interface{}{"message": "卡片ID 校验失败：" + err.Error()}, ecode.RequestErr)
		c.Abort()
		return
	}
	req.CardID = cardId
	req.AvID = avid
	req.ApplyReason = util.TrimStrSpace(req.ApplyReason)

	// 生效时间段overlap验证
	checkParams := &show.RecommendCardIntervalCheckReq{
		ID:       req.ID,
		CardPos:  req.CardPos,
		PosIndex: req.PosIndex,
		Stime:    req.Stime,
		Etime:    req.Etime,
	}
	if overlap, err = infoSvc.IntervalCheckRecommendCard(c, checkParams); err != nil {
		return
	}
	if overlap {
		c.JSONMap(map[string]interface{}{"message": "该位置已有运营卡片，请检查后重新提交"}, ecode.RequestErr)
		c.Abort()
		return
	}

	if err = infoSvc.ModifyRecommendCard(c, req); err != nil {
		res["message"] = "卡片修改失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// deleteRecommendCard
func deleteRecommendCard(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.RecommendCardOpReq{}
	if err = c.Bind(req); err != nil {
		return
	}
	// 操作者信息
	req.Uid, req.Uname = util.UserInfo(c)
	if req.Uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if err = infoSvc.DeleteRecommendCard(c, req); err != nil {
		res["message"] = "卡片删除失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// offlineRecommendCard
func offlineRecommendCard(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.RecommendCardOpReq{}
	if err = c.Bind(req); err != nil {
		return
	}
	// 操作者信息
	req.Uid, req.Uname = util.UserInfo(c)
	if req.Uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if err = infoSvc.OfflineRecommendCard(c, req); err != nil {
		res["message"] = "卡片下线失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// passRecommendCard
func passRecommendCard(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.RecommendCardOpReq{}
	if err = c.Bind(req); err != nil {
		return
	}
	// 操作者信息
	req.Uid, req.Uname = util.UserInfo(c)
	if req.Uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	req.Op = show.OpPass
	if err = infoSvc.PassRecommendCard(c, req); err != nil {
		res["message"] = "卡片通过失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// rejectRecommendCard
func rejectRecommendCard(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.RecommendCardOpReq{}
	if err = c.Bind(req); err != nil {
		return
	}
	// 操作者信息
	req.Uid, req.Uname = util.UserInfo(c)
	if req.Uname == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	req.Op = show.OpReject
	if err = infoSvc.RejectRecommendCard(c, req); err != nil {
		res["message"] = "卡片拒绝失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// validateCardId
func validateCardId(c context.Context, cardType int, cardId string) (res string, id int64, err error) {
	cardId = strings.TrimSpace(cardId)
	if cardId == "" {
		return "", 0, ecode.Error(ecode.RequestErr, "不能提交空数据")
	}
	switch cardType {
	// 视频
	case show.CardTypeAv:
		res, id, err = validateAvid(c, cardId)
	// 动态
	case show.CardTypeDynamic:
		res, id, err = validateDynamicId(c, cardId)
	// 专栏
	case show.CardTypeArticle:
		res, id, err = validateArticleId(c, cardId)
	default:
		err = ecode.Error(ecode.RequestErr, "非有效卡片类型")
	}

	return
}

// ValidateAvid .
func validateAvid(c context.Context, value string) (res string, avid int64, err error) {
	if avid, err = common.GetAvID(value); err != nil {
		return "", 0, ecode.Error(ecode.RequestErr, "视频ID非法")
	}
	log.Info("avid:%d", avid)
	if _, _, err = commonSvc.CardPreview(c, common.CardAv, avid); err != nil {
		return "", 0, ecode.Error(ecode.RequestErr, "获取视频名失败")
	}
	res = strconv.FormatInt(avid, 10)
	return
}

func validateDynamicId(c context.Context, value string) (res string, id int64, err error) {
	if id, err = strconv.ParseInt(value, 10, 64); err != nil {
		return "", 0, ecode.Error(ecode.RequestErr, "动态ID非法")
	}
	if _, _, err = commonSvc.CardPreview(c, common.CardDynamic, id); err != nil {
		return "", 0, ecode.Error(ecode.RequestErr, "获取动态名失败")
	}
	res = strconv.FormatInt(id, 10)
	return
}

func validateArticleId(c context.Context, value string) (res string, id int64, err error) {
	if id, err = strconv.ParseInt(value, 10, 64); err != nil {
		return "", 0, ecode.Error(ecode.RequestErr, "专栏ID非法")
	}
	if _, _, err = commonSvc.CardPreview(c, common.CardArticle, id); err != nil {
		return "", 0, ecode.Error(ecode.RequestErr, "获取专栏名失败")
	}
	res = strconv.FormatInt(id, 10)
	return
}
