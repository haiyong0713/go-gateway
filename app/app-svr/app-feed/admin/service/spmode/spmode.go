package spmode

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-feed/admin/model/spmode"
)

func (s *Service) Search(req *model.SearchReq) (*model.SearchRly, error) {
	rly := &model.SearchRly{}
	if req.Mid > 0 {
		users, err := s.dao.TeenagerUsersByMid(req.Mid)
		if err != nil {
			return nil, err
		}
		for _, user := range users {
			rly.List = append(rly.List, s.user2SearchItem(user))
		}
		return rly, nil
	}
	if req.DeviceToken != "" {
		devs, err := s.dao.DeviceUserModelByToken(req.DeviceToken)
		if err != nil {
			return nil, err
		}
		for _, dev := range devs {
			rly.List = append(rly.List, s.device2SearchItem(dev))
		}
		return rly, nil
	}
	return nil, ecode.RequestErr
}

func (s *Service) Relieve(c context.Context, req *model.RelieveReq, userid int64, username string) error {
	idType, id := extractRelatedKey(req.RelatedKey)
	if id == 0 {
		return ecode.RequestErr
	}
	var (
		ok  bool
		err error
	)
	switch idType {
	case model.RelatedKeyTypeUser:
		ok, err = s.relieveUser(c, id, model.OperationQuitManager)
	case model.RelatedKeyTypeDevice:
		ok, err = s.relieveDevice(c, id)
	default:
		return ecode.RequestErr
	}
	if err != nil {
		return ecode.ServerErr
	}
	if !ok {
		return ecode.Error(ecode.RequestErr, "用户状态已解除")
	}
	_ = s.worker.Do(c, func(ctx context.Context) {
		_ = s.dao.AddSpecialModeLog(&model.SpecialModeLog{
			RelatedKey:  req.RelatedKey,
			OperatorUid: userid,
			Operator:    username,
			Content:     "后台解除退出",
		})
	})
	return nil
}

func (s *Service) relieveUser(c context.Context, id, operation int64) (ok bool, err error) {
	defer func() {
		if err != nil || !ok {
			return
		}
		if user, err2 := s.dao.TeenagerUsersByID(id); err2 == nil && user != nil {
			_ = s.dao.DelCacheModelUser(c, user.Mid)
		}
	}()
	affected, err := s.dao.RelieveTeenagerUsers(id, operation)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func (s *Service) relieveDevice(c context.Context, id int64) (ok bool, err error) {
	defer func() {
		if err != nil || !ok {
			return
		}
		if device, err2 := s.dao.DeviceUserModelByID(id); err2 == nil && device != nil {
			_ = s.dao.DelCacheDevModelUser(c, device.MobiApp, device.DeviceToken)
		}
	}()
	affected, err := s.dao.RelieveDeviceUserModel(id)
	if err != nil {
		return false, err
	}
	return affected != 0, nil
}

func (s *Service) Log(req *model.LogReq) (*model.LogRly, error) {
	logs, err := s.dao.SpecialModeLogsByKey(req.RelatedKey)
	if err != nil {
		return nil, ecode.ServerErr
	}
	items := make([]*model.LogItem, 0, len(logs))
	for _, modeLog := range logs {
		items = append(items, log2LogItem(modeLog))
	}
	return &model.LogRly{List: items}, nil
}

func (s *Service) user2SearchItem(user *model.TeenagerUsers) *model.SearchItem {
	var password string
	if pwd, ok := s.EncryptedPwd[user.Password]; ok {
		password = pwd
	}
	return &model.SearchItem{
		RelatedKey: buildRelatedKey(user.ID, user.Mid, ""),
		Model:      user.Model,
		Mid:        user.Mid,
		Password:   password,
		Mtime:      user.Mtime.Time().Format("2006-01-02 15:04:05"),
		State:      user.State,
		PwdType:    user.PwdType,
	}
}

func (s *Service) device2SearchItem(dev *model.DeviceUserModel) *model.SearchItem {
	var password string
	if pwd, ok := s.EncryptedPwd[dev.Password]; ok {
		password = pwd
	}
	return &model.SearchItem{
		RelatedKey:  buildRelatedKey(dev.ID, 0, dev.DeviceToken),
		Model:       dev.Model,
		DeviceToken: dev.DeviceToken,
		Password:    password,
		Mtime:       dev.Mtime.Time().Format("2006-01-02 15:04:05"),
		State:       dev.State,
		PwdType:     dev.PwdType,
	}
}

func buildRelatedKey(id, mid int64, deviceToken string) string {
	if id <= 0 {
		return ""
	}
	if mid > 0 {
		return fmt.Sprintf("%s_%d", model.RelatedKeyTypeUser, id)
	}
	if deviceToken != "" {
		return fmt.Sprintf("%s_%d", model.RelatedKeyTypeDevice, id)
	}
	return ""
}

func extractRelatedKey(key string) (string, int64) {
	const _partsNum = 2
	parts := strings.Split(key, "_")
	if len(parts) < _partsNum {
		log.Error("related_key parts less than 2, key=%s", key)
		return "", 0
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		log.Error("Fail to ParseInt of id of RelatedKey, id=%s error=%+v", parts[1], err)
		return "", 0
	}
	return parts[0], id
}

func log2LogItem(log *model.SpecialModeLog) *model.LogItem {
	return &model.LogItem{
		Operator: log.Operator,
		Ctime:    log.Ctime.Time().Format("2006-01-02 15:04:05"),
		Content:  log.Content,
	}
}
