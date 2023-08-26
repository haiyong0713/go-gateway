package vogue

import (
	"context"

	accountAPI "git.bilibili.co/bapis/bapis-go/account/service"
	"go-gateway/app/web-svr/activity/ecode"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"

	"go-common/library/log"
)

func (s *Service) ShareInfo(c context.Context, token string) (res *model.ShareInfo, err error) {
	var (
		mid     int64
		account *accountAPI.InfoReply
		task    *model.Task
		goods   *model.Goods
	)
	if mid, err = model.TokenEncode(token); err != nil {
		err = ecode.ActivityTokenError
		return nil, err
	}
	if account, err = s.accClient.Info3(c, &accountAPI.MidReq{Mid: mid}); err != nil {
		log.Error("s.accClient.Info3(%v) error(%v)", account, err)
		return nil, err
	}
	if task, err = s.dao.Task(c, mid); err != nil {
		log.Error("s.dao.Task(%v) error(%v)", mid, err)
		return nil, err
	}
	if task == nil {
		err = ecode.ActivityTokenError
		return nil, err
	}
	if goods, err = s.dao.Goods(c, task.Goods); err != nil {
		log.Error("s.dao.Goods(%v) error(%v)", task.Goods, err)
		return nil, err
	}
	res = &model.ShareInfo{
		User: &model.ShareInfoItem{
			Name:    account.Info.GetName(),
			Picture: account.Info.GetFace(),
		},
		Prize: &model.ShareInfoItem{
			Name:    goods.Name,
			Picture: goods.Picture,
		},
	}
	return
}

func (s *Service) PrizeList(c context.Context) (res []*model.PrizeInfo, err error) {
	var data []*model.Task
	res = make([]*model.PrizeInfo, 0, 30)
	if data, err = s.dao.PrizeList(c); err != nil {
		log.Error("PrizeList(%v)", err)
		return nil, err
	} else if len(data) == 0 {
		return
	}
	var (
		userLast  = make([]int64, 0, len(data))
		userMap   = make(map[int64]*model.PrizeInfoItem)
		prizeMap  = make(map[int64]*model.PrizeInfoItem)
		prizedata []*model.Goods
	)
	for _, n := range data {
		if n.Uid >= 0 {
			userLast = append(userLast, n.Uid)
		}
	}
	if userMap, err = s.getUserInfos(c, userLast); err != nil {
		log.Error("getUserInfos(%v)", err)
		return nil, err
	}
	if prizedata, err = s.dao.GoodsList(c); err != nil {
		log.Error("GoodsList(%v)", err)
		return nil, err
	}
	for _, n := range prizedata {
		prizeMap[n.Id] = &model.PrizeInfoItem{
			Name:    n.Name,
			Picture: n.Picture,
		}
	}
	for _, n := range data {
		r := &model.PrizeInfo{
			Source: "exchange",
			Time:   int64(n.Mtime),
		}
		if v, ok := userMap[n.Uid]; ok {
			r.User = v
		}
		if v, ok := prizeMap[n.Goods]; ok {
			r.Prize = v
		}
		res = append(res, r)
		if len(res) >= 30 {
			return
		}
	}
	var winList []*model.WinListItem
	if winList, err = s.dao.WinList(c, s.c.Vogue.Sid); err != nil {
		log.Error("s.dao.WinList(%v)", err)
		return nil, err
	}
	if len(winList) <= 0 {
		return
	}
	userLast = make([]int64, 0, len(winList))
	for _, n := range winList {
		userLast = append(userLast, n.Mid)
	}
	if userMap, err = s.getUserInfos(c, userLast); err != nil {
		log.Error("getUserInfos(%v)", err)
		return nil, err
	}
	for _, n := range winList {
		r := &model.PrizeInfo{
			Prize: &model.PrizeInfoItem{
				Name:    n.GiftName,
				Picture: n.GiftImgUrl,
			},
			Source: "lottery",
			Time:   n.CTime,
		}
		if v, ok := userMap[n.Mid]; ok {
			r.User = v
		}
		res = append(res, r)
		if len(res) >= 30 {
			return
		}
	}
	return []*model.PrizeInfo{}, nil // 少于30不展示
}

func (s *Service) getUserInfos(c context.Context, mids []int64) (res map[int64]*model.PrizeInfoItem, err error) {
	res = make(map[int64]*model.PrizeInfoItem)
	var userdata *accountAPI.InfosReply
	if userdata, err = s.accClient.Infos3(c, &accountAPI.MidsReq{
		Mids: mids,
	}); err != nil {
		return nil, err
	}
	for k, v := range userdata.GetInfos() {
		res[k] = &model.PrizeInfoItem{
			Name:    v.Name,
			Picture: v.Face,
		}
	}
	return
}
