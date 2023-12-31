package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/service/model"
)

func (s *Service) arcChan(action string, nwMsg []byte, oldMsg []byte) {
	var err error
	nw := &model.Archive{}
	if err = json.Unmarshal(nwMsg, nw); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", nwMsg, err)
		return
	}
	switch action {
	case _updateAct:
		old := &model.Archive{}
		if err = json.Unmarshal(oldMsg, old); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", oldMsg, err)
			return
		}
		if old.State != nw.State {
			if nw.State == model.StateOrange || nw.State == model.StateForbidSubmit {
				// only send msg
				s.sendWechart(nw.State, nw.AID, "warn")
			} else if nw.State == model.StateForbidRecycle || nw.State == model.StateForbidLock || nw.State == model.StateForbidUpDelete {
				// send msg and off line
				if err = s.offLine(nw.AID); err == nil {
					s.sendWechart(nw.State, nw.AID, "offLine")
				}
			}
		}
	}
}

func (s *Service) sendWechart(ns int8, aid int64, titleType string) {
	var sends = make(map[string][]*model.ResWarnInfo)
	if ars, ok := s.resArchiveWarnCache[aid]; ok {
		for _, ar := range ars {
			sends[ar.UserName] = append(sends[ar.UserName], ar)
		}
	}
	for useName, send := range sends {
		log.Info("sendWechart(%v, %v, %d, %v) start send QYWX msg", aid, useName, ns, titleType)
		if e := s.alarmDao.SendWeChart(context.TODO(), ns, useName, send, titleType); e != nil {
			log.Error("sendWechart(%v, %v, %d, %v) fail err(%s)", aid, useName, ns, titleType, e.Error())
		}
	}
}

func (s *Service) offLine(aid int64) (err error) {
	// begin tran
	var tx *sql.Tx
	if tx, err = s.res.BeginTran(context.TODO()); err != nil {
		log.Error("offLine aid(%v) s.res.BeginTran() error(%v)", aid, err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			if e := tx.Rollback(); e != nil {
				log.Error("offLine Rollback error(%v)", e)
			}
			log.Error("offLine aid(%v) off line recover error(%v)", aid, r)
		}
	}()
	if ars, ok := s.resArchiveWarnCache[aid]; ok {
		var (
			applyGroupIDm = make(map[int]int)
			applyGroupIDs []string
		)
		// update resource assignment etime by id
		for _, ar := range ars {
			log.Info("offLine aid(%v) ar(%+v) start off line resource", aid, ar)
			if _, err = s.res.TxOffLine(tx, ar.AssignmentID); err != nil {
				log.Error("offLine aid(%v) s.res.TxOffLine(%v) error(%v)", aid, ar.AssignmentID, err)
				if e := tx.Rollback(); e != nil {
					log.Error("offLine aid(%v) s.res.TxOffLine(%v) rollback error(%v)", aid, ar.AssignmentID, e)
				}
				return
			}
			etime := ar.ETime.Time().Format("2006-01-02 15:04:05")
			// log for manager
			if _, err = s.res.TxInResourceLogger(tx, "material", fmt.Sprintf("批量下线 原计划投放结束时间: %v，备注: 稿件不可看，自动下线", etime), ar.MaterialID); err != nil {
				log.Error("offLine aid(%v) s.res.TxInResourceLogger(%v, %v, %v) error(%v)", aid, "material", fmt.Sprintf("批量下线 原计划投放结束时间: %v，备注: 稿件不可看，自动下线", etime), ar.MaterialID, err)
				if e := tx.Rollback(); e != nil {
					log.Error("offLine aid(%v) s.res.TxInResourceLogger(%v, %v, %v) rollback error(%v)", aid, "material", fmt.Sprintf("批量下线 原计划投放结束时间: %v，备注: 稿件不可看，自动下线", etime), ar.MaterialID, e)
				}
				return
			}
			// log for rollback db
			if _, err = s.res.TxInResourceLogger(tx, "rejob", etime, ar.AssignmentID); err != nil {
				log.Error("offLine aid(%v) s.res.TxInResourceLogger(%v, %v, %v) error(%v)", aid, "rejob", etime, ar.AssignmentID, err)
				if e := tx.Rollback(); e != nil {
					log.Error("offLine aid(%v) s.res.TxInResourceLogger(%v, %v, %v) rollback error(%v)", aid, "rejob", etime, ar.AssignmentID, e)
				}
				return
			}
			applyGroupIDm[ar.ApplyGroupID] = ar.ApplyGroupID
		}
		for _, g := range applyGroupIDm {
			applyGroupIDs = append(applyGroupIDs, strconv.Itoa(g))
			// log for manager
			if _, err = s.res.TxInResourceLogger(tx, "resource_apply", "投放被下线", g); err != nil {
				log.Error("offLine aid(%v) s.res.TxInResourceLogger(%v, %v, %v) error(%v)", aid, "resource_apply", "投放被下线", g, err)
				if e := tx.Rollback(); e != nil {
					log.Error("offLine aid(%v) s.res.TxInResourceLogger(%v, %v, %v) rollback error(%v)", aid, "resource_apply", "投放被下线", g, e)
				}
				return
			}
		}
		// update resource apply audit status by group_id
		log.Info("offLine aid(%v) apply_group_ids(%v) start free apply resource", aid, applyGroupIDs)
		if _, err = s.res.TxFreeApply(tx, applyGroupIDs); err != nil {
			log.Error("offLine aid(%v) s.res.TxFreeApply(%v) error(%v)", aid, applyGroupIDs, err)
			if e := tx.Rollback(); e != nil {
				log.Error("offLine aid(%v) s.res.TxFreeApply(%v) rollback error(%v)", aid, applyGroupIDs, e)
			}
			return
		}
	}
	if err = tx.Commit(); err != nil {
		log.Error("offLine aid(%v) tx.Commit() error(%v)", aid, err)
	}
	return
}
