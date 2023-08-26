package service

import (
	"context"
	"encoding/json"
	"strconv"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/time"

	"go-gateway/app/app-svr/ugc-season/job/model/archive"
	jobmdl "go-gateway/app/app-svr/ugc-season/job/model/databus"
	"go-gateway/app/app-svr/ugc-season/job/model/retry"
	"go-gateway/app/app-svr/ugc-season/job/model/stat"
	seasonApi "go-gateway/app/app-svr/ugc-season/service/api"
)

func (s *Service) seasonConsumer() {
	defer s.waiter.Done()
	for {
		var (
			msg *databus.Message
			ok  bool
			err error
		)
		if s.closeSub {
			log.Error("s.seasonSub.messages closed")
			return
		}
		if msg, ok = <-s.seasonSub.Messages(); !ok {
			log.Error("s.seasonSub.messages closed")
			return
		}
		_ = msg.Commit()
		m := &jobmdl.SeasonMsg{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", msg.Value, err)
			continue
		}
		log.Info("got season message key(%s) value(%s) ", msg.Key, msg.Value)
		if m.SeasonID <= 0 {
			log.Error("season_id(%d) <= 0 message(%s)", m.SeasonID, msg.Value)
			continue
		}
		switch m.Route {
		case jobmdl.RouteSeasonShow:
			s.seasonUpdate(m.SeasonID)
			log.Info("season message key(%s) value(%s) finish", msg.Key, msg.Value)
		default:
			log.Error("unknown Route(%s) message(%s)", m.Route, msg.Value)
			continue
		}
	}
}

func (s *Service) seasonUpdate(sid int64) {
	changed, action, upSnAids, rmSnAids, mid, maxPtime, err := s.tranSeason(context.Background(), sid)
	if err != nil {
		// retry
		log.Error("seasonUpdate sid(%d) error(%+v)", sid, err)
		rt := &retry.Info{Action: retry.FailSeasonAdd}
		rt.Data.SeasonID = sid
		_ = s.PushToRetryList(context.Background(), rt)
		return
	}
	if !changed {
		log.Error("sid(%d) nothing changed", sid)
		return
	}
	_ = s.updateSeasonCache(sid, mid, action, maxPtime)
	s.statCh <- &stat.SeasonResult{SeasonID: sid, Action: action} // send message to stat
	upSnAidsMap := make(map[int64]struct{}, len(upSnAids))
	if len(upSnAids) > 0 {
		msg := &jobmdl.SeasonWithArchive{
			Route:    jobmdl.SeasonRouteForUpdate,
			SeasonID: sid,
			Aids:     upSnAids,
		}
		s.SeasonNotify(sid, msg)
		for _, usAid := range upSnAids {
			upSnAidsMap[usAid] = struct{}{}
		}
	}
	if len(rmSnAids) > 0 {
		var realRmSnAids []int64
		for _, dsAid := range rmSnAids {
			if _, ok := upSnAidsMap[dsAid]; ok { //兼容换源情况update优先于delete
				continue
			}
			realRmSnAids = append(realRmSnAids, dsAid)
		}
		msg := &jobmdl.SeasonWithArchive{
			Route:    jobmdl.SeasonRouteForRemove,
			SeasonID: sid,
			Aids:     realRmSnAids,
		}
		s.SeasonNotify(sid, msg)
	}
	log.Info("seasonUpdate sid(%d) success", sid)
}

// nolint:bilirailguncheck
func (s *Service) SeasonNotify(sid int64, msg *jobmdl.SeasonWithArchive) {
	if err := s.seasonWithArchivePub.Send(context.Background(), strconv.FormatInt(sid, 10), msg); err != nil {
		// retry
		log.Error("SeasonNotify sid(%d) msg(%+v) err(%+v)", sid, msg, err)
		rt := &retry.Info{Action: retry.FailForPubArchiveDatabus}
		rt.Data.SeasonID = sid
		rt.Data.SeasonWithArchive = msg
		_ = s.PushToRetryList(context.Background(), rt)
	}
}

// nolint:errcheck,gocognit
func (s *Service) tranSeason(c context.Context, sid int64) (changed bool, action string, upSnAids, delSnAids []int64, mid int64, maxPtime time.Time, err error) {
	var (
		tx                          *sql.Tx
		season                      *archive.Season
		sections                    []*archive.SeasonSection
		eps                         []*archive.SeasonEp
		delSecIDs, delEpIDs, snAIDs []int64
		addSecs                     []*archive.SeasonSection
		addEps                      []*archive.SeasonEp
		delSecIDsMap                = make(map[int64]struct{})
		firstAid, firstSecID        int64
	)
	if season, err = s.archiveDao.Season(c, sid); err != nil || season == nil {
		log.Error("s.archiveDao.Season(%d) error(%v) or season=nil", sid, err)
		return
	}
	if sections, err = s.archiveDao.Sections(c, sid); err != nil || len(sections) == 0 {
		log.Error("s.archiveDao.Sections(%d) error(%v) or sections=nil", sid, err)
		return
	}
	if eps, err = s.archiveDao.Episodes(c, sid); err != nil || len(eps) == 0 {
		log.Error("s.archiveDao.Episodes(%d) error(%v) or eps=nil", sid, err)
		return
	}
	mid = season.Mid
	for _, sec := range sections {
		if sec.Show == archive.ShowYes && sec.State == archive.StateOpen {
			addSecs = append(addSecs, sec)
			if firstSecID == 0 {
				firstSecID = sec.SectionID
			}
		} else if sec.Show == archive.ShowNo {
			delSecIDs = append(delSecIDs, sec.SectionID)
			delSecIDsMap[sec.SectionID] = struct{}{}
		}
	}
	for _, ep := range eps {
		snAIDs = append(snAIDs, ep.AID)
		if _, ok := delSecIDsMap[ep.SectionID]; ok || ep.Show == archive.ShowNo {
			delEpIDs = append(delEpIDs, ep.EpID)
			delSnAids = append(delSnAids, ep.AID)
		} else if ep.Show == archive.ShowYes && ep.State == archive.StateOpen {
			if firstAid == 0 && firstSecID == ep.SectionID {
				firstAid = ep.AID
			}
			addEps = append(addEps, ep)
			upSnAids = append(upSnAids, ep.AID)
		}
	}
	log.Warn("seasonchange sid(%d) elEpIDs(%+v) delSnAids(%+v) upSnAids(%+v) delSecIDs(%+v)", sid, delEpIDs, delSnAids, upSnAids, delSecIDs)
	if season.Show == archive.ShowYes && len(upSnAids) > 0 {
		if maxPtime, err = s.archiveDao.SeasonMaxPtime(c, sid, upSnAids); err != nil {
			log.Error("s.archiveDao.SeasonMaxPtime(%v) error(%v)", upSnAids, err)
			return
		}
	}
	if tx, err = s.resultDao.BeginTran(c); err != nil {
		log.Error("s.result.BeginTran error(%v)", err)
		return
	}
	action = retry.ActionUp
	if season.Show == archive.ShowYes && season.State == archive.StateOpen {
		if err = s.resultDao.TxAddSeason(c, tx, season, maxPtime, firstAid, int64(len(addEps))); err != nil {
			tx.Rollback()
			log.Error("s.result.TxAddSeason error(%v)", err)
			return
		}
	} else if season.Show == archive.ShowNo {
		// 由于生产端不从上往下同步状态，所以如果season.show=0，结果库的section和episode表也删了就ok了
		if err = s.resultDao.TxDelSeasonByID(c, tx, sid); err != nil {
			tx.Rollback()
			log.Error("s.result.TxDelSeason error(%v)", err)
			return
		}
		// 删section
		if err = s.resultDao.TxDelSecBySID(c, tx, sid); err != nil {
			tx.Rollback()
			log.Error("s.result.TxDelSecBySID error(%v)", err)
			return
		}
		// 删ep
		if err = s.resultDao.TxDelEpBySID(c, tx, sid); err != nil {
			tx.Rollback()
			log.Error("s.result.TxDelEpBySID error(%v)", err)
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit error(%v)", err)
			return
		}
		delSnAids = snAIDs // when season is removed, delete all its archive's seasonID
		upSnAids = []int64{}
		action = retry.ActionDel
		changed = true
		return
	}
	//先操作ep，否则section del的ep会根据投稿端的ep.show=1插入
	//先操作删除，保证加入唯一条正确的
	if len(delEpIDs) > 0 {
		if err = s.resultDao.TxDelEpByID(c, tx, delEpIDs); err != nil {
			tx.Rollback()
			log.Error("s.result.TxDelEp error(%v)", err)
			return
		}
	}
	if len(addEps) > 0 {
		if err = s.resultDao.TxAddEp(c, tx, addEps); err != nil {
			tx.Rollback()
			log.Error("s.result.TxAddEp error(%v)", err)
			return
		}
	}
	if len(delSecIDs) > 0 {
		if err = s.resultDao.TxDelSecByID(c, tx, delSecIDs); err != nil {
			tx.Rollback()
			log.Error("s.result.TxDelSecByID error(%v)", err)
			return
		}
	}
	if len(addSecs) > 0 {
		if err = s.resultDao.TxAddSection(c, tx, addSecs); err != nil {
			tx.Rollback()
			log.Error("s.result.TxAddSection error(%v)", err)
			return
		}
	}
	// 更新season表mtime，发canal给jd-job更新缓存
	if err = s.resultDao.TxUpSeasonMtime(c, tx, sid); err != nil {
		tx.Rollback()
		log.Error("s.result.TxUpSeasonMtime sid(%d) error(%v)", sid, err)
		return
	}
	if err = tx.Commit(); err != nil {
		log.Error("tx.Commit error(%v)", err)
		return
	}
	changed = true
	return
}

func (s *Service) updateSeasonCache(sid, mid int64, action string, maxPtime time.Time) (err error) {
	defer func() {
		if err != nil {
			log.Error("updateSeasonCache error(%v)", err)
			rt := new(retry.Info)
			rt.Data.SeasonID = sid
			rt.Data.Mid = mid
			rt.Data.Ptime = maxPtime
			switch action {
			case retry.ActionUp:
				rt.Action = retry.FailUpSeasonCache
			case retry.ActionDel:
				rt.Action = retry.FailDelSeasonCache
			default:
				log.Error("updateSeason wrong action(%s)", action)
				return
			}
			_ = s.PushToRetryList(context.Background(), rt)
		}
	}()
	if _, err = s.seasonClient.UpCache(context.Background(), &seasonApi.UpCacheRequest{
		SeasonID: sid,
		Action:   action,
	}); err != nil {
		log.Error("season-service UpCache(%d, %s) error(%v)", sid, action, err)
		return
	}
	switch action {
	case retry.ActionDel:
		err = s.DelUpperSeason(context.Background(), sid, mid)
	case retry.ActionUp:
		err = s.AddUpperSeason(context.Background(), sid, mid, maxPtime)
	default:
		log.Error("updateSeasonCache wrong action(%s)", action)
	}
	log.Info("updateSeasonCache success sid(%d) action(%s)", sid, action)
	return
}
