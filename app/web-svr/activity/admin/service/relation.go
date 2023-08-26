package service

import (
	"context"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model"
	"go-gateway/app/web-svr/activity/interface/api"
	"strings"
)

const (
	followIdsLengthLimitation  = 200
	mallIdslengthLimitation    = 10
	topicIdslengthLimitation   = 10
	reserveIdslengthLimitation = 40
)

func (s *Service) ActRelationList(c context.Context, args *model.ActRelationListArgs) (listRes *model.ActRelationListRes, err error) {
	var (
		count int64
		list  []*model.ActRelationSubject
	)
	db := s.DB
	if args.Keyword != "" {
		names := "%" + args.Keyword + "%"
		db = db.Where("`id` = ? or `name` like ?", args.Keyword, names)
	}
	if err = db.Where("state=?", model.ActRelationSubjectStatusNormal).Offset((args.Page - 1) * args.PageSize).Limit(args.PageSize).Order("id desc").Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, " db.Model(&model.ActRelationSubject{}).Find() args(%+v) error(%v)", args, err)
		err = ecode.Error(ecode.RequestErr, "数据查询失败")
		return
	}
	if err = db.Model(&model.ActRelationSubject{}).Where("state=?", model.ActRelationSubjectStatusNormal).Count(&count).Error; err != nil {
		log.Errorc(c, "db.Model(&model.ActRelationSubject{}).Count() args(%v) error(%v)", args, err)
		err = ecode.Error(ecode.RequestErr, "数据查询失败")
		return
	}
	listRes = &model.ActRelationListRes{
		List:     list,
		Page:     args.Page,
		PageSize: args.PageSize,
		Count:    count,
	}
	return
}

func (s *Service) ActRelationGet(c context.Context, id int64) (res *model.ActRelationSubject, err error) {
	res = new(model.ActRelationSubject)
	if err = s.DB.Where("id=?", id).Find(&res).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, " db.Model(&model.ActRelationSubject{}).Find() id(%v) error(%v)", id, err)
		err = ecode.Error(ecode.RequestErr, "数据查询失败")
		return
	}
	if err == gorm.ErrRecordNotFound {
		err = ecode.Error(ecode.RequestErr, "id不存在")
		return
	}
	return
}

func (s *Service) ActRelationAdd(c context.Context, args *model.ActRelationSubject) (res int64, err error) {
	args.Name = strings.TrimSpace(args.Name)
	args.Description = strings.TrimSpace(args.Description)
	args.NativeIDs = strings.TrimSpace(args.NativeIDs)
	if strings.Contains(args.NativeIDs, " ") {
		err = ecode.Error(ecode.RequestErr, "NativeID存在空格，请检查")
		return
	}
	args.H5IDs = strings.TrimSpace(args.H5IDs)
	if strings.Contains(args.H5IDs, " ") {
		err = ecode.Error(ecode.RequestErr, "H5ID存在空格，请检查")
		return
	}
	args.WebIDs = strings.TrimSpace(args.WebIDs)
	if strings.Contains(args.WebIDs, " ") {
		err = ecode.Error(ecode.RequestErr, "WebID存在空格，请检查")
		return
	}
	args.LotteryIDs = strings.TrimSpace(args.LotteryIDs)
	if strings.Contains(args.LotteryIDs, " ") {
		err = ecode.Error(ecode.RequestErr, "抽奖配置存在空格，请检查")
		return
	}
	args.ReserveIDs = strings.TrimSpace(args.ReserveIDs)
	if strings.Contains(args.ReserveIDs, " ") {
		err = ecode.Error(ecode.RequestErr, "预约配置存在空格，请检查")
		return
	}
	if len(strings.Split(args.ReserveIDs, ",")) > reserveIdslengthLimitation {
		err = ecode.Error(ecode.RequestErr, "预约id数量超出限制(40)，请检查")
		return
	}
	args.VideoSourceIDs = strings.TrimSpace(args.VideoSourceIDs)
	if strings.Contains(args.VideoSourceIDs, " ") {
		err = ecode.Error(ecode.RequestErr, "视频源配置存在空格，请检查")
		return
	}
	args.FollowIDs = strings.TrimSpace(args.FollowIDs)
	if strings.Contains(args.FollowIDs, " ") {
		err = ecode.Error(ecode.RequestErr, "关注uid配置存在空格，请检查")
		return
	}
	if len(strings.Split(args.FollowIDs, ",")) > followIdsLengthLimitation {
		err = ecode.Error(ecode.RequestErr, "关注uid数量超出限制(200)，请检查")
		return
	}
	args.SeasonIDs = strings.TrimSpace(args.SeasonIDs)
	if strings.Contains(args.SeasonIDs, " ") {
		err = ecode.Error(ecode.RequestErr, "追剧ssid配置存在空格，请检查")
		return
	}
	args.MallIDs = strings.TrimSpace(args.MallIDs)
	if strings.Contains(args.MallIDs, " ") {
		err = ecode.Error(ecode.RequestErr, "会员购商品id配置存在空格，请检查")
		return
	}
	if len(strings.Split(args.MallIDs, ",")) > mallIdslengthLimitation {
		err = ecode.Error(ecode.RequestErr, "会员购id数量超出限制(10)，请检查")
		return
	}
	args.TopicIDs = strings.TrimSpace(args.TopicIDs)
	if strings.Contains(args.TopicIDs, " ") {
		err = ecode.Error(ecode.RequestErr, "话题订阅id配置存在空格，请检查")
		return
	}
	if len(strings.Split(args.TopicIDs, ",")) > topicIdslengthLimitation {
		err = ecode.Error(ecode.RequestErr, "话题订阅id数量超出限制(10)，请检查")
		return
	}
	args.FavoriteInfo = strings.TrimSpace(args.FavoriteInfo)
	if strings.Contains(args.FavoriteInfo, " ") {
		err = ecode.Error(ecode.RequestErr, "收藏配置存在空格，请检查")
		return
	}
	args.ReserveConfig = strings.TrimSpace(args.ReserveConfig)
	if strings.Contains(args.ReserveConfig, " ") {
		err = ecode.Error(ecode.RequestErr, "预约详细配置信息存在空格，请检查")
		return
	}
	args.FollowConfig = strings.TrimSpace(args.FollowConfig)
	if strings.Contains(args.FollowConfig, " ") {
		err = ecode.Error(ecode.RequestErr, "关注详细配置信息存在空格，请检查")
		return
	}
	args.SeasonConfig = strings.TrimSpace(args.SeasonConfig)
	if strings.Contains(args.SeasonConfig, " ") {
		err = ecode.Error(ecode.RequestErr, "追剧详细配置信息存在空格，请检查")
		return
	}
	args.FavoriteConfig = strings.TrimSpace(args.FavoriteConfig)
	if strings.Contains(args.FavoriteConfig, " ") {
		err = ecode.Error(ecode.RequestErr, "收藏详细配置信息存在空格，请检查")
		return
	}
	args.MallConfig = strings.TrimSpace(args.MallConfig)
	if strings.Contains(args.MallConfig, " ") {
		err = ecode.Error(ecode.RequestErr, "会员购商品详细配置存在空格，请检查")
		return
	}
	args.TopicConfig = strings.TrimSpace(args.TopicConfig)
	if strings.Contains(args.TopicConfig, " ") {
		err = ecode.Error(ecode.RequestErr, "活动订阅详细配置存在空格，请检查")
		return
	}

	if args.FavoriteInfo != "" {
		if err = validateFavorite(args.FavoriteInfo); err != nil {
			err = ecode.Error(ecode.RequestErr, "收藏相关规则配置错误")
			return
		}
	}

	if args.ReserveConfig != "" {
		if err = validateConfig(args.ReserveConfig); err != nil {
			err = ecode.Error(ecode.RequestErr, "预约相关规则配置错误")
			return
		}
	}

	if args.FollowConfig != "" {
		if err = validateConfig(args.FollowConfig); err != nil {
			err = ecode.Error(ecode.RequestErr, "关注相关规则配置错误")
			return
		}
	}

	if args.SeasonConfig != "" {
		if err = validateConfig(args.SeasonConfig); err != nil {
			err = ecode.Error(ecode.RequestErr, "追剧相关规则配置错误")
			return
		}
	}

	if args.FavoriteConfig != "" {
		if err = validateConfig(args.FavoriteConfig); err != nil {
			err = ecode.Error(ecode.RequestErr, "收藏相关规则配置错误")
			return
		}
	}

	if args.MallConfig != "" {
		if err = validateConfig(args.MallConfig); err != nil {
			err = ecode.Error(ecode.RequestErr, "会员购相关规则配置错误")
			return
		}
	}

	if args.TopicConfig != "" {
		if err = validateConfig(args.TopicConfig); err != nil {
			err = ecode.Error(ecode.RequestErr, "话题订阅相关规则配置错误")
			return
		}
	}

	if err = s.DB.Create(args).Error; err != nil {
		log.Errorc(c, " db.Model(&model.ActRelationAdd{}).Find() insert(%v) error(%v)", args, err)
		err = ecode.Error(ecode.RequestErr, "数据写入失败")
		return
	}

	res = args.ID

	// 线上数据增加 可能存在历史被胡乱访问造成的缓存，写缓存 具体值详情收拢到该地址 interface/model/like/like.go
	req := &api.InternalUpdateItemDataWithCacheReq{
		Typ:        1,
		ActionType: 1,
		Oid:        res,
	}
	if _, err = s.actClient.InternalUpdateItemDataWithCache(c, req); err != nil {
		log.Errorc(c, "InternalUpdateItemDataWithCache Req(%v) Err(%v)", req, err)
		err = ecode.Error(ecode.RequestErr, "数据更新成功，线上数据同步失败，请编辑该条目，重新提交")
		return
	}

	return
}

func validateConfig(str string) error {
	r := new(model.ActRelationConfigRule)
	return json.Unmarshal([]byte(str), r)
}

func validateFavorite(str string) error {
	r := new(model.ActRelationFavorite)
	return json.Unmarshal([]byte(str), r)
}

func (s *Service) ActRelationUpdate(c context.Context, id int64, args map[string]interface{}) (res int, err error) {
	actRelationSubject := new(model.ActRelationSubject)
	if err = s.DB.Where("id=?", id).Find(&actRelationSubject).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, " db.Model(&model.ActRelationSubject{}).Find() id(%v) error(%v)", id, err)
		err = ecode.Error(ecode.RequestErr, "数据查询失败")
		return
	}
	if err == gorm.ErrRecordNotFound {
		err = ecode.Error(ecode.RequestErr, "id不存在")
		return
	}

	update := make(map[string]interface{}, 0)
	if v, ok := args["name"]; ok {
		update["name"] = strings.TrimSpace(v.([]string)[0])
		if update["name"] == "" {
			err = ecode.Error(ecode.RequestErr, "活动名称不允许为空")
			return
		}
	}

	if v, ok := args["description"]; ok {
		update["description"] = strings.TrimSpace(v.([]string)[0])
	}

	if v, ok := args["native_ids"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "NativeID存在空格，请检查")
			return
		}
		update["native_ids"] = tmp
	}

	if v, ok := args["h5_ids"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "H5ID存在空格，请检查")
			return
		}
		update["h5_ids"] = tmp
	}

	if v, ok := args["web_ids"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "WebID存在空格，请检查")
			return
		}
		update["web_ids"] = tmp
	}

	if v, ok := args["lottery_ids"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "抽奖配置存在空格，请检查")
			return
		}
		update["lottery_ids"] = tmp
	}

	if v, ok := args["reserve_ids"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "预约配置存在空格，请检查")
			return
		}
		if len(strings.Split(tmp, ",")) > reserveIdslengthLimitation {
			err = ecode.Error(ecode.RequestErr, "预约id数量超出限制(40)，请检查")
			return
		}
		update["reserve_ids"] = tmp
	}

	if v, ok := args["video_source_ids"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "视频源配置存在空格，请检查")
			return
		}
		update["video_source_ids"] = tmp
	}

	if v, ok := args["follow_ids"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "关注uid配置存在空格，请检查")
			return
		}
		if len(strings.Split(tmp, ",")) > followIdsLengthLimitation {
			err = ecode.Error(ecode.RequestErr, "关注uid数量超出限制(200)，请检查")
			return
		}
		update["follow_ids"] = tmp
	}

	if v, ok := args["season_ids"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "追剧ssid配置存在空格，请检查")
			return
		}
		update["season_ids"] = tmp
	}

	if v, ok := args["mall_ids"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "会员购商品id配置存在空格，请检查")
			return
		}
		if len(strings.Split(tmp, ",")) > mallIdslengthLimitation {
			err = ecode.Error(ecode.RequestErr, "会员购商品id数量超出限制(10)，请检查")
			return
		}
		update["mall_ids"] = tmp
	}

	if v, ok := args["topic_ids"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "话题订阅id配置存在空格，请检查")
			return
		}
		if len(strings.Split(tmp, ",")) > topicIdslengthLimitation {
			err = ecode.Error(ecode.RequestErr, "话题订阅id数量超出限制(10)，请检查")
			return
		}
		update["topic_ids"] = tmp
	}

	if v, ok := args["favorite_info"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "收藏配置存在空格，请检查")
			return
		}
		update["favorite_info"] = tmp
	}

	if v, ok := args["reserve_config"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "预约详细配置信息存在空格，请检查")
			return
		}
		update["reserve_config"] = tmp
	}

	if v, ok := args["follow_config"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "关注详细配置信息存在空格，请检查")
			return
		}
		update["follow_config"] = tmp
	}

	if v, ok := args["season_config"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "追剧详细配置信息存在空格，请检查")
			return
		}
		update["season_config"] = tmp
	}

	if v, ok := args["favorite_config"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "收藏详细配置信息存在空格，请检查")
			return
		}
		update["favorite_config"] = tmp
	}

	if v, ok := args["mall_config"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "会员购商品详细配置信息存在空格，请检查")
			return
		}
		update["mall_config"] = tmp
	}

	if v, ok := args["topic_config"]; ok {
		tmp := strings.TrimSpace(v.([]string)[0])
		if strings.Contains(tmp, " ") {
			err = ecode.Error(ecode.RequestErr, "话题订阅详细配置信息存在空格，请检查")
			return
		}
		update["topic_config"] = tmp
	}

	if v, ok := update["favorite_info"]; ok {
		if v != "" {
			if err = validateFavorite(v.(string)); err != nil {
				err = ecode.Error(ecode.RequestErr, "收藏相关规则配置错误")
				return
			}
		}
	}

	if v, ok := update["reserve_config"]; ok {
		if v != "" {
			if err = validateConfig(v.(string)); err != nil {
				err = ecode.Error(ecode.RequestErr, "预约相关规则配置错误")
				return
			}
		}
	}

	if v, ok := update["follow_config"]; ok {
		if v != "" {
			if err = validateConfig(v.(string)); err != nil {
				err = ecode.Error(ecode.RequestErr, "关注相关规则配置错误")
				return
			}
		}
	}

	if v, ok := update["season_config"]; ok {
		if v != "" {
			if err = validateConfig(v.(string)); err != nil {
				err = ecode.Error(ecode.RequestErr, "追剧相关规则配置错误")
				return
			}
		}
	}

	if v, ok := update["favorite_config"]; ok {
		if v != "" {
			if err = validateConfig(v.(string)); err != nil {
				err = ecode.Error(ecode.RequestErr, "收藏相关规则配置错误")
				return
			}
		}
	}

	if v, ok := update["mall_config"]; ok {
		if v != "" {
			if err = validateConfig(v.(string)); err != nil {
				err = ecode.Error(ecode.RequestErr, "会员购相关规则配置错误")
				return
			}
		}
	}

	if v, ok := update["topic_config"]; ok {
		if v != "" {
			if err = validateConfig(v.(string)); err != nil {
				err = ecode.Error(ecode.RequestErr, "话题订阅相关规则配置错误")
				return
			}
		}
	}

	if err = s.DB.Model(&actRelationSubject).Where("id=?", id).Updates(update).Error; err != nil {
		log.Errorc(c, " db.Model(&model.ActRelationSubject{}).updates() id(%v) args(%v) error(%v)", id, args, err)
		err = ecode.Error(ecode.RequestErr, "数据更新失败")
		return
	}

	// 线上数据刷新 具体值详情收拢到该地址 interface/model/like/like.go
	req1 := &api.InternalUpdateItemDataWithCacheReq{
		Typ:        1,
		ActionType: 1,
		Oid:        id,
	}
	if _, err = s.actClient.InternalUpdateItemDataWithCache(c, req1); err != nil {
		log.Errorc(c, "InternalUpdateItemDataWithCache Req(%v )Err(%v)", req1, err)
		err = ecode.Error(ecode.RequestErr, "数据更新成功，线上数据同步失败，请编辑该条目，重新提交")
		return
	}

	// 刷新线上整体有效IDs
	req2 := &api.InternalSyncActRelationInfoDB2CacheReq{
		From: "op",
	}
	if _, err = s.actClient.InternalSyncActRelationInfoDB2Cache(c, req2); err != nil {
		log.Errorc(c, "InternalSyncActRelationInfoDB2Cache Req(%v )Err(%v)", req2, err)
		err = ecode.Error(ecode.RequestErr, "数据更新成功，线上数据同步失败，请编辑该条目，重新提交")
		return
	}

	return
}

func (s *Service) ActRelationState(c context.Context, id, state int64) (res int, err error) {
	if state != model.ActRelationSubjectStatusNormal && state != model.ActRelationSubjectStatusOffline {
		err = ecode.Error(ecode.RequestErr, "状态修改非法")
		return
	}
	actRelationSubject := new(model.ActRelationSubject)
	if err = s.DB.Where("id=?", id).Find(&actRelationSubject).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, " db.Model(&model.ActRelationSubject{}).Find() id(%v) error(%v)", id, err)
		err = ecode.Error(ecode.RequestErr, "数据查询失败")
		return
	}
	if err == gorm.ErrRecordNotFound {
		err = ecode.Error(ecode.RequestErr, "id不存在")
		return
	}
	doState := map[string]interface{}{
		"state": state,
	}
	if err = s.DB.Model(&actRelationSubject).Where("id=?", id).Select("state").Update(doState).Error; err != nil {
		log.Errorc(c, " db.Model(&model.ActRelationSubject{}).update() id(%v) state(%v) error(%v)", id, state, err)
		err = ecode.Error(ecode.RequestErr, "数据更新失败")
		return
	}

	// 防止主从不一致，线上缓存状态置为-1
	req := &api.InternalUpdateItemDataWithCacheReq{
		Typ:        1,
		ActionType: 3,
		Oid:        id,
	}
	if _, err = s.actClient.InternalUpdateItemDataWithCache(c, req); err != nil {
		log.Errorc(c, "InternalUpdateItemDataWithCache Req(%v) Err(%v)", req, err)
		err = ecode.Error(ecode.RequestErr, "数据删除成功，同步线上失败，请联系管理员")
		return
	}

	return
}
