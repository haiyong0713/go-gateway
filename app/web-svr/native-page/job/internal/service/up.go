package service

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/admin/model/native"
	natGRPC "go-gateway/app/web-svr/native-page/interface/api"
)

const (
	autoAuditKey         = "auto_audit"
	_pagingNatPagesLimit = 100
	_maxGetFailedNum     = 10
)

func (s *Service) UpAutoAudit(c context.Context) {
	log.Info("Start to auto audit upPages")
	lockID, err := s.tryLock(c)
	if err != nil {
		return
	}
	defer func() {
		if err = s.dao.Unlock(c, autoAuditKey, lockID); err != nil {
			log.Error("Fail to release autoAudit lock, error=%+v", err)
		}
	}()
	pages, err := s.dao.PagingAutoAuditTsPages(c, 1, 20)
	if err != nil {
		log.Error("Fail to get autoAuditPages error=%+v", err)
		return
	}
	if len(pages) == 0 {
		log.Info("No autoAuditPages to process")
		return
	}
	for _, v := range pages {
		if v.AuditType != native.TsAuditAuto || v.State != native.TsWaitOnline {
			continue
		}
		time.Sleep(10 * time.Millisecond)
		if err = s.dao.TsOnline(c, v.Id, v.Pid, v.AuditTime); err != nil {
			log.Error("Fail to auto audit upPage, tsID=%+v pid=%+v error=%+v", v.Id, v.Pid, err)
			continue
		}
		log.Info("auto audit pageID=%+v tsID=%+v success", v.Pid, v.Id)
	}
}

func (s *Service) tryLock(c context.Context) (string, error) {
	lockID, locked, err := s.dao.Lock(c, autoAuditKey, s.dao.GetCfg().Expire.AutoAuditLockExpire)
	if err != nil {
		return "", err
	}
	if !locked {
		log.Error("Fail to get autoAudit lock, lock has been taken")
		return "", errors.New("lock has been taken")
	}
	return lockID, nil
}

func (s *Service) cacheSponsoredUp(c context.Context) {
	if s.cfg.SponsoredUp == nil || !s.cfg.SponsoredUp.Open {
		return
	}
	var (
		lastID      int64
		success     int64
		getFailed   int64
		cacheFailed int64
	)
	finishedMid := make(map[int64]struct{})
	for {
		time.Sleep(20 * time.Millisecond)
		pages, err := s.dao.AttemptPagingNatPages(c, lastID, _pagingNatPagesLimit)
		if err != nil {
			lastID += _pagingNatPagesLimit
			if getFailed++; getFailed > _maxGetFailedNum {
				break
			}
			continue
		}
		if len(pages) == 0 {
			break
		}
		for _, page := range pages {
			if _, ok := finishedMid[page.RelatedUid]; ok {
				continue
			}
			if page.FromType != natGRPC.PageFromUid || page.RelatedUid == 0 || page.State == natGRPC.WaitForCommit {
				continue
			}
			if err := s.dao.AddCacheSponsoredUp(c, page.RelatedUid); err != nil {
				cacheFailed++
				continue
			}
			finishedMid[page.RelatedUid] = struct{}{}
			success++
			log.Warn("cacheSponsoredUp cache mid=%d", page.RelatedUid)
		}
		lastID = pages[len(pages)-1].ID
	}
	log.Warn("cacheSponsoredUp finished, success_cnt=%d get_failed_cnt=%d cache_failed_cnt=%d", success, getFailed, cacheFailed)
}
