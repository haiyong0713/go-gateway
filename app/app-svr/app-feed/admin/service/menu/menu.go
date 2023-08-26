package menu

import (
	"context"
	"encoding/json"
	"fmt"
	locdao "go-gateway/app/app-svr/app-feed/admin/dao/location"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/garb"
	"go-gateway/app/app-svr/app-feed/admin/dao/show"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/menu"
	"go-gateway/app/app-svr/app-feed/admin/util"
	xecode "go-gateway/app/app-svr/app-feed/ecode"

	"go-common/library/sync/errgroup.v2"
)

// Service is menu service
type Service struct {
	showDao *show.Dao
	garbDao *garb.Dao
	locDao  *locdao.Dao
}

// New new a search service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		showDao: show.New(c),
		garbDao: garb.New(c),
		locDao:  locdao.New(c),
	}
	return
}

// MenuSearch .
func (s *Service) MenuSearch(c context.Context, id int64, opType int) (rly *menu.SideReply, err error) {
	var (
		sideBars map[int64]*menu.Sidebar
		appMenus map[int64]*menu.AppMenus
	)
	if opType == menu.SideType {
		if sideBars, err = s.showDao.SideBars(c, []int64{id}); err != nil {
			return
		}
		if sv, ok := sideBars[id]; ok {
			rly = &menu.SideReply{ID: sv.ID, Name: sv.Name, Plat: sv.Plat}
		}
	} else if opType == menu.MenuType {
		if appMenus, err = s.showDao.AppMenus(c, []int64{id}); err != nil {
			return
		}
		if av, ok := appMenus[id]; ok {
			rly = &menu.SideReply{ID: av.ID, Name: av.Name}
		}
	}
	return
}

// MenuTabSave .
//
//nolint:gocognit
func (s *Service) MenuTabSave(c context.Context, arg *menu.TabSaveParam, operator string, uid int64) (rly *menu.TabSaveReply, err error) {
	var (
		tabRly    *menu.SideReply
		opAction  string
		oldTabExt *menu.TabExt
		tabExt    = &menu.TabExt{ID: arg.ID, Attribute: arg.Attribute, Type: arg.Type, TabID: arg.TabID, Inactive: arg.Inactive, InactiveIcon: arg.InactiveIcon,
			InactiveType: arg.InactiveType, Active: arg.Active, ActiveIcon: arg.ActiveIcon, ActiveType: arg.ActiveType, Stime: arg.Stime, Etime: arg.Etime,
			FontColor: arg.FontColor, BarColor: arg.BarColor, Operator: operator, TabTopColor: arg.TabTopColor, TabMiddleColor: arg.TabMiddleColor,
			TabBottomColor: arg.TabBottomColor, BgImage1: arg.BgImage1, BgImage2: arg.BgImage2}
		openImage          = tabExt.AttrVal(menu.AttrBitImage) == menu.AttrYes
		openColor          = tabExt.AttrVal(menu.AttrBitColor) == menu.AttrYes
		openBgImage        = tabExt.AttrVal(menu.AttrBitBgImage) == menu.AttrYes
		openFollowBusiness = tabExt.AttrVal(menu.AttrBitFollowBusiness) == menu.AttrYes
	)

	if !openImage && !openColor && !openBgImage && !openFollowBusiness {
		err = ecode.RequestErr
		return
	}
	// check image params
	if openImage {
		if arg.InactiveIcon == "" || arg.ActiveIcon == "" {
			err = ecode.RequestErr
			return
		}
	}
	// check color params
	if openColor {
		if arg.TabTopColor == "" || arg.TabMiddleColor == "" || arg.TabBottomColor == "" || arg.FontColor == "" {
			err = ecode.RequestErr
			return
		}
	}
	// check background image params
	if openBgImage {
		if arg.BgImage1 == "" || arg.BgImage2 == "" {
			err = ecode.RequestErr
			return
		}
	}
	// check tab allowed
	if openFollowBusiness {
		if conf.Conf.AllowedTabs.Tabs != "" {
			tabs := strings.Split(conf.Conf.AllowedTabs.Tabs, ",")
			allowed := false
			for _, tab := range tabs {
				tabId, err := strconv.ParseInt(tab, 10, 64)
				if err != nil {
					log.Error("tab id 配置有误,id(%s)", tab)
					return nil, ecode.RequestErr
				}
				if tabId == arg.TabID {
					allowed = true
					break
				}
			}
			if !allowed {
				return nil, ecode.Error(-400, "tab id 被限制")
			}
		}
	}
	// [stime,etime)
	if arg.Stime >= arg.Etime {
		err = ecode.RequestErr
		return
	}
	buildLimit := make([]*menu.BuildLimit, 0)
	if err = json.Unmarshal([]byte(arg.Limit), &buildLimit); err != nil {
		log.Error("MenuTabSave json.Unmarshal %s error(%v)", arg.Limit, err)
		return
	}
	if len(buildLimit) == 0 {
		err = ecode.RequestErr
		return
	}
	lastBuildLimit := make(map[int8]*menu.BuildLimit)
	for _, v := range buildLimit {
		if err = v.ValidateParam(); err != nil {
			return
		}
		lastBuildLimit[v.Plat] = v
	}
	// check tabid
	if tabRly, err = s.MenuSearch(c, arg.TabID, arg.Type); err != nil || tabRly == nil || tabRly.ID == 0 {
		log.Error("MenuTabSave s.MenuSearch(%d,%d) error(%v)", arg.TabID, arg.Type, err)
		err = ecode.RequestErr
		return
	}
	if arg.ID > 0 {
		if oldTabExt, err = s.showDao.RawTabExt(c, arg.ID); err != nil || oldTabExt == nil {
			log.Error("s.showDao.RawTabExt (%d) error(%v) or nil", arg.ID, err)
			return
		}
		if oldTabExt.State == menu.TabOnline {
			// check stime 和 etime
			var cOK bool
			if cOK, err = s.CheckTimeLimit(c, tabExt); err != nil {
				err = ecode.RequestErr
				return
			}
			if !cOK {
				err = xecode.TabTimeLimit
				return
			}
		}
		opAction = common.ActionUpdate
		tabExt.Ver = oldTabExt.Ver
		oldMd5 := oldTabExt.ExtMD5()
		newMd5 := tabExt.ExtMD5()
		if oldMd5 != newMd5 {
			tabExt.Ver = fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
		}
		//check ver 是否改变
	} else {
		opAction = common.ActionAdd
		//版本号
		tabExt.Ver = fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	}
	rly = &menu.TabSaveReply{}
	//nolint:staticcheck
	rly.ID, err = s.showDao.TabExtSave(c, tabExt, lastBuildLimit)
	// 操作日志
	if err = s.SaveLog(c, uid, int64(rly.ID), opAction, operator); err != nil {
		log.Error("MenuTabSave AddLog error(%v)", err)
		return
	}
	return
}

// SaveLog .
func (s *Service) SaveLog(c context.Context, uid, id int64, opAction, operator string) (err error) {
	var (
		oldTabExt *menu.TabExt
		tabLimits map[int64][]*menu.TabLimit
	)
	if oldTabExt, err = s.showDao.RawTabExt(c, id); err != nil || oldTabExt == nil {
		log.Error("SaveLog s.showDao.RawTabExt (%d) error(%v) or nil", id, err)
		return
	}
	if tabLimits, err = s.showDao.TabLimits(c, []int64{id}, 0); err != nil {
		log.Error("SaveLog s.showDao.TabLimits (%d) error(%v)", id, err)
		return
	}
	arg := make(map[string]interface{})
	arg["tab"] = oldTabExt
	if lVal, ok := tabLimits[id]; ok {
		arg["limit"] = lVal
	}
	if err = util.AddLog(common.TabBusinessID, operator, uid, int64(oldTabExt.TabID), opAction, arg); err != nil {
		log.Error("SaveLog AddLog error(%v)", err)
	}
	return
}

// CheckTimeLimie
func (s *Service) CheckTimeLimit(c context.Context, arg *menu.TabExt) (isOk bool, err error) {
	var (
		tabExts []*menu.TabExt
	)
	// 获取最新tab_id和type下所有有效的配置 order by stime
	if tabExts, err = s.showDao.RawTabExts(c, arg.TabID, arg.Type); err != nil {
		return
	}
	if len(tabExts) == 0 {
		// 没有相关的配置，直接return
		isOk = true
		return
	}
	checkTabExt := make([]*menu.TabExt, 0, len(tabExts))
	// 过滤原id
	for _, v := range tabExts {
		if v.ID != arg.ID {
			checkTabExt = append(checkTabExt, v)
		}
	}
	// 更新最新的数据
	checkTabExt = append(checkTabExt, arg)
	sort.Slice(checkTabExt, func(i, j int) bool {
		return checkTabExt[i].Stime < checkTabExt[j].Stime
	})
	var startTime xtime.Time
	for i, val := range checkTabExt {
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

// MenuTabOperate .
func (s *Service) MenuTabOperate(c context.Context, id, uid int64, state int, operator string) (err error) {
	var (
		whState   int
		upState   int
		oldTabExt *menu.TabExt
		isOk      bool
		opAction  string
	)
	if state == -1 { //删除操作,下线状态才可以删除
		whState = 0
		upState = -1
		opAction = common.ActionDelete
	} else if state == 1 { //上线
		whState = 0
		upState = 1
		if oldTabExt, err = s.showDao.RawTabExt(c, id); err != nil || oldTabExt == nil {
			log.Error("MenuTabOperate s.showDao.RawTabExt (%d) error(%v) or nil", id, err)
			return
		}
		// 上线前check时间
		if isOk, err = s.CheckTimeLimit(c, oldTabExt); err != nil {
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
	if err = s.showDao.ModifyState(c, id, whState, upState); err != nil {
		log.Error("MenuTabOperate (%d,%d) error(%v)", id, state, err)
		return
	}
	// 操作日志
	if err = s.SaveLog(c, uid, int64(id), opAction, operator); err != nil {
		log.Error("MenuTabOperate AddLog error(%v)", err)
	}
	return
}

// MenuTabList .
func (s *Service) MenuTabList(c context.Context, arg *menu.ListParam) (pager *menu.ListReply, err error) {
	var (
		rely *menu.SearchReply
	)
	if rely, err = s.showDao.TabExts(c, arg.TabID, arg.Pn, arg.Ps); err != nil || rely == nil {
		return
	}
	pager = &menu.ListReply{Page: &menu.Page{Num: arg.Pn, Size: arg.Ps, Total: rely.Total}}
	var (
		tIDs      []int64
		menusIDs  []int64
		sideIDs   []int64
		tabLimits map[int64][]*menu.TabLimit
		sideBars  map[int64]*menu.Sidebar
		appMenus  map[int64]*menu.AppMenus
	)
	for _, v := range rely.List {
		tIDs = append(tIDs, v.ID)
		if v.Type == menu.SideType {
			sideIDs = append(sideIDs, v.TabID)
		} else if v.Type == menu.MenuType {
			menusIDs = append(menusIDs, v.TabID)
		}
	}
	eg := errgroup.WithContext(c)
	if len(tIDs) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			tabLimits, e = s.showDao.TabLimits(ctx, tIDs, 0)
			return
		})
	}
	if len(sideIDs) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			sideBars, e = s.showDao.SideBars(ctx, sideIDs)
			return
		})
	}
	if len(menusIDs) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			appMenus, e = s.showDao.AppMenus(ctx, menusIDs)
			return
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	//拼接信息
	for _, v := range rely.List {
		tem := &menu.TabList{TabExt: v}
		if v.Type == menu.SideType {
			if sv, ok := sideBars[v.TabID]; ok {
				tem.MenuName = sv.Name
				tem.Plat = sv.Plat
			}
		} else if v.Type == menu.MenuType {
			if lv, k := appMenus[v.TabID]; k {
				tem.MenuName = lv.Name
			}
		}
		if lv, okk := tabLimits[v.ID]; okk {
			tem.Limit = lv
		}
		pager.List = append(pager.List, tem)
	}
	return
}
