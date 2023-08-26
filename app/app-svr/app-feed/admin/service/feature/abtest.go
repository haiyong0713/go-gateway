package feature

import (
	"context"
	"encoding/json"
	"go-common/library/ecode"
	"go-common/library/log"
	"strings"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
	featuremdl "go-gateway/app/app-svr/app-feed/admin/model/feature"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

func (s *Service) ABTestList(c context.Context, params *featuremdl.ABTestReq) (res *featuremdl.ABTestList, err error) {
	abtests, cnt, err := s.dao.SearchABTest(c, params, true)
	if err != nil {
		log.Error("s.dao.SearchABTest(%+v) error(%+v)", params, err)
		return nil, err
	}
	return &featuremdl.ABTestList{
		Page: &common.Page{
			Total: cnt,
			Num:   params.Pn,
			Size:  params.Ps,
		},
		List: abtests,
	}, nil
}

func (s *Service) ABTestSave(c context.Context, userID int, username string, params *featuremdl.ABTestSaveReq) (toast string, err error) {
	// 实验基础配置校验
	if params.Salt != "" {
		switch params.ABType {
		case "mid":
			if params.Salt != "" {
				if !strings.Contains(params.Salt, "%d") && !strings.Contains(params.Salt, "%v") {
					return "salt占位符错误", ecode.RequestErr
				}
				if strings.Contains(params.Salt, "%s") {
					return "salt占位符错误", ecode.RequestErr
				}
			}
		case "buvid":
			if params.Salt != "" {
				if !strings.Contains(params.Salt, "%s") && !strings.Contains(params.Salt, "%v") {
					return "salt格式错误", ecode.RequestErr
				}
				if strings.Contains(params.Salt, "%d") {
					return "salt格式错误", ecode.RequestErr
				}
			}
		default:
			return "实验类型不合法", ecode.RequestErr
		}
		if (strings.Count(params.Salt, "%d") + strings.Count(params.Salt, "%s") + strings.Count(params.Salt, "%v")) > 1 {
			return "salt只支持一个占位符", ecode.RequestErr
		}
	}
	if params.Bucket == 0 {
		log.Error("Fail to save abtest, because bucket 0")
		return "bucket不能为0", ecode.RequestErr
	}
	if !s.ifKeyCanUseABTest(c, params.ID, params.TreeID, params.KeyName) {
		log.Error("Fail to save abtest, because key_name is existed")
		return "key已存在", ecode.RequestErr
	}
	// 实验分组配置校验
	var expConfig []*featuremdl.ABTestExpConfig
	if err = json.Unmarshal([]byte(params.Config), &expConfig); err != nil {
		return "限制条件数据结构不合规", ecode.RequestErr
	}
	if len(expConfig) < 1 {
		return "分组不能为空", ecode.RequestErr
	}
	for _, exp := range expConfig {
		if exp.Start > exp.End {
			return "start不能大于end", ecode.RequestErr
		}
		if exp.Start > params.Bucket || exp.End > params.Bucket {
			return "start不能大于bucet", ecode.RequestErr
		}
	}
	args := &featuremdl.ABTest{
		ID:          params.ID,
		TreeID:      params.TreeID,
		KeyName:     params.KeyName,
		ABType:      params.ABType,
		Bucket:      params.Bucket,
		Salt:        params.Salt,
		Config:      params.Config,
		Creator:     username,
		CreatorUID:  userID,
		Modifier:    username,
		ModifierUID: userID,
		State:       featuremdl.StateOff,
		Relations:   params.Relations,
		Description: params.Description,
	}
	var abtest *featuremdl.ABTest
	if params.ID != 0 {
		if abtest, err = s.dao.GetABTestByID(c, params.ID); err != nil {
			log.Error("s.dao.GetABTestByID(%+v) error(%+v)", params.ID, err)
			return "", err
		}
		if abtest == nil {
			return "", ecode.NothingFound
		}
		args.TreeID = abtest.TreeID
		args.Creator = abtest.Creator
		args.CreatorUID = abtest.CreatorUID
		args.State = abtest.State
		args.Ctime = abtest.Ctime
	}
	if _, err := s.dao.SaveABTest(c, args); err != nil {
		log.Error("s.dao.SaveABTest(%+v) error(%+v)", args, params)
		return "", err
	}
	// log
	_ = s.logWorker.Do(c, func(ctx context.Context) {
		action := _actionUpdate
		before, _ := json.Marshal(abtest)
		after, _ := json.Marshal(args)
		if params.ID == 0 {
			action = _actionAdd
		}
		obj := map[string]string{
			"配置类型":   "实验基础配置",
			"before": string(before),
			"after":  string(after),
		}
		_ = util.AddLogs(common.LogFeatureABTest, username, int64(userID), int64(params.ID), action, obj)
	})
	return "", nil
}

func (s *Service) ifKeyCanUseABTest(ctx context.Context, id, treeID int, keyName string) bool {
	req := &featuremdl.ABTestReq{TreeID: treeID, KeyName: keyName}
	abtests, _, err := s.dao.SearchABTest(ctx, req, false)
	if err != nil {
		log.Error("Fail to search abtest, req=%+v error=%+v", req, err)
		return true
	}
	if len(abtests) == 0 || abtests[0].ID == id {
		return true
	}
	return false
}

func (s *Service) ABTestHandle(c context.Context, userID int, username string, params *featuremdl.ABTestHandleReq) (err error) {
	if params.State != featuremdl.StateOn && params.State != featuremdl.StateOff {
		log.Error("state(%+v) is invalid", params.State)
		return ecode.RequestErr
	}
	abtest, err := s.dao.GetABTestByID(c, params.ID)
	if err != nil {
		log.Error("s.dao.GetABTestByID(%+v) error(%+v)", params.ID, err)
		return err
	}
	if abtest == nil {
		return ecode.NothingFound
	}
	attrs := map[string]interface{}{
		"state":        params.State,
		"modifier":     username,
		"modifier_uid": userID,
	}
	if err = s.dao.UpdateABTest(c, abtest.ID, attrs); err != nil {
		log.Error("s.dao.UpdateABTest(%+v, %+v) error(%+v)", abtest.ID, attrs, err)
		return err
	}
	// log
	_ = s.logWorker.Do(c, func(ctx context.Context) {
		action := _actionOffline
		if params.State == featuremdl.StateOn {
			action = _actionOnline
		}
		obj := map[string]string{
			"before": "",
			"after":  "",
		}
		_ = util.AddLogs(common.LogFeatureABTest, username, int64(userID), int64(abtest.ID), action, obj)
	})
	return nil
}
