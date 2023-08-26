package bnj

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	xtime "go-common/library/time"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/bnj"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/interface/model/task"

	"go-common/library/sync/errgroup.v2"
)

const (
	_blockStatus      = 1
	_unlockedStatus   = 1
	_awardIsHide      = 1
	_awardHasMore     = 1
	_receivedStatus   = 1
	_rewardStatus     = 1
	_linkTextUnlocked = "蓄力中"
	_linkTextNotAward = "领取"
	_lineTextHotpot   = "已升级"
	_linkTextReward   = "已领取"
	_linkTextWait     = "待开奖"
	_linkTextResult   = "前往查看中奖结果"
	// 动态抽奖
	_awardTypeDynamic = 1
	_awardTypeHotpot  = 2
	_awardFinalType   = 7
	_maxIncrCount     = 20
	_incrMsgFmt       = "往锅里加入了【%s】"
	_incrRareMsgFmt   = "获得稀有食材【%s】"
	_decrMsgFmt       = "刚刚%s"
	_selfPrefix       = "我"
	_awardFinalID     = 100
)

// Bnj20Main bnj 2020 preview main.
func (s *Service) Bnj20Main(c context.Context, mid int64) (*bnj.MainBnj20, error) {
	nowTs := time.Now().Unix()
	if nowTs < s.c.Bnj2020.Stime.Unix() {
		return nil, ecode.ActivityNotStart
	}
	reservedCnt := s.bnj20Mem.AppointCnt
	data := &bnj.MainBnj20{
		Sid:           s.c.Bnj2020.Sid,
		ReservedCount: reservedCnt,
		Award:         new(bnj.AwardBnj20),
		Value:         s.bnj20Mem.HotpotValue,
		HotPotLevel:   s.bnj20Mem.HotpotLevel,
	}
	// 游戏主动屏蔽
	if s.c.Bnj2020.BlockGame != 0 {
		data.BlockGame = _blockStatus
		data.BlockGameAction = _blockStatus
	}
	if s.c.Bnj2020.BlockGameAction != 0 {
		data.BlockGameAction = _blockStatus
	}
	// 游戏结束
	if s.bnj20Mem.GameFinish == 1 {
		data.TimelinePic = s.c.Bnj2020.TimelinePic
		data.H5TimelinePic = s.c.Bnj2020.H5TimelinePic
		data.ShareTimelinePic = s.c.Bnj2020.ShareTimelinePic
	}
	arcs := s.bnj20Mem.Arcs
	for _, v := range s.c.Bnj2020.Info {
		if v.Publish.Unix() > nowTs {
			break
		}
		tmp := &bnj.InfoBnj20{
			Name:         v.Name,
			Pic:          v.Pic,
			H5Pic:        v.H5Pic,
			Detail:       v.Detail,
			H5Detail:     v.H5Detail,
			SharePic:     v.SharePic,
			H5SharePic:   v.H5SharePic,
			DynamicPic:   v.DynamicPic,
			H5DynamicPic: v.H5DynamicPic,
		}
		for _, aid := range v.Aids {
			if arc, ok := arcs[aid.Aid]; ok && arc.IsNormal() {
				tmp.Arcs = append(tmp.Arcs, &bnj.ArcBnj20{
					Aid:        arc.Aid,
					Title:      arc.Title,
					Pic:        arc.Pic,
					Owner:      arc.Author,
					Stat:       bnj.ArcStatBnj20{View: arc.Stat.View},
					RcmdReason: aid.RcmdReason,
				})
			}
		}
		data.Infos = append(data.Infos, tmp)
	}
	if len(data.Infos) == 0 {
		data.Infos = make([]*bnj.InfoBnj20, 0)
	}
	var (
		awardState   map[int64]int
		reserveState *like.HasReserve
		taskState    map[string]*task.UserTask
	)
	if mid > 0 {
		group := errgroup.WithContext(c)
		group.Go(func(ctx context.Context) error {
			var reserveErr error
			reserveState, reserveErr = s.likeDao.ReserveOnly(ctx, s.c.Bnj2020.Sid, mid)
			if reserveErr != nil {
				log.Error("Bnj20Main s.likeDao.ReserveOnly mid(%d) error(%v)", mid, reserveErr)
			}
			if reserveState != nil {
				data.HasReserved = reserveState.State
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			var awardErr error
			if awardState, awardErr = s.dao.CacheRewards(ctx, mid, s.c.Bnj2020.Sid); awardErr != nil {
				log.Error("Bnj20Main s.dao.CacheRewards mid(%d) error(%v)", mid, awardErr)
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			var taskErr error
			if taskState, taskErr = s.taskDao.UserTaskState(ctx, s.previewTasks, mid, task.BusinessAct, s.c.Bnj2020.Sid, nowTs); taskErr != nil {
				log.Error("Bnj20Material s.taskDao.UserTaskState mid(%d) sid(%d) error(%v)", mid, s.c.Bnj2020.Sid, taskErr)
			}
			return nil
		})
		group.Wait()
	}
	preUnlock := false
	for _, v := range s.c.Bnj2020.Award {
		tmp := &bnj.AwardItemBnj20{
			ID:      v.ID,
			Name:    v.Name,
			Pic:     v.Pic,
			CardPic: v.CardPic,
			Type:    v.Type,
			Count:   v.Count,
		}
		if reservedCnt >= v.Count {
			tmp.HasUnlocked = _unlockedStatus
			preUnlock = true
			switch v.Type {
			case _awardTypeDynamic:
				tmp.LinkText = _linkTextWait
				if v.LinkURL != "" {
					tmp.HasReward = _rewardStatus
					tmp.LinkText = _linkTextResult
					tmp.LinkURL = v.LinkURL
				}
			case _awardTypeHotpot:
				tmp.LinkText = _lineTextHotpot
			default:
				tmp.LinkURL = v.LinkURL
				tmp.LinkText = _linkTextNotAward
				state := awardState[v.ID]
				taskState := taskState[bnjTaskKey(v.TaskID)]
				if state == _rewardStatus || (taskState != nil && taskState.Count > 0) {
					tmp.HasReward = _rewardStatus
					tmp.LinkText = _linkTextReward
				}
			}
		} else {
			// 隐藏奖励未解锁不展示详情
			if v.IsHide == _awardIsHide {
				if !preUnlock {
					data.Award.HasMore = _awardHasMore
					break
				}
			}
			preUnlock = false
			tmp.LinkText = _linkTextUnlocked
		}
		data.Award.List = append(data.Award.List, tmp)
	}
	finalState := awardState[_awardFinalID]
	finalTaskState := taskState[bnjTaskKey(s.c.Bnj2020.FinalTaskID)]
	if finalState == _rewardStatus || (finalTaskState != nil && finalTaskState.Count > 0) {
		data.Award.FinalHasAward = _rewardStatus
	}
	return data, nil
}

// Bnj20Reward reward bnj 2020 award.
func (s *Service) Bnj20Reward(c context.Context, mid, id int64) (err error) {
	nowTs := time.Now().Unix()
	if nowTs > s.c.Bnj2020.AwardEndTime.Unix() {
		err = ecode.ActivityOverEnd
		return
	}
	award, err := s.awardCheck(id)
	if err != nil {
		return
	}
	if err = s.reserveStateCheck(c, mid); err != nil {
		return
	}
	reservedSid := s.c.Bnj2020.Sid
	var awardState, taskState int
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		hasAward, hasAwardErr := s.dao.AddCacheRewards(ctx, mid, reservedSid, award.ID)
		if hasAwardErr != nil {
			return ecode.ActivityBnjRewardFail
		}
		if !hasAward {
			awardState = _rewardStatus
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		receiveState, stateErr := s.taskDao.UserTaskState(ctx, s.previewTasks, mid, task.BusinessAct, s.c.Bnj2020.Sid, nowTs)
		if stateErr != nil {
			log.Error("Bnj20Reward s.taskDao.UserTaskState mid(%d) sid(%d) error(%v)", mid, s.c.Bnj2020.Sid, err)
			return nil
		}
		if state, ok := receiveState[bnjTaskKey(award.TaskID)]; ok && state != nil && state.Count > 0 {
			taskState = _rewardStatus
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		log.Error("Bnj20Reward group.Wait() mid(%d) award(%+v) error(%v)", mid, award, err)
		return
	}
	if awardState == _rewardStatus || taskState == _rewardStatus {
		err = ecode.ActivityBnjHasReward
		return
	}
	action := &bnj.AwardAction{
		Mid:          mid,
		ID:           award.ID,
		Type:         award.Type,
		SourceID:     award.SourceID,
		SourceExpire: award.SourceExpire,
		TaskID:       award.TaskID,
		Mirror:       metadata.String(c, metadata.Mirror),
	}
	if err = s.bnjAwardPub.Send(context.Background(), fmt.Sprintf("award_%d", mid), action); err != nil {
		log.Error("Bnj20Reward award(%+v) mid(%d) error(%+v)", award, mid, err)
		err = ecode.ActivityBnjRewardFail
		if e := s.dao.DelCacheRewards(c, mid, reservedSid, award.ID); e != nil {
			log.Error("Bnj20Reward s.dao.DelCacheRewards mid(%d) sid(%d) ID(%d) error(%v)", mid, reservedSid, award.ID, e)
		}
	}
	return
}

// Bnj20Material get bnj 20 material list.
func (s *Service) Bnj20Material(c context.Context, mid int64) *bnj.MaterialRes {
	res := &bnj.MaterialRes{
		Blocked:    s.c.Bnj2020.BlockMaterial,
		NormalList: s.c.Bnj2020.NormalList,
		SpecialList: &bnj.MaterialSpec{
			Good:     s.c.Bnj2020.SpecialList.Good,
			GoodDesc: s.c.Bnj2020.SpecialList.GoodDesc,
			Bad:      s.c.Bnj2020.SpecialList.Bad,
			BadDesc:  s.c.Bnj2020.SpecialList.BadDesc,
		},
	}
	nowTs := time.Now().Unix()
	for _, v := range s.c.Bnj2020.RareList {
		material := &bnj.Material{
			ID:       v.ID,
			Pic:      v.Pic,
			H5Pic:    v.H5Pic,
			Name:     v.Name,
			Desc:     v.Desc,
			SharePic: v.SharePic,
			CardPic:  v.CardPic,
			TaskID:   v.TaskID,
		}
		if nowTs >= v.Publish.Unix() {
			material.HasUnlocked = _unlockedStatus
			res.RareList = append(res.RareList, material)
		}
	}
	if mid > 0 && len(res.RareList) > 0 {
		var receiveState map[string]*task.UserTask
		group := errgroup.WithContext(c)
		// has received
		group.Go(func(ctx context.Context) error {
			var err error
			if receiveState, err = s.taskDao.UserTaskState(ctx, s.previewTasks, mid, task.BusinessAct, s.c.Bnj2020.Sid, nowTs); err != nil {
				log.Error("Bnj20Material s.taskDao.UserTaskState mid(%d) sid(%d) error(%v)", mid, s.c.Bnj2020.Sid, err)
			}
			return nil
		})
		// has red dot
		group.Go(func(ctx context.Context) error {
			redDotState, err := s.dao.CacheClearRedDot(ctx, mid, s.c.Bnj2020.Sid)
			if err != nil {
				log.Error("Bnj20Material s.dao.CacheClearRedDot mid(%d) sid(%d) error(%v)", mid, s.c.Bnj2020.Sid, err)
				return nil
			}
			if redDotState > 0 {
				res.RareHotDot = 1
			}
			return nil
		})
		group.Wait()
		for _, v := range res.RareList {
			if v.HasUnlocked != _unlockedStatus {
				continue
			}
			if idState, ok := receiveState[bnjTaskKey(v.TaskID)]; ok && idState != nil && idState.Count > 0 {
				v.HasReceived = _receivedStatus
			}
		}
	}
	return res
}

// Bnj20MaterialUnlock receive special material.
func (s *Service) Bnj20MaterialUnlock(c context.Context, mid, id int64) (err error) {
	nowTs := time.Now().Unix()
	material, err := s.materialIDCheck(id, nowTs)
	if err != nil {
		return
	}
	if err = s.reserveStateCheck(c, mid); err != nil {
		return
	}
	err = s.unlockMaterial(c, mid, nowTs, material)
	return
}

// Bnj20MaterialRedDot clear bnj 20 material red dot.
func (s *Service) Bnj20MaterialRedDot(c context.Context, mid int64) (err error) {
	err = s.dao.DelCacheClearRedDot(c, mid, s.c.Bnj2020.Sid)
	if err != nil {
		err = ecode.ActivityRedDotClearFail
	}
	return
}

// Bnj20HotpotIncrease increase hotpot value.
func (s *Service) Bnj20HotpotIncrease(c context.Context, mid, count int64) (res *bnj.IncreaseMaterial, err error) {
	if atomic.LoadInt64(&s.bnj20Mem.GameFinish) == 1 {
		err = ecode.ActivityGameFinish
		return
	}
	// ensure submit times
	if count := s.dao.HotpotIncreaseCount(c, mid); count >= 86400 {
		log.Warn("Reject to increase hotpot value: mid: %d, count: %d", mid, count)
		return
	}
	defer s.dao.IncreaseHotpotCount(c, mid)
	if err = s.reserveStateCheck(c, mid); err != nil {
		return
	}
	if count > _maxIncrCount {
		count = _maxIncrCount
	}
	nowTs := time.Now().Unix()
	var chance []*conf.IncreaseBnj20
	// get now day increase conf
	for _, v := range s.c.Bnj2020.Info {
		if nowTs >= v.Publish.Unix() {
			chance = v.Increase
		}
	}
	var (
		randMax   int64 = 10000000
		material  *conf.Bnj20Material
		materials []string
	)
	randNum := rand.Int63n(randMax)
	materials = append(materials, s.c.Bnj2020.SpecialList.Good...)
	materials = append(materials, s.c.Bnj2020.SpecialList.Bad...)
	for _, v := range chance {
		right := int64(float64(randMax) - (math.Pow(1-v.P, float64(count)) * float64(randMax)))
		// 命中食材
		if randNum < right {
			// 命中稀有食材
			if len(v.IDs) > 0 {
				id := v.IDs[int(randNum%int64(len(v.IDs)))]
				if id == 0 {
					name := materials[int(randNum%int64(len(materials)))]
					material = &conf.Bnj20Material{Name: name}
					break
				}
				material = s.rareMaterial[id]
				break
			}
		}
	}
	var msg string
	increaseNum := atomic.LoadInt64(&s.bnj20Mem.HotpotLevel) * count
	if material != nil {
		msg = fmt.Sprintf(_incrMsgFmt, material.Name)
		if material.ID > 0 {
			msg = fmt.Sprintf(_incrRareMsgFmt, material.Name)
		}
		res = &bnj.IncreaseMaterial{
			ID:       material.ID,
			Pic:      material.Pic,
			H5Pic:    material.H5Pic,
			CardPic:  material.CardPic,
			SharePic: material.SharePic,
			Name:     material.Name,
			Desc:     material.Desc,
		}
		if material.ID > 0 {
			s.cache.Do(c, func(ctx context.Context) {
				s.unlockMaterial(ctx, mid, nowTs, material)
			})
		}
	}
	action := &bnj.Action{Mid: mid, Type: bnj.ActionTypeIncr, Num: increaseNum, Message: msg, Ts: nowTs}
	if err = s.bnjPub.Send(c, fmt.Sprintf("increase_%d", mid), action); err != nil {
		log.Error("Bnj20HotpotIncrease s.bnjPub.Send mid(%d) action(%+v) error(%v)", mid, action, err)
		err = nil
	}
	return
}

// Bnj20HotpotIncrease decrease hotpot value.
func (s *Service) Bnj20HotpotDecrease(c context.Context, mid int64) (ttl, decreaseNum int64, msg string, err error) {
	if atomic.LoadInt64(&s.bnj20Mem.GameFinish) == 1 {
		err = ecode.ActivityGameFinish
		return
	}
	value, err := s.dao.AddCacheDecreaseCD(c, mid, s.c.Bnj2020.DecreaseCD)
	if err != nil {
		log.Error("Bnj20HotpotDecrease s.dao.CacheResetCD(%d) error(%v) value(%v)", mid, err, value)
		err = nil
		return
	}
	ttl = int64(s.c.Bnj2020.DecreaseCD)
	if !value {
		if ttl, err = s.dao.TTLCacheDecreaseCD(c, mid); err != nil {
			log.Error("Bnj20HotpotDecrease s.dao.TTLCacheDecreaseCD(%d) error(%v)", mid, err)
			err = nil
		}
		// redis expire set fail
		if ttl == -1 {
			ttl = int64(s.c.Bnj2020.DecreaseCD)
			s.cache.Do(c, func(ctx context.Context) {
				s.dao.DelCacheDecreaseCD(ctx, mid)
			})
		}
		return
	}
	if err = s.reserveStateCheck(c, mid); err != nil {
		return
	}
	randNum := rand.Intn(10000)
	nowTs := time.Now().Unix()
	var chance []*conf.DecreaseBnj20
	// get now day
	for _, v := range s.c.Bnj2020.Info {
		if nowTs >= v.Publish.Unix() {
			chance = v.Decrease
		}
	}
	var actionMsg string
	for _, v := range chance {
		if randNum >= v.Left && randNum < v.Right {
			decreaseNum = v.Value
			msg = bnj.DecreaseMsgTypes[v.Type]
			switch {
			case v.Type == bnj.DecrLevelOne:
				msg = _selfPrefix + msg
			// 普通食材展示食材名
			case v.Type == bnj.DecrLevelTwo:
				msg = fmt.Sprintf(_selfPrefix+msg, s.c.Bnj2020.NormalList[randNum%len(s.c.Bnj2020.NormalList)])
			// 广播彩蛋type过滤
			case v.Type >= bnj.DecrLevelThree:
				actionMsg = fmt.Sprintf(_decrMsgFmt, msg)
			}
			break
		}
	}
	// 无广播msg过滤
	if actionMsg == "" {
		return
	}
	action := &bnj.Action{Mid: mid, Type: bnj.ActionTypeDecr, Num: decreaseNum, Message: actionMsg, Ts: nowTs}
	if err = s.bnjPub.Send(c, fmt.Sprintf("decrease_%d", mid), action); err != nil {
		log.Error("Bnj20HotpotDecrease s.bnjPub.Send mid(%d) action(%+v) error(%v)", mid, action, err)
		err = nil
	}
	return
}

func (s *Service) materialIDCheck(id, nowTs int64) (material *conf.Bnj20Material, err error) {
	material, ok := s.rareMaterial[id]
	if !ok || material == nil {
		err = xecode.RequestErr
		return
	}
	if material.Publish.Unix() > nowTs {
		err = ecode.ActivityUnlocked
	}
	return
}

func (s *Service) unlockMaterial(c context.Context, mid, nowTs int64, material *conf.Bnj20Material) (err error) {
	previewTask, ok := s.previewTasks[material.TaskID]
	if !ok {
		log.Error("unlockMaterial material(%+v) conf error", material)
		err = xecode.RequestErr
		return
	}
	taskState, err := s.taskDao.UserTaskState(c, map[int64]*task.Task{previewTask.ID: previewTask}, mid, task.BusinessAct, s.c.Bnj2020.Sid, nowTs)
	if err != nil {
		log.Error("unlockMaterial s.taskDao.UserTaskState mid(%d) sid(%d) error(%v)", mid, s.c.Bnj2020.Sid, err)
		return
	}
	state, ok := taskState[bnjTaskKey(material.TaskID)]
	if ok && state != nil && state.Count > 0 {
		err = ecode.ActivityHasReceived
		return
	}
	// add log
	if err = s.taskDao.AddUserTaskLog(c, mid, task.BusinessAct, previewTask.ID, s.c.Bnj2020.Sid, 0); err != nil {
		log.Error("unlockMaterial s.taskDao.AddUserTaskLog mid(%d) taskID(%d) foreignID(%d) error(%v)", mid, previewTask.ID, s.c.Bnj2020.Sid, err)
		return
	}
	var (
		count  int64 = 1
		finish int64
	)
	if previewTask.FinishCount == count {
		finish = task.HasFinish
	}
	if err = s.taskDao.AddUserTaskState(c, mid, task.BusinessAct, previewTask.ID, s.c.Bnj2020.Sid, 0, count, finish, 0); err != nil {
		log.Error("unlockMaterial s.taskDao.AddUserTaskState mid(%d) taskID(%d) count(%d) finish(%d) error(%v)", mid, s.c.Bnj2020.Sid, count, finish, err)
		return
	}
	s.cache.Do(c, func(ctx context.Context) {
		upTask := &task.UserTask{
			Mid:        mid,
			BusinessID: task.BusinessAct,
			TaskID:     previewTask.ID,
			ForeignID:  s.c.Bnj2020.Sid,
			Round:      0,
			Count:      count,
			Finish:     finish,
			Award:      0,
			Ctime:      xtime.Time(nowTs),
		}
		s.taskDao.SetCacheUserTaskState(ctx, upTask, mid, upTask.BusinessID, upTask.ForeignID)
		// 新材料，恢复红点
		s.dao.AddCacheClearRedDot(ctx, mid, s.c.Bnj2020.Sid)
	})
	return
}

func (s *Service) awardCheck(id int64) (award *conf.Bnj20Award, err error) {
	if id == _awardFinalID {
		// 游戏未完成
		if s.bnj20Mem.GameFinish != 1 {
			err = ecode.ActivityBnjSubLow
			return
		}
		award = &conf.Bnj20Award{ID: _awardFinalID, Type: _awardFinalType, TaskID: s.c.Bnj2020.FinalTaskID}
		return
	}
	for _, v := range s.c.Bnj2020.Award {
		if v.ID == id {
			award = v
			break
		}
	}
	if award == nil || award.Type == _awardTypeDynamic || award.Type == _awardTypeHotpot {
		err = xecode.RequestErr
		return
	}
	reservedCnt := s.bnj20Mem.AppointCnt
	if award.Count > reservedCnt {
		err = ecode.ActivityBnjSubLow
	}
	return
}

func (s *Service) reserveStateCheck(c context.Context, mid int64) (err error) {
	reserveState, err := s.likeDao.ReserveOnly(c, s.c.Bnj2020.Sid, mid)
	if err != nil {
		log.Error("s.dao.ReserveOnly(%d,%d) error(%+v)", s.c.Bnj2020.Sid, mid, err)
		err = ecode.ActivityBnjNotSub
		return
	}
	if reserveState == nil || reserveState.ID == 0 || reserveState.State != 1 {
		err = ecode.ActivityBnjNotSub
		return
	}
	return nil
}

func (s *Service) loadBnj20() {
	nowTs := time.Now().Unix()
	var aids []int64
	for _, v := range s.c.Bnj2020.Info {
		if v.Publish.Unix() < nowTs {
			for _, aid := range v.Aids {
				aids = append(aids, aid.Aid)
			}
		}
	}
	if len(aids) > 0 {
		arcsReply, err := client.ArchiveClient.Arcs(context.Background(), &arcapi.ArcsRequest{Aids: aids})
		if err != nil {
			log.Error("bnj20proc s.arcClient.Arcs(%v) error(%v)", aids, err)
			return
		}
		if len(arcsReply.GetArcs()) > 0 {
			tmp := make(map[int64]*arcapi.Arc, len(aids))
			for _, aid := range aids {
				if arc, ok := arcsReply.Arcs[aid]; ok && arc != nil {
					tmp[aid] = arc
				} else {
					log.Error("bnj20proc aid(%d) data(%v)", aid, arc)
					continue
				}
			}
			s.bnj20Mem.Arcs = tmp
		}
	} else {
		log.Error("bnj20proc aids conf error")
		return
	}
	log.Info("loadBnj20() success")
}

func (s *Service) loadBnj20Task() {
	ids, err := s.taskDao.TaskIDs(context.Background(), task.BusinessAct, s.c.Bnj2020.Sid)
	if err != nil {
		log.Error("loadBnj20Task s.taskDao.TaskIDs(%d) error(%v)", s.c.Bnj2020.Sid, err)
		return
	}
	if len(ids) == 0 {
		log.Warn("loadBnj20Task len(ids) == 0")
		return
	}
	tasks, err := s.taskDao.Tasks(context.Background(), ids)
	if err != nil {
		log.Error("loadBnj20Task s.taskDao.Tasks(%v) error(%v)", ids, err)
		return
	}
	if len(tasks) != len(ids) {
		log.Warn("loadBnj20Task len(tasks) != len(ids)")
		return
	}
	s.previewTasks = tasks
	log.Info("loadBnj20Task() success")
}

func bnjTaskKey(taskID int64) string {
	return fmt.Sprintf("%d_0", taskID)
}
