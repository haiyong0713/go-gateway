package like

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	premdl "go-gateway/app/web-svr/activity/interface/model/prediction"
)

const (
	_upState  = 1
	_delState = 0
)

// PreJudge .
func (s *Service) PreJudge(c context.Context, arg *premdl.PreParams) (res map[int64]*premdl.PredictionItem, err error) {
	var (
		subject     *likemdl.SubjectItem
		preIds      []int64
		preInfos    map[int64]*premdl.Prediction
		parentInfos map[int64]*premdl.Prediction
		preDeals    map[int64][]*premdl.Prediction
		lock        sync.Mutex
	)
	if subject, err = s.dao.ActSubject(c, arg.Sid); err != nil {
		log.Error("s.dao.ActSubject(%d) error(%+v)", arg.Sid, err)
		return
	}
	if subject.ID == 0 || subject.Type != likemdl.PREDICTION {
		return
	}
	if preIds, err = s.preDao.PreList(c, arg.Sid); err != nil {
		log.Error("s.dao.PreList(%d) error(%+v)", arg.Sid, err)
		return
	}
	if len(preIds) == 0 {
		s.cache.Do(c, func(ctx context.Context) {
			if e := s.PreListSet(ctx, arg.Sid); e != nil {
				log.Error(" s.PreListSet(%d) error(%v)", arg.Sid, e)
			}
		})
		return
	}
	if preInfos, err = s.preDao.Predictions(c, preIds); err != nil {
		log.Error("s.dao.Predictions(%v) error(%+v)", preIds, err)
		return
	}
	parentInfos = make(map[int64]*premdl.Prediction)
	preDeals = make(map[int64][]*premdl.Prediction)
	for _, v := range preInfos {
		if v.ID == 0 {
			continue
		}
		if v.Pid == 0 {
			parentInfos[v.ID] = v
		} else {
			if _, ok := preDeals[v.Pid]; !ok {
				preDeals[v.Pid] = make([]*premdl.Prediction, 0)
			}
			preDeals[v.Pid] = append(preDeals[v.Pid], v)
		}
	}
	res = make(map[int64]*premdl.PredictionItem, len(parentInfos))
	group, ctx := errgroup.WithContext(c)
	for _, v := range parentInfos {
		pid := v.ID
		if _, ok := preDeals[pid]; ok && len(preDeals[pid]) > 0 {
			group.Go(func() error {
				temp, e := s.singleJudge(ctx, preDeals[pid], arg.Point)
				if e == nil && temp != nil {
					lock.Lock()
					res[pid] = temp
					lock.Unlock()
				}
				return e
			})
		} else {
			log.Info("PreJudge:There is no subset (%v)", v)
		}
	}
	if err = group.Wait(); err != nil {
		log.Error("PreJudge:group.Wait() error(%v)", err)
	}
	return
}

// singleJudge .
func (s *Service) singleJudge(c context.Context, list []*premdl.Prediction, point int64) (res *premdl.PredictionItem, err error) {
	var (
		preRange  *premdl.Prediction
		temps     []int64
		tempInfo  map[int64]*premdl.PredictionItem
		tempCount = 10
	)
	for _, v := range list {
		if v.Min <= point && v.Max >= point {
			preRange = v
			break
		}
	}
	if preRange == nil {
		log.Error("singleJudge preRange is nil point(%d)", point)
		return
	}
	if temps, err = s.preDao.ItemRandMember(c, preRange.ID, tempCount); err != nil {
		log.Error("s.dao.ItemRandMember(%d) error(%v)", preRange.ID, err)
		return
	}
	if len(temps) == 0 {
		s.cache.Do(c, func(ctx context.Context) {
			if e := s.ItemListSet(ctx, preRange.ID); e != nil {
				log.Error("s.ItemListSet(%d) error(%v)", preRange.ID, e)
			}
		})
		return
	}
	if tempInfo, err = s.preDao.PredItems(c, temps); err != nil {
		log.Error("s.dao.PredItems(%v) error(%+v)", temps, err)
		return
	}
	if len(tempInfo) == 0 {
		log.Info("singleJudge tempInfo is nil")
		return
	}
	for _, v := range tempInfo {
		if v.ID != 0 {
			res = v
			break
		}
	}
	return
}

// PreItemUp .
func (s *Service) PreItemUp(c context.Context, id int64) (err error) {
	var (
		item map[int64]*premdl.PredictionItem
	)
	if item, err = s.preDao.RawPredItems(c, []int64{id}); err != nil {
		log.Error("s.dao.RawPredItems(%d) error(%+v)", id, err)
		return
	}
	if _, ok := item[id]; !ok {
		s.preDao.DelCachePredItems(c, []int64{id})
	} else {
		s.preDao.AddCachePredItems(c, item)
	}
	return
}

// PreUp .
func (s *Service) PreUp(c context.Context, id int64) (err error) {
	var (
		item map[int64]*premdl.Prediction
	)
	if item, err = s.preDao.RawPredictions(c, []int64{id}); err != nil {
		log.Error("s.dao.RawPredictions(%d) error(%+v)", id, err)
		return
	}
	if _, ok := item[id]; !ok {
		s.preDao.DelCachePredictions(c, []int64{id})
	} else {
		s.preDao.AddCachePredictions(c, item)
	}
	return
}

// PreSetItem .
func (s *Service) PreSetItem(c context.Context, id, pid int64, state int) (err error) {
	var (
		list    map[int64]*premdl.PredictionItem
		delFlag int
	)
	if state == _upState {
		if list, err = s.preDao.RawPredItems(c, []int64{id}); err != nil {
			log.Error("s.dao.RawPredItems(%d) error(%+v)", id, err)
			return
		}
		delFlag = 0
		if _, ok := list[id]; ok {
			delFlag = 1
		}
	} else if state == _delState {
		delFlag = 0
	} else {
		return
	}
	if delFlag == 1 {
		err = s.preDao.AddItemPreSet(c, []int64{id}, pid)
	} else {
		err = s.preDao.DelItemPreSet(c, []int64{id}, pid)
	}
	return
}

// ItemListSet prediction_item set reload.
func (s *Service) ItemListSet(c context.Context, pid int64) (err error) {
	id := int64(0)
	for {
		var (
			addIDs, delIDs []int64
			list           []*premdl.PredictionItem
		)
		if list, err = s.preDao.ItemListSet(c, id, pid); err != nil {
			log.Error("s.dao.ItemListSet(%d) error(%+v)", id, pid)
			return
		}
		if len(list) == 0 {
			log.Info("PreSet pid(%d) success", pid)
			break
		}
		addIDs = make([]int64, 0, len(list))
		delIDs = make([]int64, 0, len(list))
		for _, v := range list {
			if v.State == 1 {
				addIDs = append(addIDs, v.ID)
			} else {
				delIDs = append(delIDs, v.ID)
			}
			if v.ID > id {
				id = v.ID
			}
		}
		eg, ctx := errgroup.WithContext(c)
		if len(delIDs) > 0 {
			eg.Go(func() (e error) {
				if e = s.preDao.DelItemPreSet(ctx, delIDs, pid); e != nil {
					log.Error("s.dao.DelItemPreSet(%v,%d) error(%v)", delIDs, pid, e)
				}
				return
			})
		}
		if len(addIDs) > 0 {
			eg.Go(func() (e error) {
				if e = s.preDao.AddItemPreSet(ctx, addIDs, pid); e != nil {
					log.Error("s.dao.AddItemPreSet(%v,%d) error(%v)", addIDs, pid, e)
				}
				return
			})
		}
		if err = eg.Wait(); err != nil {
			log.Error("ItemListSet:eg.Wait() error(%v)", err)
			break
		}
	}
	return
}

// PreSet .
func (s *Service) PreSet(c context.Context, id, sid int64, state int) (err error) {
	var (
		delFlag int
		list    map[int64]*premdl.Prediction
	)
	if state == _upState {
		if list, err = s.preDao.RawPredictions(c, []int64{id}); err != nil {
			log.Error("s.dao.RawPredictions(%d) error(%v)", id, err)
			return
		}
		delFlag = 0
		if _, ok := list[id]; ok {
			delFlag = 1
		}
	} else if state == _delState {
		delFlag = 0
	} else {
		return
	}
	if delFlag == 1 {
		err = s.preDao.AddPreSet(c, []int64{id}, sid)
	} else {
		err = s.preDao.DelPreSet(c, []int64{id}, sid)
	}
	return
}

// PreListSet prediction set reload.
func (s *Service) PreListSet(c context.Context, sid int64) (err error) {
	id := int64(0)
	for {
		var (
			addIDs, delIDs []int64
			list           []*premdl.Prediction
		)
		if list, err = s.preDao.ListSet(c, id, sid); err != nil {
			log.Error("s.dao.ListSet(%d) error(%+v)", id, sid)
			return
		}
		if len(list) == 0 {
			log.Info("PreSet sid(%d) success", sid)
			break
		}
		addIDs = make([]int64, 0, len(list))
		delIDs = make([]int64, 0, len(list))
		for _, v := range list {
			if v.State == 1 {
				addIDs = append(addIDs, v.ID)
			} else {
				delIDs = append(delIDs, v.ID)
			}
			if v.ID > id {
				id = v.ID
			}
		}
		eg, ctx := errgroup.WithContext(c)
		if len(delIDs) > 0 {
			eg.Go(func() (e error) {
				if e = s.preDao.DelPreSet(ctx, delIDs, sid); e != nil {
					log.Error("s.dao.DelPreSet(%v,%d) error(%v)", delIDs, sid, e)
				}
				return
			})
		}
		if len(addIDs) > 0 {
			eg.Go(func() (e error) {
				if e = s.preDao.AddPreSet(ctx, addIDs, sid); e != nil {
					log.Error("s.dao.AddPreSet(%v,%d) error(%v)", addIDs, sid, e)
				}
				return
			})
		}
		if err = eg.Wait(); err != nil {
			log.Error("PreSet:eg.Wait() error(%v)", err)
			break
		}
	}
	return
}
