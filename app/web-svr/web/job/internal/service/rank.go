package service

import (
	"context"

	"go-common/library/log"
	archiveapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/job/internal/model"

	"github.com/pkg/errors"
)

var (
	rankIndexDay       = []int64{1, 3, 7}
	rankRegionDay      = []int64{3, 7}
	rankRegionDayAll   = []int64{1, 3, 7}
	rankRegionOriginal = []int64{0, 1}
)

func (s *Service) setRankIndex() {
	ctx := context.Background()
	for _, day := range rankIndexDay {
		func() {
			var aids []int64
			if err := retry(func() (err error) {
				aids, err = s.dao.RankIndex(ctx, day)
				return err
			}); err != nil {
				log.Error("日志告警 RankIndex day:%d error:%+v", day, err)
				return
			}
			if err := retry(func() (err error) {
				return s.dao.AddCacheRankIndex(ctx, day, aids)
			}); err != nil {
				log.Error("日志告警 AddCacheRankIndex day:%d error:%+v", day, err)
				return
			}
		}()
	}
}

func (s *Service) setRankRecommend() {
	ctx := context.Background()
	region := s.ac.RankRids.Recommend
	if len(region) == 0 {
		log.Error("日志告警 setRankRecommend len(s.ac.RankRids.Recommend) == 0")
		return
	}
	for _, rid := range region {
		func() {
			var aids []int64
			if err := retry(func() (err error) {
				aids, err = s.dao.RankRecommend(ctx, rid)
				return err
			}); err != nil {
				log.Error("日志告警 RankRecommend rid:%d error:%+v", rid, err)
				return
			}
			if cnt := len(aids); cnt < s.ac.Rule.RcmdMinCnt {
				log.Error("日志告警 RankRecommend rid:%d len(aids):%d < min:%d", rid, cnt, s.ac.Rule.RcmdMinCnt)
				return
			}
			if err := retry(func() (err error) {
				return s.dao.AddCacheRankRecommend(ctx, rid, aids)
			}); err != nil {
				log.Error("日志告警 AddCacheRankRecommend day:%d error:%+v", rid, err)
				return
			}
		}()
	}
}

func (s *Service) setLpRankRecommend() {
	ctx := context.Background()
	lp := s.ac.RankRids.LandingPage
	if len(lp) == 0 {
		log.Error("日志告警 setLpRankRecommend len(s.ac.LandingPage) == 0")
		return
	}
	for _, business := range lp {
		func() {
			var aids []int64
			if err := retry(func() (err error) {
				aids, err = s.dao.LpRankRecommend(ctx, business)
				return err
			}); err != nil {
				log.Error("日志告警 LpRankRecommend business:%s error:%+v", business, err)
				return
			}
			if cnt := len(aids); cnt < s.ac.Rule.RcmdMinCnt {
				log.Error("日志告警 LpRankRecommend business:%s len(aids):%d < min:%d", business, cnt, s.ac.Rule.RcmdMinCnt)
				return
			}
			if err := retry(func() (err error) {
				return s.dao.AddCacheLpRankRecommend(ctx, business, aids)
			}); err != nil {
				log.Error("日志告警 AddCacheRankRecommend business:%s error:%+v", business, err)
				return
			}
		}()
	}
}

func (s *Service) setFirstRankRegion() {
	ctx := context.Background()
	region := s.ac.RankRids.FirstRegion
	if len(region) == 0 {
		log.Error("日志告警 setFirstRankRegion len(s.ac.RankRids.FirstRegion) == 0")
		return
	}
	for _, rid := range region {
		for _, day := range rankRegionDay {
			func() {
				var list []*model.RankAid
				if err := retry(func() (err error) {
					list, err = s.dao.RankRegion(ctx, rid, day, 0)
					return err
				}); err != nil {
					log.Error("日志告警 FirstRankRegion rid:%d error:%+v", rid, err)
					return
				}
				if err := retry(func() (err error) {
					return s.dao.AddCacheRankRegion(ctx, rid, day, 0, list)
				}); err != nil {
					log.Error("日志告警 FirstRankRegion AddCacheRankRegion rid:%d day:%d original:%d error:%+v", rid, day, 0, err)
					return
				}
			}()
		}
	}
}

func (s *Service) setSecondRankRegion() {
	ctx := context.Background()
	if len(s.arcTypes) == 0 {
		log.Error("日志告警 setFirstRankRegion len(s.arcTypes) == 0")
		return
	}
	for _, region := range s.arcTypes {
		if region == nil || region.Pid == 0 {
			continue
		}
		// 去除下线的分区
		if s.offlineRegion(region.ID) {
			continue
		}
		regionDay := func() []int64 {
			if s.regionDayAll(region.ID) {
				return rankRegionDayAll
			}
			return rankRegionDay
		}()
		for _, day := range regionDay {
			for _, original := range rankRegionOriginal {
				if s.noOriginalRegion(region.ID) && original == 1 {
					continue
				}
				func() {
					var list []*model.RankAid
					if err := retry(func() (err error) {
						list, err = s.dao.RankRegion(ctx, int64(region.ID), day, original)
						return err
					}); err != nil {
						log.Error("日志告警 RankRegion rid:%d day:%d original:%d error:%+v", region.ID, day, original, err)
						return
					}
					if err := retry(func() (err error) {
						return s.dao.AddCacheRankRegion(ctx, int64(region.ID), day, original, list)
					}); err != nil {
						log.Error("日志告警 AddCacheRankRegion rid:%d day:%d error:%+v", region.ID, day, err)
						return
					}
				}()
			}
		}
	}
}

func (s *Service) setRankTag() {
	ctx := context.Background()
	if len(s.arcTypes) == 0 {
		log.Error("日志告警 setRankTag len(s.arcTypes) == 0")
		return
	}
	for _, region := range s.arcTypes {
		// 去除下线的分区
		if s.offlineRegion(region.ID) {
			continue
		}
		// 只获取二级分区热门tag列表
		if region == nil || region.Pid == 0 {
			continue
		}
		func() {
			// 获取分区tag热门列表
			var tagIDs []int64
			if err := retry(func() (err error) {
				tagIDs, err = s.dao.TagHots(ctx, int64(region.ID))
				return err
			}); err != nil {
				log.Error("日志告警 TagHots rid:%d error:%+v", region.ID, err)
				return
			}
			if len(tagIDs) == 0 {
				log.Warn("setRankTag rid:%d hot tags nil", region.ID)
				return
			}
			for _, tagID := range tagIDs {
				var rankAids []*model.RankAid
				if err := retry(func() (err error) {
					rankAids, err = s.dao.RankTag(ctx, int64(region.ID), tagID)
					return err
				}); err != nil {
					log.Error("日志告警 RankTag rid:%d tagID:%d error:%+v", region.ID, tagID, err)
					return
				}
				if err := retry(func() (err error) {
					return s.dao.AddCacheRankTag(ctx, int64(region.ID), tagID, rankAids)
				}); err != nil {
					log.Error("日志告警 AddCacheRankTag rid:%d tagID:%d error:%+v", region.ID, tagID, err)
					return
				}
			}
		}()
	}
}

func (s *Service) setRankList() {
	ctx := context.Background()
	region := s.ac.RankRids.RankV2Rids
	if len(region) == 0 {
		log.Warn("setRankList len(region) == 0")
		return
	}
	for _, rid := range region {
		func() {
			var rankList *model.RankList
			if err := retry(func() (err error) {
				rankList, err = s.dao.RankList(ctx, model.RankListTypeNone, rid)
				return err
			}); err != nil {
				log.Error("日志告警 RankList rid:%d error:%+v", rid, err)
				return
			}
			if err := retry(func() (err error) {
				return s.dao.AddCacheRankList(ctx, model.RankListTypeNone, rid, rankList)
			}); err != nil {
				log.Error("日志告警 AddCacheRankList rid:%d error:%+v", rid, err)
				return
			}
		}()
	}
	for _, typ := range []model.RankListType{model.RankListTypeAll, model.RankListTypeOrigin, model.RankListTypeRookie} {
		func() {
			var rankList *model.RankList
			if err := retry(func() (err error) {
				rankList, err = s.dao.RankList(ctx, typ, 0)
				return err
			}); err != nil {
				log.Error("日志告警 RankList typ:%d error:%+v", typ, err)
				return
			}
			if err := retry(func() (err error) {
				return s.dao.AddCacheRankList(ctx, typ, 0, rankList)
			}); err != nil {
				log.Error("日志告警 AddCacheRankList typ:%d error:%+v", typ, err)
				return
			}
		}()
	}
	for _, rid := range s.ac.RankRids.RankOldRids {
		func() {
			var rankList *model.RankList
			if err := retry(func() (err error) {
				rankList, err = s.dao.RankListOld(ctx, rid)
				return err
			}); err != nil {
				log.Error("日志告警 RankListOld rid:%d error:%+v", rid, err)
				return
			}
			if err := retry(func() (err error) {
				return s.dao.AddCacheRankList(ctx, model.RankListTypeNone, rid, rankList)
			}); err != nil {
				log.Error("日志告警 AddCacheRankList RankListOld rid:%d error:%+v", rid, err)
				return
			}
		}()
	}
}

func (s *Service) cronArcTypes() {
	err := s.loadArcTypes()
	if err != nil {
		log.Error("日志告警 cronArcTypes error:%v", err)
	}
}

func (s *Service) loadArcTypes() error {
	reply, err := s.archiveGRPC.Types(context.Background(), &archiveapi.NoArgRequest{})
	if err != nil {
		return err
	}
	if len(reply.GetTypes()) == 0 {
		return errors.New("loadArcTypes Types len 0")
	}
	s.arcTypes = reply.GetTypes()
	return nil
}

func (s *Service) offlineRegion(rid int32) bool {
	for _, offline := range s.ac.RankRids.Offline {
		if rid == offline {
			return true
		}
	}
	return false
}

func (s *Service) noOriginalRegion(rid int32) bool {
	for _, noOriginal := range s.ac.RankRids.NoOriginal {
		if rid == noOriginal {
			return true
		}
	}
	return false
}

func (s *Service) regionDayAll(rid int32) bool {
	for _, allDayRid := range s.ac.RankRids.DayAll {
		if rid == allDayRid {
			return true
		}
	}
	return false
}
