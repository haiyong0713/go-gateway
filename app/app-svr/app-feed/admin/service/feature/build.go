package feature

import (
	"context"
	"encoding/json"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/feature"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	_actionAdd     = "add"
	_actionUpdate  = "update"
	_actionOnline  = "online"
	_actionOffline = "offline"
)

func (s *Service) BuildList(ctx context.Context, req *feature.BuildListReq) (*feature.BuildListRly, error) {
	buildLts, cnt, err := s.dao.SearchBuildLt(ctx, req, true, true)
	if err != nil {
		log.Error("s.dao.SearchBuildLt(%+v) error(%+v)", req, err)
		return nil, err
	}
	return &feature.BuildListRly{
		Page: &common.Page{
			Total: cnt,
			Num:   req.Pn,
			Size:  req.Ps,
		},
		List: buildLts,
	}, nil
}

func (s *Service) SaveBuild(ctx context.Context, userID int, username string, req *feature.SaveBuildReq) (string, error) {
	if err := s.validateBuildConf(ctx, req.Config, req.TreeID); err != nil {
		log.Error("s.validateBuildConf(%+v) error(%+v)", req.Config, err)
		return "", err
	}
	if !s.ifKeyCanUse(ctx, req.ID, req.TreeID, req.KeyName) {
		log.Error("Fail to save build_limit, because key_name is existed")
		return "key已存在", ecode.RequestErr
	}
	var buildLt = new(feature.BuildLimit)
	attrs := &feature.BuildLimit{
		ID:          req.ID,
		TreeID:      req.TreeID,
		KeyName:     req.KeyName,
		Config:      strings.TrimSpace(req.Config),
		Creator:     username,
		CreatorUID:  userID,
		Modifier:    username,
		ModifierUID: userID,
		State:       feature.StateOff,
		Description: req.Description,
	}
	if req.ID != 0 {
		buildLt, err := s.dao.GetBuildLtByID(ctx, req.ID)
		if err != nil {
			log.Error("s.dao.GetBuildLtByID(%+v) error(%+v)", req.ID, err)
			return "", err
		}
		if buildLt == nil {
			return "", ecode.NothingFound
		}
		attrs.TreeID = buildLt.TreeID
		attrs.Creator = buildLt.Creator
		attrs.CreatorUID = buildLt.CreatorUID
		attrs.State = buildLt.State
		attrs.Ctime = buildLt.Ctime
	}
	if _, err := s.dao.SaveBuildLt(ctx, attrs); err != nil {
		log.Error("s.dao.SaveBuildLt(%+v) error(%+v)", attrs, req)
		return "", err
	}
	// log
	_ = s.logWorker.Do(ctx, func(ctx context.Context) {
		action := _actionUpdate
		before := buildLt.Config
		after := attrs.Config
		if req.ID == 0 {
			action = _actionAdd
		}
		obj := map[string]string{
			"before": before,
			"after":  after,
		}
		_ = util.AddLogs(common.LogFeatureBuild, username, int64(userID), int64(attrs.ID), action, obj)
	})
	return "", nil
}

func (s *Service) validateBuildConf(ctx context.Context, config string, treeID int) error {
	var confItems []*feature.BuildConfItem
	err := json.Unmarshal([]byte(config), &confItems)
	if err != nil {
		log.Error("json.Unmarshal(%+v) error(%+v)", config, err)
		return ecode.RequestErr
	}

	//nolint:ineffassign,staticcheck
	mobiApps, err := s.dao.GetSvrAttrPlats(ctx, treeID)
	for _, confItem := range confItems {
		if _, ok := mobiApps[confItem.MobiApp]; !ok {
			log.Error("tree(%+v) don't have mobi_app(%+v)", treeID, confItem.MobiApp)
			return ecode.RequestErr
		}
		for _, condition := range confItem.Conditions {
			if _, ok := feature.OpList[condition.Op]; !ok || condition.Build <= 0 {
				log.Error("condition(%+v) is invalid", condition)
				return ecode.RequestErr
			}
		}
	}
	return nil
}

func (s *Service) HandleBuild(ctx context.Context, userID int, username string, req *feature.HandleBuildReq) error {
	if req.State != feature.StateOn && req.State != feature.StateOff {
		log.Error("state(%+v) is invalid", req.State)
		return ecode.RequestErr
	}
	buildLt, err := s.dao.GetBuildLtByID(ctx, req.ID)
	if err != nil {
		log.Error("s.dao.GetBuildLtByID(%+v) error(%+v)", req.ID, err)
		return err
	}
	if buildLt == nil {
		return ecode.NothingFound
	}
	attrs := map[string]interface{}{
		"state":        req.State,
		"modifier":     username,
		"modifier_uid": userID,
	}
	if err := s.dao.UpdateBuildLt(ctx, buildLt.ID, attrs); err != nil {
		log.Error("s.dao.UpdateBuildLt(%+v, %+v) error(%+v)", buildLt.ID, attrs, err)
		return err
	}
	// log
	_ = s.logWorker.Do(ctx, func(ctx context.Context) {
		action := _actionOffline
		before, after := buildLt.Config, ""
		if req.State == feature.StateOn {
			action = _actionOnline
			before, after = after, before
		}
		obj := map[string]string{
			"before": before,
			"after":  after,
		}
		_ = util.AddLogs(common.LogFeatureBuild, username, int64(userID), int64(buildLt.ID), action, obj)
	})
	return nil
}

func (s *Service) ifKeyCanUse(ctx context.Context, id, treeID int, keyName string) bool {
	req := &feature.BuildListReq{TreeID: treeID, KeyName: keyName}
	buildLts, _, err := s.dao.SearchBuildLt(ctx, req, false, false)
	if err != nil {
		log.Error("Fail to search build_limit, req=%+v error=%+v", req, err)
		return false
	}
	for _, buildLt := range buildLts {
		if buildLt != nil && buildLt.ID != id {
			return false
		}
	}
	return true
}
