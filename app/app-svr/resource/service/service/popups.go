package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
	"strconv"
)

func checkBuilds(req *pb.PopUpsReq, builds string) (valid bool, err error) {
	var (
		builds_limit []*model.BuildLimit
	)
	if err = json.Unmarshal([]byte(builds), &builds_limit); err != nil {
		log.Error("resource.PopUps.checkBuilds (%v) error(%v)", builds, err)
		return false, err
	}
	for _, limit := range builds_limit {
		if limit.Plat != req.Plat {
			continue
		}
		switch limit.Conditions {
		case "gt":
			{
				valid = req.Build > limit.Build
				break
			}
		case "lt":
			{
				valid = req.Build < limit.Build
				break
			}
		case "eq":
			{
				valid = req.Build == limit.Build
				break
			}
		case "ne":
			{
				valid = req.Build != limit.Build
				break
			}
		default:
			valid = false
		}
		if valid {
			break
		}
	}
	return
}

func (s *Service) PopUps(c context.Context, req *pb.PopUpsReq) (reply *pb.PopUpsReply, err error) {
	var (
		valid      bool
		is_poped   bool
		popupItems []*model.PopUps
	)
	popupItems, err = s.popups.GetEffectivePopUps(c)
	if err != nil {
		log.Error("resource.PopUps GetEffectivePopUps(%v) error(%v)", popupItems, err)
		return nil, err
	}

	for _, item := range popupItems {
		// 判断版本
		if valid, err = checkBuilds(req, item.Builds); err != nil {
			log.Error("resource.PopUps checkBuilds(%v) error(%v)", item.Builds, err)
			return
		}
		if !valid {
			continue
		}
		log.Info("Current PopUps(%+v), valid(%+v)", item, valid)
		// 判断人群包
		if item.CrowdType != -1 {
			var crowd_value int64
			if crowd_value, err = strconv.ParseInt(item.CrowdValue, 10, 64); err != nil {
				log.Error("resource.PopUps convert crowd_value(%+v) error(%+v)", item.CrowdValue, err)
				return
			}
			if valid, err = s.popups.CheckCrowd(req, item.CrowdBase, crowd_value); err != nil {
				log.Error("resource.PopUps checkCrowd(%v) error(%v)", req, err)
				return
			}
			log.Info("After checkCrowd PopUps(%+v), valid(%+v)", item, valid)
		}
		if !valid {
			continue
		}
		// 判断是否弹窗过
		if is_poped, err = s.popups.GetIsPopFromTaishan(c, model.PopUpsKey(item.ID, req.Mid, req.Buvid)); err != nil {
			log.Error("resource.PopUps GetIsPopFromTaishan req(%+v) error(%+v)", model.PopUpsKey(item.ID, req.Mid, req.Buvid), err)
			return
		}
		log.Info("After GetIsPopFromTaishan PopUps(%+v), valid(%+v), is_poped(%+v)", item, valid, is_poped)

		if is_poped {
			return &pb.PopUpsReply{IsPoped: true}, nil
		}

		reply = &pb.PopUpsReply{
			Id:             item.ID,
			Pic:            item.Pic,
			PicIpad:        item.PicIpad,
			Description:    item.Description,
			LinkType:       item.LinkType,
			Link:           item.Link,
			TeenagePush:    int32(item.TeenagePush),
			AutoHideStatus: int32(item.AutoHideStatus),
			CloseTime:      item.CloseTime,
		}
		is_poped = true
		val, err := json.Marshal(is_poped)
		if err != nil {
			log.Error("popups.Marshal is_poped(%+v) error(%+v)", is_poped, err)
			return nil, err
		}
		if err = s.popups.PutReq([]byte(model.PopUpsKey(item.ID, req.Mid, req.Buvid)), val, 0); err != nil {
			log.Error("popups.PutReq(%+v) error(%+v)", val, err)
			return nil, err
		}
		log.Error("item(%+v) is_poped(%+v)", item, is_poped)
	}
	if !valid {
		err = fmt.Errorf("无有效配置")
		return nil, err
	}
	return reply, nil
}
