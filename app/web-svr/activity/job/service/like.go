package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/job/model/like"
	l "go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/job/model/match"
	mdlRank "go-gateway/app/web-svr/activity/job/model/rank"

	"go-common/library/sync/errgroup.v2"
)

const (
	// maxArcBatchLikeLimit 一次从表中获取稿件数量
	maxArcBatchLikeLimit = 1000
	// dbChannelLength db channel 长度
	dbChannelLength = 100
	// arcChannelLength 稿件channel长度
	arcChannelLength = 100
)

// AddLike like add data update cache .
func (s *Service) AddLike(c context.Context, addMsg json.RawMessage) (err error) {
	var (
		likeObj = new(l.Item)
	)
	if err = json.Unmarshal(addMsg, likeObj); err != nil {
		log.Error("AddLike json.Unmarshal(%s) error(%+v)", addMsg, err)
		return
	}
	if err = s.dao.LikeUp(c, likeObj.ID); err != nil {
		log.Error("s.dao.LikeUp(%d) error(%+v)", likeObj.ID, err)
		return
	}
	if err = s.dao.AddLikeCtimeCache(c, likeObj.ID); err != nil {
		log.Error("s.dao.AddLikeCtimeCache(%d) error(%+v)", likeObj.ID, err)
		return
	}
	log.Info("AddLike success s.actRPC.LikeUp(%d)", likeObj.ID)
	return
}

// UpLike update likes data update cahce
func (s *Service) UpLike(c context.Context, newMsg, oldMsg json.RawMessage) (err error) {
	var (
		likeObj = new(l.Item)
		oldObj  = new(l.Item)
	)
	if err = json.Unmarshal(newMsg, likeObj); err != nil {
		log.Error("UpLike json.Unmarshal(%s) error(%+v)", newMsg, err)
		return
	}
	if err = json.Unmarshal(oldMsg, oldObj); err != nil {
		log.Error("UpLike json.Unmarshal(%s) error(%+v)", oldMsg, err)
		return
	}
	if err = s.dao.LikeUp(c, likeObj.ID); err != nil {
		log.Error(" s.dao.LikeUp(%d) error(%+v)", likeObj.ID, err)
		return
	}
	if oldObj.State != likeObj.State {
		if likeObj.State == 1 {
			//add ctime cache
			if err = s.dao.AddLikeCtimeCache(c, likeObj.ID); err != nil {
				log.Error("s.dao.AddLikeCtimeCache(%d) error(%+v)", likeObj.ID, err)
				return
			}
		} else {
			if err = s.dao.DelLikeCtimeCache(c, likeObj.ID, likeObj.Sid, likeObj.Type); err != nil {
				log.Error("s.actRPC.DelLikeCtimeCache(%v) error(%+v)", likeObj, err)
				return
			}
		}
		if oldObj.State == -1 && likeObj.State == 1 {
			//点赞数从key->value 缓存中回源
			if err = s.dao.ActSetReload(c, likeObj.ID); err != nil {
				log.Error("s.dao.ActSetReload(%v) error(%+v)", likeObj, err)
				return
			}
		}
	}
	log.Info("UpLike success s.actRPC.LikeUp(%d)", likeObj.ID)
	return
}

// DelLike delete like update cache
func (s *Service) DelLike(c context.Context, oldMsg json.RawMessage) (err error) {
	var (
		likeObj = new(l.Item)
	)
	if err = json.Unmarshal(oldMsg, likeObj); err != nil {
		log.Error("DelLike json.Unmarshal(%s) error(%+v)", oldMsg, err)
		return
	}
	if err = s.dao.LikeUp(c, likeObj.ID); err != nil {
		log.Error("s.dao.LikeUp(%d) error(%+v)", likeObj.ID, err)
		return
	}
	if err = s.dao.DelLikeCtimeCache(c, likeObj.ID, likeObj.Sid, likeObj.Type); err != nil {
		log.Error("s.dao.DelLikeCtimeCache(%v) error(%+v)", likeObj, err)
		return
	}
	log.Info("DelLike success s.actRPC.LikeUp(%d)", likeObj.ID)
	return
}

func (s *Service) ResetLikeTypeCount(_ context.Context, sid int64) (err error) {
	go func() {
		if err := s.dao.ResetCacheLikeTypeCount(context.Background(), sid); err != nil {
			log.Error("ResetLikeTypeCount sid:%d error(%v)", sid, err)
		}
	}()
	return nil
}

// upLikeContent .
func (s *Service) upLikeContent(c context.Context, upMsg json.RawMessage) (err error) {
	var (
		cont = new(l.Content)
	)
	if err = json.Unmarshal(upMsg, cont); err != nil {
		log.Error("upLikeContent json.Unmarshal(%s) error(%+v)", upMsg, err)
		return
	}
	if err = s.dao.SetLikeContent(c, cont.ID); err != nil {
		log.Error("s.dao.SetLikeContent(%d) error(%+v)", cont.ID, err)
	}
	log.Info("upLikeContent success s.dao.SetLikeContent(%d)", cont.ID)
	return
}

// archiveCanal .
func (s *Service) archiveCanal() {
	defer s.waiter.Done()
	if s.archiveSub == nil {
		return
	}
	var (
		err error
		c   = context.Background()
	)
	for {
		msg, ok := <-s.archiveSub.Messages()
		if !ok {
			log.Info("databus: activity-job binlog archive exit!")
			return
		}
		msg.Commit()
		m := &match.Message{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("archiveCanal json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		switch m.Table {
		case _archiveTable:
			// archive data update
			if m.Action == match.ActUpdate {
				newArc := &l.Archive{}
				oldArc := &l.Archive{}
				if err = json.Unmarshal(m.New, newArc); err != nil {
					log.Error("archiveCanal json.Unmarshal(%s) error(%+v)", m.New, err)
					continue
				}
				// 判断是否是活动稿件
				if newArc.MissionID > 0 && newArc.Aid > 0 {
					if err = json.Unmarshal(m.Old, oldArc); err != nil {
						log.Error("archiveCanal json.Unmarshal(%s) error(%+v)", m.Old, err)
						continue
					}
					//判断稿件是否是下架处理
					if newArc.State < 0 && oldArc.State >= 0 {
						s.arcActionCh <- newArc
					}
					// 审核通过增加抽奖机会
					if newArc.MissionID > 0 && newArc.State >= 0 && oldArc.State < 0 && newArc.Mid > 0 {
						s.upSubjectStickTop(c, newArc.MissionID, newArc)
						s.lotteryActionch <- &l.LotteryMsg{MissionID: newArc.MissionID, Mid: newArc.Mid, ObjID: newArc.Aid}
						s.arcPassTasks(c, newArc)
					}
					// 联合投稿特殊判断
					if newArc.MissionID == s.c.Staff.Sid && newArc.State >= 0 && newArc.Attribute != oldArc.Attribute && newArc.Mid > 0 {
						if ((newArc.Attribute >> api.AttrBitIsCooperation) & 1) == api.AttrYes {
							s.staffPassTask(c, newArc.Aid, newArc.Mid)
						}
					}
				}
			} else if m.Action == match.ActInsert {
				newArc := &l.Archive{}
				if err = json.Unmarshal(m.New, newArc); err != nil {
					log.Error("archiveCanal json.Unmarshal(%s) error(%+v)", m.New, err)
					continue
				}
				// 审核通过增加抽奖机会
				if newArc.MissionID > 0 && newArc.Mid > 0 && newArc.State >= 0 {
					s.upSubjectStickTop(c, newArc.MissionID, newArc)
					s.lotteryActionch <- &l.LotteryMsg{MissionID: newArc.MissionID, Mid: newArc.Mid, ObjID: newArc.Aid}
					s.arcPassTasks(c, newArc)
				}
			}
		}
		log.Info("archiveCanal success key:%s partition:%d offset:%d value:%s", msg.Key, msg.Partition, msg.Offset, msg.Value)
	}
}

// actLikeproc .
func (s *Service) actLikeproc() {
	defer s.waiter.Done()
	var (
		ch = s.arcActionCh
	)
	for {
		ms, ok := <-ch
		if !ok {
			log.Warn("actLikeproc s.archiveProc() quit")
			return
		}
		//获取lid
		list, err := s.dao.LikeAidItem(context.Background(), ms.MissionID, ms.Aid)
		if err != nil {
			log.Error("s.dao.LikeAidItem(%d,%d) error(%v)", ms.MissionID, ms.Aid, err)
			continue
		}
		if list == nil || list.ID == 0 {
			continue
		}
		if err := s.dao.DelLikeState(context.Background(), ms.MissionID, []int64{list.ID}, -1, "稿件下架"); err != nil {
			log.Error("actLikeproc s.dao.DelLikeState(%d) error(%v)", list.ID, err)
			continue
		}
		log.Info("actLikeproc DelLikeState(%d) success (%d,%d)  ", list.ID, ms.MissionID, ms.Aid)
	}
}

func (s *Service) arcPassTasks(c context.Context, newArc *l.Archive) {
	switch newArc.MissionID {
	case s.c.Image.TaskArchiveSID:
		s.upDoTask(c, newArc)
	case s.c.Staff.Sid: // 联合投稿达成任务
		if ((newArc.Attribute >> api.AttrBitIsCooperation) & 1) == api.AttrYes {
			s.staffPassTask(c, newArc.Aid, newArc.Mid)
		}
	default:
		// 春节红包活动 start
		if yTaskID, yok := s.c.Image.YearTaskIDs[strconv.FormatInt(newArc.MissionID, 10)]; yok && yTaskID > 0 {
			s.singleDoTask(c, newArc.Mid, yTaskID)
		}
		// 通用活动
		if taskID, yok := s.c.Image.CommonTaskIDs[strconv.FormatInt(newArc.MissionID, 10)]; yok && taskID > 0 {
			s.singleDoTask(c, newArc.Mid, taskID)
		}
	}
}

func (s *Service) addUpListHis() {
	var err error
	for i := 0; i < 3; i++ {
		if err = s.dao.UpListHisAdd(context.Background(), s.c.Taaf.Sidv2); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		log.Error("s.dao.UpListHisAdd sid(%d) error(%v)", s.c.Taaf.Sidv2, err)
	} else {
		log.Warn("s.dao.UpListHisAdd sid(%d) success", s.c.Taaf.Sidv2)
	}
}

// midArchiveInfo 用户活动稿件信息
func (s *Service) midArchiveInfo(c context.Context, sid int64) (*mdlRank.ArchiveStatMap, error) {
	dbCh := make(chan []*like.Like, dbChannelLength)
	arcCh := make(chan *api.ArcsReply, arcChannelLength)
	eg := errgroup.WithContext(c)
	var (
		archiveStateMap *mdlRank.ArchiveStatMap
	)
	eg.Go(func(ctx context.Context) (err error) {
		err = s.archiveIntoChannel(c, sid, dbCh)
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		archiveStateMap, err = s.archiveInfoDetail(c, dbCh, arcCh)
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, err
	}
	return archiveStateMap, nil
}

// archiveInfo 稿件信息获取
func (s *Service) archiveIntoChannel(c context.Context, sid int64, ch chan []*like.Like) error {
	var (
		err   error
		batch int
	)
	defer close(ch)
	for {
		likeList, err := s.dao.LikeList(c, sid, s.mysqlOffset(batch), maxArcBatchLikeLimit)
		if err != nil {
			log.Error("s.dao.LikeList: error(%v)", err)
			break
		}
		if len(likeList) > 0 {
			ch <- likeList
		}
		if len(likeList) < maxArcBatchLikeLimit {
			break
		}
		time.Sleep(100 * time.Microsecond)
		batch++
	}
	return err
}

// archiveInfoDetail 稿件点赞信息详情
func (s *Service) archiveInfoDetail(c context.Context, ch chan []*like.Like, arcCh chan *api.ArcsReply) (*mdlRank.ArchiveStatMap, error) {
	var memberArchive = mdlRank.ArchiveStatMap{}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		defer close(arcCh)
		for v := range ch {
			aids := []int64{}
			for _, item := range v {
				aids = append(aids, item.Wid)
			}
			err = s.archiveInfo(c, aids, arcCh)
		}
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		aidsMap := make(map[int64]bool)
		for v := range arcCh {
			if v == nil || v.Arcs == nil {
				err = ecode.ActivityWriteHandArchiveErr
			}
			for aid, arc := range v.Arcs {
				if arc == nil {
					err = ecode.ActivityWriteHandArchiveErr
				}
				// 防止aid重复
				if _, ok := aidsMap[aid]; !ok && arc.IsNormal() {
					memberArchive[arc.Author.Mid] = append(memberArchive[arc.Author.Mid], &mdlRank.ArchiveStat{
						Mid:     arc.Author.Mid,
						Aid:     aid,
						View:    arc.Stat.View,
						Danmaku: arc.Stat.Danmaku,
						Reply:   arc.Stat.Reply,
						Fav:     arc.Stat.Fav,
						Coin:    arc.Stat.Coin,
						Share:   arc.Stat.Share,
						NowRank: arc.Stat.NowRank,
						Like:    arc.Stat.Like,
						Videos:  arc.Videos,
					})
					aidsMap[aid] = true
				}
			}
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%v)", err)
		return nil, err
	}
	return &memberArchive, nil
}

// archiveInfo 从channel中获取稿件id，并获取详情
func (s *Service) archiveInfo(c context.Context, aids []int64, arcCh chan *api.ArcsReply) error {
	var times int
	patch := maxArcsLength
	concurrency := concurrencyArchiveDb
	times = len(aids) / patch / concurrency
	for index := 0; index <= times; index++ {
		eg := errgroup.WithContext(c)
		for batch := 0; batch < concurrency; batch++ {
			i := index
			b := batch
			eg.Go(func(ctx context.Context) error {
				start := i*patch*concurrency + b*patch
				if start >= len(aids) {
					return nil
				}
				reqAids := aids[start:]
				end := start + patch
				if end < len(aids) {
					reqAids = aids[start:end]
				}
				if len(reqAids) > 0 {
					reply, err := s.arcClient.Arcs(c, &api.ArcsRequest{Aids: reqAids})
					if err != nil {
						log.Error("s.arcClient.Arcs: error(%v)", err)
						return err
					}
					arcCh <- reply
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			log.Error("eg.Wait error(%v)", err)
			return ecode.ActivityWriteHandArchiveErr
		}
	}
	return nil
}

// mysqlOffset count mysql offset
func (s *Service) mysqlOffset(batch int) int {
	return batch * maxArcBatchLikeLimit
}

func (s *Service) webDataView(c context.Context, vid int64, offset, limit, retryCnt int) (list []*l.WebData, err error) {
	for i := 0; i < retryCnt; i++ {
		if list, err = s.dao.WebDataView(c, vid, offset, limit); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

//func (s *Service) loadOperationData() {
//	ctx := context.Background()
//	ticker := time.NewTicker(time.Second * s.c.OperationSource.UpdateTicker)
//	defer func() {
//		ticker.Stop()
//	}()
//	for {
//		select {
//		case <-ticker.C:
//			s.OperationDataDo(ctx)
//		}
//	}
//}

func (s *Service) OperationDataDo(ctx context.Context) {
	for _, sid := range s.c.OperationSource.OperationSids {
		viewData, err := s.webDataView(ctx, sid, 0, _objectPieceSize, _retryTimes)
		if err != nil {
			log.Errorc(ctx, "loadOperationData s.webDataView(%d,%d,%d) error(%+v)", sid, 0, _objectPieceSize, err)
			continue
		}
		if len(viewData) == 0 {
			log.Infoc(ctx, "loadOperationData s.webDataView(%d,%d,%d) count 0", sid, 0, _objectPieceSize)
			continue
		}
		if err = s.setViewDataCache(ctx, sid, viewData); err != nil {
			log.Errorc(ctx, "loadOperationData setViewDataCache sid(%v) error(%+v)", sid, err)
		}
	}
}

func (s *Service) setViewDataCache(ctx context.Context, sid int64, data []*l.WebData) (err error) {
	for i := 0; i < _retry; i++ {
		if err = s.dao.AddViewDataCache(ctx, sid, data); err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(ctx, "loadOperationData setViewDataCache s.dao.AddViewDataCache sid(%v) ")
	}
	return nil
}
