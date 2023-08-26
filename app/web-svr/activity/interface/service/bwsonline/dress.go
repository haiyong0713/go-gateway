package bwsonline

import (
	"context"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
)

func (s *Service) MyDressList(ctx context.Context, mid int64) ([]*bwsonline.Dress, error) {
	list, err := s.userDress(ctx, mid, false)
	if err != nil {
		log.Errorc(ctx, "MyDressList s.userDress mid:%d error:%v", mid, err)
	}
	return list, err
}

func (s *Service) DressUp(ctx context.Context, mid int64, ids []int64) error {
	// check type
	afIDs := filterIDs(ids)
	if len(afIDs) > 0 {
		data, err := s.dao.DressByIDs(ctx, afIDs)
		if err != nil {
			log.Errorc(ctx, "DressUp s.dao.DressByIDs ids:%v error:%v", afIDs, err)
			return err
		}
		for _, id := range afIDs {
			if item, ok := data[id]; !ok || item == nil {
				return ecode.BwsOnlineDressNotExist
			}
		}
		// 过滤重复位置图
		posMap := make(map[int64]struct{})
		for _, v := range data {
			if _, ok := posMap[v.Pos]; ok {
				return ecode.BwsOnlineDressPosRepeat
			}
		}
		// 判断是否已拥有
		userDress, err := s.dao.UserDress(ctx, mid)
		if err != nil {
			return err
		}
		for _, id := range afIDs {
			var idCheck bool
			for _, v := range userDress {
				if v != nil && id == v.DressId {
					idCheck = true
					break
				}
			}
			if !idCheck {
				return xecode.Errorf(ecode.BwsOnlineDressNotHave, ecode.BwsOnlineDressNotHave.Message(), data[id].Title)
			}
		}
	}
	if _, err := s.dao.DressOff(ctx, mid); err != nil {
		return err
	}
	if len(afIDs) > 0 {
		if _, err := s.dao.DressUp(ctx, mid, afIDs); err != nil {
			return err
		}
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		s.dao.DelCacheUserDress(ctx, mid)
	})
	return nil
}

func (s *Service) userDress(ctx context.Context, mid int64, hasEquip bool) ([]*bwsonline.Dress, error) {
	data, err := s.dao.UserDress(ctx, mid)
	if err != nil {
		return nil, err
	}
	var dressIDs []int64
	for _, v := range data {
		if v != nil && v.DressId > 0 {
			if hasEquip && v.State != bwsonline.DressHasEquip {
				continue
			}
			dressIDs = append(dressIDs, v.DressId)
		}
	}
	if len(dressIDs) == 0 {
		return []*bwsonline.Dress{}, nil
	}
	dresses, err := s.dao.DressByIDs(ctx, dressIDs)
	if err != nil {
		return nil, err
	}
	var list []*bwsonline.Dress
	for _, v := range dressIDs {
		if dress, ok := dresses[v]; ok && dress != nil {
			list = append(list, dress)
		}
	}
	return list, nil
}
