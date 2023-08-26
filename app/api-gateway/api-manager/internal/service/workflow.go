package service

import (
	"context"
	"errors"
	"strings"

	"go-common/library/ecode"

	"go-gateway/app/api-gateway/api-manager/internal/model"
	"go-gateway/app/api-gateway/delay"

	"go.uber.org/multierr"
)

// 创建发布 检查是否已生成代码
func (s *Service) CreateWF(ctx context.Context, id int64, apiName string, onlyDeploy bool) (url string, err error) {
	info, err := s.delay.GetLatestWF(ctx, apiName)
	if err != nil || info == nil {
		return
	}
	if info.ID != id {
		return "", ecode.RequestErr
	}
	var wfName string
	if onlyDeploy && info.WFName != "" && info.Image != "" {
		wfName, url, err = s.delay.CreateWorkflow(ctx, apiName, info.Boss, info.Version, info.Image, delay.OnlyDeploy)
	} else if info.WFName == "" {
		wfName, url, err = s.delay.CreateWorkflow(ctx, apiName, info.Boss, info.Version, "", delay.NormalDeploy)
	} else {
		err = ecode.RequestErr
	}
	if err != nil {
		return
	}
	if err = s.delay.UpdateWFName(ctx, info.ID, wfName); err != nil {
		return
	}
	if err = s.delay.UpdateWFState(ctx, info.ID, int(delay.WFStateNormal)); err != nil {
		return
	}
	if err = s.delay.UpdateWFDisplay(ctx, info.ID, "", delay.DisplayStateNormal); err != nil {
		return
	}
	return
}

// 获取发布状态 进入发布页面或前端轮询时调用
// 如果没有发布记录 则从生成代码开始
// 主要是为了拿到db里的主键id
//
//nolint:gocognit
func (s *Service) GetWFStatus(ctx context.Context, apiName string) (res *model.WFStatus, err error) {
	info, err := s.delay.GetLatestWF(ctx, apiName)
	if err != nil || info == nil {
		return
	}
	//如果没有记录或者已结单 返回空
	res = &model.WFStatus{
		ID:           info.ID,
		WFName:       info.WFName,
		DiscoveryID:  info.DiscoveryID,
		DisplayName:  info.DisplayName,
		DisplayState: info.DisplayState,
		CodeAddress:  info.Boss,
		CodeVersion:  info.Version,
		State:        info.State,
		Logs:         info.Log,
	}
	if info.ID == 0 || info.State == delay.WFStateFinished {
		res = &model.WFStatus{}
		return
	}
	//处于编译阶段 用db里存的状态返回即可
	if info.WFName == "" || info.State == delay.WFStateFailed || info.State == delay.WFStateManualStop {
		return
	}
	//已经存在发布单 查发布单状态
	phase, displayName, err := s.delay.GetWorkflowStatus(ctx, info.WFName)
	res.DisplayName = displayName
	res.DisplayState = delay.DisplayStatusMap[phase]
	if err != nil {
		if phase == delay.NodeStatusFailed {
			errLog, err1 := s.delay.GetLogWorkflow(ctx, info.WFName, displayName)
			if err1 != nil {
				err = multierr.Append(err, err1)
				return
			} else {
				res.Logs = errLog
				err = multierr.Append(err, errors.New(errLog))
			}
			if err1 = s.delay.UpdateFailedWF(ctx, info.ID, res.DisplayName, res.DisplayState, delay.DisplayStateFailed, errLog); err1 != nil {
				err = multierr.Append(err, err1)
				return
			}
			res.State = delay.DisplayStateFailed
		}
		return
	}
	if err = s.delay.UpdateWFDisplay(ctx, info.ID, res.DisplayName, res.DisplayState); err != nil {
		return
	}
	if displayName == delay.DisplayNameProdDeploy && phase == delay.NodeStatusSucceeded {
		if err = s.delay.UpdateWFState(ctx, info.ID, int(delay.WFStateFinished)); err != nil {
			return
		}
		res.State = delay.WFStateFinished
	}

	//更新镜像地址
	if info.Image == "" && delay.DisplayNameMap[displayName] > delay.DisplayBuild {
		_ = s.wfFanout.Do(ctx, func(ctx context.Context) {
			var data delay.OutputsData
			if data, err = s.delay.OutputWorkflow(ctx, info.WFName, delay.DisplayNameBuild); err != nil || len(data.Parameters) == 0 {
				return
			}
			for _, p := range data.Parameters {
				if p.Name != "image_name" || p.Value == "" {
					continue
				}
				_ = s.delay.UpdateWFImage(ctx, info.ID, p.Value)
				break
			}
		})
	}
	//更新discoveryID
	if info.DiscoveryID == "" && delay.DisplayNameMap[displayName] > delay.DisplayUatCreatApp {
		_ = s.wfFanout.Do(ctx, func(ctx context.Context) {
			var data delay.OutputsData
			if data, err = s.delay.OutputWorkflow(ctx, info.WFName, delay.DisplayNameUatCreatApp); err != nil || len(data.Parameters) == 0 {
				return
			}
			for _, p := range data.Parameters {
				if p.Name != "discovery_id" || p.Value == "" {
					continue
				}
				var pathInfo *model.DynpathParam
				if pathInfo, err = s.DynPath(ctx, apiName); err != nil {
					return
				}
				if err = s.dao.GWAddPath(ctx, pathInfo, p.Value); err != nil {
					return
				}
				_ = s.delay.UpdateWFDis(ctx, info.ID, p.Value)
				break
			}
		})
	}
	return
}

// 继续发布
func (s *Service) ResumeWF(ctx context.Context, id int64, apiName string) (err error) {
	info, err := s.delay.GetLatestWF(ctx, apiName)
	if err != nil || info == nil {
		return
	}
	if info.ID != id || info.WFName == "" {
		return ecode.RequestErr
	}
	if !strings.Contains(info.DisplayName, "suspend") {
		return ecode.RequestErr
	}
	err = s.delay.ResumeWorkflow(ctx, info.WFName, info.DisplayName)
	return
}

// 手动停止发布
func (s *Service) StopWF(ctx context.Context, id int64, apiName string) (err error) {
	info, err := s.delay.GetLatestWF(ctx, apiName)
	if err != nil || info == nil {
		return
	}
	if info.ID != id || info.WFName == "" {
		return ecode.RequestErr
	}
	return s.delay.StopWorkflow(ctx, info.WFName)
	//err = s.delay.UpdateWFState(ctx, info.ID, int(delay.WFStateManualStop))
}
