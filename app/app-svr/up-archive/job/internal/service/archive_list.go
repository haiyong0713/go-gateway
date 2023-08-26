package service

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/app-svr/up-archive/job/internal/model"
	"go-gateway/app/app-svr/up-archive/service/api"
)

func (s *Service) initArcList() {
	defer s.waiter.Done()
	lastAid := s.ac.BuildArchiveList.LastAid
	for {
		if s.closed || !s.ac.BuildArchiveList.Switch {
			close(s.aidsChan)
			return
		}
		var aids []int64
		if err := retry(func() (err error) {
			aids, err = s.dao.RawUpper(context.Background(), lastAid, s.ac.BuildArchiveList.Limit)
			return err
		}); err != nil {
			log.Error("日志告警 initBuildArcList RawUpper lastAid:%d error:%+v", lastAid, err)
			continue
		}
		if len(aids) == 0 {
			close(s.aidsChan)
			break
		}
		lastAid = aids[len(aids)-1]
		s.aidsChan <- aids
	}
	log.Warn("initBuildArcList success")
}

func (s *Service) buildArc() {
	defer s.waiter.Done()
	for {
		aids, ok := <-s.aidsChan
		if !ok {
			return
		}
		for _, aid := range aids {
			s.buildArcNoSpace(context.Background(), aid)
		}
	}
}

func (s *Service) buildArcNoSpace(ctx context.Context, mid int64) { // 必须在up_on_space未上线前使用
	exists, err := s.dao.CacheArcPassedExists(ctx, mid, api.Without_no_space)
	if err != nil {
		log.Error("日志告警 buildArcNoSpace CacheArcPassedExists mid:%d error:%+v", mid, err)
		return
	}
	if exists { // 缓存存在时退出
		return
	}
	var arcs []*model.UpArc
	if err := retry(func() (err error) {
		arcs, err = s.dao.RawArcPassed(ctx, mid)
		return err
	}); err != nil {
		log.Error("日志告警 buildArcNoSpace RawArcPassed mid:%d error:%+v", mid, err)
		return
	}
	var withoutNoSpace []*model.UpArc
	for _, v := range arcs {
		if v != nil && v.IsOldAllowed() {
			v.RandScoreNumber()
			withoutNoSpace = append(withoutNoSpace, v)
		}
	}
	var staffAids []int64
	if err := retry(func() (err error) {
		staffAids, err = s.dao.RawStaffAids(ctx, mid)
		return err
	}); err != nil {
		log.Error("日志告警 buildArcNoSpace RawStaffAids mid:%d error:%+v", mid, err)
		return
	}
	if len(staffAids) > 0 {
		staffArcMap, _, err := s.arcs(ctx, staffAids, true)
		if err != nil {
			log.Error("日志告警 buildArcNoSpace arcs mid:%d error:%+v", mid, err)
			return
		}
		for _, v := range staffAids {
			if arc, ok := staffArcMap[v]; ok && arc != nil && arc.IsOldAllowed() {
				arc.RandScoreNumber()
				withoutNoSpace = append(withoutNoSpace, arc)
			}
		}
	}
	// 初始化缓存先删
	if err := retry(func() error {
		return s.dao.DelCacheArcNoSpace(ctx, mid)
	}); err != nil {
		log.Error("日志告警 buildArcNoSpace DelCacheArcNoSpace mid:%d error:%+v", mid, err)
		return
	}
	s.addCacheArc(withoutNoSpace, func(arcs []*model.UpArc) error {
		if err := retry(func() error {
			return s.dao.AddCacheArcPassed(ctx, mid, arcs, api.Without_no_space)
		}); err != nil {
			log.Error("日志告警 buildArcNoSpace AddCacheArcPassed mid:%d without:%+d error:%v", mid, api.Without_no_space, err)
			return err
		}
		return nil
	})
	log.Warn("buildArcNoSpace success mid:%d,length:without_no_space:%d", mid, len(withoutNoSpace))
}
