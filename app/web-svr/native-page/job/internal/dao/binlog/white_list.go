package binlog

import (
	"context"
	"encoding/json"

	"go-gateway/app/web-svr/native-page/job/internal/model"

	"github.com/pkg/errors"
	"go-common/library/log"
)

type WhiteListProcessor struct {
	dao *Dao
}

func (m *WhiteListProcessor) ParseMsg(c context.Context, msg json.RawMessage) (interface{}, error) {
	whiteList := &model.WhiteList{}
	if err := json.Unmarshal(msg, whiteList); err != nil {
		log.Errorc(c, "Fail to unmarshal whiteList, data=%+v error=%+v", string(msg), err)
		return nil, err
	}
	return whiteList, nil
}

func (m *WhiteListProcessor) HandleInsert(c context.Context, msg json.RawMessage) error {
	log.Info("handle-whiteList-insert, msg=%s", string(msg))
	data, err := m.ParseMsg(c, msg)
	if err != nil {
		return err
	}
	whiteList := data.(*model.WhiteList)
	if whiteList.State != model.StateValid {
		return errors.Errorf("whiteList is invalid")
	}
	return m.dao.AddCacheWhiteListByMid(c, whiteList.Mid, whiteList)
}

func (m *WhiteListProcessor) HandleUpdate(c context.Context, new json.RawMessage, old json.RawMessage) error {
	log.Info("handle-whiteList-update, new=%s old=%s", string(new), string(old))
	data, err := m.ParseMsg(c, new)
	if err != nil {
		return err
	}
	whiteList := data.(*model.WhiteList)
	switch whiteList.State {
	case model.StateInvalid:
		return m.dao.DelCacheWhiteListByMid(c, whiteList.Mid)
	case model.StateValid:
		return m.dao.AddCacheWhiteListByMid(c, whiteList.Mid, whiteList)
	}
	return nil
}

func (m *WhiteListProcessor) HandleDelete(c context.Context, msg json.RawMessage) error {
	data, err := m.ParseMsg(c, msg)
	if err != nil {
		return err
	}
	whiteList := data.(*model.WhiteList)
	return m.dao.DelCacheWhiteListByMid(c, whiteList.Mid)
}
