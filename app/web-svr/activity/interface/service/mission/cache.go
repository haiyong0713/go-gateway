package mission

import (
	"context"
	"github.com/bluele/gcache"
	"go-common/library/log"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/conf"
	"time"
)

const (
	_defaultCacheTtl = 300 * time.Second
)

func (s *Service) storeActivityCacheTicker(ctx context.Context) (err error) {
	duration := time.Duration(_defaultRefreshTicker) * time.Second
	if conf.Conf.MissionActivityConf != nil && conf.Conf.MissionActivityConf.CacheRule.RefreshActivityCacheSeconds != 0 {
		duration = time.Duration(conf.Conf.MissionActivityConf.CacheRule.RefreshActivityCacheSeconds) * time.Second
	}
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			s.storeValidActivityCache(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) storeActivityTasksCacheTicker(ctx context.Context) (err error) {
	duration := time.Duration(_defaultRefreshTicker) * time.Second
	if conf.Conf.MissionActivityConf != nil && conf.Conf.MissionActivityConf.CacheRule.RefreshActivityTasksCacheSeconds != 0 {
		duration = time.Duration(conf.Conf.MissionActivityConf.CacheRule.RefreshActivityTasksCacheSeconds) * time.Second
	}
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ticker.C:
			s.storeActivityTaskCache(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) getMissionActivityFromLocal(ctx context.Context, actId int64) (activityDetail *v1.MissionActivityDetail, err error) {
	cacheValue, err := s.activityCache.Get(actId)
	if err != nil {
		log.Errorc(ctx, "[getMissionActivityFromCache][FromLocal][Error], err:%+v", err)
		return
	}
	activityDetail, ok := cacheValue.(*v1.MissionActivityDetail)
	if !ok {
		err = gcache.KeyNotFoundError
		return
	}
	return
}

func (s *Service) getActivityTasksFromLocal(ctx context.Context, actId int64) (tasks []*v1.MissionTaskDetail, err error) {
	cacheValue, err := s.activityCache.Get(actId)
	if err != nil {
		return
	}
	tasks, ok := cacheValue.([]*v1.MissionTaskDetail)
	if !ok {
		err = gcache.KeyNotFoundError
		return
	}
	return
}

func (s *Service) storeValidActivityCache(ctx context.Context) {
	// 获取有效的活动
	validsIds, err := s.GetValidActivityIds(ctx, false)
	if err != nil {
		log.Errorc(ctx, "[storeValidActivityCache][Error], err:%+v", err)
		return
	}
	for _, actId := range validsIds {
		actDetail, errG := s.getMissionActivityInfo(ctx, actId, true, false)
		if errG != nil {
			log.Errorc(ctx, "[storeValidActivityCache][getMissionActivityInfo][Error], actId:%d, err:%+v", actId, err)
			continue
		}
		err = s.activityCache.SetWithExpire(actId, actDetail, _defaultCacheTtl)
		if err != nil {
			log.Errorc(ctx, "[storeValidActivityCache][CacheSet][Error], actId:%d, err:%+v", actId, err)
			continue
		}
	}
	return
}

func (s *Service) storeActivityTaskCache(ctx context.Context) {
	// 获取有效的活动
	validIds, err := s.GetValidActivityIds(ctx, false)
	if err != nil {
		log.Errorc(ctx, "[storeActivityTaskCache][Error], err:%+v", err)
		return
	}
	for _, actId := range validIds {
		actDetail, errG := s.getMissionActivityInfo(ctx, actId, false, false)
		if errG != nil {
			log.Errorc(ctx, "[storeActivityTaskCache][getMissionActivityInfo][Error], actId:%d, err:%+v", actId, err)
			continue
		}
		tasks, errG := s.getActivityTasks(ctx, actDetail.Id, true, false)
		if errG != nil {
			log.Errorc(ctx, "[storeActivityTaskCache][getActivityTasks][Error], actId:%d, err:%+v", actId, err)
			continue
		}
		err = s.activityTasksCache.SetWithExpire(actId, tasks, _defaultCacheTtl)
		if err != nil {
			log.Errorc(ctx, "[storeActivityTaskCache][CacheSet][Error], actId:%d, err:%+v", actId, err)
			continue
		}
		s.storeTaskInfoCache(ctx, tasks)
		s.storeGroupActivityTaskCache(ctx, tasks)
	}
	return
}

func (s *Service) storeTaskInfoCache(ctx context.Context, tasks []*v1.MissionTaskDetail) {
	for _, task := range tasks {
		err := s.activityTaskInfoCache.SetWithExpire(task.TaskId, task, _defaultCacheTtl)
		if err != nil {
			log.Errorc(ctx, "[storeTaskInfoCache][CacheSet][Error], task:%+v, err:%+v", task, err)
			continue
		}
	}
}

func (s *Service) storeGroupActivityTaskCache(ctx context.Context, tasks []*v1.MissionTaskDetail) {
	for _, task := range tasks {
		for _, group := range task.Groups {
			err := s.groupTaskMappingCache.SetWithExpire(group.GroupId, task, _defaultCacheTtl)
			if err != nil {
				log.Errorc(ctx, "[storeGroupActivityTaskCache][CacheSet][Error], groupId:%d, err:%+v", group.GroupId, err)
				continue
			}
		}
	}
}

func (s *Service) RefreshActivityCache(ctx context.Context, actId int64) (err error) {
	// 活动基础详情
	_, err = s.getMissionActivityInfo(ctx, actId, true, true)
	if err != nil {
		log.Errorc(ctx, "[RefreshActivityCache][getMissionActivityInfo][Error], actId:%d, err:%+v", actId, err)
		return
	}
	_, err = s.getActivityTasks(ctx, actId, true, true)
	if err != nil {
		log.Errorc(ctx, "[RefreshActivityCache][getMissionActivityInfo][Error], actId:%d, err:%+v", actId, err)
		return
	}
	return
}
