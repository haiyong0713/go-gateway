package http

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func eventTopicList(c *bm.Context) {
	var (
		err   error
		pager *show.EventTopicPager
	)
	res := map[string]interface{}{}
	req := &show.EventTopicLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if pager, err = popularSvc.EventTopicList(req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

func addEventTopic(c *bm.Context) {
	var (
		err error
		//title string
	)
	res := map[string]interface{}{}
	req := &show.EventTopicAP{}
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
	if err = popularSvc.AddEventTopic(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func upEventTopic(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.EventTopicUP{}
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
	if err = popularSvc.UpdateEventTopic(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func delEventTopic(c *bm.Context) {
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
	if err = popularSvc.DeleteEventTopic(req.ID, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}
func popularStarsList(c *bm.Context) {
	var (
		err   error
		pager *show.PopularStarsPager
	)
	res := map[string]interface{}{}
	req := &show.PopularStarsLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	if pager, err = popularSvc.PopularStarsList(req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(pager, nil)
}

// ValidateUpid .
func ValidateUpid(c context.Context, idStr string) (name string, err error) {
	var (
		id int64
	)
	if id, err = strconv.ParseInt(idStr, 10, 64); err != nil {
		return
	}
	if name, _, err = commonSvc.CardPreview(c, common.CardUp, id); err != nil {
		return
	}
	return
}

// ValidateAvid .
func ValidateAvid(c context.Context, values string) (res string, err error) {
	type Content struct {
		ID    interface{} `json:"id"`
		Title string      `json:"title"`
	}
	type ContentAid struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}
	var contents []*Content
	var contentAids []*ContentAid
	if err = json.Unmarshal([]byte(values), &contents); err == nil {
		dup := make(map[int64]bool)
		for _, v := range contents {
			if v == nil {
				return "", fmt.Errorf("不能提交空数据！")
			}
			var aid int64
			switch v.ID.(type) {
			case float64:
				aid = int64(v.ID.(float64))
			case string:
				if aid, err = common.GetAvID(v.ID.(string)); err != nil {
					return "", fmt.Errorf("视频ID非法！")
				}
			default:
				return "", fmt.Errorf("视频ID非法！")
			}
			if dup[aid] {
				return "", fmt.Errorf("重复视频ID (%d)", aid)
			}
			dup[aid] = true
			if _, _, err = commonSvc.CardPreview(c, common.CardAv, aid); err != nil {
				return
			}
			contentAids = append(contentAids, &ContentAid{
				ID:    aid,
				Title: v.Title,
			})
		}
		//nolint:gomnd
		if len(contents) < 3 {
			if len(contents) != 1 {
				return "", fmt.Errorf("单视频模式视频组成不能超过1个，多视频模式视频组成不能少于3个")
			}
		}
		//nolint:gomnd
		if len(contents) > 5 {
			return "", fmt.Errorf("多视频模式视频组成不能超过5个")
		}
		b, _ := json.Marshal(contentAids)
		res = string(b)
		return
	} else {
		err = nil
	}
	return
}

func addPopularStars(c *bm.Context) {
	var (
		err    error
		upName string
	)
	res := map[string]interface{}{}
	req := &show.PopularStarsAP{}
	if err = c.Bind(req); err != nil {
		return
	}
	uid, name := util.UserInfo(c)
	if name == "" {
		c.JSONMap(map[string]interface{}{"message": "请重新登录"}, ecode.Unauthorized)
		c.Abort()
		return
	}
	if req.Content, err = ValidateAvid(c, req.Content); err != nil {
		c.JSONMap(map[string]interface{}{"message": "视频ID 校验失败：" + err.Error()}, ecode.RequestErr)
		c.Abort()
		return
	}
	req.Value = util.TrimStrSpace(req.Value)
	if upName, err = ValidateUpid(c, req.Value); err != nil {
		c.JSONMap(map[string]interface{}{"message": "up主ID 校验失败：" + err.Error()}, ecode.RequestErr)
		c.Abort()
		return
	}
	req.Person = name
	req.UID = uid
	req.LongTitle = upName
	if err = popularSvc.AddPopularStars(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func updatePopularStars(c *bm.Context) {
	var (
		err    error
		upName string
	)
	res := map[string]interface{}{}
	req := &show.PopularStarsUP{}
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
	if req.Content, err = ValidateAvid(c, req.Content); err != nil {
		c.JSONMap(map[string]interface{}{"message": err.Error()}, ecode.RequestErr)
		c.Abort()
		return
	}
	req.Value = util.TrimStrSpace(req.Value)
	if upName, err = ValidateUpid(c, req.Value); err != nil {
		c.JSONMap(map[string]interface{}{"message": err.Error()}, ecode.RequestErr)
		c.Abort()
		return
	}
	req.LongTitle = upName
	if err = popularSvc.UpdatePopularStars(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func deletePopularStars(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &struct {
		ID   int64  `form:"id" validate:"required"`
		Type string `form:"type" validate:"required"`
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
	if err = popularSvc.DeletePopularStars(req.ID, req.Type, name, uid); err != nil {
		res["message"] = "卡片删除失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func rejectPopularStars(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &struct {
		ID   int64  `form:"id" validate:"required"`
		Type string `form:"type" validate:"required"`
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
	if err = popularSvc.RejectPopularStars(req.ID, req.Type, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func aiAddPopularStars(c *bm.Context) {
	var (
		err        error
		aids       []byte
		addPopStar []*show.PopularStarsAP
	)
	res := map[string]interface{}{}
	req := &struct {
		Data string `form:"data" validate:"required"`
	}{}
	if err = c.Bind(req); err != nil {
		return
	}
	log.Info("aiAddPopularStars value(%v)", req.Data)
	values := make([]*show.PopularStarsAIAP, 0)
	if err = json.Unmarshal([]byte(req.Data), &values); err != nil {
		log.Error("aiAddPopularStars.Unmarshal value(%v) error(%v)", req.Data, err)
		res["message"] = "数据解析失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	for _, v := range values {
		var (
			AiValues []*show.AiValue
			upName   string
		)
		for _, aid := range v.Aids {
			aiValue := &show.AiValue{
				ID: aid,
			}
			AiValues = append(AiValues, aiValue)
		}
		if aids, err = json.Marshal(AiValues); err != nil {
			log.Error("aiAddPopularStars.Marshal value(%v) error(%v)", v.Aids, err)
			res["message"] = "数据encode失败 " + err.Error()
			c.JSONMap(res, ecode.RequestErr)
		}
		mid := strconv.FormatInt(v.Mid, 10)
		if upName, err = ValidateUpid(c, mid); err != nil {
			c.JSONMap(map[string]interface{}{"message": "up主ID 校验失败：" + err.Error()}, ecode.RequestErr)
			c.Abort()
			return
		}

		tmp := &show.PopularStarsAP{
			LongTitle: upName,
			Value:     strconv.FormatInt(v.Mid, 10),
			Content:   string(aids),
		}
		addPopStar = append(addPopStar, tmp)
	}
	if err = popularSvc.AIAddPopularStars(c, addPopStar); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func popRecommendList(c *bm.Context) {
	var (
		err   error
		pager *show.PopRecommendPager
	)
	res := map[string]interface{}{}
	req := &show.PopRecommendLP{}
	if err = c.Bind(req); err != nil {
		return
	}
	req.AID, _ = common.GetAvID(req.ID)
	if pager, err = popularSvc.PopRecommendList(req); err != nil {
		res["message"] = "列表获取失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	for k, v := range pager.Item {
		var (
			id    int64
			title string
			bvid  string
		)
		id, err = strconv.ParseInt(v.CardValue, 10, 64)
		if err != nil {
			res["message"] = err.Error()
			c.JSONMap(res, ecode.RequestErr)
			return
		}
		if bvid, _ = common.GetBvID(id); bvid != "" {
			pager.Item[k].Bvid = bvid
		}
		if title, _, err = commonSvc.CardPreview(c, common.CardAv, id); err != nil {
			pager.Item[k].Title = fmt.Sprintf("id为%d错误:", id) + err.Error()
			continue
		}
		pager.Item[k].Title = title
	}
	c.JSON(pager, nil)
}

func validateID(c context.Context, idStr string) (err error) {
	var (
		id int64
	)
	id, err = common.GetAvID(idStr)
	if err != nil {
		return
	}
	if _, _, err = commonSvc.CardPreview(c, common.CardAv, id); err != nil {
		return
	}
	return
}

func addPopRecommend(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.PopRecommendAP{}
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
	if err = validateID(c, req.CardValue); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if req.CardValue, err = common.GetAvIDStr(req.CardValue); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = popularSvc.AddPopRecommend(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func upPopRecommend(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	req := &show.PopRecommendUP{}
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
	if err = validateID(c, req.CardValue); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if req.CardValue, err = common.GetAvIDStr(req.CardValue); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = popularSvc.UpdatePopRecommend(c, req, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func delPopRecommend(c *bm.Context) {
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
	if err = popularSvc.DeletePopRecommend(req.ID, name, uid); err != nil {
		res["message"] = "卡片创建失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}
