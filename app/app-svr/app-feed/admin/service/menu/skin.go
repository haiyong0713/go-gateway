package menu

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	menuModel "go-gateway/app/app-svr/app-feed/admin/model/menu"
	"go-gateway/app/app-svr/app-feed/admin/util"
	xecode "go-gateway/app/app-svr/app-feed/ecode"

	"go-common/library/sync/errgroup.v2"

	garb "git.bilibili.co/bapis/bapis-go/garb/service"
)

// MenuSkinList .
func (s *Service) MenuSkinList(c context.Context, arg *menuModel.SkinListParam) (pager *menuModel.SkinListReply, err error) {
	var (
		rely  *menuModel.SkinSearchReply
		skins *garb.SkinListReply
	)
	if rely, err = s.showDao.SkinExts(c, arg.SID, arg.Pn, arg.Ps); err != nil || rely == nil {
		return
	}
	pager = &menuModel.SkinListReply{Page: &menuModel.Page{Num: arg.Pn, Size: arg.Ps, Total: rely.Total}}
	var (
		tIDs       []int64
		skinLimits map[int64][]*menuModel.SkinLimit
		skinIDs    []int64
	)
	for _, v := range rely.List {
		tIDs = append(tIDs, v.ID)
		skinIDs = append(skinIDs, v.SkinID)
	}
	if len(tIDs) > 0 {
		eg := errgroup.WithContext(c)
		eg.Go(func(ctx context.Context) (e error) {
			skinLimits, e = s.showDao.SkinLimits(ctx, tIDs)
			return
		})
		eg.Go(func(ctx context.Context) (e error) {
			skins, e = s.garbDao.SkinInfos(ctx, skinIDs)
			return
		})
		if err = eg.Wait(); err != nil {
			log.Error("MenuSkinList eg.wait error(%v)", err)
			return
		}
	}
	//拼接信息
	for _, v := range rely.List {
		tem := &menuModel.SkinList{SkinExt: v}
		if skins != nil {
			if sv, sok := skins.Skins[v.SkinID]; sok {
				tem.Image = sv.Preview
			}
		}
		if lv, okk := skinLimits[v.ID]; okk {
			tem.Limit = lv
		}
		pager.List = append(pager.List, tem)
	}
	return
}

// MenuSkinSave .
func (s *Service) MenuSkinSave(c context.Context, arg *menuModel.SkinSaveParam, operator string, uid int64) (rly *menuModel.SkinSaveReply, err error) {
	var (
		opAction   string
		skinRly    *garb.SkinListReply
		oldSkinExt *menuModel.SkinExt
		skinExt    = &menuModel.SkinExt{
			ID:             arg.ID,
			Attribute:      arg.Attribute,
			Stime:          arg.Stime,
			Etime:          arg.Etime,
			Operator:       operator,
			SkinID:         arg.SkinID,
			SkinName:       arg.SkinName,
			UserScopeType:  arg.UserScopeType,
			UserScopeValue: arg.UserScopeValue,
			DressUpValue:   arg.DressUpValue,
			DressUpType:    arg.DressUpType,
		}
	)
	// [stime,etime)
	if arg.Stime >= arg.Etime {
		err = ecode.RequestErr
		return
	}
	buildLimit := make([]*menuModel.SkinBuildLimit, 0)
	if err = json.Unmarshal([]byte(arg.Limit), &buildLimit); err != nil {
		log.Error("MenuSkinSave json.Unmarshal %s error(%v)", arg.Limit, err)
		return
	}
	if len(buildLimit) == 0 {
		err = ecode.RequestErr
		return
	}
	lastBuildLimit := make(map[int8]*menuModel.SkinBuildLimit)
	for _, v := range buildLimit {
		if err = v.ValidateParam(); err != nil {
			return
		}
		lastBuildLimit[v.Plat] = v
	}
	// check skinid
	if skinRly, err = s.garbDao.SkinInfos(c, []int64{arg.SkinID}); err != nil || skinRly == nil {
		log.Error("MenuSkinSave s.garbDao.SkinInfo(%d) error(%v)", arg.SkinID, err)
		err = xecode.SkinResourceNotFound
		return
	}
	if _, sok := skinRly.Skins[arg.SkinID]; !sok {
		err = xecode.SkinResourceNotFound
		return
	}
	if arg.ID > 0 {
		if oldSkinExt, err = s.showDao.RawSkinExt(c, arg.ID); err != nil || oldSkinExt == nil || oldSkinExt.ID == 0 {
			log.Error("s.showDao.RawSkinExt (%d) error(%v) or nil", arg.ID, err)
			err = xecode.SkinConfigNotFound
			return
		}
		if oldSkinExt.State == 1 {
			// 上线后不可修改主题id
			if oldSkinExt.SkinID != skinExt.SkinID {
				err = xecode.SkinOnlineConfigResourceNotModifiable
				return
			}

			// check stime 和 etime
			var cOK bool
			if cOK, err = s.CheckSkinTimeLimit(c, skinExt); err != nil {
				log.Error("feed-admin.Service.menu.MenuSkinSave.CheckSkinTimeLimit Error (%v)", err)
				return
			}
			if !cOK {
				err = xecode.TabTimeLimit
				return
			}
		}
		opAction = common.ActionUpdate
	} else {
		opAction = common.ActionAdd
	}
	// 新建策略组
	var oldSkinLocationPGID int64 = 0
	if oldSkinExt != nil {
		oldSkinLocationPGID = oldSkinExt.LocationPolicyGroupID
	}
	var addedPolicyGroupID int64
	if addedPolicyGroupID, err = s.SkinAddLocationPolicy(c, oldSkinLocationPGID, arg.AreaIDs, operator); err != nil {
		log.Error("MenuSkinSave SkinAddLocationPolicy Error (%v)", err)
		return
	}
	skinExt.LocationPolicyGroupID = addedPolicyGroupID
	// 最终保存
	rly = &menuModel.SkinSaveReply{}
	//nolint:staticcheck
	rly.ID, err = s.showDao.SkinExtSave(c, skinExt, lastBuildLimit)
	// 操作日志
	if err = s.SkinSaveLog(c, uid, rly.ID, opAction, operator); err != nil {
		log.Error("MenuSkinSave AddLog error(%v)", err)
		return
	}
	return
}

// SkinAddLocationPolicy 主题下发新增区域限制策略组
func (s *Service) SkinAddLocationPolicy(ctx context.Context, policyGroupID int64, areaIDsStr string, username string) (pgid int64, err error) {
	// 若不传Area IDs或传`all`，则为选择了`全球`，即不做限制
	if areaIDsStr == "" || areaIDsStr == "all" {
		return 0, nil
	}
	policyGroupName := fmt.Sprintf("%s%d", menuModel.SkinLocationPolicyGroupNamePrefix, time.Now().Unix())
	policyGroupBusinessSource := menuModel.SkinLocationBusinessSource
	policyGroupType := menuModel.SkinLocationPolicyGroupTypeSkin
	policyGroupRemark := menuModel.SkinLocationPolicyGroupRemark
	policyPlayAuth := menuModel.SkinLocationPolicyPlayAuth
	policyDownAuth := menuModel.SkinLocationPolicyDownAuth
	var areaIDs []int64
	areaIDStrs := strings.Split(areaIDsStr, ",")
	areaIDs = make([]int64, 0, len(areaIDStrs))
	for _, areaIDStr := range areaIDStrs {
		areaID, err := strconv.ParseInt(areaIDStr, 10, 64)
		if err != nil {
			return 0, ecode.New(148002)
		}
		areaIDs = append(areaIDs, areaID)
	}
	if pgid, err = s.locDao.AddGroupWithItems(ctx, policyGroupName, policyGroupType, policyGroupRemark, areaIDs, policyPlayAuth, policyDownAuth, policyGroupBusinessSource, username); err != nil {
		//nolint:govet
		log.Error("", err)
		return
	}

	// 若已存在PolicyGroup，删除不需要的数据
	if policyGroupID != 0 {
		//nolint:biligowordcheck
		go func() {
			if derr := s.locDao.DeleteGroup(ctx, policyGroupID, policyGroupBusinessSource, username); derr != nil {
				log.Error("feed-admin.Service.menu.SkinAddLocationPolicy.DeleteGroup(%d) Error (%v)", policyGroupID, derr)
			}
			log.Info("feed-admin.Service.menu.SkinAddLocationPolicy.DeleteGroup(%d) Done", policyGroupID)
		}()
	}

	return
}

// SkinSaveLog .
func (s *Service) SkinSaveLog(c context.Context, uid, id int64, opAction, operator string) (err error) {
	var (
		oldSkinExt *menuModel.SkinExt
		skinLimits map[int64][]*menuModel.SkinLimit
	)
	if oldSkinExt, err = s.showDao.RawSkinExt(c, id); err != nil || oldSkinExt == nil {
		log.Error("SkinSaveLog s.showDao.RawSkinExt (%d) error(%v) or nil", id, err)
		return
	}
	if skinLimits, err = s.showDao.SkinLimits(c, []int64{id}); err != nil {
		log.Error("SaveLog s.showDao.SkinLimits (%d) error(%v)", id, err)
		return
	}
	arg := make(map[string]interface{})
	arg["skin"] = oldSkinExt
	if lVal, ok := skinLimits[id]; ok {
		arg["limit"] = lVal
	}
	if err = util.AddLog(common.SkinBusinessID, operator, uid, int64(oldSkinExt.SkinID), opAction, arg); err != nil {
		log.Error("SaveLog AddLog error(%v)", err)
	}
	return
}

// CheckTimeLimie
func (s *Service) CheckSkinTimeLimit(c context.Context, arg *menuModel.SkinExt) (isOk bool, err error) {
	var (
		skinExts []*menuModel.SkinExt
	)
	// 获取最新所有有效的配置 order by stime
	if skinExts, err = s.showDao.RawSkinExts(c); err != nil {
		return
	}
	if len(skinExts) == 0 {
		// 没有相关的配置，直接return
		isOk = true
		return
	}
	checkSkinExt := make([]*menuModel.SkinExt, 0, len(skinExts))
	// 过滤原id
	for _, v := range skinExts {
		if v.ID != arg.ID {
			checkSkinExt = append(checkSkinExt, v)
		}
	}
	// 更新最新的数据
	checkSkinExt = append(checkSkinExt, arg)
	sort.Slice(checkSkinExt, func(i, j int) bool {
		return checkSkinExt[i].Stime < checkSkinExt[j].Stime
	})
	var startTime xtime.Time
	for i, val := range checkSkinExt {
		if i != 0 {
			if val.Stime <= startTime {
				return
			}
		}
		startTime = val.Etime
	}
	isOk = true
	return
}

// MenuSkinOperate .
func (s *Service) MenuSkinOperate(c context.Context, id, uid int64, state int, operator string) (err error) {
	var (
		whState    int
		upState    int
		oldSkinExt *menuModel.SkinExt
		isOk       bool
		opAction   string
	)
	if state == -1 { //删除操作,下线状态才可以删除
		whState = 0
		upState = -1
		opAction = common.ActionDelete
	} else if state == 1 { //上线
		whState = 0
		upState = 1
		if oldSkinExt, err = s.showDao.RawSkinExt(c, id); err != nil || oldSkinExt == nil {
			log.Error("MenuSkinOperate s.showDao.RawSkinExt (%d) error(%v) or nil", id, err)
			return
		}
		// 上线前check时间
		if isOk, err = s.CheckSkinTimeLimit(c, oldSkinExt); err != nil {
			return
		}
		if !isOk {
			err = xecode.TabTimeLimit
			return
		}
		opAction = common.ActionOnline
	} else { //下线
		whState = 1
		upState = 0
		opAction = common.ActionOffline
	}
	if err = s.showDao.SkinModifyState(c, id, whState, upState); err != nil {
		log.Error("MenuSkinOperate (%d,%d) error(%v)", id, state, err)
		return
	}
	// 操作日志
	if err = s.SkinSaveLog(c, uid, int64(id), opAction, operator); err != nil {
		log.Error("MenuSkinOperate AddLog error(%v)", err)
	}
	return
}

// SkinSearch .
func (s *Service) SkinSearch(c context.Context, sid int64) (rly *menuModel.SkinReply, err error) {
	var (
		skinInfo *garb.SkinListReply
	)
	skinInfo, err = s.garbDao.SkinInfos(c, []int64{sid})
	if err != nil || skinInfo == nil {
		return
	}
	skinVal, ok := skinInfo.Skins[sid]
	if !ok {
		return
	}
	rly = &menuModel.SkinReply{ID: skinVal.ID, Name: skinVal.Name, Image: skinVal.Preview}
	return
}
