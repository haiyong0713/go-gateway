package show

import (
	"context"
	"time"

	"go-common/library/log"

	"go-gateway/app/app-svr/resource/service/model"
)

const (
	_selSideSQL = `SELECT
			s.id, s.plat, s.module, s.name, s.logo,
 			s.logo_white, s.param, s.rank, l.build, l.conditions,
			s.tip, s.need_login, s.white_url, s.logo_selected, s.tab_id,
			s.red_dot_url, lang.name, s.global_red_dot, s.red_dot_limit, s.animate,
			s.white_url_show, s.area, s.show_purposed,s.area_policy,s.red_dot_for_new,
			s.op_load_condition, s.op_title, s.op_sub_title, s.op_title_icon, s.op_link_type, 
			s.op_link_text, s.op_link_icon, s.op_fans_limit, s.tab_id, s.animate,
			s.logo_selected, s.gray_token, s.dynamic_conf_url, s.op_title_color, s.op_background_color, s.op_link_container_color
		FROM
			sidebar AS s,sidebar_limit AS l,language AS lang
		WHERE
			s.state=1 AND s.id=l.s_id AND lang.id=s.lang_id AND s.online_time<=?
			and (s.offline_time = '0000-00-00 00:00:00' or s.offline_time >= ?)
		ORDER BY s.rank DESC,l.id ASC
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

// SideBar get side bar.
func (d *Dao) SideBar(ctx context.Context, now time.Time) (ss []*model.SideBar, limits map[int64][]*model.SideBarLimit, err error) {
	rows, err := d.db.Query(ctx, _selSideSQL, now, now)
	if err != nil {
		log.Error("d.db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	limits = make(map[int64][]*model.SideBarLimit)
	for rows.Next() {
		s := &model.SideBar{}
		redDotForNew := 0
		if err = rows.Scan(
			&s.ID, &s.Plat, &s.Module, &s.Name, &s.Logo,
			&s.LogoWhite, &s.Param, &s.Rank, &s.Build, &s.Conditions,
			&s.Tip, &s.NeedLogin, &s.WhiteURL, &s.LogoSelected, &s.TabID,
			&s.Red, &s.Language, &s.GlobalRed, &s.RedLimit, &s.Animate,
			&s.WhiteURLShow, &s.Area, &s.ShowPurposed, &s.AreaPolicy, &redDotForNew,
			&s.OpLoadCondition, &s.OpTitle, &s.OpSubTitle, &s.OpTitleIcon, &s.OpLinkType,
			&s.OpLinkText, &s.OpLinkIcon, &s.OpFansLimit, &s.TabID, &s.Animate,
			&s.LogoSelected, &s.GrayToken, &s.DynamicConfUrl, &s.OpTitleColor, &s.OpBackgroundColor, &s.OpLinkContainerColor); err != nil {
			log.Error("row.Scan error(%v)", err)
			return
		}
		if redDotForNew == 1 {
			s.RedDotForNew = true
		}
		if _, ok := limits[s.ID]; !ok {
			ss = append(ss, s)
		}
		limit := &model.SideBarLimit{
			ID:        s.ID,
			Build:     s.Build,
			Condition: s.Conditions,
		}
		limits[s.ID] = append(limits[s.ID], limit)
	}
	err = rows.Err()
	return
}

var ALL_PLATS = []int32{
	int32(0), int32(1), int32(2), int32(5), int32(8), int32(9), int32(10), int32(20),
}

// SideBarModules get modules.
func (d *Dao) SideBarModules(ctx context.Context, moduleType int32) (sm map[int32][]*model.ModuleInfo, err error) {
	rows, err := d.db.Query(ctx, _moduleSQL, moduleType)
	if err != nil {
		log.Error("d.db.Query error(%v)", err)
		return
	}
	sm = make(map[int32][]*model.ModuleInfo)
	defer rows.Close()
	for rows.Next() {
		m := &model.ModuleInfo{}
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
				tempM := &model.ModuleInfo{
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
