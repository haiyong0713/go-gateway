package binlog

import (
	"context"
	"encoding/json"

	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/job/internal/model"
)

type UserSpaceProcessor struct {
	dao *Dao
}

func (m *UserSpaceProcessor) ParseMsg(c context.Context, msg json.RawMessage) (interface{}, error) {
	userSpace := &model.NativeUserSpace{}
	if err := json.Unmarshal(msg, userSpace); err != nil {
		log.Errorc(c, "Fail to unmarshal userSpace, data=%+v error=%+v", string(msg), err)
		return nil, err
	}
	return userSpace, nil
}

func (m *UserSpaceProcessor) HandleInsert(c context.Context, msg json.RawMessage) error {
	log.Info("handle-NativeUserSpace-insert, msg=%s", string(msg))
	return m.deleteCache(c, msg)
}

func (m *UserSpaceProcessor) HandleUpdate(c context.Context, new json.RawMessage, old json.RawMessage) error {
	log.Info("handle-NativeUserSpace-update, new=%s old=%s", string(new), string(old))
	// mid唯一
	return m.deleteCache(c, old)
}

func (m *UserSpaceProcessor) HandleDelete(c context.Context, msg json.RawMessage) error {
	log.Info("handle-NativeUserSpace-delete, msg=%s", string(msg))
	return m.deleteCache(c, msg)
}

func (m *UserSpaceProcessor) deleteCache(c context.Context, msg json.RawMessage) error {
	data, err := m.ParseMsg(c, msg)
	if err != nil {
		return err
	}
	userSpace := data.(*model.NativeUserSpace)
	return m.dao.DelCacheUserSpaceByMid(c, userSpace.Mid)
}
