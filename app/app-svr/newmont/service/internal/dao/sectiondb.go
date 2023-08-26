package dao

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"go-gateway/app/app-svr/newmont/service/api"
	secmdl "go-gateway/app/app-svr/newmont/service/internal/model/section"

	"go-common/library/log"
)

const (
	_sidebarSQL = `
		SELECT
			id, plat, module, name, logo,
 			logo_white, param, rank,
			tip, need_login, white_url, logo_selected, tab_id,
			red_dot_url, lang_id,global_red_dot, red_dot_limit, animate,
			white_url_show, area, show_purposed,area_policy,red_dot_for_new,
			op_load_condition, op_title, op_sub_title, op_title_icon, op_link_type, 
			op_link_text, op_link_icon, op_fans_limit, tab_id, animate,
			logo_selected, gray_token, dynamic_conf_url, op_title_color, op_background_color, op_link_container_color, tus_value
		FROM
			sidebar
		WHERE
			state=1 AND online_time<=?
			and (offline_time = '0000-00-00 00:00:00' or offline_time >= ?)
		ORDER BY rank DESC
	`

	_sidebarLimitSQL = `
		SELECT
			s_id, build, conditions
		FROM
			sidebar_limit
	`

	_languageSQL = `
		SELECT
			id, name
		FROM
			language
	`

	_moduleSQL = `SELECT
			id, plat, title, style, button_name,
			button_url,button_icon, button_style, white_url, title_color,
			subtitle, subtitle_url, subtitle_color, background, background_color,
			mtype, audit_show, is_mng, op_style_type, op_load_condition
		FROM
			sidebar_module
		WHERE 
			state = 1 AND mtype = ?
		ORDER BY rank DESC,id DESC
	`
)

func (d *sectionDao) SideBar(ctx context.Context, now time.Time) ([]*secmdl.SideBar, error) {
	rows, err := d.db.Query(ctx, _sidebarSQL, now, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sidebars []*secmdl.SideBar
	for rows.Next() {
		s := &secmdl.SideBar{}
		redDotForNew := 0
		if err := rows.Scan(
			&s.ID, &s.Plat, &s.Module, &s.Name, &s.Logo,
			&s.LogoWhite, &s.Param, &s.Rank,
			&s.Tip, &s.NeedLogin, &s.WhiteURL, &s.LogoSelected, &s.TabID,
			&s.Red, &s.LanguageID, &s.GlobalRed, &s.RedLimit, &s.Animate,
			&s.WhiteURLShow, &s.Area, &s.ShowPurposed, &s.AreaPolicy, &redDotForNew,
			&s.OpLoadCondition, &s.OpTitle, &s.OpSubTitle, &s.OpTitleIcon, &s.OpLinkType,
			&s.OpLinkText, &s.OpLinkIcon, &s.OpFansLimit, &s.TabID, &s.Animate,
			&s.LogoSelected, &s.GrayToken, &s.DynamicConfUrl, &s.OpTitleColor, &s.OpBackgroundColor, &s.OpLinkContainerColor, &s.TusValue); err != nil {
			return nil, err
		}
		if redDotForNew == 1 {
			s.RedDotForNew = true
		}
		sidebars = append(sidebars, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sidebars, nil
}

func (d *sectionDao) SidebarLimit(ctx context.Context) (map[int64][]*secmdl.SideBarLimit, error) {
	rows, err := d.db.Query(ctx, _sidebarLimitSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sidebarLimits := make(map[int64][]*secmdl.SideBarLimit)
	for rows.Next() {
		sl := &secmdl.SideBarLimit{}
		if err := rows.Scan(&sl.SideBarID, &sl.Build, &sl.Condition); err != nil {
			return nil, err
		}
		sidebarLimits[sl.SideBarID] = append(sidebarLimits[sl.SideBarID], sl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sidebarLimits, nil
}

func (d *sectionDao) SidebarLang(ctx context.Context) (map[int64]string, error) {
	rows, err := d.db.Query(ctx, _languageSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	languages := make(map[int64]string)
	for rows.Next() {
		l := struct {
			ID   int64
			Name string
		}{}
		if err := rows.Scan(&l.ID, &l.Name); err != nil {
			return nil, err
		}
		languages[l.ID] = l.Name
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return languages, nil
}

var ALL_PLATS = []int32{
	int32(0), int32(1), int32(2), int32(5), int32(8), int32(9), int32(10), int32(20),
}

// SideBarModules get modules.
func (d *sectionDao) SideBarModules(ctx context.Context, moduleType int32) (sm map[int32][]*secmdl.ModuleInfo, err error) {
	rows, err := d.db.Query(ctx, _moduleSQL, moduleType)
	if err != nil {
		log.Error("d.db.Query error(%v)", err)
		return
	}
	sm = make(map[int32][]*secmdl.ModuleInfo)
	defer rows.Close()
	for rows.Next() {
		m := &secmdl.ModuleInfo{}
		if err = rows.Scan(
			&m.ID, &m.Plat, &m.Title, &m.Style, &m.ButtonName,
			&m.ButtonURL, &m.ButtonIcon, &m.ButtonStyle, &m.WhiteURL, &m.TitleColor,
			&m.Subtitle, &m.SubtitleURL, &m.SubtitleColor, &m.Background, &m.BackgroundColor,
			&m.MType, &m.AuditShow, &m.IsMng, &m.OpStyleType, &m.OpLoadCondition); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		// 对于8、9、10跨平台的module，目前使用枚举来做全plat设置
		if m.ID == 8 || m.ID == 9 || m.ID == 10 {
			for _, plat := range ALL_PLATS {
				tempPlat := plat
				tempM := &secmdl.ModuleInfo{
					ID:              m.ID,
					Plat:            plat,
					Title:           m.Title,
					Style:           m.Style,
					ButtonName:      m.ButtonName,
					ButtonURL:       m.ButtonURL,
					ButtonIcon:      m.ButtonIcon,
					ButtonStyle:     m.ButtonStyle,
					WhiteURL:        m.WhiteURL,
					TitleColor:      m.TitleColor,
					Subtitle:        m.Subtitle,
					SubtitleURL:     m.SubtitleURL,
					SubtitleColor:   m.SubtitleColor,
					Background:      m.Background,
					BackgroundColor: m.BackgroundColor,
					MType:           m.MType,
					AuditShow:       m.AuditShow,
					IsMng:           m.IsMng,
					OpStyleType:     m.OpStyleType,
					OpLoadCondition: m.OpLoadCondition,
				}
				sm[tempPlat] = append(sm[tempPlat], tempM)
			}
			continue
		}
		sm[m.Plat] = append(sm[m.Plat], m)
	}
	err = rows.Err()
	return
}

const (
	_iconsSQL = `
		SELECT 
		id,module,icon,global_red_dot,effect_group,effect_url,stime,etime FROM mng_icon 
		WHERE stime<=? AND etime>=? AND state=1 
		ORDER BY id DESC
	`
)

func (d *sectionDao) Icons(c context.Context, startTime, endTime time.Time) (map[int64]*api.MngIcon, error) {
	rows, err := d.db.Query(c, _iconsSQL, endTime, startTime)
	if err != nil {
		return nil, err
	}
	icons := make(map[int64]*api.MngIcon)
	defer rows.Close()
	for rows.Next() {
		var (
			module string
			ic     = &api.MngIcon{}
		)
		if err = rows.Scan(&ic.Id, &module, &ic.Icon, &ic.GlobalRed, &ic.EffectGroup, &ic.EffectUrl, &ic.Stime, &ic.Etime); err != nil {
			log.Error("rows.Scan err(%+v)", err)
			err = nil
			continue
		}
		if err = json.Unmarshal([]byte(module), &ic.Module); err != nil {
			log.Error("json.Unmarshal err(%+v) module(%s)", err, module)
			err = nil
			continue
		}
		// 每个模块只会展示一个icon，有多个生效时优先取后配置的（按ID倒序
		for _, v := range ic.Module {
			if _, ok := icons[v.Oid]; ok {
				continue
			}
			icons[v.Oid] = ic
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return icons, nil
}

const (
	_hiddensSQL = `SELECT
	id,sid,rid,cid,module_id,channel,pid,stime,etime,hidden_condition,hide_dynamic
	FROM entrance_hidden
	WHERE stime<=? AND etime>=? AND state=1`
	_hiddenLimitsSQL = `SELECT
	oid,build,conditions,plat
	FROM  entrance_hidden_limit
	WHERE state = 1
	`
)

// Hiddens is
func (d *sectionDao) Hiddens(c context.Context, now time.Time) ([]*api.Hidden, error) {
	rows, err := d.db.Query(c, _hiddensSQL, now, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hiddens []*api.Hidden
	for rows.Next() {
		h := &api.Hidden{}
		if err = rows.Scan(&h.Id, &h.Sid, &h.Rid, &h.Cid, &h.ModuleId, &h.Channel, &h.Pid, &h.Stime, &h.Etime, &h.HiddenCondition, &h.HideDynamic); err != nil {
			return nil, err
		}
		channelArr := strings.Split(h.Channel, ",")
		channelMap := make(map[string]string, len(channelArr))
		var channelFuzzy []string
		for _, v := range channelArr {
			if strings.Contains(v, "%") { //如果有%则要单独处理包含逻辑
				channelFuzzy = append(channelFuzzy, v)
				continue
			}
			channelMap[v] = v
		}
		h.ChannelMap = channelMap
		h.ChannelFuzzy = channelFuzzy
		hiddens = append(hiddens, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hiddens, nil
}

func (d *sectionDao) HiddenLimits(c context.Context) (map[int64][]*api.HiddenLimit, error) {
	rows, err := d.db.Query(c, _hiddenLimitsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hiddensLimits := make(map[int64][]*api.HiddenLimit)
	for rows.Next() {
		hl := &api.HiddenLimit{}
		if err := rows.Scan(&hl.Oid, &hl.Build, &hl.Conditions, &hl.Plat); err != nil {
			return nil, err
		}
		hiddensLimits[hl.Oid] = append(hiddensLimits[hl.Oid], hl)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hiddensLimits, nil
}
