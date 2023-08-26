package tribe

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/ecode"

	"github.com/golang/protobuf/ptypes/empty"

	"go-gateway/app/app-svr/fawkes/service/tools/utils"

	"go-gateway/app/app-svr/fawkes/service/api/app/tribe"
	mngmdl "go-gateway/app/app-svr/fawkes/service/model/manager"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

type VersionPackList []*tribe.VersionPack
type PackList []*tribe.PackInfo

func (s *Service) ListTribePack(ctx context.Context, req *tribe.ListTribePackReq) (resp *tribe.ListTribePackResp, err error) {
	var (
		total                          int64
		versionIds, packIds, gitJobIds []int64
		tpv                            []*tribemdl.PackVersion
		packs                          []*tribemdl.Pack
		packInfos                      VersionPackList
		flowMap                        map[int64]*tribemdl.ConfigFlow
		filterMap                      map[int64]*tribemdl.ConfigFilter
		versionNameMap                 map[string][]*tribemdl.PackVersion
		versionPackMap                 map[int64]*tribemdl.Pack
	)
	appInfo, err := s.fkDao.AppPass(ctx, req.AppKey)
	if err != nil {
		log.Errorc(ctx, "%v", err)
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("app [%s] not found", req.AppKey))
		return
	}
	if total, err = s.fkDao.CountPackVersionByOptions(ctx, req.TribeId, req.Env); err != nil {
		log.Errorc(ctx, "error: %v", err)
		return
	}
	if total == 0 {
		return
	}
	if tpv, err = s.fkDao.TribePackVersionListByOptions(ctx, req.TribeId, req.Env, req.Pn, req.Ps); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	for _, v := range tpv {
		versionIds = append(versionIds, v.Id)
	}
	if packs, err = s.fkDao.SelectTribePackByVersions(ctx, req.TribeId, req.Env, versionIds); err != nil {
		log.Errorc(ctx, "SelectTribePackByVersions err: %v", err)
		return
	}
	for _, v := range packs {
		packIds = append(packIds, v.Id)
		gitJobIds = append(gitJobIds, v.GlJobId)
	}
	// 包升级过滤信息 key-tribe_pack_id
	if filterMap, err = s.packConfigFilter(ctx, req.TribeId, req.Env, packIds); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	// 流量信息  key-gl_job_id
	if flowMap, err = s.packConfigFlow(ctx, req.TribeId, req.Env, gitJobIds); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	versionNameMap = groupByVersionName(tpv)
	versionPackMap = groupByVersionId(packs)

	for name, versions := range versionNameMap {
		vp := &tribe.VersionPack{}
		vp.VersionName = name
		var packList PackList
		for _, v := range versions {
			packInfo := tribePackPO2DTO(versionPackMap[v.Id])
			var (
				filter           *tribemdl.ConfigFilter
				flow             *tribemdl.ConfigFlow
				filterOk, flowOk bool
			)
			if filter, filterOk = filterMap[packInfo.Id]; filterOk {
				packInfo.PackUpgrade = convertFilter(filter)
			}
			if flow, flowOk = flowMap[packInfo.GlJobId]; flowOk {
				packInfo.Flow = convertFlow(flow)
			}
			packInfo.JobUrl = MakeGitPath(appInfo.GitPath, packInfo.GlJobId)
			packInfo.VersionInfo = convertVersion(v)
			packList = append(packList, packInfo)
			operator, mtime := lastOperator(packInfo, filter, flow, v)
			packInfo.LastOperator = operator
			packInfo.LastMtime = mtime
		}
		sort.Sort(packList)
		vp.PackInfo = packList
		packInfos = append(packInfos, vp)
	}
	sort.Sort(packInfos)
	resp = &tribe.ListTribePackResp{
		PageInfo:        &tribe.PageInfo{Total: total, Pn: req.Pn, Ps: req.Ps},
		VersionPackInfo: packInfos,
	}
	return
}

func groupByVersionId(packs []*tribemdl.Pack) map[int64]*tribemdl.Pack {
	m := make(map[int64]*tribemdl.Pack)
	for _, v := range packs {
		m[v.VersionId] = v
	}
	return m
}

// 按照version_name分组
func groupByVersionName(tpv []*tribemdl.PackVersion) map[string][]*tribemdl.PackVersion {
	var m = make(map[string][]*tribemdl.PackVersion)
	for _, v := range tpv {
		if pv, ok := m[v.VersionName]; ok {
			tmp := pv
			tmp = append(tmp, v)
			m[v.VersionName] = tmp
		} else {
			m[v.VersionName] = []*tribemdl.PackVersion{v}
		}
	}
	return m
}

func (s *Service) EvolutionTribePack(ctx context.Context, req *tribe.EvolutionTribeReq) (resp *empty.Empty, err error) {
	var (
		p       *tribemdl.Pack
		version *tribemdl.PackVersion
	)
	resp = new(empty.Empty)
	if p, err = s.fkDao.SelectTribePackById(ctx, req.TribePackId); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	if p == nil {
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("pack id[%d] not found", req.TribePackId))
		log.Errorc(ctx, err.Error())
		return
	}
	if version, err = s.fkDao.SelectTribePackVersionById(ctx, p.VersionId); err != nil || version == nil {
		err = ecode.Error(ecode.NothingFound, fmt.Sprintf("tribe pack version id[%d] not found", p.VersionId))
		log.Errorc(ctx, "%v", err)
		return
	}
	var packVersionId int64
	if err = s.fkDao.Transact(ctx, func(tx *sql.Tx) error {
		var prodVersion *tribemdl.PackVersion
		// select version id if not exist insert
		if prodVersion, err = s.fkDao.TxSelectTribePackVersionForUpdate(tx, p.TribeId, tribemdl.ProdEnv, version.VersionCode); err != nil {
			log.Errorc(ctx, "error: %v", err)
			return err
		}
		if prodVersion == nil || prodVersion.Id == 0 {
			if packVersionId, err = s.fkDao.TxSetTribePackVersion(tx, p.TribeId, tribemdl.ProdEnv, version.VersionCode, version.VersionName, false); err != nil {
				log.Errorc(ctx, "error: %v", err)
				return err
			}
		} else {
			err = ecode.Error(ecode.RequestErr, "已推送到正式环境，请勿重复推送")
			log.Errorc(ctx, err.Error())
			return err
		}
		// 同步到prod环境
		if _, err = s.fkDao.TxCopyTribePack(tx, p, packVersionId, tribemdl.ProdEnv, req.Description, utils.GetUsername(ctx)); err != nil {
			log.Errorc(ctx, "error: %v", err)
			return err
		}
		return err
	}); err != nil {
		log.Errorc(ctx, "error: %v", err)
		return
	}
	_, _ = s.fkDao.AddLog(ctx, p.AppKey, tribemdl.ProdEnv, mngmdl.ModelCD, mngmdl.OperationCDPushProd, fmt.Sprintf("构建ID: %v", p.GlJobId), utils.GetUsername(ctx))
	return
}

// key-gl_job_id
func (s *Service) packConfigFlow(ctx context.Context, tribeId int64, env string, gitlabJobIds []int64) (flowMap map[int64]*tribemdl.ConfigFlow, err error) {
	// 版本下各个pack的流量设置
	var flow []*tribemdl.ConfigFlow
	if flow, err = s.fkDao.SelectTribePackConfigFlow(ctx, tribeId, env, gitlabJobIds); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	flowMap = make(map[int64]*tribemdl.ConfigFlow)
	for _, v := range flow {
		flowMap[v.GlJobId] = v
	}
	return
}

// key tribe_pack_id
func (s *Service) packConfigFilter(ctx context.Context, tribeId int64, env string, packIds []int64) (filterMap map[int64]*tribemdl.ConfigFilter, err error) {
	// 包维度设置
	var filter []*tribemdl.ConfigFilter
	if filter, err = s.fkDao.SelectTribeConfigPackFilterByPackIds(ctx, tribeId, env, packIds); err != nil {
		log.Errorc(ctx, "%v", err)
		return
	}
	filterMap = make(map[int64]*tribemdl.ConfigFilter)
	for _, v := range filter {
		filterMap[v.TribePackId] = v
	}
	return
}

func tribePackPO2DTO(po *tribemdl.Pack) (tribeInfo *tribe.PackInfo) {
	tribeInfo = &tribe.PackInfo{
		Id:          po.Id,
		AppId:       po.AppId,
		AppKey:      po.AppKey,
		Env:         po.Env,
		TribeId:     po.TribeId,
		GlJobId:     po.GlJobId,
		DepGlJobId:  po.DepGlJobId,
		DepFeature:  po.DepFeature,
		VersionId:   po.VersionId,
		GitType:     int32(po.GitType),
		GitName:     po.GitName,
		Commit:      po.Commit,
		PackType:    int32(po.PackType),
		ChangeLog:   po.ChangeLog,
		Operator:    po.Operator,
		Size_:       po.Size,
		Md5:         po.Md5,
		PackPath:    po.PackPath,
		PackUrl:     po.PackUrl,
		MappingUrl:  po.MappingUrl,
		BbrUrl:      po.BbrUrl,
		CdnUrl:      po.CdnUrl,
		Description: po.Description,
		Mtime:       po.Mtime.Unix(),
		Ctime:       po.Ctime.Unix(),
	}
	return
}

func convertFlow(f *tribemdl.ConfigFlow) *tribe.Flow {
	ft := strings.SplitAfter(f.Flow, tribemdl.Comma)
	from, _ := strconv.ParseInt(ft[0], 10, 64)
	to, _ := strconv.ParseInt(ft[1], 10, 64)
	return &tribe.Flow{
		From:     from,
		To:       to,
		GitJobId: f.GlJobId,
		Ctime:    f.Ctime.Unix(),
		Mtime:    f.Mtime.Unix(),
		Operator: f.Operator,
	}
}

func convertFilter(f *tribemdl.ConfigFilter) *tribe.GetConfigPackUpgradeFilterResp {
	return &tribe.GetConfigPackUpgradeFilterResp{
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
		Mtime:          f.Mtime.Unix(),
		Ctime:          f.Ctime.Unix(),
		Operator:       f.Operator,
	}
}

func convertVersion(v *tribemdl.PackVersion) *tribe.VersionInfo {
	return &tribe.VersionInfo{
		Env:         v.Env,
		IsActive:    v.IsActive == tribemdl.CdActive,
		VersionCode: strconv.FormatInt(v.VersionCode, 10),
		VersionId:   v.Id,
		Mtime:       v.Mtime.Unix(),
		Ctime:       v.Ctime.Unix(),
		Operator:    v.Operator,
	}
}

// 最近操作人
func lastOperator(info *tribe.PackInfo, filter *tribemdl.ConfigFilter, flow *tribemdl.ConfigFlow, version *tribemdl.PackVersion) (operator string, mtime int64) {
	operator = info.Operator
	mtime = info.Mtime
	if filter != nil && len(filter.Operator) > 0 && mtime < filter.Mtime.Unix() {
		operator = filter.Operator
		mtime = filter.Mtime.Unix()
	}
	if flow != nil && len(flow.Operator) > 0 && mtime < flow.Mtime.Unix() {
		operator = flow.Operator
		mtime = flow.Mtime.Unix()
	}
	if version != nil && len(version.Operator) > 0 && mtime < version.Mtime.Unix() {
		operator = version.Operator
		mtime = version.Mtime.Unix()
	}
	return
}

func (vpl VersionPackList) Len() int {
	return len(vpl)
}

func (vpl VersionPackList) Less(i int, j int) bool {
	var packMaxJobIdI int64
	var packMaxJobIdJ int64
	packMaxJobIdI = max(vpl[i].PackInfo)
	packMaxJobIdJ = max(vpl[j].PackInfo)
	return packMaxJobIdI > packMaxJobIdJ
}

func (vpl VersionPackList) Swap(i int, j int) {
	vpl[i], vpl[j] = vpl[j], vpl[i]
}

func max(p []*tribe.PackInfo) int64 {
	var maxJobId int64
	for _, v := range p {
		if v.GlJobId > maxJobId {
			maxJobId = v.GlJobId
		}
	}
	return maxJobId
}

func (pl PackList) Len() int {
	return len(pl)
}

func (pl PackList) Less(i int, j int) bool {
	return pl[i].GlJobId > pl[j].GlJobId
}

func (pl PackList) Swap(i int, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}
