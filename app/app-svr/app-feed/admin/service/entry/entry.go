package entry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/robfig/cron"
	"go-common/library/conf/env"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	pb "go-gateway/app/app-svr/app-feed/admin/api/entry"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	entry "go-gateway/app/app-svr/app-feed/admin/dao/show"
	model "go-gateway/app/app-svr/app-feed/admin/model/entry"
	"go-gateway/app/app-svr/app-feed/ecode"
	"strconv"
	"time"
)

var (
	_timeFormat = "2006-01-02 15:04:05"
)

// Service is entry service
type Service struct {
	dao *entry.Dao
	c   *conf.Config
}

// New new a entry service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao: entry.New(c),
		c:   c,
	}
	if env.DeployEnv != env.DeployEnvPre {
		s.startRegularCheck()
	}
	return
}

func (s *Service) startRegularCheck() {
	var err error
	c := cron.New()

	if err = s.pushNewStateToDataBus(context.Background()); err != nil {
		panic(err)
	}
	err = c.AddFunc("0 */1 * * *", func() {
		//nolint:errcheck
		s.pushNewStateToDataBus(context.Background())
	})
	if err != nil {
		panic(err)
	}
	c.Start()
}

func (s *Service) pushNewStateToDataBus(ctx context.Context) (err error) {
	if data, err := s.GetAppEntryStateForDataBus(ctx); err != nil {
		log.Error("cron running 0: %v", err)
	} else if data == nil {
		log.Info("cron running success: nil-data")
		return err
	} else {
		log.Info("cron get data success")
		eg := errgroup.WithContext(ctx)

		for _, r := range data {
			record := r
			eg.Go(func(ctx context.Context) (e error) {
				if !s.dao.GetEntryPubTaskLock(ctx, record.ID) {
					return nil
				}

				pubId := "app-entry-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + strconv.FormatInt(int64(record.ID), 10)
				if e = s.dao.PubEntryState(pubId, record); e != nil {
					log.Error("[app-entry]cron running err - PubEntryState: %d, %v", record.ID, e)
					return e
				}

				tx := s.dao.DB.Begin()
				if e = s.dao.ToggleEntryTimeSetting(record.ID, record.LoopCount); e != nil {
					log.Error("[app-entry]cron running err - ToggleEntryTimeSetting: %d, %v", record.ID, e)
					return e
				}

				defer (func() {
					if e != nil {
						tx.Rollback()
					} else {
						log.Warn("[app-entry]cron running success")
						tx.Commit()
					}
				})()
				return nil
			})
		}

		if err = eg.Wait(); err != nil {
			return err
		}
	}
	return err
}

func (s *Service) CreateEntry(_ context.Context, req *pb.CreateEntryReq) (reply *empty.Empty, err error) {
	// 检查结束时间间是否早于起效时
	if req.Etime.Time().Unix() <= req.Stime.Time().Unix() {
		err = errors.New("活动开始时间不得晚于结束时间")
		log.Error("create entry error 0: %s", err.Error())
		return nil, ecode.EntryTimeSettingError
	}

	// 检查起效时间是否早于当前时间
	if req.Stime.Time().Unix() < time.Now().Unix() {
		err = errors.New("活动开始时间不得早于当前时间")
		log.Error("create entry error 1: %s", err.Error())
		return nil, ecode.EntryTimeSettingError
	}

	// 检查传入的版本平台参数
	//nolint:ineffassign
	var platforms = ""
	if platBytes, err := json.Marshal(req.Platforms); err != nil {
		log.Error("create entry error 2: %s", err.Error())
		return reply, ecode.EntryParamsError
	} else {
		platforms = string(platBytes)
	}

	tx := s.dao.DB.Begin()
	newEntry := &model.AppEntry{
		EntryName:    req.EntryName,
		STime:        req.Stime,
		ETime:        req.Etime,
		CreatedBy:    req.CreatedBy,
		OnlineStatus: 0,
		Platforms:    platforms,
		TotalLoop:    req.TotalLoop,
	}

	if err = s.dao.CreateEntry(newEntry); err != nil {
		log.Error("create entry error 3: %s", err.Error())
		tx.Rollback()
		return
	}

	for _, state := range req.States {
		newState := &model.AppEntryState{
			StateName:   state.StateName,
			DynamicIcon: state.DynamicIcon,
			StaticIcon:  state.StaticIcon,
			Url:         state.Url,
			EntryID:     newEntry.ID,
			LoopCount:   state.LoopCount,
		}
		// 检查状态的设定
		if state.LoopCount < 0 {
			log.Error("create entry error 4: %s", ecode.EntryParamsError)
			return reply, ecode.EntryParamsError
		}

		if err = s.dao.CreateEntryState(newState); err != nil {
			log.Error("create entry error 5: %s", err.Error())
			tx.Rollback()
			return
		}
	}

	defer tx.Commit()
	return
}

//nolint:gocognit
func (s *Service) EditEntry(ctx context.Context, req *pb.EditEntryReq) (reply *empty.Empty, err error) {
	// 检查结束时间间是否早于起效时
	if req.Etime.Time().Unix() <= req.Stime.Time().Unix() {
		err = errors.New("活动开始时间不得晚于结束时间")
		log.Error("edit entry error 0: %s", err.Error())
		return nil, ecode.EntryTimeSettingError
	}

	// 检查结束时间是否早于当前时间
	if req.Etime.Time().Unix() < time.Now().Unix() {
		err = errors.New("活动结束时间不得早于当前时间")
		log.Error("edit entry error 1: %s", err.Error())
		return nil, ecode.EntryTimeSettingError
	}

	// 检查传入的版本平台信息
	//nolint:ineffassign
	var platforms = ""
	if platBytes, err := json.Marshal(req.Platforms); err != nil {
		log.Error("edit entry error 2: %s", err.Error())
		return reply, ecode.EntryParamsError
	} else {
		platforms = string(platBytes)
	}

	// 获取原有的活动配置
	appEntry := model.AppEntry{}
	if err := s.dao.GetEntryById(req.Id, &appEntry); err != nil {
		log.Error("edit entry error 3: %s", err.Error())
	}

	// 如果原有活动在线，则进行相关检查
	if appEntry.OnlineStatus == 1 {
		checkEG := errgroup.WithContext(ctx)

		// 检查新的时间设定是否会和线上已有的造成冲突, 并且检查是否对plat配置有冲突
		checkEG.Go(func(ctx context.Context) (err error) {
			if conflicts, err := s.dao.CheckEffectiveEntryWithTime(req.Id, req.Stime.Time(), req.Etime.Time()); err != nil {
				log.Error("edit entry online error 4-0: %s", err.Error())
				return err
			} else if len(conflicts) > 0 {
				var platMap = make(map[int32]bool)
				for i := range req.Platforms {
					platMap[req.Platforms[i].Platform] = true
				}
				for _, c := range conflicts {
					var temp []*pb.EffectivePlatform
					if err = json.Unmarshal([]byte(c.Platforms), &temp); err != nil {
						log.Error("edit entry online error 4-1: %s", err.Error())
						return err
					}
					for _, p := range temp {
						if v, ok := platMap[p.Platform]; v && ok {
							err = ecode.EntryOnlineLimit
							log.Error("edit entry online check error 4-2: %s", err.Error())
							return err
						}
					}
				}
			}
			return nil
		})

		// 检查是否有活动正在生效
		checkEG.Go(func(ctx context.Context) (err error) {
			now := time.Now().Unix()
			if appEntry.STime.Time().Unix() <= now && now < appEntry.ETime.Time().Unix() {
				// 如果正在生效则寻找是否有生效的状态
				if ts, err := s.dao.GetPushedTimeSettingsByEntryID(req.Id); err != nil {
					if err.Error() != "-404" {
						log.Error("edit entry online check error 5-0: %s", err)
						return err
					}
				} else if ts != nil && ts.ID != 0 {
					hasOldState := false
					for _, state := range req.States {
						if state.Id == ts.StateID {
							hasOldState = true
							break
						}
					}
					if !hasOldState {
						log.Error("edit entry online check error 5-1: %s", ecode.EntryDeleteOnlineStateError)
						return ecode.EntryDeleteOnlineStateError
					}
				}
			}
			return nil
		})

		if err = checkEG.Wait(); err != nil {
			return nil, err
		}
	}

	tx := s.dao.DB.Begin()
	newEntry := &model.AppEntry{
		BaseModel: model.BaseModel{
			ID: req.Id,
		},
		EntryName: req.EntryName,
		STime:     req.Stime,
		ETime:     req.Etime,
		CreatedBy: req.CreatedBy,
		Platforms: platforms,
		TotalLoop: req.TotalLoop,
	}
	// 修改活动
	if err = s.dao.EditEntry(ctx, newEntry); err != nil {
		log.Error("edit entry error 6: %s", err.Error())
		tx.Rollback()
		return
	}

	// 如果编辑的时候，活动已经结束,则消除所有已推送的状态切换
	if time.Now().Unix() >= appEntry.ETime.Time().Unix() {
		if err = s.dao.DisableTimeSettingByEntryID(req.Id); err != nil {
			log.Error("edit entry error 7: %s", err.Error())
			tx.Rollback()
			return
		}
	}

	// 消除所有未推送且待推送时间大于新的结束时间的状态切换：为了缩短时间
	if req.Etime.Time().Unix() < appEntry.ETime.Time().Unix() {
		if err = s.dao.DisableTimeSettingByEntryIDWithEtime(req.Id, req.Etime.Time()); err != nil {
			log.Error("edit entry error 8: %s", err.Error())
			tx.Rollback()
			return
		}
	}

	// 消除所有未推送且待推送时间大于新的结束时间的状态切换：为了缩短时间
	if req.Stime.Time().Unix() > appEntry.STime.Time().Unix() {
		if err = s.dao.DisableTimeSettingByEntryIDWithStime(req.Id, req.Stime.Time()); err != nil {
			log.Error("edit entry error 9: %s", err.Error())
			tx.Rollback()
			return
		}
	}

	// 默认先关闭所有状态，再根据是否是旧的再打开
	if err = s.dao.DeleteEntryState(req.Id); err != nil {
		log.Error("edit entry error 10: %s", err.Error())
		tx.Rollback()
		return
	}
	for _, state := range req.States {
		newState := &model.AppEntryState{
			StateName:   state.StateName,
			DynamicIcon: state.DynamicIcon,
			StaticIcon:  state.StaticIcon,
			Url:         state.Url,
			EntryID:     newEntry.ID,
			LoopCount:   state.LoopCount,
		}
		// 检查活动的相关配置
		if state.LoopCount < 0 {
			err = errors.New("loop can not be negative")
			log.Error("edit entry error 11: %s", err.Error())
			return reply, err
		}

		if state.Id != 0 {
			// 判定活动是否为新建的，如果不是则修改状态
			newState.BaseModel = model.BaseModel{
				ID: state.Id,
			}
			newState.IsDeprecated = 0
			// 判定活动和状态是否为一对
			if isPair, err := s.dao.CheckEntryStatePair(req.Id, state.Id); err != nil {
				log.Error("edit entry error 12: %s", err)
				return reply, err
			} else {
				if !isPair {
					log.Error("edit entry error 13: %s", fmt.Sprintf("entry id: %d和state id: %d不匹配", req.Id, state.Id))
					return reply, ecode.EntryParamsError
				}
			}
			if err = s.dao.EditState(newState); err != nil {
				log.Error("edit entry error 14: %s", err.Error())
				tx.Rollback()
				return reply, err
			}
		} else {
			// 如果是，则直接新建
			if err = s.dao.CreateEntryState(newState); err != nil {
				log.Error("edit entry error 15: %s", err.Error())
				tx.Rollback()
				return reply, err
			}
		}
	}

	defer tx.Commit()
	return
}

func (s *Service) DeleteEntry(_ context.Context, req *pb.DeleteEntryReq) (reply *empty.Empty, err error) {
	appEntry := &model.AppEntry{}
	// 检查是否存在线上的入口配置
	if err = s.dao.GetEntryById(req.Id, appEntry); err != nil {
		return
	} else {
		if appEntry.OnlineStatus == 1 {
			err = errors.New("请先将在线的任务下线后再删除")
			log.Error("delete entry error 0: %s", err.Error())
			return nil, ecode.EntryOfflineBeforeDelete
		}
	}
	tx := s.dao.DB.Begin()
	// 删除活动
	if err = s.dao.DeleteEntry(req.Id); err != nil {
		log.Error("delete entry error 1: %s", err.Error())
		tx.Rollback()
		return
	}
	// 删除相关状态
	if err = s.dao.DeleteEntryState(req.Id); err != nil {
		log.Error("delete entry error 2: %s", err.Error())
		tx.Rollback()
		return
	}
	// 删除相关状态切换
	if err = s.dao.DisableTimeSettingByEntryID(req.Id); err != nil {
		log.Error("delete entry error 3: %s", err.Error())
		tx.Rollback()
		return
	}
	defer tx.Commit()
	return
}

func (s *Service) ToggleEntry(ctx context.Context, req *pb.ToggleEntryOnlineStatusReq) (reply *empty.Empty, err error) {
	// 检查传入参数是否正常
	if req.OnlineStatus != 1 && req.OnlineStatus != 0 {
		err = ecode.EntryParamsError
		log.Error("toggle entry error 0: %s", err.Error())
		return reply, err
	}

	// 如果原有活动在线，则进行相关检查
	if req.OnlineStatus == 1 {
		appEntry := &model.AppEntry{}
		if err := s.dao.GetEntryById(req.Id, appEntry); err != nil {
			log.Error("toggle entry error 1: %s", err.Error())
			return reply, ecode.EntryParamsError
		}
		if appEntry.OnlineStatus == int32(req.OnlineStatus) {
			err = ecode.EntryParamsError
			log.Error("toggle entry error 2: %s", err.Error())
			return reply, err
		}

		// 检查新的时间设定是否会和线上已有的造成冲突, 并且检查是否对plat配置有冲突
		if conflicts, err := s.dao.CheckEffectiveEntryWithTime(req.Id, appEntry.STime.Time(), appEntry.ETime.Time()); err != nil {
			log.Error("toggle entry error 3: %s", err.Error())
			return nil, err
		} else if len(conflicts) > 0 {
			var (
				platMap   = make(map[int32]bool)
				mainPlats []*pb.EffectivePlatform
			)
			if err = json.Unmarshal([]byte(appEntry.Platforms), &mainPlats); err != nil {
				log.Error("toggle entry online error 4: %s", err.Error())
				return nil, err
			}
			for _, p := range mainPlats {
				platMap[p.Platform] = true
			}
			for _, c := range conflicts {
				var temp []*pb.EffectivePlatform
				if err = json.Unmarshal([]byte(c.Platforms), &temp); err != nil {
					log.Error("toggle entry online error 4-1: %s", err.Error())
					return nil, err
				}
				for _, p := range temp {
					if v, ok := platMap[p.Platform]; v && ok {
						err = ecode.EntryOnlineLimit
						log.Error("toggle entry online check error 4-2: %s", err)
						return nil, err
					}
				}
			}
		}
	}

	tx := s.dao.DB.Begin()
	// 切换活动线上状态
	if err = s.dao.ToggleEntry(ctx, req.Id, int32(req.OnlineStatus)); err != nil {
		log.Error("toggle entry error 5: %s", err.Error())
		tx.Rollback()
		return
	}
	// 每次切换状态使得之前所有的切换设定都失效
	if err = s.dao.DisableTimeSettingByEntryID(req.Id); err != nil {
		log.Error("toggle entry error 6: %s", err.Error())
		tx.Rollback()
		return
	}
	defer tx.Commit()
	return
}

func (s *Service) GetEntryList(_ context.Context, req *pb.GetEntryListReq) (reply *pb.GetEntryListRep, err error) {
	var count int32
	if count, err = s.dao.GetEntryCount(); err != nil {
		return
	} else if count == 0 {
		reply = &pb.GetEntryListRep{
			Items: make([]*pb.AppEntry, 0),
			Page: &pb.Page{
				Total:    0,
				PageSize: req.PageSize,
				PageNum:  req.PageNum,
			},
		}
		return
	}
	if entries, err := s.dao.GetEntryList(req.PageSize, req.PageNum); err != nil {
		log.Error("[app-entry]get entry list error - GetEntryList: %s", err)
		return nil, err
	} else {
		reply = &pb.GetEntryListRep{
			Items: make([]*pb.AppEntry, len(entries)),
			Page: &pb.Page{
				Total:    count,
				PageSize: req.PageSize,
				PageNum:  req.PageNum,
			},
		}
		now := time.Now()
		for i, e := range entries {
			var (
				plats    []*pb.EffectivePlatform
				states   []*model.AppEntryState
				sentLoop int32
			)
			if err = json.Unmarshal([]byte(e.Platforms), &plats); err != nil {
				log.Error("[app-entry]get entry list error - Unmarshal: %s", err)
				return nil, err
			}
			if states, err = s.dao.GetStatesByEntryID(e.ID); err != nil {
				log.Error("[app-entry]get entry list error - GetStatesByEntryID: %s", err)
				return nil, err
			}

			// 获取该入口已经推送过的动画次数
			if sentLoop, err = s.dao.GetSentLoopByEntryID(e.ID); err != nil {
				log.Error("[app-entry]get entry list error - GetSentLoopByEntryID: %s, %v", err, e.ID)
				return nil, err
			}
			reply.Items[i] = &pb.AppEntry{
				Id:           e.ID,
				EntryName:    e.EntryName,
				OnlineStatus: pb.OnlineStatus(e.OnlineStatus),
				States:       make([]*pb.AppEntryState, len(states)),
				STime:        e.STime,
				ETime:        e.ETime,
				Platforms:    plats,
				CreatedBy:    e.CreatedBy,
				TotalLoop:    e.TotalLoop,
				SentLoop:     sentLoop,
			}
			for j, s := range states {
				reply.Items[i].States[j] = &pb.AppEntryState{
					Id:          s.ID,
					StateName:   s.StateName,
					Url:         s.Url,
					StaticIcon:  s.StaticIcon,
					DynamicIcon: s.DynamicIcon,
					LoopCount:   s.LoopCount,
				}
			}
			if e.OnlineStatus == 1 && e.STime.Time().Unix() <= now.Unix() && e.ETime.Time().Unix() > now.Unix() {
				if lastST, err := s.dao.GetPushedTimeSettingsByEntryID(e.ID); err != nil {
					if err.Error() == "-404" {
						reply.Items[i].CurrentState = 0
					} else {
						log.Error("[app-entry]get entry list error - GetPushedTimeSettingsByEntryID: %s", err)
						return nil, err
					}
				} else {
					reply.Items[i].CurrentState = lastST.StateID
				}
			}
		}
	}
	return
}

func (s *Service) SetNextState(_ context.Context, req *pb.SetNextStateReq) (reply *empty.Empty, err error) {
	// 检查起效时间是否早于当前时间
	if req.Stime.Time().Unix() < time.Now().Unix() {
		err = errors.New("起效时间不能早于当前时间")
		log.Error("SetNextState error 0: %s", err)
		return reply, ecode.EntryTimeSettingError
	}

	currentEntry := model.AppEntry{}
	if err := s.dao.GetEntryById(req.EntryID, &currentEntry); err != nil {
		log.Error("SetNextState error 1: %s", err)
		return reply, err
	} else {
		if currentEntry.OnlineStatus == 1 {
			// 已在线的，检查切换时间是否在活动周期内
			stime := req.Stime.Time().Unix()
			if !(currentEntry.STime.Time().Unix() <= stime && stime < currentEntry.ETime.Time().Unix()) {
				err = errors.New("时间设定冲突，请确认入口配置起效时间！")
				log.Error("SetNextState error 2: %s", err)
				return reply, ecode.EntryTimeSettingError
			}
		} else {
			// 未在线的提醒要上线
			err = errors.New("该入口尚未上线，请先上线！")
			log.Error("SetNextState error 3: %s", err)
			return reply, ecode.EntryIsOffline
		}
	}

	// 检查状态是否和活动是匹配的
	if isPair, err := s.dao.CheckEntryStatePair(req.EntryID, req.StateID); err != nil {
		log.Error("SetNextState error 4: %s", err)
		return reply, err
	} else {
		if !isPair {
			err = errors.New("entry id和state id不匹配")
			log.Error("SetNextState error 5: %s", err)
			return reply, ecode.EntryParamsError
		}
	}

	// 删除该活动之前的推送状态
	tx := s.dao.DB.Begin()
	if err = s.dao.DisableNotPushedTimeSettingByEntryID(req.EntryID); err != nil {
		log.Error("SetNextState error 6: %s", err)
		tx.Rollback()
		return reply, err
	}

	// 新建活动的推送状态
	newTimeSetting := &model.AppEntryTimeSetting{
		CreatedBy: req.CreatedBy,
		STime:     req.Stime,
		StateID:   req.StateID,
		EntryID:   req.EntryID,
	}
	if err = s.dao.CreateTimeSetting(newTimeSetting); err != nil {
		log.Error("SetNextState error 7: %s", err)
		tx.Rollback()
		return reply, err
	}
	defer tx.Commit()
	return
}

func (s *Service) GetTimeSettingList(_ context.Context, req *pb.GetTimeSettingListReq) (reply *pb.GetTimeSettingListRep, err error) {
	stateNameMap := make(map[int32]string)
	if states, err := s.dao.GetStatesByEntryID(req.EntryID); err != nil {
		return reply, err
	} else {
		for _, s := range states {
			stateNameMap[s.ID] = s.StateName
		}
	}
	if timeSettings, err := s.dao.GetTimeSettingByEntryID(req.EntryID); err != nil {
		return reply, err
	} else {
		reply = &pb.GetTimeSettingListRep{
			Items: make([]*pb.AppEntryTimeSetting, len(timeSettings)),
		}
		for i, s := range timeSettings {
			reply.Items[i] = &pb.AppEntryTimeSetting{
				Id:           s.ID,
				StateName:    stateNameMap[s.StateID],
				STime:        s.STime,
				CreatedBy:    s.CreatedBy,
				PushTime:     s.PushTime,
				CTime:        s.CTime,
				EntryId:      s.EntryID,
				StateId:      s.StateID,
				IsDeprecated: s.IsDeprecated,
			}
			formattedPushTime := s.PushTime.Time().Format(_timeFormat)
			if formattedPushTime != "2009-12-31 23:59:59" {
				reply.Items[i].PushTime = s.PushTime
			}
		}
	}
	return
}

func (s *Service) GetAppEntryStateForDataBus(ctx context.Context) (result []*pb.AppEntryForDataBus, err error) {
	if entries, err := s.dao.GetEffectiveEntry(ctx); err != nil {
		if err.Error() == "-404" {
			return nil, nil
		} else {
			log.Error("get current app entry for databus error 0: %s", err)
			return nil, err
		}
	} else {
		eg := errgroup.WithContext(ctx)

		for i := range entries {
			en := entries[i]
			eg.Go(func(ctx context.Context) (e error) {
				var (
					ts       *model.AppEntryTimeSetting
					state    *model.AppEntryState
					plats    []*pb.EffectivePlatform
					sentLoop int32
				)
				if ts, e = s.dao.GetNotPushedTimeSettingByEntryID(en.ID); e != nil {
					if e.Error() == "-404" {
						return nil
					} else {
						log.Error("GetAppEntryStateForDataBus error - GetNotPushedTimeSettingByEntryID: %s", e)
						return e
					}
				}
				if e = json.Unmarshal([]byte(en.Platforms), &plats); e != nil {
					log.Error("GetAppEntryStateForDataBus error - Unmarshal: %s", e)
					return e
				}

				// 尚未起效返回空
				if ts.STime.Time().Unix() > time.Now().Unix() {
					return nil
				}

				if state, e = s.dao.GetStateByID(ts.StateID); e != nil {
					// 该状态已经删除返回空
					if e.Error() == "-404" {
						return nil
					} else {
						log.Error("GetAppEntryStateForDataBus error -GetNotPushedTimeSettingByEntryID: %s, %v", e, ts.StateID)
						return e
					}
				}

				// 获取该入口已经推送过的动画次数
				if sentLoop, e = s.dao.GetSentLoopByEntryID(en.ID); err != nil {
					log.Error("GetAppEntryStateForDataBus error - GetSentLoopByEntryID: %s, %v", e, en.ID)
					return e
				}
				loop := state.LoopCount
				restLoop := en.TotalLoop - sentLoop
				if loop > restLoop {
					loop = restLoop
				}

				// 已经生效则推送消息
				result = append(result, &pb.AppEntryForDataBus{
					ID:          ts.ID,
					StateID:     ts.StateID,
					StateName:   state.StateName,
					StaticIcon:  state.StaticIcon,
					DynamicIcon: state.DynamicIcon,
					Url:         state.Url,
					LoopCount:   loop,
					STime:       ts.STime,
					ETime:       en.ETime,
					Platforms:   plats,
					EntryName:   en.EntryName,
				})
				return nil
			})
		}
		err = eg.Wait()
	}
	return result, err
}
