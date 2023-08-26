package feature

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/feature"
	featureMdl "go-gateway/app/app-svr/app-feed/admin/model/feature"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func (s *Service) BusinessConfigList(c context.Context, req *featureMdl.BusinessConfigListReq) (*featureMdl.BusinessConfigReply, error) {
	businessConfigs, total, err := s.dao.SearchBusinessConfig(c, req, true, true)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return &featureMdl.BusinessConfigReply{
		Page: &common.Page{
			Total: total,
			Num:   req.Pn,
			Size:  req.Ps,
		},
		List: businessConfigs,
	}, nil
}

func (s *Service) BusinessConfigSave(c context.Context, req *featureMdl.BusinessConfigSaveReq, userID int64, username string) (string, error) {
	businessConfigs, _, err := s.dao.SearchBusinessConfig(c, &featureMdl.BusinessConfigListReq{
		TreeID:  req.TreeID,
		KeyName: req.KeyName,
	}, false, false)
	if err != nil {
		log.Error("%+v", err)
		return "唯一性校验失败(数据库)", err
	}
	for _, businessConfig := range businessConfigs {
		if businessConfig != nil && businessConfig.ID != req.ID {
			log.Error("Fail to save build_limit, because key_name is existed")
			return "key已存在", ecode.RequestErr
		}
	}
	param := &featureMdl.BusinessConfig{
		ID:            req.ID,
		TreeID:        req.TreeID,
		KeyName:       req.KeyName,
		Config:        req.Config,
		Description:   req.Description,
		Relations:     req.Relations,
		Creator:       username,
		CreatorUID:    int(userID),
		Modifier:      username,
		ModifierUID:   int(userID),
		State:         featureMdl.StateOff,
		WhiteListType: req.WhiteListType,
		WhiteList:     req.WhiteList,
	}
	var businessConfig = new(featureMdl.BusinessConfig)
	if req.ID != 0 {
		if businessConfig, err = s.dao.BusinessConfigByID(c, req.ID); err != nil {
			log.Error("%+v", err)
			return "", err
		}
		param.KeyName = businessConfig.KeyName
		param.TreeID = businessConfig.TreeID
		param.Creator = businessConfig.Creator
		param.CreatorUID = businessConfig.CreatorUID
		param.State = businessConfig.State
		param.Ctime = businessConfig.Ctime
	}
	if _, err = s.dao.BusinessConfigSave(c, param); err != nil {
		log.Error("%+v", err)
		return "", err
	}
	// log
	_ = s.logWorker.Do(c, func(ctx context.Context) {
		action := _actionUpdate
		before := businessConfig.Config
		after := param.Config
		if req.ID == 0 {
			action = _actionAdd
		}
		obj := map[string]string{
			"before": before,
			"after":  after,
		}
		_ = util.AddLogs(common.LogFeatureBusinessConfig, username, userID, int64(param.ID), action, obj)
	})
	return "", nil
}

func (s *Service) BusinessConfigAct(c context.Context, req *featureMdl.BusinessConfigActReq, userID int64, username string) error {
	if req.State != featureMdl.StateOn && req.State != featureMdl.StateOff {
		log.Error("state(%+v) is invalid", req.State)
		return ecode.RequestErr
	}
	businessConfig, err := s.dao.BusinessConfigByID(c, req.ID)
	if err != nil {
		log.Error("s.dao.BusinessConfigByID(%+v) error(%+v)", req.ID, err)
		return err
	}
	if businessConfig == nil {
		return ecode.NothingFound
	}
	attrs := map[string]interface{}{
		"state":        req.State,
		"modifier":     username,
		"modifier_uid": userID,
	}
	if err := s.dao.UpdateBusinessConfig(c, businessConfig.ID, attrs); err != nil {
		log.Error("s.dao.UpdateBuildLt(%+v, %+v) error(%+v)", businessConfig.ID, attrs, err)
		return err
	}
	// log
	_ = s.logWorker.Do(c, func(ctx context.Context) {
		action := _actionOffline
		before, after := businessConfig.Config, ""
		if req.State == feature.StateOn {
			action = _actionOnline
			before, after = after, before
		}
		obj := map[string]string{
			"before": before,
			"after":  after,
		}
		_ = util.AddLogs(common.LogFeatureBusinessConfig, username, userID, int64(businessConfig.ID), action, obj)
	})
	return nil
}
