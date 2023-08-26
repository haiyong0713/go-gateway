package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/archive-extra/service/api"
)

func (s *Service) AddArchiveExtraValue(c context.Context, aid int64, key, value string) (err error) {
	// 权限控制
	if !s.canGetExtraInfo(c, key) {
		err = ecode.AccessDenied
		log.Warn("AddArchiveExtraValue Not Authorised Caller %s", metadata.String(c, metadata.Caller))
		return
	}
	_, err = s.d.ExtraUpdate(c, aid, key, value)
	if err != nil {
		log.Error("s.d.ExtraUpdate aid(%d) key(%s) value(%s) err(%v)", aid, key, value, err)
	}
	return
}

func (s *Service) BatchAddArchiveExtraValue(c context.Context, key string, aidValues map[int64]string) (err error) {
	// 权限控制
	if !s.canGetExtraInfo(c, key) {
		err = ecode.AccessDenied
		log.Warn("AddArchiveExtraValue Not Authorised Caller %s", metadata.String(c, metadata.Caller))
		return
	}
	for aid, value := range aidValues {
		_, err = s.d.ExtraUpdate(c, aid, key, value)
		if err != nil {
			log.Error("s.d.ExtraUpdate aid(%d) key(%s) value(%s) err(%v)", aid, key, value, err)
		}
	}
	return
}

func (s *Service) GetArchiveExtraValue(c context.Context, aid int64) (reply map[string]string, err error) {
	reply = make(map[string]string)
	reply, err = s.d.ExtraByAid(c, aid)
	if err != nil {
		log.Error("s.d.ExtraByAid aid(%d) err(%v)", aid, err)
		return nil, err
	}

	return
}

func (s *Service) RemoveArchiveExtraValue(c context.Context, aid int64, key string) (err error) {
	// 权限控制
	if !s.canGetExtraInfo(c, key) {
		err = ecode.AccessDenied
		log.Warn("AddArchiveExtraValue Not Authorised Caller %s", metadata.String(c, metadata.Caller))
		return
	}
	err = s.d.ExtraDel(c, aid, key)
	if err != nil {
		log.Error("s.d.ExtraDel aid(%d) key(%s) err(%v)", aid, key, err)
	}
	return
}

func (s *Service) BatchRemoveArchiveExtraValue(c context.Context, aids []int64, key string) (err error) {
	// 权限控制
	if !s.canGetExtraInfo(c, key) {
		err = ecode.AccessDenied
		log.Warn("AddArchiveExtraValue Not Authorised Caller %s", metadata.String(c, metadata.Caller))
		return
	}
	for _, aid := range aids {
		err = s.d.ExtraDel(c, aid, key)
		log.Error("s.d.ExtraDel aid(%d) key(%s) err(%v)", aid, key, err)
	}
	return
}

func (s *Service) BatchGetArchiveExtraValue(c context.Context, aids []int64) (reply map[int64]*api.ArchiveExtraValueReply, err error) {
	reply, err = s.d.ExtraByAids(c, aids)
	if err != nil {
		log.Error("s.d.ExtraDel aids(%+v)", aids)
	}
	return
}

func (s *Service) GetArchiveExtraBasedOnKeys(c context.Context, aid int64, keys []string) (reply map[string]string, err error) {
	reply = make(map[string]string)
	reply, err = s.d.ExtraByKeys(c, aid, keys)
	if err != nil {
		log.Error("s.d.ExtraByAid aid(%d) err(%v)", aid, err)
		return nil, err
	}

	return
}

func (s *Service) canGetExtraInfo(c context.Context, key string) bool {
	// 权限控制
	if s.c.Custom.ExtraCallersSwitch {
		caller := metadata.String(c, metadata.Caller)
		if biz, ok := s.authorisedCallers[caller]; !ok || biz != key {
			return false
		}
	}

	return true
}
