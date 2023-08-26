package mine

import accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"

type MineWeb struct {
	Mid     int64  `json:"mid"`
	Name    string `json:"name"`
	Face    string `json:"face"`
	VipType int32  `json:"vip_type"`
}

func (m *MineWeb) FromMineWeb(p *accountgrpc.Profile) {
	m.Mid = p.Mid
	m.Name = p.Name
	m.Face = p.Face
	if m.Face == "" {
		m.Face = _defaultFace
	}
	if p.Vip.Status == _vipStatusNormal { //1-正常
		m.VipType = p.Vip.Type
	}
}
