package mine

import (
	"go-gateway/app/app-svr/app-car/interface/model"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

const (
	_defaultFace     = "http://static.hdslb.com/images/member/noface.gif"
	_vipStatusNormal = 1
)

type MineParam struct {
	model.DeviceInfo
	Env string `form:"env"`
}

type Mine struct {
	Mid           int64  `json:"mid"`
	Name          string `json:"name"`
	ShowNameGuide bool   `json:"show_name_guide"`
	Face          string `json:"face"`
	ShowFaceGuide bool   `json:"show_face_guide"`
	Sex           int32  `json:"sex"`
	Rank          int32  `json:"rank"`
	Silence       int32  `json:"silence"`
	Level         int32  `json:"level"`
	VipType       int32  `json:"vip_type"`
	Official      struct {
		Type int8   `json:"type"`
		Desc string `json:"desc"`
	} `json:"official_verify"`
	Vip  accountgrpc.VipInfo `json:"vip,omitempty"`
	Fans string              `json:"fans,omitempty"`
}

func (m *Mine) FromMine(p *accountgrpc.Profile) {
	m.Silence = p.Silence
	m.Mid = p.Mid
	m.Name = p.Name
	m.Face = p.Face
	if m.Face == "" {
		m.Face = _defaultFace
	}
	switch p.Sex {
	case "男":
		m.Sex = 1
	case "女":
		m.Sex = 2
	default:
		m.Sex = 0
	}
	m.Rank = p.Rank
	m.Level = p.Level
	m.Vip = p.Vip
	if p.Vip.Status == _vipStatusNormal { //1-正常
		m.VipType = p.Vip.Type
	}
	if p.Official.Role == 0 {
		m.Official.Type = -1
	} else {
		if p.Official.Role <= 2 || p.Official.Role == 7 {
			m.Official.Type = 0
		} else {
			m.Official.Type = 1
		}
		m.Official.Desc = p.Official.Title
	}
}
