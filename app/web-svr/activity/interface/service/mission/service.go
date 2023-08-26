package mission

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	errgroup2 "go-common/library/sync/errgroup.v2"
	v1 "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/mission"
	model "go-gateway/app/web-svr/activity/interface/model/mission"
	"go-gateway/app/web-svr/activity/interface/rewards"
	"go-gateway/app/web-svr/activity/interface/service/like"
	stockserver "go-gateway/app/web-svr/activity/interface/service/stock_server"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"time"

	bind "go-gateway/app/web-svr/activity/interface/service/bind"

	"github.com/bluele/gcache"
)

var localS *Service

type Service struct {
	conf                  *conf.Config
	dao                   *mission.Dao
	activityCache         gcache.Cache
	activityTasksCache    gcache.Cache
	activityTaskInfoCache gcache.Cache
	groupTaskMappingCache gcache.Cache
	stockSvr              *stockserver.Service
	reserveSvr            *like.Service
	bindSvr               *bind.Service
}

const (
	_defaultRefreshTicker = 5
)

func New(c *conf.Config) (s *Service) {
	if localS != nil {
		return localS
	}
	s = &Service{
		conf:                  c,
		dao:                   mission.New(c),
		activityCache:         gcache.New(c.MissionActivityConf.CacheRule.ValidActivitySize).LFU().Build(),
		activityTasksCache:    gcache.New(c.MissionActivityConf.CacheRule.ValidActivitySize).LFU().Build(),
		groupTaskMappingCache: gcache.New(c.MissionActivityConf.CacheRule.ValidActivitySize).LFU().Build(),
		activityTaskInfoCache: gcache.New(c.MissionActivityConf.CacheRule.ValidActivitySize).LFU().Build(),
		bindSvr:               bind.New(c),
		stockSvr:              stockserver.New(c),
		reserveSvr:            like.New(c),
	}
	go initialize.CallC(s.storeActivityCacheTicker)
	go initialize.CallC(s.storeActivityTasksCacheTicker)
	localS = s
	return
}

func (s *Service) GetMissionActivityList(ctx context.Context, req *v1.GetMissionActivityListReq) (resp *v1.GetMissionActivityListResp, err error) {
	resp = new(v1.GetMissionActivityListResp)
	acts, total, err := s.dao.GetActivityListByPage(ctx, req.Pn, req.Ps)
	if err != nil && err != xsql.ErrNoRows {
		log.Errorc(ctx, "[GetMissionActivityList][GetActivityListByPage][Error], err:%+v", err)
		return
	}
	resp.Total = total
	resp.List = acts
	return
}

// GetMissionActivityInfo 获取活动详情
func (s *Service) GetMissionActivityInfo(ctx context.Context, req *v1.GetMissionActivityInfoReq) (resp *v1.MissionActivityDetail, err error) {
	resp = new(v1.MissionActivityDetail)
	if req == nil || req.ActId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	actId := req.ActId
	if req.SkipCache != 0 {
		resp, err = s.getMissionActivityInfo(ctx, actId, true, true)
	} else {
		resp, err = s.getMissionActivityInfo(ctx, actId, false, false)
	}
	if err != nil {
		log.Errorc(ctx, "[Server][GetMissionActivityInfo][GetActivityInfo][Error], err:%+v", err)
		return
	}
	return
}

// ChangeMissionActivityStatus 更改活动状态
func (s *Service) ChangeMissionActivityStatus(ctx context.Context, req *v1.ChangeMissionActivityStatusReq) (resp *v1.NoReply, err error) {
	resp = new(v1.NoReply)
	if req == nil || req.ActId == 0 ||
		(int(req.Status) != model.ActivityNormalStatus && int(req.Status) != model.ActivityAbnormalStatus) {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	activity, err := s.dao.GetActivityInfo(ctx, req.ActId)
	if err != nil {
		if err == xsql.ErrNoRows {
			err = xecode.Errorf(xecode.RequestErr, "活动不存在")
			return
		}
		return
	}
	if activity.Status == model.ActivityLockStatus {
		err = xecode.Errorf(xecode.RequestErr, "当前状态不允许编辑")
		return
	}
	err = s.dao.ActivityStatusUpdate(ctx, req.ActId, req.Status)
	_ = s.dao.DelActivityCache(ctx, req.ActId)
	return
}

// SaveMissionActivity 保存任务活动
func (s *Service) SaveMissionActivity(ctx context.Context, req *v1.MissionActivityDetail) (resp *v1.NoReply, err error) {
	resp = new(v1.NoReply)
	if req == nil || req.ActName == "" {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	if req.Id != 0 {
		// 更新
		err = s.updateMissionActivity(ctx, req)
	} else {
		// 新增
		err = s.addMissionActivity(ctx, req)
	}
	return
}

func (s *Service) addMissionActivity(ctx context.Context, activity *v1.MissionActivityDetail) (err error) {
	err = s.dao.AddActivity(ctx, activity)
	// TODO 处理活动对应的用户维度表
	return
}

func (s *Service) updateMissionActivity(ctx context.Context, activity *v1.MissionActivityDetail) (err error) {
	oldAct, err := s.dao.GetActivityInfo(ctx, activity.Id)
	if err != nil {
		return
	}
	if oldAct.UidCount != activity.UidCount {
		err = xecode.Errorf(xecode.RequestErr, "用户数不可变更")
		return
	}
	err = s.dao.UpdateActivity(ctx, activity)
	if err != nil {
		return
	}
	err = s.dao.DelActivityCache(ctx, activity.Id)
	if err != nil && err != redis.ErrNil {
		return
	}
	return
}

// GetMissionTasks 获取活动的任务列表
func (s *Service) GetMissionTasks(ctx context.Context, req *v1.GetMissionTasksReq) (resp *v1.GetMissionTasksResp, err error) {
	resp = new(v1.GetMissionTasksResp)
	if req == nil || req.Id == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	tasks, err := s.dao.GetActivityTasks(ctx, req.Id)
	if err != nil {
		log.Errorc(ctx, "[GetMissionTasks][GetActivityTasks][Error], err:%+v", err)
		return
	}
	if len(tasks) == 0 {
		return
	}
	stockIds := make([]int64, 0)
	for _, v := range tasks {
		stockIds = append(stockIds, v.StockId)
	}
	if len(stockIds) == 0 {
		return
	}
	stockList, err := s.stockSvr.BatchQueryStockRecord(ctx, stockIds, true)
	if err != nil {
		log.Errorc(ctx, "[GetMissionTasks][GetStocksByIds][Error], err:%+v", err)
		return
	}
	//if len(stockList) == 0 {
	//	err = xecode.Errorf(xecode.RequestErr, "库存信息获取失败")
	//	return
	//}
	stockMap := make(map[int64]*v1.CreateStockRecordReq)
	for _, stock := range stockList {
		stockMap[stock.StockId] = stock
	}
	for _, task := range tasks {
		task.StockConfig = &v1.TaskStockConfig{
			CycleLimit: "",
		}
		stock, ok := stockMap[task.StockId]
		if ok {
			task.StockConfig.CycleLimit = stock.CycleLimit
		}
	}
	resp.TaskList = tasks
	return
}

// SaveMissionTasks 活动下的任务全量保存
func (s *Service) SaveMissionTasks(ctx context.Context, req *v1.SaveMissionTasksReq) (resp *v1.NoReply, err error) {
	resp = new(v1.NoReply)
	if req == nil || req.ActId == 0 || len(req.Tasks) == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	actId := req.ActId
	tasks := req.Tasks
	err = s.taskSaveCommonCheck(ctx, actId, tasks)
	if err != nil {
		return
	}
	oldTasks, err := s.dao.GetActivityTasks(ctx, actId)
	if err != nil {
		return
	}
	errgroup := errgroup2.WithContext(ctx)
	for _, v := range tasks {
		task := v
		errgroup.Go(func(ctx context.Context) (err error) {
			err = s.doSingleTaskSave(ctx, actId, task)
			if err != nil {
				log.Errorc(ctx, "[SaveMissionTasks][ErrGroup][doSingleTaskSave][Error], err:%+v, task:%+v", err, task)
				return
			}
			return
		})
	}
	err = errgroup.Wait()
	if err != nil {
		log.Errorc(ctx, "[SaveMissionTasks][ErrGroup][Wait][Error], err:%+v", err)
		return
	}
	// 删除不需要的任务
	delTasks := s.getDelTasks(oldTasks, tasks)
	if len(delTasks) > 0 {
		err = s.dao.RemoveTasks(ctx, actId, delTasks)
	}
	return
}

func (s *Service) getDelTasks(oldTasks []*v1.MissionTaskDetail, tasks []*v1.MissionTaskDetail) (delTaskIds []int64) {
	delTaskIds = make([]int64, 0)
	updateTaskIds := make(map[int64]bool)
	for _, v := range tasks {
		if v.TaskId != 0 {
			updateTaskIds[v.TaskId] = true
		}
	}
	for _, v := range oldTasks {
		if _, ok := updateTaskIds[v.TaskId]; !ok {
			delTaskIds = append(delTaskIds, v.TaskId)
		}
	}
	return
}

func (s *Service) taskGroupMappingCheck(ctx context.Context, oldTasks []*v1.MissionTaskDetail, tasks []*v1.MissionTaskDetail) (err error) {
	// task 和 group的映射关系变更是否做判断 待定
	return
}

func (s *Service) taskSaveCommonCheck(ctx context.Context, actId int64, tasks []*v1.MissionTaskDetail) (err error) {
	activity, err := s.dao.GetActivityInfo(ctx, actId)
	if err != nil {
		if err == xsql.ErrNoRows {
			err = xecode.Errorf(xecode.RequestErr, "活动不存在")
		}
		log.Errorc(ctx, "[Service][taskSaveCommonCheck][GetActivityInfo][Error], err:%+v", err)
		return
	}
	groupIds := make([]int64, 0)
	for _, task := range tasks {
		for _, group := range task.Groups {
			groupIds = append(groupIds, group.GroupId)
		}
		// 校验奖品
		rewardResp, errG := rewards.Client.GetAwardConfigById(ctx, task.RewardId)
		if errG != nil || rewardResp == nil || rewardResp.Id == 0 {
			err = errG
			log.Errorc(ctx, "[TaskSaveCheck][GetAwardConfigById][Error], err:%+v", err)
			err = xecode.Errorf(xecode.RequestErr, "奖品填写有误")
			return
		}
	}
	groupActId := activity.GroupsActId
	res, err := s.reserveSvr.GetReserveCounterGroupIDBySid(ctx, groupActId)
	if err != nil {
		log.Errorc(ctx, "[Service][taskSaveCommonCheck][GetReserveCounterGroupIDBySid][Error], err:%+v", err)
		return
	}
	if len(res) == 0 {
		err = xecode.Errorf(xecode.RequestErr, "预约活动的节点组列表为空")
		return
	}
	reserveGroupMap := make(map[int64]bool)
	for _, v := range res {
		reserveGroupMap[v] = true
	}
	for _, groupId := range groupIds {
		if _, ok := reserveGroupMap[groupId]; !ok {
			err = xecode.Errorf(xecode.RequestErr, fmt.Sprintf("节点组 %d 不存在预约活动内", groupId))
			return
		}
	}
	return
}

// SaveMissionTask 保存活动下的某个任务
func (s *Service) SaveMissionTask(ctx context.Context, req *v1.MissionTaskDetail) (resp *v1.NoReply, err error) {
	resp = new(v1.NoReply)
	if req.ActId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	err = s.taskSaveCommonCheck(ctx, req.ActId, []*v1.MissionTaskDetail{req})
	if err != nil {
		return
	}
	if req.TaskId != 0 {
		taskInfo, errG := s.dao.GetActivityTaskInfo(ctx, req.ActId, req.TaskId)
		if errG != nil {
			log.Errorc(ctx, "[SaveMissionTask][GetActivityTaskInfo][Error], err:%+v", errG)
			err = xecode.Errorf(xecode.RequestErr, "任务信息获取失败")
			return
		}
		req.StockId = taskInfo.StockId
	}
	err = s.doSingleTaskSave(ctx, req.ActId, req)
	return
}

func (s *Service) doSingleTaskSave(ctx context.Context, actId int64, task *v1.MissionTaskDetail) (err error) {
	stockId := int64(0)
	var activityDetail *v1.MissionActivityDetail
	if activityDetail, err = s.getMissionActivityInfo(ctx, actId, true, true); err != nil {
		return
	}
	if task.TaskId != 0 {
		err = s.dao.UpdateActivityTask(ctx, actId, task)
	} else {
		err = s.dao.AddActivityTask(ctx, actId, task)
	}
	if err != nil {
		log.Errorc(ctx, "[doSingleTaskSave][AddOrUpdate][Error], err:%+v", err)
		return
	}
	if task.StockId == 0 {
		cResp, errG := s.stockSvr.CreateStockRecord(ctx, &v1.CreateStockRecordReq{
			StockId:        task.StockId,
			ResourceId:     "mission",
			ResourceVer:    time.Now().Unix(),
			ForeignActId:   fmt.Sprintf("mission-%d-%d", actId, task.TaskId),
			CycleLimit:     task.StockConfig.CycleLimit,
			DescInfo:       "",
			StockStartTime: activityDetail.BeginTime,
			StockEndTime:   activityDetail.EndTime,
		})
		if errG != nil {
			err = errG
			log.Errorc(ctx, "[doSingleTaskSave][CreateStockRecord][Error], err:%+v", err)
			// 删除任务
			err = s.dao.RemoveTasks(ctx, actId, []int64{task.TaskId})
			if err != nil {
				log.Errorc(ctx, "[doSingleTaskSave][RemoveTaskAfterStock][Error], err:%+v", err)
				err = xecode.Errorf(xecode.RequestErr, "对应关系创建失败，请删除后再次添加")
			}
			err = xecode.Errorf(xecode.RequestErr, "对应关系创建失败，请重试")
			return
		}
		stockId = cResp.StockId
	} else {
		_, err = s.stockSvr.UpdateStockServerConf(ctx, &v1.CreateStockRecordReq{
			StockId:        task.StockId,
			ResourceId:     "mission",
			ResourceVer:    time.Now().Unix(),
			ForeignActId:   fmt.Sprintf("mission-%d-%d", actId, task.TaskId),
			CycleLimit:     task.StockConfig.CycleLimit,
			DescInfo:       "",
			StockStartTime: activityDetail.BeginTime,
			StockEndTime:   activityDetail.EndTime,
		})
		if err != nil {
			log.Errorc(ctx, "[doSingleTaskSave][CreateStockRecord][Error], err:%+v", err)
			err = xecode.Errorf(xecode.RequestErr, "库存信息更新失败，请重试")
			return
		}
		stockId = task.StockId
	}
	err = s.dao.UpdateTaskStockId(ctx, actId, task.TaskId, stockId)
	if err != nil {
		log.Errorc(ctx, "[doSingleTaskSave][UpdateTaskStockId][Error], err:%+v", err)
		return
	}
	_ = s.dao.DelActivityTasksCache(ctx, task.ActId)
	_ = s.dao.DelActivityTaskCache(ctx, task.TaskId)
	return
}

// GetMissionTaskInfo 获取活动下某个任务详情
func (s *Service) GetMissionTaskInfo(ctx context.Context, req *v1.GetMissionTaskInfoReq) (resp *v1.MissionTaskDetail, err error) {
	resp = new(v1.MissionTaskDetail)
	if req == nil || req.ActId == 0 || req.TaskId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	resp, err = s.dao.GetActivityTaskInfo(ctx, req.ActId, req.TaskId)
	if err != nil {
		if err == xsql.ErrNoRows {
			err = xecode.Errorf(xecode.RequestErr, "任务不存在")
		}
		log.Errorc(ctx, "[GetMissionTaskInfo][GetActivityTaskInfo][Error], err:%+v", err)
		return
	}
	stockResp, err := s.stockSvr.QueryStockRecord(ctx, resp.StockId, true)
	if err != nil {
		log.Errorc(ctx, "[GetMissionTaskInfo][GetStocksByIds][Error], err:%+v", err)
		return
	}
	resp.StockConfig = &v1.TaskStockConfig{
		CycleLimit: stockResp.CycleLimit,
	}
	return
}

// GetMissionTaskCompleteStatus 任务活动下某个用户的完成状态
func (s *Service) GetMissionTaskCompleteStatus(ctx context.Context, req *v1.GetMissionTaskCompleteStatusReq) (resp *v1.GetMissionTaskCompleteStatusResp, err error) {
	resp = new(v1.GetMissionTaskCompleteStatusResp)
	return
}

func (s *Service) DelMissionTask(ctx context.Context, req *v1.DelMissionTaskReq) (resp *v1.NoReply, err error) {
	resp = new(v1.NoReply)
	if req == nil || req.ActId == 0 || req.TaskId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	_, err = s.dao.GetActivityTaskInfo(ctx, req.ActId, req.TaskId)
	if err != nil {
		if err == xsql.ErrNoRows {
			err = xecode.Errorf(xecode.RequestErr, "任务不存在")
			return
		}
		return
	}
	err = s.dao.RemoveTasks(ctx, req.ActId, []int64{req.TaskId})
	if err != nil {
		return
	}
	err = s.dao.DelActivityTasksCache(ctx, req.ActId)
	if err != nil && err != redis.ErrNil {
		return
	}
	return
}
