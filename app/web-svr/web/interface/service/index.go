package service

import (
	"context"
	"encoding/json"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/model"
)

const (
	_maxLength = 100
	_emptySort = "{}"
)

// IndexSort get index sort.
// nolint:gomnd
func (s *Service) IndexSort(c context.Context, mid int64, version int64) (json.RawMessage, error) {
	set, err := s.dao.IndexSortCache(c, mid, version)
	if err != nil {
		log.Error("s.dao.IndexSortCache mid(%d) error(%v)", mid, err)
		return json.RawMessage(_emptySort), nil
	}
	count := len(set)
	if count < 2 {
		return json.RawMessage(_emptySort), nil
	}
	set = strings.TrimPrefix(set, `"`)
	set = strings.TrimSuffix(set, `"`)
	set = strings.Replace(set, "\\", "", -1)
	setData := new(model.IndexSet)
	if err = json.Unmarshal([]byte(set), &setData); err != nil {
		log.Warn("IndexSort mid:%d json.Unmarshal:%s warn:%v", mid, set, err)
		return json.RawMessage(_emptySort), nil
	}
	return json.RawMessage(set), nil
}

func (s *Service) IndexSortSet(c context.Context, uid int64, settings string, version int64) (err error) {
	var set model.IndexSet
	if err = json.Unmarshal([]byte(settings), &set); err != nil {
		log.Error("IndexSortSet json.Unmarshal  setting(%s) error(%v)", settings, err)
		err = ecode.RequestErr
		return
	}
	if len(set.Sort) > _maxLength {
		log.Warn("IndexSortSet length warn(%d)", len(set.Sort))
		err = ecode.RequestErr
		return
	}
	err = s.dao.SetIndexSortCache(c, uid, set, version)
	return
}
