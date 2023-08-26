package service

import (
	"context"

	acccli "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/admin/dao"
	"go-gateway/app/web-svr/native-page/admin/model"
)

func (s *Service) AddWhiteList(c context.Context, req *model.AddWhiteListReq, uid int64, username, from string) error {
	if req == nil || req.Mid == 0 {
		return ecode.RequestErr
	}
	whiteList, _, err := s.dao.WhiteList(c, req.Mid, 0, 1)
	if err != nil {
		return err
	}
	if len(whiteList) != 0 {
		return nil
	}
	item := &model.WhiteListRecord{
		Mid:         req.Mid,
		Creator:     username,
		CreatorUID:  int(uid),
		Modifier:    username,
		ModifierUID: int(uid),
		FromType:    from,
		State:       dao.StateValid,
	}
	if _, err := s.dao.AddWhiteList(c, item); err != nil {
		return err
	}
	return nil
}

func (s *Service) BatchAddWhiteList(c context.Context, req *model.BatchAddWhiteListReq, uid int64, username, from string) error {
	whiteList, err := s.dao.WhiteListByMids(c, req.Mids)
	if err != nil {
		return err
	}
	insertMids := make([]int64, 0, len(req.Mids))
	for _, mid := range req.Mids {
		if _, ok := whiteList[mid]; ok {
			continue
		}
		insertMids = append(insertMids, mid)
	}
	if len(insertMids) == 0 {
		return nil
	}
	attrs := make([]*model.WhiteListRecord, 0, len(insertMids))
	for _, mid := range insertMids {
		item := &model.WhiteListRecord{
			Mid:         mid,
			Creator:     username,
			CreatorUID:  int(uid),
			Modifier:    username,
			ModifierUID: int(uid),
			FromType:    from,
			State:       dao.StateValid,
		}
		attrs = append(attrs, item)
	}
	if err := s.dao.BatchAddWhiteList(c, attrs); err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteWhiteList(c context.Context, req *model.DeleteWhiteListReq, uid int64, username string) error {
	attrs := map[string]interface{}{
		"modifier":     username,
		"modifier_uid": int(uid),
		"state":        dao.StateInvalid,
	}
	if err := s.dao.UpdateWhiteList(c, req.ID, attrs); err != nil {
		return err
	}
	return nil
}

func (s *Service) WhiteList(c context.Context, req *model.GetWhiteListReq) (*model.GetWhiteListRly, error) {
	whiteList, total, err := s.dao.WhiteList(c, req.Mid, req.Pn, req.Ps)
	if err != nil {
		return nil, err
	}
	rly := &model.GetWhiteListRly{Total: total}
	if len(whiteList) == 0 {
		return rly, nil
	}
	var mids []int64
	list := make([]*model.ListItem, 0, len(whiteList))
	for _, v := range whiteList {
		if v == nil {
			continue
		}
		mids = append(mids, v.Mid)
		list = append(list, &model.ListItem{WhiteListRecord: v})
	}
	func() {
		if len(mids) == 0 {
			return
		}
		rly, err := s.accClient.Infos3(c, &acccli.MidsReq{Mids: mids})
		if err != nil {
			log.Error("Fail to get accountInfos, mids=%+v error=%+v", mids, err)
			return
		}
		if rly == nil || rly.Infos == nil {
			return
		}
		accounts := rly.Infos
		for _, v := range list {
			if account, ok := accounts[v.Mid]; ok && account != nil {
				v.UserName = account.Name
			}
		}
	}()
	rly.List = list
	return rly, nil
}
