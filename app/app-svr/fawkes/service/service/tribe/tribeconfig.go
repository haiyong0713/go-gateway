package tribe

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"

	"github.com/golang/protobuf/ptypes/empty"

	"go-gateway/app/app-svr/fawkes/service/api/app/tribe"
	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// ActiveVersion 版本-生效
func (s *Service) ActiveVersion(ctx context.Context, req *tribe.ActiveVersionReq) (resp *empty.Empty, err error) {
	var isActive int8
	op := utils.GetUsername(ctx)
	if req.Active {
		isActive = tribemdl.CdActive
	} else {
		isActive = tribemdl.CdNotActive
	}
	if _, err = s.fkDao.UpdatePackVersionStatus(ctx, req.VersionId, isActive, op); err != nil {
		log.Errorc(ctx, "%v", err)
	}
	return
}

// ConfigVersionFlow 版本-流量配置
func (s *Service) ConfigVersionFlow(ctx context.Context, req *tribe.ConfigVersionFlowReq) (resp *empty.Empty, err error) {
	var flowInfo tribemdl.FlowInfo
	var gitJobIds []string
	op := utils.GetUsername(ctx)
	for _, v := range req.Flow {
		flowInfo.Flows = append(flowInfo.Flows, strconv.FormatInt(v.From, 10)+tribemdl.Comma+strconv.FormatInt(v.To, 10))
		flowInfo.GitlabJobIds = append(flowInfo.GitlabJobIds, v.GitJobId)
		gitJobIds = append(gitJobIds, strconv.FormatInt(v.GitJobId, 10))
	}
	if err = s.fkDao.BatchAddPackFlowConfig(ctx, req.TribeId, req.Env, req.VersionId, op, &flowInfo); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	_, _ = s.fkDao.AddLog(ctx, req.AppKey, req.Env, mngmdl.ModelCD, mngmdl.OperationCDFlowConfig, fmt.Sprintf("构建ID: %v", strings.Join(gitJobIds, ",")), op)
	return
}

func (s *Service) GetVersionFlow(ctx context.Context, req *tribe.GetVersionFlowReq) (resp *tribe.GetVersionFlowResp, err error) {
	var (
		packs     []*tribemdl.Pack
		gitJobIds []int64
		flows     []*tribemdl.ConfigFlow
		tribeFlow []*tribe.Flow
	)
	if packs, err = s.fkDao.SelectTribePackByVersions(ctx, req.TribeId, req.Env, []int64{req.VersionId}); err != nil {
		log.Errorc(ctx, "%v or find no rows", err)
		return
	}
	for _, v := range packs {
		gitJobIds = append(gitJobIds, v.GlJobId)
	}
	if flows, err = s.fkDao.SelectTribePackConfigFlow(ctx, req.TribeId, req.Env, gitJobIds); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	for _, v := range flows {
		ft := strings.SplitAfter(v.Flow, tribemdl.Comma)
		from, _ := strconv.ParseInt(ft[0], 10, 64)
		to, _ := strconv.ParseInt(ft[1], 10, 64)
		tribeFlow = append(tribeFlow, &tribe.Flow{
			From:     from,
			To:       to,
			GitJobId: v.GlJobId,
		})
	}
	resp = &tribe.GetVersionFlowResp{
		Flows: tribeFlow,
	}
	return
}

// ConfigVersionUpgrade 版本-升级配置 应用配置
func (s *Service) ConfigVersionUpgrade(ctx context.Context, req *tribe.ConfigVersionUpgradeReq) (resp *empty.Empty, err error) {
	resp = &empty.Empty{}
	if _, err = s.fkDao.AddTribeConfigUpgrade(ctx, req.TribeId, req.Env, req.TribePackId, strings.Join(req.StartingVersionCode, tribemdl.Comma), strings.Join(req.ChosenVersionCode, tribemdl.Comma)); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	return
}

func (s *Service) GetConfigVersionUpgrade(ctx context.Context, req *tribe.GetConfigVersionUpgradeReq) (resp *tribe.GetConfigVersionUpgradeResp, err error) {
	var upgrade *tribemdl.PackUpgrade
	if upgrade, err = s.fkDao.SelectTribeConfigUpgrade(ctx, req.TribeId, req.Env, req.TribePackId); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if upgrade == nil {
		return
	}
	resp = &tribe.GetConfigVersionUpgradeResp{}
	if upgrade.ChosenVersionCode != "" {
		resp.ChosenVersionCode = strings.Split(upgrade.ChosenVersionCode, tribemdl.Comma)
	}
	if upgrade.StartVersionCode != "" {
		resp.StartingVersionCode = strings.Split(upgrade.StartVersionCode, tribemdl.Comma)
	}
	return
}

// ConfigPackUpgradeFilter 包-升级配置
func (s *Service) ConfigPackUpgradeFilter(ctx context.Context, req *tribe.ConfigPackUpgradeFilterReq) (resp *empty.Empty, err error) {
	resp = &empty.Empty{}
	op := utils.GetUsername(ctx)
	if req.Type == tribemdl.PackFilterTypeCustom {
		if req.Percent == 0 && req.DeviceId == "" {
			err = ecode.Error(ecode.RequestErr, "自定义模式下，升级比例和设备需至少配置一项")
			log.Errorc(ctx, err.Error())
			return
		}
	}
	salt := salt()
	if _, err = s.fkDao.AddTribePackFilterConfig(ctx, req.TribeId, req.Env, req.TribePackId, req.Isp, req.Network, req.Channel, req.City, req.DeviceId, int64(req.Type), req.Percent, salt, req.ExcludesSystem, op); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	// add log
	_, _ = s.fkDao.AddLog(ctx, req.AppKey, req.Env, mngmdl.ModelCD, mngmdl.OperationCDFilterConfig, fmt.Sprintf("tribe 构建ID: %v", req.TribePackId), op)
	return
}

func (s *Service) GetConfigPackUpgradeFilter(ctx context.Context, req *tribe.GetConfigPackUpgradeFilterReq) (resp *tribe.GetConfigPackUpgradeFilterResp, err error) {
	var f *tribemdl.ConfigFilter
	if f, err = s.fkDao.SelectTribeConfigPackFilter(ctx, req.TribeId, req.Env, req.TribePackId); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if f == nil {
		return
	}
	resp = &tribe.GetConfigPackUpgradeFilterResp{
		TribeId:        f.TribeId,
		Env:            f.Env,
		BuildId:        f.TribePackId,
		Network:        f.Network,
		Isp:            f.Isp,
		Channel:        f.Channel,
		City:           f.City,
		Type:           tribe.UpgradeType(f.Type),
		Percent:        int64(f.Percent),
		DeviceId:       f.Device,
		Salt:           f.Salt,
		ExcludesSystem: f.ExcludesSystem,
		Operator:       f.Operator,
	}
	return
}

func salt() string {
	kinds := [][]int{{10, 48}, {26, 97}}
	keyb := make([]byte, 8)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 8; i++ {
		ikind := rand.Intn(2)
		scope, base := kinds[ikind][0], kinds[ikind][1]
		keyb[i] = uint8(base + rand.Intn(scope))
	}
	return string(keyb)
}
