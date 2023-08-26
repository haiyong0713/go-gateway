package service

import (
	"context"

	"go-common/library/ecode"

	"go-gateway/app/web-svr/activity/admin/dao"
	"go-gateway/app/web-svr/activity/admin/model"
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
		accounts, err := s.Accounts(c, mids)
		if err != nil {
			return
		}
		for _, v := range list {
			if account, ok := accounts[v.Mid]; ok && account != nil {
				v.UserName = account.Name
			}
		}
	}()
	rly.List = list
	return rly, nil
}
