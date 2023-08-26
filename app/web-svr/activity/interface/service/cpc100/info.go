package cpc100

import (
	"context"
	"encoding/json"
	"fmt"
	xecode "go-common/library/ecode"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/tool"
	"time"
)

func (s *Service) Info(c context.Context, mid int64) (res *likemdl.Cpc100Info, err error) {
	var items []*likemdl.WebDataItem
	items, err = s.dao.GetWebViewDataByVid(c, s.c.Cpc100.Vid)
	if err != nil {
		return nil, err
	}
	info := make(map[string]int)
	if mid > 0 {
		info, _ = s.dao.Cpc100UnlockInfo(c, mid)
	}
	res = new(likemdl.Cpc100Info)
	now := time.Now()
	for _, item := range items {
		one := &likemdl.Cpc100Egg{
			Name:  item.Name,
			STime: item.STime,
			ETime: item.ETime,
			Key:   tool.MD5(fmt.Sprint(item.ID, "20210614")),
		}
		json.Unmarshal(item.Raw, &one.Data)
		if now.Before(item.STime.Time()) {
			continue
		}
		one.Unlocked = info[one.Key] == 1
		res.List = append(res.List, one)
	}
	return
}

func (s *Service) Unlock(c context.Context, mid int64, key string) (err error) {
	var items []*likemdl.WebDataItem
	items, err = s.dao.GetWebViewDataByVid(c, s.c.Cpc100.Vid)
	if err != nil {
		return
	}
	for _, item := range items {
		itemKey := tool.MD5(fmt.Sprint(item.ID, "20210614"))
		if itemKey != key {
			continue
		}
		if time.Now().After(item.STime.Time()) {
			return s.dao.Cpc100Unlock(c, mid, key)
		} else {
			return xecode.RequestErr
		}
	}
	return xecode.RequestErr
}

func (s *Service) Reset(c context.Context) (err error) {
	return s.dao.DelCacheGetWebViewDataByVid(c, s.c.Cpc100.Vid)
}

func (s *Service) PageView(ctx context.Context) (int64, error) {
	return s.dao.CpcGetPV(ctx)
}

func (s *Service) TotalView(ctx context.Context) (total int64, err error) {
	pv, err := s.dao.CpcGetPV(ctx)
	if err != nil {
		return
	}
	tv, err := s.dao.CpcGetTopicView(ctx)
	if err != nil {
		return
	}
	total = pv + tv
	return
}
