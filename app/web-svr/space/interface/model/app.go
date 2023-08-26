package model

import (
	api "git.bilibili.co/bapis/bapis-go/account/service"
)

// AppAccInfo app acc info struct.
type AppAccInfo struct {
	Mid       int64           `json:"mid"`
	Name      string          `json:"name"`
	Sex       string          `json:"sex"`
	Face      string          `json:"face"`
	Sign      string          `json:"sign"`
	Rank      int32           `json:"rank"`
	Level     int32           `json:"level"`
	LevelInfo api.LevelInfo   `json:"level_info"`
	Pendant   api.PendantInfo `json:"pendant"`
	Silence   int32           `json:"silence"`
	Vip       struct {
		api.VipInfo
		// TODO 以后可删除
		VipType   int32 `json:"vipType"`
		VipStatus int32 `json:"vipStatus"`
	} `json:"vip"`
	OfficialInfo struct {
		Type int    `json:"type"`
		Desc string `json:"desc"`
	} `json:"official_info"`
	Following  int64       `json:"following"`
	Follower   int64       `json:"follower"`
	Relation   interface{} `json:"relation"`
	BeRelation interface{} `json:"be_relation"`
	FansGroup  int         `json:"fans_group"`
	Audio      int         `json:"audio"`
	Shop       interface{} `json:"shop"`
	Elec       interface{} `json:"elec"`
	Live       interface{} `json:"live"`
	FansBadge  bool        `json:"fans_badge"`
	TopPhoto   string      `json:"top_photo"`
	Block      *AccBlock   `json:"block,omitempty"`
}

// FromProfile from account profile.
func (ai *AppAccInfo) FromProfile(p *api.ProfileStatReply) {
	ai.Mid = p.Profile.Mid
	ai.Name = p.Profile.Name
	ai.Sex = p.Profile.Sex
	ai.Face = p.Profile.Face
	ai.Sign = p.Profile.Sign
	ai.Rank = p.Profile.Rank
	ai.Face = p.Profile.Face
	ai.Level = p.Profile.Level
	ai.Silence = p.Profile.Silence
	ai.Vip.VipType = p.Profile.Vip.Type
	ai.Vip.VipStatus = p.Profile.Vip.Status
	ai.Vip.VipInfo = p.Profile.Vip
	ai.Pendant = p.Profile.Pendant
	if p.Profile.Official.Role == 0 {
		ai.OfficialInfo.Type = -1
	} else {
		if p.Profile.Official.Role <= 2 || p.Profile.Official.Role == 7 {
			ai.OfficialInfo.Type = 0
			ai.OfficialInfo.Desc = p.Profile.Official.Title
		} else {
			ai.OfficialInfo.Type = 1
			ai.OfficialInfo.Desc = p.Profile.Official.Title
		}
	}
	ai.Following = p.Following
	ai.Follower = p.Follower
}

// AppTab tab if show.
type AppTab struct {
	Dynamic  bool `json:"dynamic"`
	Shop     bool `json:"shop"`
	Archive  bool `json:"video"`
	Article  bool `json:"article"`
	Audio    bool `json:"audio"`
	Album    bool `json:"album"`
	Favorite bool `json:"favorite"`
	Bangumi  bool `json:"bangumi"`
	Game     bool `json:"game"`
}

// AppIndex app index data.
type AppIndex struct {
	Info    *AppAccInfo `json:"info"`
	Tab     *AppTab     `json:"tab"`
	Dynamic *DyTotal    `json:"dynamic"`
	Archive *UpArc      `json:"archive"`
}

// AppIndexArg .
type AppIndexArg struct {
	Mid      int64
	Vmid     int64  `form:"mid" validate:"min=1"`
	Qn       int    `form:"qn" default:"16" validate:"min=1"`
	Platform string `form:"platform" default:"android"`
	Ps       int32  `form:"ps" default:"16" validate:"min=1"`
	Device   string `form:"device"`
}
