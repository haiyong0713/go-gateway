package search

import (
	"encoding/json"
	"strconv"

	arcmdl "go-gateway/app/app-svr/archive/service/api"

	pangugsgrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
)

// SearchUserRes .
type SearchUserRes struct {
	Code           int             `json:"code,omitempty"`
	SeID           string          `json:"seid,omitempty"`
	Page           int             `json:"page,omitempty"`
	PageSize       int             `json:"pagesize,omitempty"`
	Total          int             `json:"total,omitempty"`
	NumResults     int             `json:"numResults"`
	NumPages       int             `json:"numPages"`
	SuggestKeyword string          `json:"suggest_keyword"`
	RqtType        string          `json:"rqt_type,omitempty"`
	CostTime       json.RawMessage `json:"cost_time,omitempty"`
	ExpList        json.RawMessage `json:"exp_list,omitempty"`
	EggHit         int             `json:"egg_hit"`
	PageInfo       json.RawMessage `json:"pageinfo,omitempty"`
	Result         []*SearchUser   `json:"result,omitempty"`
	ShowColumn     int             `json:"show_column"`
	InBlackKey     int8            `json:"in_black_key"`
	InWhiteKey     int8            `json:"in_white_key"`
}

type SearchUser struct {
	Type           string                    `json:"type"`
	Mid            int64                     `json:"mid"`
	Uname          string                    `json:"uname"`
	Usign          string                    `json:"usign"`
	Fans           int64                     `json:"fans"`
	Videos         int                       `json:"videos"`
	Upic           string                    `json:"upic"`
	FaceNft        int32                     `json:"face_nft"` // face_nft,从账号服务获取
	FaceNftType    pangugsgrpc.NFTRegionType `json:"face_nft_type"`
	VerifyInfo     string                    `json:"verify_info"`
	Level          int                       `json:"level"`
	Gender         int                       `json:"gender"`
	IsUpuser       int                       `json:"is_upuser"`
	IsLive         int                       `json:"is_live"`
	RoomID         int64                     `json:"room_id"`
	Res            []*UserVideo              `json:"res"`
	OfficialVerify OfficialVerify            `json:"official_verify"`
	HitColumns     []string                  `json:"hit_columns"`
	IsSeniorMember int32                     `json:"is_senior_member"`
}

type OfficialVerify struct {
	Type int32  `json:"type"`
	Desc string `json:"desc"`
}

type UserVideo struct {
	Aid          int64  `json:"aid"`
	Bvid         string `json:"bvid"`
	Title        string `json:"title"`
	Pubdate      int64  `json:"pubdate"`
	Arcurl       string `json:"arcurl"`
	Pic          string `json:"pic"`
	Play         string `json:"play"`
	Dm           int64  `json:"dm"`
	Coin         int64  `json:"coin"`
	Fav          int64  `json:"fav"`
	Desc         string `json:"desc"`
	Duration     string `json:"duration"`
	IsPay        int    `json:"is_pay"`
	IsUnionVideo int    `json:"is_union_video"`
}

// Fill fill search user data.
func (v *UserVideo) Fill(arc *arcmdl.Arc) {
	if arc == nil {
		return
	}
	v.Play = strconv.FormatInt(int64(arc.Stat.View), 10)
	v.Pubdate = int64(arc.PubDate)
}
