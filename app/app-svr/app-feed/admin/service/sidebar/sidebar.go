package sidebar

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf8"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/show"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	smdl "go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// Service is sidebar service
type Service struct {
	showDao *show.Dao
}

// New new a search service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		showDao: show.New(c),
	}
	return
}

// ModuleList is
func (s *Service) ModuleList(c context.Context, plat int32) ([]*smdl.SidebarModuleSimple, error) {
	ics, err := s.showDao.ModulesByPlat(c, plat)
	if err != nil {
		log.Error("s.showDao.ModulesByPlat plat(%d) err(%+v)", plat, err)
		return nil, err
	}
	return ics, nil
}

// ModuleDetail is
func (s *Service) ModuleDetail(c context.Context, id int64) (*smdl.SidebarModule, error) {
	m, err := s.showDao.Module(c, id)
	if err != nil || m == nil {
		log.Error("s.showDao.Module id(%d) err(%+v) or m=nil", id, err)
		return nil, err
	}
	return m, nil
}

// SaveLog .
func (s *Service) SaveLog(c context.Context, uid, id int64, opAction, operator string, entity interface{}) error {
	var err error
	arg := make(map[string]interface{})
	if strings.HasPrefix(opAction, "sidebar_") {
		if entity == nil {
			var side *smdl.SidebarORM
			side, err = s.showDao.ModuleItemsDetail(c, id)
			if err != nil {
				log.Error("SaveLog s.showDao.ModuleItemsDetail (%d) error(%v) or nil", id, err)
				return err
			}
			arg["sidebar_info"] = side
		} else {
			arg["sidebar_info"] = entity
		}
	} else {
		if entity == nil {
			var ic *smdl.SidebarModule
			ic, err = s.showDao.Module(c, id)
			if err != nil {
				log.Error("SaveLog s.showDao.Modules (%d) error(%v) or nil", id, err)
				return err
			}
			arg["module_info"] = ic
		} else {
			arg["module_info"] = entity
		}
	}

	if err = util.AddLog(common.LogSidebar, operator, uid, id, opAction, arg); err != nil {
		log.Error("SaveLog AddLog error(%v)", err)
		return err
	}
	return nil
}

// ModuleSave .
func (s *Service) ModuleSave(c context.Context, arg *smdl.SaveModuleParam, operator string, uid int64) error {
	moduleParam := &smdl.SidebarModule{
		ID:              arg.ID,
		MType:           arg.MType,
		Plat:            arg.Plat,
		Name:            arg.Name,
		Title:           arg.Title,
		Style:           arg.Style,
		Rank:            arg.Rank,
		ButtonName:      arg.ButtonName,
		ButtonURL:       arg.ButtonURL,
		ButtonIcon:      arg.ButtonIcon,
		ButtonStyle:     arg.ButtonStyle,
		WhiteURL:        arg.WhiteURL,
		TitleColor:      arg.TitleColor,
		Subtitle:        arg.Subtitle,
		SubtitleColor:   arg.SubtitleColor,
		SubtitleURL:     arg.SubtitleURL,
		Background:      arg.Background,
		BackgroundColor: arg.BackgroundColor,
		State:           smdl.SModuleNormal,
		AuditShow:       arg.AuditShow,
		IsMng:           arg.IsMng,
		OpStyleType:     arg.OpStyleType,
		OpLoadCondition: arg.OpLoadCondition,
		AreaPolicy:      arg.AreaPolicy,
		ShowPurposed:    arg.ShowPurposed,
	}
	id, err := s.showDao.ModuleSave(c, moduleParam)
	if err != nil {
		log.Error("ModuleSave s.showDao.ModuleSave err(%+v)", err)
		return err
	}
	// 操作日志
	opAction := common.ActionAdd
	if arg.ID > 0 {
		opAction = common.ActionUpdate
	}
	if err = s.SaveLog(c, uid, id, opAction, operator, moduleParam); err != nil {
		log.Error("ModuleSave s.SaveLog error(%v)", err)
		return err
	}
	return nil
}

// ModuleOpt is
func (s *Service) ModuleOpt(c context.Context, id, uid int64, state int, operator string) error {
	rows, err := s.showDao.UpdateModuleState(c, id, state)
	if err != nil {
		log.Error("ModuleOpt s.showDao.UpdateModuleState id:%d state:%d error(%v)", id, state, err)
		return err
	}
	if rows == 0 {
		return ecode.NotModified
	}
	// 操作日志
	if err = s.SaveLog(c, uid, id, common.ActionDelete, operator, nil); err != nil {
		log.Error("ModuleOpt s.SaveLog error(%v)", err)
		return err
	}
	return nil
}

//const (
//	No_Offline_UnixTime = "0000-00-00 00:00:00"
//)

// ModuleItemList is
func (s *Service) ModuleItemList(c context.Context, req *smdl.ModuleItemListReq) (result []*smdl.SidebarEntity, err error) {
	if req.Module == 0 {
		return nil, ecode.Error(-400, "缺少必要参数信息")
	}

	result = make([]*smdl.SidebarEntity, 0)

	var sidebars []*smdl.SidebarORM
	if sidebars, err = s.showDao.ModuleItemsByPlat(c, req); err != nil {
		return nil, err
	}
	sids := make([]int64, 0)
	for _, s := range sidebars {
		sids = append(sids, s.ID)
	}

	// 获取所有sidebar的版本限制
	var limits map[int64][]*smdl.SidebarLimit
	if limits, err = s.showDao.SideBarLimits(c, sids); err != nil {
		return nil, err
	}

	for _, s := range sidebars {
		if s.OfflineTime <= 0 {
			s.OfflineTime = 0
		}

		item := &smdl.SidebarEntity{
			SidebarORM: *s,
		}

		if limitArr, ok := limits[s.ID]; ok {
			var isBuildMatch bool
			if req.Build > 0 {
				for _, limit := range limitArr {
					switch limit.Conditions {
					case "gt":
						{
							isBuildMatch = req.Build > limit.Build
						}
					case "lt":
						{
							isBuildMatch = req.Build < limit.Build
						}
					case "eq":
						{
							isBuildMatch = req.Build == limit.Build
						}
					case "ne":
						{
							isBuildMatch = req.Build != limit.Build
						}
					}
					if !isBuildMatch {
						break
					}
				}
				if !isBuildMatch {
					continue
				}
			}
			item.Limit = limitArr
		}
		result = append(result, item)
	}

	return result, err
}

// ModuleItemDetail is
func (s *Service) ModuleItemDetail(c context.Context, sidebarID int64) (result *smdl.SidebarEntity, err error) {
	if sidebarID == 0 {
		return nil, ecode.Error(-400, "缺少必要参数信息")
	}
	var raw *smdl.SidebarORM

	if raw, err = s.showDao.ModuleItemsDetail(c, sidebarID); err != nil {
		return nil, err
	}
	if raw.OfflineTime <= 0 {
		raw.OfflineTime = 0
	}
	result = &smdl.SidebarEntity{
		SidebarORM: *raw,
	}
	sids := []int64{sidebarID}

	// 获取所有sidebar的版本限制
	var limits map[int64][]*smdl.SidebarLimit
	if limits, err = s.showDao.SideBarLimits(c, sids); err != nil {
		return nil, err
	}
	if limitArr, ok := limits[sidebarID]; ok {
		result.Limit = limitArr
	}
	return result, err
}

// 模块item状态变更
func (s *Service) ModuleItemOpt(c context.Context, sidebarID int64, state int32, username string, uid int64) (err error) {
	var sidebar *smdl.SidebarORM
	if sidebar, err = s.showDao.ModuleItemsDetail(c, sidebarID); err != nil {
		return err
	}
	if sidebar.State == smdl.SidebarORM_State_Normal && state != smdl.SidebarORM_State_Banned {
		return ecode.Error(-400, "状态切换参数错误")
	}
	if sidebar.State == smdl.SidebarORM_State_Banned && (state != smdl.SidebarORM_State_Normal && state != smdl.SidebarORM_State_Deleted) {
		return ecode.Error(-400, "状态切换参数错误")
	}

	if err = s.showDao.UpdateModuleItemState(c, sidebarID, state); err != nil {
		return err
	}

	action := ""
	switch state {
	case smdl.SidebarORM_State_Normal:
		{
			action = common.ActionOnline
		}
	case smdl.SidebarORM_State_Banned:
		{
			action = common.ActionOffline
		}
	case smdl.SidebarORM_State_Deleted:
		{
			action = common.ActionDelete
		}
	}

	// 操作日志
	if err = s.SaveLog(c, uid, sidebarID, "sidebar_"+action, username, sidebar); err != nil {
		log.Error("ModuleOpt s.SaveLog error(%v)", err)
		return err
	}
	return err
}

const (
	NameMaxLen             = 10
	ParamMaxLen            = 128
	TabIDMaxLen            = 30
	OpTitleMaxLen          = 10
	OpTitleWideMaxLen      = 20
	OpSubTitleMaxLen       = 20
	OpLinkTextLinkMaxLen   = 6
	OpLinkTextButtonMaxLen = 4

	// 按钮样式
	OpLinkTypeBUTTON = 0
	// 带文案跳链
	OpLinkTypeLinkWithText = 1
	// 底tab大加号
	OpLinkTypeDialogOpener = 3
	// 底tabNA页
	OpLinkTypeNAPageID = 4

	// 首页Tab-图标
	ModuleStyleHomeIcon = 0
	// 我的页-图标
	// 运营位
	ModuleStyleOP       = 3
	ModuleStyleCommonOp = 7

	// 运营位-默认
	ModuleOpStyleTypeCommon = 0
	// 运营位-投稿引导强化卡
	ModuleOpStyleTypeUpCard = 1

	OfflineTimeDefault = -62135596800
)

// nolint:gocognit
func (s *Service) ModuleItemSave(c context.Context, req *smdl.SidebarEntity, username string, uid int64) (err error) {
	sidebar := req.SidebarORM
	if sidebar.TusValue != "" {
		if _, err := strconv.ParseInt(sidebar.TusValue, 10, 64); err != nil {
			return ecode.Error(-400, "人群包id输入错误")
		}
	}
	if sidebar.Module == 0 {
		return ecode.Error(-400, "请选择模块")
	}
	if sidebar.Name == "" {
		return ecode.Error(-400, "请填写名称")
	}
	if utf8.RuneCountInString(sidebar.Name) > NameMaxLen {
		return ecode.Error(-400, "名称不超过10个字符")
	}

	if sidebar.OpLinkType != OpLinkTypeDialogOpener && sidebar.Param == "" {
		return ecode.Error(-400, "请填写参数")
	}

	if utf8.RuneCountInString(sidebar.Param) > ParamMaxLen {
		return ecode.Error(-400, "参数不超过128个字符")
	}

	if utf8.RuneCountInString(sidebar.TabID) > TabIDMaxLen {
		return ecode.Error(-400, "上报参数不超过30个字符")
	}
	if req.LimitStr == "" {
		return ecode.Error(-400, "请填写版本限制")
	}

	if sidebar.OnlineTime == 0 {
		return ecode.Error(-400, "请选择上线时间")
	}

	if sidebar.OfflineTime > 0 && sidebar.OfflineTime <= sidebar.OnlineTime {
		return ecode.Error(-400, "上线时间必须早于下线时间")
	}
	if sidebar.OfflineTime <= 0 {
		sidebar.OfflineTime = OfflineTimeDefault
	}

	// icon类型，使用NA页ID,oplinktype =4 ,并且是老的首页顶底tab
	if req.ModuleStyle == ModuleStyleHomeIcon && req.ModuleOpStyleType == OpLinkTypeNAPageID {
		if !isNum(sidebar.Param) {
			return ecode.Error(-400, "NA页面id必须为数字")
		}
	}
	if req.ModuleStyle == ModuleStyleOP || req.ModuleStyle == ModuleStyleCommonOp {
		if sidebar.OpTitle == "" {
			return ecode.Error(-400, "请填写主标题")
		}
	} else {
		if sidebar.Module != 8 && sidebar.OpLinkType != OpLinkTypeDialogOpener && sidebar.Logo == "" {
			return ecode.Error(-400, "请上传图标")
		}
	}
	// 投稿引导强化卡
	if req.ModuleStyle == ModuleStyleOP && req.ModuleOpStyleType == ModuleOpStyleTypeUpCard {
		if utf8.RuneCountInString(sidebar.OpTitle) > OpTitleMaxLen {
			return ecode.Error(-400, "主标题不超过10个字符")
		}
		if sidebar.OpSubTitle == "" {
			return ecode.Error(-400, "请填写副标题")
		}
		if utf8.RuneCountInString(sidebar.OpSubTitle) > OpSubTitleMaxLen {
			return ecode.Error(-400, "副标题不超过20个字符")
		}
		if sidebar.OpLinkType != 0 {
			return ecode.Error(-400, "请确认运营样式参数是否正确")
		}
		if sidebar.OpLinkText == "" {
			return ecode.Error(-400, "请填写按钮文案")
		}
		if utf8.RuneCountInString(sidebar.OpLinkText) > OpLinkTextButtonMaxLen {
			return ecode.Error(-400, "按钮文案不超过4个字符")
		}
	}

	// 通用运营位
	if req.ModuleStyle == ModuleStyleOP && req.ModuleOpStyleType == ModuleOpStyleTypeCommon {
		if utf8.RuneCountInString(sidebar.OpTitle) > OpTitleWideMaxLen {
			return ecode.Error(-400, "主标题不超过20个字符")
		}
		if sidebar.OpLinkType < 0 || (sidebar.OpLinkType > 3 && sidebar.OpLinkType != 5) {
			return ecode.Error(-400, "请确认运营样式参数是否正确")
		}
		if sidebar.OpTitleIcon == "" {
			return ecode.Error(-400, "请补充文案图标")
		}
		if sidebar.OpLinkType == OpLinkTypeBUTTON {
			if sidebar.OpLinkText == "" {
				return ecode.Error(-400, "请填写按钮文案")
			}
			if utf8.RuneCountInString(sidebar.OpLinkText) > OpLinkTextButtonMaxLen {
				return ecode.Error(-400, "按钮文案不超过4个字符")
			}
			if sidebar.OpLinkIcon == "" {
				return ecode.Error(-400, "请补充按钮图标")
			}
		}
		if sidebar.OpLinkType == OpLinkTypeLinkWithText {
			if sidebar.OpLinkText == "" {
				return ecode.Error(-400, "请填写跳链文案")
			}
			if utf8.RuneCountInString(sidebar.OpLinkText) > OpLinkTextLinkMaxLen {
				return ecode.Error(-400, "跳链文案不超过6个字符")
			}
		}
	}

	sidebarLimits := make([]*smdl.SidebarLimitORM, 0)

	limits := strings.Split(req.LimitStr, "|")

	const limitPairLen = 2

	for _, limit := range limits {
		limit = strings.Replace(limit, " ", "", -1)
		temp := strings.Split(limit, ",")
		var build int64
		if len(temp) != limitPairLen {
			return ecode.Error(-400, "limit配置错误")
		}
		if temp[0] != "gt" && temp[0] != "lt" && temp[0] != "eq" && temp[0] != "ne" {
			return ecode.Error(-400, "limit配置错误")
		}
		if build, err = strconv.ParseInt(temp[1], 10, 64); err != nil {
			return ecode.Error(-400, "limit配置错误")
		}

		item := &smdl.SidebarLimitORM{
			Build:      build,
			Conditions: temp[0],
		}
		sidebarLimits = append(sidebarLimits, item)
	}

	tx := s.showDao.DB.Begin()

	defer (func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	})()

	if sidebar.ID > 0 {
		var oldSidebar *smdl.SidebarORM
		if oldSidebar, err = s.showDao.ModuleItemsDetail(c, sidebar.ID); err != nil {
			return err
		}
		sidebar.State = oldSidebar.State
	}

	// 更新或保存sidebar
	if err = s.showDao.SaveModuleItem(c, &sidebar); err != nil {
		log.Error("保存sidebar失败， info: (%+v), err: (%s)", sidebar, err.Error())
		return
	}

	// 删除build
	if err = s.showDao.DeleteSidebarLimit(c, sidebar.ID); err != nil {
		log.Error("删除limit失败， sid: (%d), err: (%s)", sidebar.ID, err.Error())
		return
	}
	// 新建limit
	for i, limit := range sidebarLimits {
		limit.SID = sidebar.ID
		if err = s.showDao.SaveSidebarLimit(c, sidebarLimits[i]); err != nil {
			log.Error("新建limit失败， sid: (%d), err: (%s)", sidebar.ID, err.Error())
			return
		}
	}

	// 操作日志
	opAction := common.ActionAdd
	if req.ID > 0 {
		opAction = common.ActionUpdate
	}

	if err = s.SaveLog(c, uid, sidebar.ID, "sidebar_"+opAction, username, sidebar); err != nil {
		log.Error("ModuleSave s.SaveLog error(%v)", err)
		return err
	}

	return err
}

func isNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
