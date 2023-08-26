package service

import (
	"context"
	"fmt"

	"go-common/library/log"
	archiveapi "go-gateway/app/app-svr/archive/service/api"
	dynamicapi "go-gateway/app/web-svr/dynamic/service/api/v1"
	"go-gateway/app/web-svr/web/job/internal/model"
	"go-gateway/pkg/idsafe/bvid"

	"github.com/pkg/errors"
)

const (
	_firstPn        = 1
	_newListCachePs = 100
)

var newListType = []int32{0, 1}

func (s *Service) setNewListFirstRegion() {
	ctx := context.Background()
	for _, region := range s.arcTypes {
		if region == nil || region.Pid != 0 {
			continue
		}
		// 去除下线的分区
		if s.offlineRegion(region.ID) {
			continue
		}
		var list []*archiveapi.Arc
		if err := retry(func() (err error) {
			// 一级分区
			list, err = func() ([]*archiveapi.Arc, error) {
				res, regionErr := s.dynamicGRPC.RecentThrdRegArc(ctx, &dynamicapi.RecentThrdRegArcReq{Rid: region.ID, Pn: _firstPn, Ps: _newListCachePs})
				if regionErr != nil {
					return nil, regionErr
				}
				if len(res.GetArchives()) == 0 {
					return nil, errors.New(fmt.Sprintf("newlist first rid:%d is nil", region.ID))
				}
				return res.GetArchives(), nil
			}()
			return err
		}); err != nil {
			log.Error("日志告警 Newlist rid:%d error:%+v", region.ID, err)
			return
		}
		var bvList []*model.BvArc
		for _, v := range list {
			bvidStr, _ := bvid.AvToBv(v.Aid)
			bvList = append(bvList, &model.BvArc{Arc: v, Bvid: bvidStr})
		}
		if err := retry(func() (err error) {
			return s.dao.AddCacheNewList(ctx, int64(region.ID), 0, bvList, 0)
		}); err != nil {
			log.Error("日志告警 AddCacheNewList rid:%d error:%+v", region.ID, err)
			return
		}
	}
}

func (s *Service) setNewListSecondRegion() {
	ctx := context.Background()
	for _, region := range s.arcTypes {
		if region == nil || region.Pid == 0 {
			continue
		}
		// 去除下线的分区
		if s.offlineRegion(region.ID) {
			continue
		}
		for _, typ := range newListType {
			var (
				list  []*archiveapi.Arc
				total int
			)
			if err := retry(func() (err error) {
				// 二级分区
				list, total, err = func() ([]*archiveapi.Arc, int, error) {
					res, regionErr := s.dynamicGRPC.RegAllArcs(ctx, &dynamicapi.RegAllReq{Rid: int64(region.ID), Type: typ, Pn: _firstPn, Ps: _newListCachePs})
					if regionErr != nil {
						return nil, 0, regionErr
					}
					if len(res.GetArchives()) == 0 {
						return nil, 0, errors.New(fmt.Sprintf("newlist second rid:%d typ:%d is nil", region.ID, typ))
					}
					return res.GetArchives(), int(res.GetCount()), nil
				}()
				return err
			}); err != nil {
				log.Error("日志告警 NewList rid:%d error:%+v", region.ID, err)
				return
			}
			var bvList []*model.BvArc
			for _, v := range list {
				bvidStr, _ := bvid.AvToBv(v.Aid)
				bvList = append(bvList, &model.BvArc{Arc: v, Bvid: bvidStr})
			}
			if err := retry(func() (err error) {
				return s.dao.AddCacheNewList(ctx, int64(region.ID), int64(typ), bvList, total)
			}); err != nil {
				log.Error("日志告警 AddCacheNewList error:%+v", err)
				return
			}
		}
	}
}
