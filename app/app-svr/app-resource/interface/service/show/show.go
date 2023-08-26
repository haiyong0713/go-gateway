package show

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	farm "go-farm"
	"go-gateway/app/app-svr/app-card/interface/model/i18n"

	"go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/exp/ab"
	"go-common/library/log"
	"go-common/library/net/metadata"
	xtime "go-common/library/time"

	"go-common/library/log/infoc.v2"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/abtest"
	"go-gateway/app/app-svr/app-resource/interface/model/location"
	"go-gateway/app/app-svr/app-resource/interface/model/show"
	"go-gateway/app/app-svr/app-resource/interface/model/tab"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	locmdl "git.bilibili.co/bapis/bapis-go/community/service/location"
	garb "git.bilibili.co/bapis/bapis-go/garb/service"
	bGroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
	resourceApi "git.bilibili.co/bapis/bapis-go/resource/service"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

const (
	_initTabKey = "tab_%d_%s"

	// 	_initVersion         = "showtab_version"
	_defaultLanguageHans = "hans"
	_defaultLanguageHant = "hant"
	_initTtExtKey        = "tt_ext_%d_%d"
	_dialogStyle         = 4
	_topMoreStyle        = 5
	_tabStyle            = 0
	_recommendStyle      = 6
	_publishBubble       = 8
)

var (
	_showAbtest = map[string]string{
		"bilibili://pegasus/hottopic": "home_tabbar_server_1",
	}
	_deafaultTab = map[string]*show.Tab{

		"bilibili://pegasus/promo": {
			DefaultSelected: 1,
		},
	}
	_showWhiteList = map[string]struct{}{
		"bilibili://live/home":     {},
		"bilibili://pegasus/promo": {},
		"bilibili://pgc/home":      {},
		// ========/=======
		"bilibili://main/home/":      {},
		"bilibili://following/home/": {},
		"bilibili://user_center/":    {},
		// ========/=======
		"bilibili://main/home":            {},
		"bilibili://following/home":       {},
		"bilibili://user_center":          {},
		"bilibili://link/im_home":         {},
		"bilibili://user_center/download": {},
		"bilibili://user_center/mine":     {},
	}
	_moduleMap = map[int64]struct{}{

		_top:    {},
		_tab:    {},
		_bottom: {},
	}
	_topLeftNewUserFlag = ab.Int("app_change_mode", "首页左上角新用户实验", -1)
)

// Tabs show tabs
//
//nolint:gocognit
func (s *Service) Tabs(c context.Context, plat int8, build, teenagersMode, lessonsMode int, buvid, mobiApp, platform, language, channel string, mid int64, slocale, clocale string) (res *show.Show, config *show.Config, a *abtest.List, err error) {
	//nolint:staticcheck
	if key := fmt.Sprintf(_initTabKey, plat, language); len(s.tabCache[fmt.Sprintf(key)]) == 0 || language == "" {
		if model.IsOverseas(plat) {
			var key = fmt.Sprintf(_initTabKey, plat, _defaultLanguageHant)
			//nolint:staticcheck
			if len(s.tabCache[fmt.Sprintf(key)]) > 0 {
				language = _defaultLanguageHant
			} else {
				language = _defaultLanguageHans
			}
		} else {
			language = _defaultLanguageHans
		}
	}
	var (
		key           = fmt.Sprintf(_initTabKey, plat, language)
		tmpTabs       = []*show.Tab{}
		baseTabs      = []*show.Tab{}
		tabIDs        = []int64{}
		tabExtIDs     = []*resourceApi.Tab{}
		noLoginAvatar string
		// pids          []string
		// auths  map[int64]*locmdl.Auth
		ipMeta *location.Info
		// showByDef     = map[string]bool{}
		showAuthDef      = map[int64]*locmdl.ZoneLimitAuth{}
		policiesAuth     map[int64]*locmdl.ZoneLimitAuth
		deniedByWhiteURL map[int64]struct{}
	)
	if tabs, ok := s.tabCache[key]; ok {
		for _, v := range tabs {
			// 重新赋值逻辑不可删除，否则会影响extend正确性
			var _v = *v
			vp := &_v
			buildForbid := s.buildForbid(vp.ID, build)
			if buildForbid {
				continue
			}
			if !s.c.ShowHotAll {
				if ab, ok := s.abtestCache[vp.Group]; ok {
					if _, ok := s.showTabMids[mid]; !ab.AbTestIn(buvid) && !ok {
						continue
					}
					a = &abtest.List{}
					a.ListChange(ab)
				}
			}
			if _, ok := _showWhiteList[vp.URI]; (teenagersMode != 0 || lessonsMode != 0) && !ok {
				continue
			}
			// tab 固定资源位的资源配置
			if vp.Module == _tab {
				tabExtIDs = append(tabExtIDs, &resourceApi.Tab{TabId: vp.ID, TType: tab.SideType})
			}
			// if vp.Area != "" {
			// 	pids = append(pids, vp.Area)
			// }
			if vp.AreaPolicy > 0 {
				showAuthDef[vp.AreaPolicy] = &locmdl.ZoneLimitAuth{}
				switch vp.ShowPurposed {
				case 0:
					showAuthDef[vp.AreaPolicy].Play = locmdl.Status_Allow
				case 1:
					showAuthDef[vp.AreaPolicy].Play = locmdl.Status_Forbidden
				}
			}
			tabIDs = append(tabIDs, vp.ID)
			tmpTabs = append(tmpTabs, vp)
		}
	}
	// 获取运营tab的地区ID
	if !s.auditTab(mobiApp, build, plat) && teenagersMode == 0 {
		for _, m := range s.menuCache {
			if _, ok := m.Versions[model.PlatAPPBuleChange(plat)]; ok {
				// if m.Area != "" {
				// 	pids = append(pids, m.Area)
				// }
				// 获取所有有效运营tabID
				if m.TabID > 0 {
					tabExtIDs = append(tabExtIDs, &resourceApi.Tab{TabId: m.TabID, TType: tab.MenuType})
				}
				if m.AreaPolicy > 0 {
					showAuthDef[m.AreaPolicy] = &locmdl.ZoneLimitAuth{}
					switch m.ShowPurposed {
					case 0:
						showAuthDef[m.AreaPolicy].Play = locmdl.Status_Allow
					case 1:
						showAuthDef[m.AreaPolicy].Play = locmdl.Status_Forbidden
					}
				}
			}
		}
	}
	eg := errgroup.WithContext(c)
	rdMap := make(map[int64]*show.Red)
	hiddenMap := make(map[int64]bool, len(tabIDs))
	tabExtMap := make(map[string]*resourceApi.TabExt, len(tabExtIDs))
	if len(tabExtIDs) > 0 {
		// 获取tab下特定颜色图片等配置
		eg.Go(func(ctx context.Context) error {
			tExtRly, err := s.rdao.GetTabExt(ctx, int64(plat), int64(build), buvid, tabExtIDs)
			if err != nil {
				log.Error("s.rdao.GetTabEx error(%+v)", err)
				return nil
			}
			for _, kv := range tExtRly {
				if kv == nil {
					continue
				}
				tabExtMap[fmt.Sprintf(_initTtExtKey, kv.TabId, kv.TType)] = kv
			}
			return nil
		})
	}
	if len(s.redDot) > 0 {
		// redDot
		var mutex sync.Mutex
		for _, v := range s.redDot[plat] {
			tmpID := v.ID
			tmpURL := v.URL
			buildForbid := s.buildForbid(tmpID, build)
			if buildForbid {
				continue
			}
			eg.Go(func(ctx context.Context) error {
				red, err := s.redDao.RedDot(ctx, mid, tmpURL, platform, model.GotoGame)
				if err != nil {
					log.Error("s.accDao.RedDot error(%+v)", err)
					return nil
				}
				if red != nil {
					mutex.Lock()
					rdMap[tmpID] = red
					mutex.Unlock()
				}
				return nil
			})
		}
	}
	if len(tabIDs) > 0 && model.IsAndroid(plat) { // 安卓市场才屏蔽
		eg.Go(func(ctx context.Context) error {
			reply, err := s.rdao.EntrancesIsHidden(ctx, tabIDs, build, plat, channel)
			if err != nil {
				log.Error("s.rdao.EntrancesIsHidden err(%+v)", err)
				return nil
			}
			if reply != nil {
				hiddenMap = reply.Infos
			}
			return nil
		})
	}
	// if len(pids) > 0 {
	// 	eg.Go(func(ctx context.Context) (err error) {
	// 		req := &locmdl.AuthPIDsReq{
	// 			Pids:              strings.Join(pids, ","),
	// 			IpAddr:            metadata.String(ctx, metadata.RemoteIP),
	// 			DefaultAuthStatus: map[int64]*locmdl.AuthPIDsReqDefaultStatus{},
	// 			InvertedMode:      true,
	// 		}
	// 		for area, toShow := range showByDef {
	// 			areaV, _ := strconv.ParseInt(area, 10, 64)
	// 			if areaV > 0 {
	// 				req.DefaultAuthStatus[areaV] = &locmdl.AuthPIDsReqDefaultStatus{
	// 					Play: locmdl.Status_Forbidden,
	// 				}
	// 				if toShow {
	// 					req.DefaultAuthStatus[areaV].Play = locmdl.Status_Allow
	// 				}
	// 			}
	// 		}
	// 		if auths, err = s.loc.RawAuthPIDs(ctx, req); err != nil {
	// 			log.Error("%+v", err)
	// 			err = nil
	// 			return
	// 		}
	// 		return nil
	// 	})
	// }
	if len(showAuthDef) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			req := &locmdl.ZoneLimitPoliciesReq{
				UserIp:       metadata.String(ctx, metadata.RemoteIP),
				DefaultAuths: showAuthDef,
			}
			var reply *locmdl.ZoneLimitPoliciesReply
			if reply, err = s.loc.ZoneLimitPolicies(ctx, req); err != nil {
				log.Error("Failed to get ZoneLimitPolicies: %+v", err)
				err = nil
				return
			}
			policiesAuth = reply.Auths
			return nil
		})
	}
	eg.Go(func(ctx context.Context) (err error) {
		ipMeta, err = s.loc.Info(ctx, metadata.String(ctx, metadata.RemoteIP))
		if err != nil {
			log.Error("Failed to request location ip info: %+v", err)
			return nil
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		deniedByWhiteURL = s.tabDeniedByWhiteURL(ctx, mid, tmpTabs)
		return nil
	})
	//nolint:errcheck
	eg.Wait()
	for _, v := range tmpTabs {
		// 判断是否在入口屏蔽配置里
		if isHidden, ok := hiddenMap[v.ID]; ok && isHidden {
			continue
		}
		if v.AreaPolicy > 0 {
			auth, ok := policiesAuth[v.AreaPolicy]
			if ok {
				if auth.Play == locmdl.Status_Forbidden {
					continue
				}
			}
		}
		// 不在白名单里所以不展示
		if _, denied := deniedByWhiteURL[v.ID]; denied {
			continue
		}
		// areaInt, _ := strconv.ParseInt(v.Area, 10, 64)
		// if areaInt > 0 {
		// 	if auth, ok := auths[areaInt]; ok && auth.Play == int64(locmdl.Status_Forbidden) {
		// 		continue
		// 	}
		// }
		if eVal, ok := tabExtMap[fmt.Sprintf(_initTtExtKey, v.ID, tab.SideType)]; ok && eVal != nil {
			v.Extension = &show.Extension{}
			v.Extension.BuildExt(eVal)
		}
		baseTabs = append(baseTabs, v)
	}
	if !s.auditTab(mobiApp, build, plat) && teenagersMode == 0 {
		if menus := s.menus(plat, build, lessonsMode, policiesAuth); len(menus) > 0 {
			// tab 运营资源位的资源配置
			for _, mVal := range menus {
				if mID, mErr := strconv.ParseInt(mVal.TabID, 10, 64); mErr == nil && mID > 0 {
					if eVal, ok := tabExtMap[fmt.Sprintf(_initTtExtKey, mID, tab.MenuType)]; ok && eVal != nil {
						mVal.Extension = &show.Extension{}
						mVal.Extension.BuildExt(eVal)
					}
				}
			}
			baseTabs = append(baseTabs, menus...)
		}
	}
	res = &show.Show{}
	for _, v := range baseTabs {
		t := &show.Tab{}
		*t = *v
		if r, ok := rdMap[t.ID]; ok && t.Red != "" && r != nil {
			if r.RedDot {
				t.RedDot = &show.RedDot{Type: r.Type, Number: r.Number}
			}
			if r.Icon != "" {
				t.AnimateIcon = &show.AnimateIcon{Icon: r.Icon, Json: t.Animate}
			}
		}
		switch v.ModuleStr {
		case model.ModuleTop:
			t.Pos = len(res.Top) + 1
			res.Top = append(res.Top, t)
		case model.ModuleTab:
			t.Pos = len(res.Tab) + 1
			res.Tab = append(res.Tab, t)
		case model.ModuleBottom:
			t.Pos = len(res.Bottom) + 1
			res.Bottom = append(res.Bottom, t)
		}
	}
	adjustInternational(c, res, ipMeta, slocale, clocale)
	if s.c.Custom.NoLoginAvatarAll {
		noLoginAvatar = "all"
	} else {
		//nolint:staticcheck,gomnd
		if buvidInt := crc32.ChecksumIEEE([]byte(buvid)) % 100; buvidInt >= 0 && buvidInt < 20 {
			noLoginAvatar = "1"
		} else if buvidInt >= 20 && buvidInt < 60 {
			noLoginAvatar = "2"
		} else {
			noLoginAvatar = "3"
		}
	}
	config = &show.Config{PopupStyle: 1}
	if avatar, ok := s.c.Custom.NoLoginAvatar[noLoginAvatar]; ok {
		config.NoLoginAvatar = avatar.URL
		config.NoLoginAvatarType = avatar.Type
	}
	return
}

func adjustInternational(ctx context.Context, dst *show.Show, ipMeta *location.Info, slocale, clocale string) {
	if i18n.PreferTraditionalChinese(ctx, slocale, clocale) {
		for _, v := range dst.Tab {
			i18n.TranslateAsTCV2(&v.Name)
		}
		for _, v := range dst.Top {
			i18n.TranslateAsTCV2(&v.Name)
		}
		for _, v := range dst.Bottom {
			i18n.TranslateAsTCV2(&v.Name)
		}
	}
	if ipMeta != nil {
		// 非大陆 ip 且使用繁体用户移除底部会员购
		// zone_id 计算参考 https://info.bilibili.co/pages/viewpage.action?pageId=4530206#location项目文档-接口数据说明
		//nolint:gomnd
		if i18n.PreferTraditionalChinese(ctx, slocale, clocale) && (ipMeta.Country != "局域网" && ((ipMeta.ZoneId >> 22) != 1)) {
			bottom := dst.Bottom
			dst.Bottom = make([]*show.Tab, 0, len(dst.Bottom))
			for _, t := range bottom {
				if t.TabID == "会员购Bottom" {
					continue
				}
				dst.Bottom = append(dst.Bottom, t)
			}
		}
	}
}

func adaptSLocale(ctx context.Context, dst *show.Show, slocale, clocale string) {
	if !i18n.PreferTraditionalChinese(ctx, slocale, clocale) {
		return
	}
	for _, v := range dst.Tab {
		i18n.TranslateAsTCV2(&v.Name)
	}
	for _, v := range dst.Top {
		i18n.TranslateAsTCV2(&v.Name)
	}
	for _, v := range dst.Bottom {
		i18n.TranslateAsTCV2(&v.Name)
		if v.Type != int64(resourceApi.SectionItemOpLinkType_DIALOG_OPENER) {
			continue
		}
		for _, item := range v.DialogItems {
			i18n.TranslateAsTCV2(&item.Name)
		}
	}
}

func (s *Service) menus(plat int8, build, lessonsMode int, policyAuths map[int64]*locmdl.ZoneLimitAuth) (res []*show.Tab) {
	memuCache := s.menuCache
LOOP:
	for _, m := range memuCache {
		if vs, ok := m.Versions[model.PlatAPPBuleChange(plat)]; ok {
			for _, v := range vs {
				if model.InvalidBuild(build, v.Build, v.Condition) {
					continue LOOP
				}
			}
			// 课堂模式下，非课堂模式的卡片过滤
			if lessonsMode == 1 {
				if !m.LessonsMode {
					continue
				}
			} else if !m.NormalMode {
				continue
			}
			// 地区限制
			if m.AreaPolicy > 0 {
				auth, ok := policyAuths[m.AreaPolicy]
				if ok {
					if auth.Play == locmdl.Status_Forbidden {
						continue
					}
				}
			}
			// areaInt, _ := strconv.ParseInt(m.Area, 10, 64)
			// if auth, ok := auths[areaInt]; ok && auth.Play == int64(locmdl.Status_Forbidden) {
			// 	continue
			// }
			t := &show.Tab{}
			t.TabMenuChange(m)
			res = append(res, t)
		}
	}
	return
}

func (s *Service) buildForbid(sid int64, build int) bool {
	for _, l := range s.limitsCahce[sid] {
		if model.InvalidBuild(build, l.Build, l.Condition) {
			return true
		}
	}
	return false
}

//nolint:gocognit
func (s *Service) TabBubble(c context.Context, plat int8, build, teenagersMode, lessonsMode int, buvid,
	language string, mid int64) (res map[int64]*show.TabBubble, err error) {
	// 与tab过滤逻辑保持一致
	//nolint:staticcheck
	if key := fmt.Sprintf(_initTabKey, plat, language); len(s.tabCache[fmt.Sprintf(key)]) == 0 || language == "" {
		if model.IsOverseas(plat) {
			var key = fmt.Sprintf(_initTabKey, plat, _defaultLanguageHant)
			//nolint:staticcheck
			if len(s.tabCache[fmt.Sprintf(key)]) > 0 {
				language = _defaultLanguageHant
			} else {
				language = _defaultLanguageHans
			}
		} else {
			language = _defaultLanguageHans
		}
	}
	var (
		key     = fmt.Sprintf(_initTabKey, plat, language)
		tmptabs = []*show.Tab{}
	)
	if tabs, ok := s.tabCache[key]; ok {
		for _, v := range tabs {
			if v.ModuleStr != "bottom" {
				continue
			}
			buildForbid := s.buildForbid(v.ID, build)
			if buildForbid {
				continue
			}
			if !s.c.ShowHotAll {
				if ab, ok := s.abtestCache[v.Group]; ok {
					if _, ok := s.showTabMids[mid]; !ab.AbTestIn(buvid) && !ok {
						continue
					}
				}
			}
			if _, ok := _showWhiteList[v.URI]; (teenagersMode != 0 || lessonsMode != 0) && !ok {
				continue
			}
			tmptabs = append(tmptabs, v)
		}
	}
	res = make(map[int64]*show.TabBubble)
	for _, tab := range tmptabs {
		if bcache, ok := s.bubbleCache[tab.ID]; ok {
			if bcache == nil {
				continue
			}
			state, err := s.bubbleDao.BubbleConfig(c, bcache.ID, mid)
			if err != nil {
				log.Error("TabBubble(%d) get mc error %v", bcache.ID, err)
				//nolint:ineffassign
				err = nil
				continue
			}
			if state == -1 {
				continue
			}
			var isPush bool
			if state == model.BubblePushing {
				expire := int32(bcache.ETime.Time().Unix() - time.Now().Unix())
				if expire <= 0 {
					log.Error("bubble expire error bid(%d) etime(%v)", bcache.ID, bcache.ETime.Time())
					continue
				}
				if err := s.bubbleDao.SetBubbleConfig(c, bcache.ID, mid, model.BubblePushed, expire); err != nil {
					log.Error("bubble(%d) set mc error %v", bcache.ID, err)
					continue
				}
				isPush = true
			}
			// 刷新mc成功才推送
			if !isPush {
				continue
			}
			bubble := &show.TabBubble{
				Key:   key,
				ID:    bcache.ID,
				Title: bcache.Desc,
				Cover: bcache.Icon,
				URI:   bcache.URL,
				Param: strconv.FormatInt(bcache.ID, 10),
				STime: bcache.STime,
				ETime: bcache.ETime,
			}
			res[tab.ID] = bubble
		}
	}
	return
}

// isOkSkin
func (s *Service) isOkSkin(data *resourceApi.SkinInfo, plat int8, build int) (isOk bool) {
	var (
		nowTime = xtime.Time(time.Now().Unix())
	)
	if data == nil || data.Info == nil || len(data.Limit) == 0 {
		return
	}
	// 判断满足条件的皮肤
	if data.Info.Stime > nowTime || nowTime > data.Info.Etime {
		return
	}
	for _, vLt := range data.Limit {
		// 需要满足当前plat下所有的条件，才算校验通过
		if vLt.Plat == int32(plat) {
			// 有配置限制才下发皮肤
			isOk = true
			if model.InvalidBuild(build, int(vLt.Build), vLt.Conditions) {
				// 有一个版本校验不通过时，则认为不满足条件
				isOk = false
				break
			}
		}
	}
	return
}

// zlimitInfo .
func (s *Service) zlimitInfo(c context.Context, gids []int64) (int64, map[int64]*locmdl.GroupAuth, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	zlimitRly, e := s.loc.ZlimitInfo(c, []string{ip}, gids)
	if e != nil {
		log.Error("zlimitInfo s.loc.ZlimitInfo(%s,%v) error(%v)", ip, gids, e)
		// IP服务异常，视作不下发的皮肤
		return 0, nil, e
	}
	// IP服务异常，视作没有满足
	if zlimitRly == nil {
		return 0, nil, ecode.NothingFound
	}
	// 获取ip所在的zone_id
	var ipZoneID int64
	if zVal, zok := zlimitRly.Infos[ip]; zok && zVal != nil && len(zVal.ZoneId) >= 2 {
		// ZoneId 第0位全地区、第1位国家、第2位省份、第3位城市
		ipZoneID = zVal.ZoneId[1]
	} else {
		// IP服务异常,未找到当前ip对应的zoneid
		log.Error("zlimitInfo s.loc.ZlimitInf(%s) not found zoneid", ip)
		return 0, nil, ecode.NothingFound
	}
	return ipZoneID, zlimitRly.Policy, nil
}

func filterByLocation(in []*resourceApi.SkinInfo,
	ipErr error,
	ipZoneID int64,
	zPolicy map[int64]*locmdl.GroupAuth) []*resourceApi.SkinInfo {
	out := make([]*resourceApi.SkinInfo, 0, len(in))
	for _, sv := range in {
		if sv == nil || sv.Info == nil {
			continue
		}
		// 全球生效
		if sv.Info.LocationPolicyGID == 0 {
			out = append(out, sv)
			continue
		}
		// ip服务发生错误，则不下发该皮肤
		if ipErr != nil {
			continue
		}
		gval, gok := zPolicy[sv.Info.LocationPolicyGID]
		if !gok || gval == nil {
			// 没有找到对应的规则不下发该皮肤
			continue
		}
		// 遍历当前组id下的规则
		var isSuccess bool
		for _, rule := range gval.PolicyAuths {
			if rule == nil {
				continue
			}
			rval, rok := rule.ZoneAuths[ipZoneID]
			// 在所有规则中未找到对应zoneid的限制
			if !rok || rval == nil {
				continue
			}
			// 存在对应的规则，判断是否允许下发play 的：1是禁止，2是允许
			//nolint:gomnd
			if rval.Play == 2 {
				isSuccess = true
			} else { // 规则中有一项不满足，则该皮肤不下发
				isSuccess = false
				break
			}
		}
		if isSuccess {
			out = append(out, sv)
			continue
		}
	}
	return out
}

func memberInMap(bgRes *bGroup.MemberInReply) map[string]bool {
	if bgRes == nil {
		return nil
	}
	out := make(map[string]bool)
	for _, v := range bgRes.Results {
		out[v.Name] = v.In
	}
	return out
}

func attrVal(attr int64, bit int64) bool {
	return (attr>>bit)&int64(1) == int64(1)
}

func filterByDressUp(sv *resourceApi.SkinInfo, mid int64, isFreeTheme, isSkinDress bool,
	bgErr error,
	bgRes *bGroup.MemberInReply) int64 {
	if sv == nil || sv.Info == nil {
		return 0
	}
	attr := sv.Info.Attribute
	//未开启主动位用户装扮 || 不是高版本
	if !attrVal(sv.Info.Attribute, 0) || !isSkinDress {
		return attr
	}
	memberIn := memberInMap(bgRes)
	switch sv.Info.DressUpType {
	// 仅纯色用户
	case "only_pure":
		if !isFreeTheme {
			attr = sv.Info.Attribute & (math.MaxInt64 - 1<<0)
		}
	// 人群包
	case "mid_scope":
		if mid <= 0 || bgErr != nil { //(未登录 || 人群包出错)不开启强制用户装扮
			attr = sv.Info.Attribute & (math.MaxInt64 - 1<<0)
		} else {
			tr := cvtAsMemberInBGroup(sv.Info.DressUpValue)
			// 人群包中不存在该mid，不开启强制用户装扮
			if !memberIn[tr.Name] {
				attr = sv.Info.Attribute & (math.MaxInt64 - 1<<0)
			}
		}
	default: // 无人群限制
	}
	return attr
}

func filterByUserScope(in []*resourceApi.SkinInfo, mid int64, isFreeTheme bool,
	bgErr error,
	bgRes *bGroup.MemberInReply) []*resourceApi.SkinInfo {
	memberIn := memberInMap(bgRes)
	out := make([]*resourceApi.SkinInfo, 0, len(in))
	for _, sv := range in {
		if sv == nil || sv.Info == nil {
			continue
		}
		switch sv.Info.UserScopeType {
		// 无人群限制
		case "":
			out = append(out, sv)
		// 仅纯色用户
		// 以前纯色即免费，但是还有部分运营主题也需要被覆盖，这里改成免费主题即下发
		case "only_pure":
			if isFreeTheme {
				out = append(out, sv)
			}
		// 人群包
		case "mid_scope":
			if mid <= 0 {
				continue
			}
			// 人群包出错不返回
			if bgErr != nil {
				continue
			}
			tr := cvtAsMemberInBGroup(sv.Info.UserScopeValue)
			// 人群包中存在该mid，追加至结果
			if memberIn[tr.Name] {
				out = append(out, sv)
			}
		default:
			log.Warn("Unreconginzed user scope type: %+v", sv.Info)
			continue
		}
	}
	return out
}

func cvtAsMemberInBGroup(in string) *bGroup.MemberInReq_MemberInReqSingle {
	if in == "" {
		return nil
	}
	parts := strings.Split(in, "|")
	//nolint:gomnd
	if len(parts) < 2 {
		return nil
	}
	return &bGroup.MemberInReq_MemberInReqSingle{
		Business: parts[0],
		Name:     parts[1],
	}
}

// commonSkin .
func (s *Service) commonSkin(c context.Context, plat int8, build int, mid int64, isFreeTheme bool) (res *show.SkinConf, err error) {
	var (
		skinsInfo   = s.skinCache
		succAllSkin []*resourceApi.SkinInfo
		skinsReply  *garb.SkinListReply
		gids        []int64
		bgGroups    []*bGroup.MemberInReq_MemberInReqSingle
	)
	//是否是62及以上版本
	var isSkinDress bool
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.SkinDress, nil) {
		isSkinDress = true
	}
	for _, val := range skinsInfo {
		if !s.isOkSkin(val, plat, build) {
			continue
		}
		// 判断所有满足条件的皮肤
		succAllSkin = append(succAllSkin, val)
		// 限制生效区域
		if val.Info != nil && val.Info.LocationPolicyGID > 0 {
			gids = append(gids, val.Info.LocationPolicyGID)
		}

		if val.Info != nil && val.Info.UserScopeType == "mid_scope" {
			bg := cvtAsMemberInBGroup(val.Info.UserScopeValue)
			if bg == nil {
				continue
			}
			bgGroups = append(bgGroups, bg)
		}
		//开启强制装扮&&开启下发人群限制 start
		if isSkinDress && val.Info != nil && attrVal(val.Info.Attribute, 0) && val.Info.DressUpType == "mid_scope" {
			bg := cvtAsMemberInBGroup(val.Info.DressUpValue)
			if bg == nil {
				continue
			}
			bgGroups = append(bgGroups, bg)
		}
		//开启强制装扮&&开启下发人群限制 end
	}

	// 没有获取满足条件的皮肤配置
	if len(succAllSkin) == 0 {
		return
	}

	eg := errgroup.WithContext(c)
	var (
		ipErr    error
		ipZoneID int64
		zPolicy  map[int64]*locmdl.GroupAuth
	)
	if len(gids) > 0 {
		eg.Go(func(ctx context.Context) error {
			if ipZoneID, zPolicy, ipErr = s.zlimitInfo(c, gids); ipErr != nil {
				// ip服务获取失败,降级处理
				log.Error("s.zlimitInf(%v) error(%v)", gids, ipErr)
				return nil
			}
			return nil
		})
	}
	var (
		bgErr error
		bgRes *bGroup.MemberInReply
	)
	if mid > 0 && len(bgGroups) > 0 {
		eg.Go(func(ctx context.Context) error {
			bgRes, bgErr = s.bgroupDao.MemberIn(c, &bGroup.MemberInReq{
				Member:    strconv.FormatInt(mid, 10),
				Groups:    bgGroups,
				Dimension: bGroup.Mid,
			})
			if bgErr != nil {
				log.Error("s.bgroupDao.MemberIn(%v) error(%v)", bgGroups, bgErr)
				return nil
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// 过滤满足条件的皮肤
	succAllSkin = filterByLocation(succAllSkin, ipErr, ipZoneID, zPolicy)        // 地区限制
	succAllSkin = filterByUserScope(succAllSkin, mid, isFreeTheme, bgErr, bgRes) // 用户限制

	var succSkin *resourceApi.SkinInfo
	if len(succAllSkin) > 0 {
		succSkin = succAllSkin[0]
	}
	if succSkin == nil || succSkin.Info == nil {
		return
	}
	if skinsReply, err = s.garbDao.SkinList(c, []int64{succSkin.Info.SkinID}); err != nil {
		log.Error("s.garbDao.SkinList(%d) error(%v)", succSkin.Info.SkinID, err)
		return
	}
	if skinsReply == nil {
		return
	}
	if skinVal, ok := skinsReply.Skins[succSkin.Info.SkinID]; ok {
		attr := filterByDressUp(succSkin, mid, isFreeTheme, isSkinDress, bgErr, bgRes) //是否主动为用户装扮
		res = &show.SkinConf{
			ID:         skinVal.ID,
			Name:       skinVal.Name,
			Preview:    skinVal.Preview,
			Ver:        skinVal.Ver,
			PackageUrl: skinVal.PackageUrl,
			PackageMd5: skinVal.PackageMd5,
			Data:       skinVal.Data,
			Conf: &show.SingleConf{
				Alias:     succSkin.Info.SkinName,
				Attribute: attr,
				STime:     succSkin.Info.Stime,
				ETime:     succSkin.Info.Etime,
			},
		}
	}
	return
}

// skinColors .
func (s *Service) skinColors(c context.Context, mid int64, build int, mobiApp string) (res []*show.SkinColor, err error) {
	var (
		skinsReply *garb.SkinColorUserListReply
	)
	if skinsReply, err = s.garbDao.SkinColorUserList(c, mid, int64(build), mobiApp); err != nil {
		log.Error(" s.garbDao.SkinColorUserList(%d) error(%v)", mid, err)
		return
	}
	if skinsReply == nil || len(skinsReply.List) == 0 {
		return
	}
	for _, val := range skinsReply.List {
		res = append(res, &show.SkinColor{
			ID:        val.ID,
			Name:      val.Name,
			IsFree:    val.IsFree,
			Price:     val.Price,
			IsBought:  val.IsBought,
			Status:    val.Status,
			BuyTime:   val.BuyTime,
			DueTime:   val.DueTime,
			ColorName: val.ColorName,
		})
	}
	return
}

// userEquip .
func (s *Service) userEquip(c context.Context, mid int64) (res *show.SkinConf, err error) {
	var (
		skinsReply *garb.SkinUserEquipReply
	)
	if skinsReply, err = s.garbDao.SkinUserEquip(c, mid); err != nil {
		log.Error(" s.garbDao.SkinUserEquip(%d) error(%v)", mid, err)
		return
	}
	if skinsReply == nil || skinsReply.Skin == nil {
		return
	}
	res = &show.SkinConf{
		ID:         skinsReply.Skin.ID,
		Name:       skinsReply.Skin.Name,
		Preview:    skinsReply.Skin.Preview,
		Ver:        skinsReply.Skin.Ver,
		PackageUrl: skinsReply.Skin.PackageUrl,
		PackageMd5: skinsReply.Skin.PackageMd5,
		Data:       skinsReply.Skin.Data,
	}
	return
}

// Skin 依赖接口下发err时，直接返回，与客户端协商后决定不做降级，客户端保持上次冷起皮肤特效.
func (s *Service) Skin(c context.Context, plat int8, build int, mid int64, mobiApp string, isFreeTheme bool) (reply *show.SkinReply, err error) {
	rly := &show.SkinReply{}
	eg := errgroup.WithContext(c)
	// 获取运营资源
	// 只支持安卓，iphone,不支持ipad上的粉版，已和产品确定
	if plat == model.PlatAndroid || plat == model.PlatIPhone {
		// 554以上版本才下发皮肤资源
		if (plat == model.PlatAndroid && build >= 5540500) || (plat == model.PlatIPhone && build >= 9200) {
			// 错误不降级与客户端协定
			eg.Go(func(ctx context.Context) (e error) {
				if rly.CommonEquip, e = s.commonSkin(ctx, plat, build, mid, isFreeTheme); e != nil {
					log.Error("s.CommonSkin(%d,%d) error(%v)", plat, build, e)
				}
				return
			})
			if mid > 0 {
				// 获取个人配置
				eg.Go(func(ctx context.Context) (e error) {
					if rly.UserEquip, e = s.userEquip(ctx, mid); e != nil {
						log.Error(" s.userEquip(%d) error(%v)", mid, e)
					}
					return
				})
			}
		}
		// 获取皮肤资源
		eg.Go(func(ctx context.Context) (e error) {
			if rly.SkinColors, e = s.skinColors(ctx, mid, build, mobiApp); e != nil {

				//nolint:govet
				log.Error("s.skinColors(%d,%s,%s) error(%v)", mid, build, mobiApp, e)
			}
			return
		})
	}
	// 没有版本限制
	if mid > 0 {
		eg.Go(func(ctx context.Context) error {
			// 获取装扮信息 LoadingUserEquip
			loadEqRly, e := s.garbDao.LoadingUserEquip(ctx, mid)
			if e != nil {
				log.Error("s.garbDao.LoadingUserEquip mid(%d) error(%v)", mid, e)
				return e
			}
			if loadEqRly == nil || loadEqRly.Loading == nil {
				return nil
			}
			rly.LoadEquip = &show.LoadEquip{ID: loadEqRly.Loading.ID, Name: loadEqRly.Loading.Name, Ver: loadEqRly.Loading.Ver, LoadingUrl: loadEqRly.Loading.LoadingUrl}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	reply = rly
	ios2021NationalDayPatch(c, reply)
	return
}

// 2021.10.07 23:59:59 后可移除
func ios2021NationalDayPatch(ctx context.Context, in *show.SkinReply) {
	dev, ok := device.FromContext(ctx)
	if !ok {
		return
	}
	if !dev.IsIOS() {
		return
	}

	if in.CommonEquip == nil {
		return
	}
	if in.CommonEquip.Name != "喜迎国庆" {
		return
	}
	if in.CommonEquip.Data == nil {
		return
	}
	in.CommonEquip.Data.TailColor = "61666d"
	in.CommonEquip.Data.TailColorSelected = "cd382f"
	in.CommonEquip.Data.TailIconMode = "img"

	in.CommonEquip.Data.TailIconColor = ""
	in.CommonEquip.Data.TailIconColorDark = ""
	in.CommonEquip.Data.TailIconColorSelected = ""
	in.CommonEquip.Data.TailIconColorSelectedDark = ""

	in.CommonEquip.PackageMd5 = "9abc265063681cb236a68afa3379257b"
	in.CommonEquip.PackageUrl = "http://i0.hdslb.com/bfs/garb/zip/42cb3c4999513afa406584bd7104f004416ecccd.zip"
}

// ClickTab .
func (s *Service) ClickTab(c context.Context, id int64, buvid, ver, cType string) error {
	if cType == show.ClearVer {
		if err := s.rdao.AddMenuExtVer(c, id, buvid, ver); err != nil {
			log.Error("s.rdao.AddMenuExtVer(%d,%s,%s) error(%v)", id, buvid, ver, err)
		}
	}
	// 不下发错误，可降级处理
	return nil
}

// TopActivity is top activity entrance
func (s *Service) TopActivity(c context.Context, params *show.TopActivityReq, mid int64, buvid, ua string) (*show.TopActivityReply, error) {
	if s.c.Custom.TopActivityMngSwitch { // 是否请求后台入口配置开关
		// get s10 区分安卓 ios粉 ipad粉 iPadHD
		plat := model.Plat2(params.MobiApp, params.Device)
		mngActRly, err := s.rdao.TopActivity(c, params.Build, plat)
		mngAct := mngActRly.GetItem()
		if err == nil && mngAct != nil {
			online := &show.TopOnline{
				Icon:     mngAct.StaticIcon,
				Uri:      mngAct.Url,
				UniqueID: mngAct.StateName,
				Interval: s.c.Custom.TopActivityInterval,
				Animate: &show.TopAnimate{
					Svg:  mngAct.DynamicIcon,
					Loop: mngAct.LoopCount,
				},
				Type: show.TopActivityMng,
				Name: mngAct.EntryName,
			}
			bs, _ := json.Marshal(online)
			hash := strconv.FormatUint(farm.Hash64(bs), 10)
			return &show.TopActivityReply{Online: online, Hash: hash}, nil
		}
		if err != nil {
			if !ecode.EqualError(ecode.NothingFound, err) {
				log.Error("TopActivity s.rdao.TopActivity error(%+v) params(%+v)", err, params)
			}
			//nolint:ineffassign
			err = nil
		}
	}
	if s.c.Custom.TopActivityFissionSwitch { // 是否请求裂变入口配置开关
		// get fission
		fission, err := s.fissionDao.Entrance(c, mid, params.Build, buvid, params.MobiApp, params.Device, params.Platform, ua)
		if err != nil {
			log.Error("TopActivity s.fissionDao.Entrance params(%+v) mid(%d) buvid(%s) ua(%s) error(%+v)", params, mid, buvid, ua, err)
			return nil, ecode.NothingFound
		}
		if fission == nil || fission.Icon == "" {
			log.Error("TopActivity s.fissionDao.Entrance is empty fission(%+v) params(%+v) mid(%d) buvid(%s) ua(%s) error(%+v)", fission, params, mid, buvid, ua, err)
			return nil, ecode.NothingFound
		}
		var animate *show.TopAnimate
		if fission.AnimateIcon != nil {
			animate = &show.TopAnimate{
				Icon: fission.AnimateIcon.Icon,
				Json: fission.AnimateIcon.Json,
				Loop: 2, // 七日活动动效循环2次
			}
		}
		online := &show.TopOnline{
			Icon:     fission.Icon,
			Uri:      fission.Url,
			Interval: s.c.Custom.TopActivityInterval,
			Animate:  animate,
			Type:     show.TopActivityFission,
			Name:     fission.Name,
		}
		bs, _ := json.Marshal(online)
		hash := strconv.FormatUint(farm.Hash64(bs), 10)
		log.Warn("TopActivity params(%+v) mid(%d) buvid(%s) ua(%s) response(%+v)", params, mid, buvid, ua, &show.TopActivityReply{Online: online, Hash: hash})
		return &show.TopActivityReply{Online: online, Hash: hash}, nil
	}
	return nil, ecode.NothingFound
}

func hasSchoolTab(in []*resourceApi.Section) bool {
	for _, v := range in {
		//不是首页tab
		if v == nil || v.Style != _tabStyle || v.Id != _tab {
			continue
		}
		for _, item := range v.Items {
			if item.Title == "校园" {
				return true
			}
		}
	}
	return false
}

//nolint:gocognit
func (s *Service) TabsV2(c context.Context, plat int8, buvid string, mid int64, params *show.TabsV2Params) (*show.Show, *show.Config, error) {
	if params.Lang == "" {
		params.Lang = _defaultLanguageHans
		if model.IsOverseas(plat) {
			params.Lang = _defaultLanguageHant
		}
	}
	var (
		tempTabs  []*show.Tab
		tabExtIDs []*resourceApi.Tab
		redDot    []*show.SectionURL
	)
	homeSections, err := s.rdao.HomeSections(c, mid, int32(plat), int32(params.Build), params.Lang, params.Channel, buvid)
	if err != nil {
		log.Error("s.rdao.HomeSections error(%+v)", err)
		return nil, nil, err
	}
	if homeSections == nil {
		return nil, nil, ecode.NothingFound
	}
	for _, section := range homeSections.Sections {
		if section == nil {
			continue
		}
		if section.Style == _tabStyle || section.Style == _topMoreStyle || section.Style == _dialogStyle || section.Style == _recommendStyle || section.Style == _publishBubble { //顶底tab||分区按钮||弹窗 || 港澳台垂类tab
			for _, item := range section.Items {
				if _, ok := _showWhiteList[item.Uri]; (params.TeenagersMode != 0 || params.LessonsMode != 0) && !ok {
					continue
				}
				if section.Id == _tab && section.Style == _tabStyle { //固定资源位配置,id=8-->顶部tab
					tabExtIDs = append(tabExtIDs, &resourceApi.Tab{TabId: section.Id, TType: tab.SideType})
				}
				if _, ok := _moduleMap[section.Id]; ok {
					if item.RedDotUrl != "" {
						redDot = append(redDot, &show.SectionURL{ID: item.Id, URL: item.RedDotUrl})
					}
				}
				tempTab := buildHomeTabItem(item, _deafaultTab, section.Id, section.Style)
				tempTabs = append(tempTabs, tempTab)
			}
		}
	}
	var showAuthDef = map[int64]*locmdl.ZoneLimitAuth{}
	if !s.auditTab(params.MobiApp, int(params.Build), plat) && params.TeenagersMode == 0 {
		for _, m := range s.menuCache {
			if _, ok := m.Versions[model.PlatAPPBuleChange(plat)]; !ok {
				continue
			}
			if m.TabID > 0 {
				tabExtIDs = append(tabExtIDs, &resourceApi.Tab{TabId: m.TabID, TType: tab.MenuType})
			}
			if m.AreaPolicy > 0 {
				showAuthDef[m.AreaPolicy] = &locmdl.ZoneLimitAuth{}
				switch m.ShowPurposed {
				case 0:
					showAuthDef[m.AreaPolicy].Play = locmdl.Status_Allow
				case 1:
					showAuthDef[m.AreaPolicy].Play = locmdl.Status_Forbidden
				}
			}
		}
	}
	var (
		mutex sync.Mutex
		eg    = errgroup.WithContext(c)
	)
	var tabExtMap = make(map[string]*resourceApi.TabExt, len(tabExtIDs))
	if len(tabExtIDs) > 0 {
		// 获取tab下特定颜色图片等配置
		eg.Go(func(ctx context.Context) error {
			tExtRly, err := s.rdao.GetTabExt(ctx, int64(plat), params.Build, buvid, tabExtIDs)
			if err != nil {
				log.Error("s.rdao.GetTabEx error(%+v)", err)
				return nil
			}
			for _, kv := range tExtRly {
				if kv == nil {
					continue
				}
				mutex.Lock()
				tabExtMap[fmt.Sprintf(_initTtExtKey, kv.TabId, kv.TType)] = kv
				mutex.Unlock()
			}
			return nil
		})
	}
	var policiesAuth map[int64]*locmdl.ZoneLimitAuth
	if len(showAuthDef) > 0 {
		eg.Go(func(ctx context.Context) error {
			req := &locmdl.ZoneLimitPoliciesReq{
				UserIp:       metadata.String(ctx, metadata.RemoteIP),
				DefaultAuths: showAuthDef,
			}
			reply, err := s.loc.ZoneLimitPolicies(ctx, req)
			if err != nil {
				log.Error("Failed to get ZoneLimitPolicies: %+v", err)
				return nil
			}
			policiesAuth = reply.Auths
			return nil
		})
	}
	var rdMap = make(map[int64]*show.Red, len(redDot))
	if len(redDot) > 0 {
		for _, v := range redDot {
			tmpID := v.ID
			tmpURL := v.URL
			eg.Go(func(ctx context.Context) error {
				red, err := s.redDao.RedDot(ctx, mid, tmpURL, params.Platform, model.GotoGame)
				if err != nil {
					log.Error("s.accDao.RedDot error(%+v)", err)
					return nil
				}
				if red != nil {
					mutex.Lock()
					rdMap[tmpID] = red
					mutex.Unlock()
				}
				return nil
			})
		}
	}
	var tempTopLeft = &show.TopLeft{Url: s.c.Custom.TopLeftDefaultUrl, Goto: 1}
	if showTopLeft(c, mid) {
		eg.Go(func(ctx context.Context) error {
			if s.hitWhiteList(ctx, mid) {
				s.buildTopLeftToStory(c, tempTopLeft, 0)
				return nil
			}
			if s.redis.HitTopLeftBlackList(ctx, mid) {
				log.Info("hit topleft black list mid(%+v)", mid)
				return nil
			}
			newUser, err := s.checkNewUser(ctx, mid)
			if err != nil {
				log.Error("checkNewUser error(%+v) mid(%d)", err, mid)
				return nil
			}
			var userRegState int64
			if newUser {
				userRegState = 1
			}
			if s.topLeftExp(ctx, newUser, mid) == 1 { //命中实验则展示特殊头标+刷视频跳链
				s.buildTopLeftToStory(c, tempTopLeft, userRegState)
			}
			return nil
		})
	}
	var changeSchoolTab bool
	if hasSchoolTab(homeSections.Sections) {
		eg.Go(func(ctx context.Context) error {
			changeSchoolTab = s.schoolDao.ChangeSchoolTabPosition(ctx, mid)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait error(%+v)", err)
	}
	for _, v := range tempTabs { //资源配置
		if eVal, ok := tabExtMap[fmt.Sprintf(_initTtExtKey, v.ID, tab.SideType)]; ok && eVal != nil {
			v.Extension = &show.Extension{}
			v.Extension.BuildExt(eVal)
		}
	}
	if !s.auditTab(params.MobiApp, int(params.Build), plat) && params.TeenagersMode == 0 {
		if menus := s.menus(plat, int(params.Build), int(params.LessonsMode), policiesAuth); len(menus) > 0 {
			// tab 运营资源位的资源配置
			for _, mVal := range menus {
				mID, mErr := strconv.ParseInt(mVal.TabID, 10, 64)
				if mErr != nil || mID <= 0 {
					continue
				}
				eVal, ok := tabExtMap[fmt.Sprintf(_initTtExtKey, mID, tab.MenuType)]
				if !ok || eVal == nil {
					continue
				}
				mVal.Extension = &show.Extension{}
				mVal.Extension.BuildExt(eVal)
			}
			tempTabs = append(tempTabs, menus...)
		}
	}
	var (
		dialogItems    []*show.DialogItems
		recommendItems []*show.Tab
		publishBubbles []*show.PublishBubble
	)
	for _, v := range tempTabs {
		switch v.ModuleStyle {
		case _dialogStyle:
			dialogItems = append(dialogItems, &show.DialogItems{
				ID:     v.ID,
				Icon:   v.Icon,
				Name:   v.Name,
				Uri:    v.URI,
				OpIcon: v.OpIcon,
			})
		case _recommendStyle:
			recommendItems = append(recommendItems, v)
		case _publishBubble:
			publishBubbles = append(publishBubbles, &show.PublishBubble{
				ID:   v.ID,
				Icon: v.Icon,
				Url:  v.URI,
			})
		default:
			continue
		}
	}
	func() {
		if len(publishBubbles) == 0 {
			return
		}
		bubblebs, err := json.Marshal(publishBubbles)
		if err != nil {
			log.Error("publish bubble json marshal error:%+v", err)
			return
		}
		payload := infoc.NewLogStreamV(s.c.Custom.ShowTabLogID,
			log.Int64(mid),
			log.Int64(params.Build),
			log.String(params.MobiApp),
			log.String(params.Device),
			log.String(buvid),
			log.String(string(bubblebs)),
		)
		if err := s.infoc.Info(c, payload); err != nil {
			log.Error("show tab s.infoc.Info error: %+v", err)
			return
		}
	}()
	res := &show.Show{
		TopLeft: tempTopLeft,
	}
	for _, v := range tempTabs {
		if r, ok := rdMap[v.ID]; ok && v.Red != "" && r != nil {
			if r.RedDot {
				v.RedDot = &show.RedDot{Type: r.Type, Number: r.Number}
			}
			if r.Icon != "" {
				v.AnimateIcon = &show.AnimateIcon{Icon: r.Icon, Json: v.Animate}
			}
		}
		switch v.ModuleStr {
		case model.ModuleTop:
			v.Pos = len(res.Top) + 1
			res.Top = append(res.Top, v)
		case model.ModuleTab:
			v.Pos = len(res.Tab) + 1
			res.Tab = append(res.Tab, v)
		case model.ModuleBottom:
			if v.Type == int64(resourceApi.SectionItemOpLinkType_DIALOG_OPENER) {
				v.DialogItems = dialogItems
				v.PublishBubble = publishBubbles
			}
			v.Pos = len(res.Bottom) + 1
			res.Bottom = append(res.Bottom, v)
		case model.ModuleTopMore:
			v.Pos = len(res.TopMore) + 1
			res.TopMore = append(res.TopMore, v)
		}
	}
	if len(recommendItems) > 0 {
		for _, v := range recommendItems {
			v.Pos = len(res.Tab) + 1
			res.Tab = append(res.Tab, v)
		}
		//nolint:gomnd
		if len(res.Tab) > 20 { //首页tab最多下发20个
			res.Tab = res.Tab[:20]
		}
	}
	adaptRcmd(params.DisableRcmd == 1, res)
	if changeSchoolTab {
		changePositionBetweenSchoolAndHotTab(res.Tab)
	}
	adaptSLocale(c, res, params.Slocale, params.Clocale)
	config := &show.Config{PopupStyle: 1}
	if avatar, ok := s.c.Custom.NoLoginAvatar["all"]; ok {
		config.NoLoginAvatar = avatar.URL
		config.NoLoginAvatarType = avatar.Type
	}
	return res, config, nil
}

func changePositionBetweenSchoolAndHotTab(tabs []*show.Tab) {
	var schoolTabIndex, hotTabIndex int
	for index, v := range tabs {
		switch v.Name {
		case "校园":
			schoolTabIndex = index
		case "热门":
			hotTabIndex = index
		}
	}
	if schoolTabIndex == 0 || hotTabIndex == 0 {
		return
	}
	//先换pos,再换位置
	tabs[schoolTabIndex].Pos, tabs[hotTabIndex].Pos = tabs[hotTabIndex].Pos, tabs[schoolTabIndex].Pos
	tabs[schoolTabIndex], tabs[hotTabIndex] = tabs[hotTabIndex], tabs[schoolTabIndex]
}

func adaptRcmd(disableRcmd bool, tab *show.Show) {
	if !disableRcmd {
		return
	}
	var disableRcmdMap = map[int64]struct{}{
		984: {}, //android校园tab
		985: {}, //ios校园tab
	}
	TabFilter := show.TabsFilter(tab.Tab)
	res := TabFilter.Filter(func(in *show.Tab) bool {
		_, ok := disableRcmdMap[in.ID]
		return ok
	})
	tab.Tab = res
}

func (s *Service) buildTopLeftToStory(ctx context.Context, in *show.TopLeft, userRegState int64) {
	in.HeadTag = s.c.Custom.TopLeftHeadTag
	in.StoryBackgroundImage = s.c.Custom.IosTopLeftStoryBackgroundImage
	in.StoryForegroundImage = s.c.Custom.IosTopLeftStoryForegroundImage
	in.ListenBackgroundImage = s.c.Custom.IosTopLeftListenBackgroundImage
	in.ListenForegroundImage = s.c.Custom.IosTopLeftListenForegroundImage
	//默认使用ios图片，如果是安卓则使用安卓的图片
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid()
	}).MustFinish() {
		in.StoryBackgroundImage = s.c.Custom.AndroidTopLeftStoryBackgroundImage
		in.StoryForegroundImage = s.c.Custom.AndroidTopLeftStoryForegroundImage
		in.ListenBackgroundImage = s.c.Custom.AndroidTopLeftListenBackgroundImage
		in.ListenForegroundImage = s.c.Custom.AndroidTopLeftListenForegroundImage
	}
	in.Url = fmt.Sprintf("%s?user_reg_state=%d", s.c.Custom.TopLeftSpecialUrl, userRegState)
	in.Exp = 1
	in.Goto = 2
}

func (s *Service) hitWhiteList(ctx context.Context, mid int64) bool {
	var (
		business = "DF"
		name     = "lifeng_over"
	)
	bGroupReply, err := s.bgroupDao.MemberIn(ctx, &bGroup.MemberInReq{
		Member: strconv.FormatInt(mid, 10),
		Groups: []*bGroup.MemberInReq_MemberInReqSingle{
			{
				Business: business,
				Name:     name,
			},
		},
	})
	if err != nil {
		log.Error("s.bgroupDao.MemberIn error(%+v) mid(%d)", err, mid)
		return false
	}
	if len(bGroupReply.Results) == 0 {
		log.Error("s.bgroupDao.MemberIn error(%+v) mid(%d)", ecode.NothingFound, mid)
		return false
	}
	if bGroupReply.Results[0].GetBusiness() != business || bGroupReply.Results[0].GetName() != name {
		log.Error("s.bgroupDao.MemberIn business or name mismatched mid(%d)", mid)
		return false
	}
	return bGroupReply.Results[0].GetIn()
}

func (s *Service) checkNewUser(ctx context.Context, mid int64) (bool, error) {
	duration, err := getDurationBetweenExpAndNow()
	if err != nil {
		return false, err
	}
	reply, err := s.accountClient.CheckRegTime(ctx, &account.CheckRegTimeReq{Mid: mid, Periods: fmt.Sprintf("0-%d", int64(duration.Hours()))})
	if err != nil {
		return false, err
	}
	return reply.Hit, nil
}

func (s *Service) topLeftExp(ctx context.Context, newUser bool, mid int64) int64 {
	if newUser {
		return abtestRun(ctx, _topLeftNewUserFlag)
	}
	//nolint:gomnd
	group := crc32.ChecksumIEEE([]byte(fmt.Sprintf("%d_app_change_mode_old", mid))) % 10
	if _, ok := s.c.TopLeftExpGroup[strconv.FormatInt(int64(group), 10)]; ok {
		return 1
	}
	return 0
}

func abtestRun(ctx context.Context, flag *ab.IntFlag) int64 {
	t, ok := ab.FromContext(ctx)
	if !ok {
		return -1
	}
	return flag.Value(t)
}

func showTopLeft(ctx context.Context, mid int64) bool {
	if mid == 0 {
		return false
	}
	return pd.WithContext(ctx).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPhone().And().Build(">=", 66600000)
	}).OrWhere(func(pdContext *pd.PDContext) {
		pdContext.IsPlatAndroid().And().Build(">=", 6660000)
	}).FinishOr(false)
}

func getDurationBetweenExpAndNow() (time.Duration, error) {
	expStartTime, err := time.ParseInLocation("2006-01-02", "2022-03-24", time.Local)
	if err != nil {
		return 0, err
	}
	return time.Since(expStartTime), nil
}

func buildHomeTabItem(rsb *resourceApi.SectionItem, defaultTab map[string]*show.Tab, moduleID int64, moduleStyle int32) *show.Tab {
	var (
		_top    = 10
		_tab    = 8
		_bottom = 9
		t       = &show.Tab{}
	)
	t.ID = rsb.Id
	t.Icon = rsb.Icon
	t.IconSelected = rsb.LogoSelected
	t.Name = rsb.Title
	t.URI = rsb.Uri
	t.Module = int(moduleID)
	t.ModuleStyle = moduleStyle
	t.Type = int64(rsb.OpLinkType)
	t.Red = rsb.RedDotUrl
	t.Animate = rsb.Animate
	if rsb.GetMngIcon().GetIcon() != "" {
		t.OpIcon = &show.OpIcon{
			Id:    rsb.GetMngIcon().GetId(),
			Icon:  rsb.GetMngIcon().GetIcon(),
			ETime: rsb.GetMngIcon().GetStime().Time().Unix(),
			FTime: rsb.GetMngIcon().GetEtime().Time().Unix(),
		}
	}
	if t.ModuleStyle == _tabStyle {
		switch t.Module {
		case _top:
			t.ModuleStr = model.ModuleTop
		case _tab:
			t.ModuleStr = model.ModuleTab
			t.Icon = ""
			t.IconSelected = ""
		case _bottom:
			t.ModuleStr = model.ModuleBottom
		}
	}
	if t.ModuleStyle == _topMoreStyle {
		t.ModuleStr = model.ModuleTopMore
	}
	if t.Type == int64(resourceApi.SectionItemOpLinkType_NA_PAGE_ID) {
		if t.URI != "" {
			t.URI = fmt.Sprintf("bilibili://following/home_bottom_tab_activity_tab/%s", t.URI)
		}
	}
	if len(defaultTab) > 0 {
		if dt, ok := defaultTab[t.URI]; ok && dt != nil {
			t.DefaultSelected = dt.DefaultSelected
			t.TabID = dt.TabID
		}
		if rsb.TabId != "" {
			t.TabID = rsb.TabId
		}
	}
	return t
}

func (s *Service) tabDeniedByWhiteURL(ctx context.Context, mid int64, tabs []*show.Tab) map[int64]struct{} {
	lock := &sync.Mutex{}
	deniedByWhiteURL := map[int64]struct{}{}

	eg := errgroup.WithContext(ctx)
	for _, v := range tabs {
		v := v
		if v.WhiteURL == "" {
			continue
		}
		// 未登录
		if mid <= 0 {
			// 也下发
			if v.WhiteURLShow > 0 {
				continue
			}
			// 不下发
			deniedByWhiteURL[v.ID] = struct{}{}
			continue
		}
		eg.Go(func(ctx context.Context) error {
			inWhiteList, err := s.rdao.UserCheck(ctx, mid, v.WhiteURL)
			if err != nil {
				log.Error("Failed to check user is in white list: %d with url: %q: %+v", mid, v.WhiteURL, err)
				if v.WhiteURLShow > 0 {
					return nil
				}
				lock.Lock()
				defer lock.Unlock()
				deniedByWhiteURL[v.ID] = struct{}{}
				return nil
			}
			if !inWhiteList {
				lock.Lock()
				defer lock.Unlock()
				deniedByWhiteURL[v.ID] = struct{}{}
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("Failed to execute error group: %+v", err)
		return nil
	}
	return deniedByWhiteURL
}
