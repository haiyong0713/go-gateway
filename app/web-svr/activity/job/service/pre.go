package service

import (
	"context"
	"encoding/json"

	"go-common/library/log"
	"go-common/library/sync/errgroup"
)

const (
	_upState  = 1
	_delState = 0
)

// DelPreMc .
func (s *Service) DelPreMc(c context.Context, msg json.RawMessage) {
	var (
		pre struct {
			ID  int64 `json:"id"`
			Sid int64 `json:"sid"`
		}
		err error
	)
	if err = json.Unmarshal(msg, &pre); err != nil {
		log.Error("DelPreMc:json.Unmarshal(%s) error(%v)", msg, err)
		return
	}
	eg, errCtx := errgroup.WithContext(c)
	eg.Go(func() (e error) {
		e = s.dao.UpPre(errCtx, pre.ID)
		return
	})
	eg.Go(func() (e error) {
		e = s.dao.PreSetUp(errCtx, pre.ID, pre.Sid, _delState)
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("DelPreMc:eg.Wait() error(%+v)", err)
	}
	log.Info("DelPreMc success %d", pre.ID)
}

// UpPreMc .
func (s *Service) UpPreMc(c context.Context, msg json.RawMessage) {
	var (
		pre struct {
			ID  int64 `json:"id"`
			Sid int64 `json:"sid"`
		}
		err error
	)
	if err = json.Unmarshal(msg, &pre); err != nil {
		log.Error("UpPreMc:json.Unmarshal(%s) error(%v)", msg, err)
		return
	}
	eg, errCtx := errgroup.WithContext(c)
	eg.Go(func() (e error) {
		e = s.dao.UpPre(errCtx, pre.ID)
		return
	})
	eg.Go(func() (e error) {
		e = s.dao.PreSetUp(errCtx, pre.ID, pre.Sid, _upState)
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("UpPreMc:eg.Wait() error(%+v)", err)
		return
	}
	log.Info("UpPreMc success %d", pre.ID)
}

// UpItemPreMc .
func (s *Service) UpItemPreMc(c context.Context, msg json.RawMessage) {
	var (
		pre struct {
			ID  int64 `json:"id"`
			Pid int64 `json:"pid"`
		}
		err error
	)
	if err = json.Unmarshal(msg, &pre); err != nil {
		log.Error("UpItemPreMc:json.Unmarshal(%s) error(%v)", msg, err)
		return
	}
	eg, errCtx := errgroup.WithContext(c)
	eg.Go(func() (e error) {
		e = s.dao.UpItemPre(errCtx, pre.ID)
		return
	})
	eg.Go(func() (e error) {
		e = s.dao.PreItemSetUp(errCtx, pre.ID, pre.Pid, _upState)
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("UpItemPreMc:eg.Wait() error(%+v)", err)
		return
	}
	log.Info("UpItemPreMc success %d", pre.ID)
}

// DelItemPreMc .
func (s *Service) DelItemPreMc(c context.Context, msg json.RawMessage) {
	var (
		pre struct {
			ID  int64 `json:"id"`
			Pid int64 `json:"pid"`
		}
		err error
	)
	if err = json.Unmarshal(msg, &pre); err != nil {
		log.Error("DelItemPreMc:json.Unmarshal(%s) error(%v)", msg, err)
		return
	}
	eg, errCtx := errgroup.WithContext(c)
	eg.Go(func() (e error) {
		e = s.dao.UpItemPre(errCtx, pre.ID)
		return
	})
	eg.Go(func() (e error) {
		e = s.dao.PreItemSetUp(errCtx, pre.ID, pre.Pid, _delState)
		return
	})
	if err = eg.Wait(); err != nil {
		log.Error("DelItemPreMc:eg.Wait() error(%+v)", err)
		return
	}
	log.Info("DelItemPreMc success %d", pre.ID)
}
