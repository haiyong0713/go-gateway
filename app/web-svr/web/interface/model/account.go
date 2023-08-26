package model

import (
	"strconv"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	pangugsgrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
)

// Space space top photo
type Space struct {
	SImg string `json:"s_img"`
	LImg string `json:"l_img"`
}

// Card  Card  and Space and Relation and Archive Count.
type Card struct {
	Card         *AccountCard `json:"card"`
	Space        *Space       `json:"space,omitempty"`
	Following    bool         `json:"following"`
	ArchiveCount int64        `json:"archive_count"`
	ArticleCount int32        `json:"article_count"`
	Follower     int64        `json:"follower"`
	LikeNum      int64        `json:"like_num"`
}

// AccountCard struct.
type AccountCard struct {
	Mid         string                    `json:"mid"`
	Name        string                    `json:"name"`
	Approve     bool                      `json:"approve"`
	Sex         string                    `json:"sex"`
	Rank        string                    `json:"rank"`
	Face        string                    `json:"face"`
	FaceNft     int32                     `json:"face_nft"`
	FaceNftType pangugsgrpc.NFTRegionType `json:"face_nft_type"`
	DisplayRank string                    `json:"DisplayRank"`
	Regtime     int64                     `json:"regtime"`
	Spacesta    int32                     `json:"spacesta"`
	Birthday    string                    `json:"birthday"`
	Place       string                    `json:"place"`
	Description string                    `json:"description"`
	Article     int                       `json:"article"`
	Attentions  []int64                   `json:"attentions"`
	Fans        int64                     `json:"fans"`
	Friend      int64                     `json:"friend"`
	Attention   int64                     `json:"attention"`
	Sign        string                    `json:"sign"`
	LevelInfo   struct {
		Cur     int32 `json:"current_level"`
		Min     int   `json:"current_min"`
		NowExp  int   `json:"current_exp"`
		NextExp int   `json:"next_exp"`
	} `json:"level_info"`
	Pendant        accmdl.PendantInfo   `json:"pendant"`
	Nameplate      accmdl.NameplateInfo `json:"nameplate"`
	Official       accmdl.OfficialInfo
	OfficialVerify OfficialVerify `json:"official_verify"`
	Vip            VipInfo        `json:"vip"`
	IsSeniorMember int32          `json:"is_senior_member"`
}

type VipInfo struct {
	accmdl.VipInfo
	// TODO 以后可删除
	VipType   int32 `json:"vipType"`
	VipStatus int32 `json:"vipStatus"`
}

// OfficialVerify old official verify
type OfficialVerify struct {
	Type int32  `json:"type"`
	Desc string `json:"desc"`
}

// Relation .
type Relation struct {
	Relation   interface{} `json:"relation"`
	BeRelation interface{} `json:"be_relation"`
}

// FromCard from account catd.
func (ac *AccountCard) FromCard(c *accmdl.Card) {
	ac.Mid = strconv.FormatInt(c.Mid, 10)
	ac.Name = c.Name
	// ac.Approve =
	ac.Sex = c.Sex
	ac.Rank = strconv.FormatInt(int64(c.Rank), 10)
	ac.DisplayRank = "0"
	ac.Face = c.Face
	ac.FaceNft = c.FaceNftNew
	// ac.Regtime =
	if c.Silence == 1 {
		ac.Spacesta = -2
	}
	// ac.Birthday =
	// ac.Place =
	// ac.Description =
	// ac.Article =
	// ac.Attentions = []int64{}
	// ac.Fans =
	// ac.Friend
	// ac.Attention =
	ac.Sign = c.Sign
	ac.LevelInfo.Cur = c.Level
	ac.LevelInfo.NextExp = 0
	// ac.LevelInfo.Min =
	ac.Pendant = c.Pendant
	ac.Nameplate = c.Nameplate
	ac.OfficialVerify = FromOfficial(c.Official)
	ac.Official = c.Official
	ac.Vip.VipType = c.Vip.Type
	ac.Vip.VipStatus = c.Vip.Status
	ac.Vip.VipInfo = c.Vip
	ac.IsSeniorMember = c.IsSeniorMember
}

// FromOfficial from official to official verify.
func FromOfficial(info accmdl.OfficialInfo) (d OfficialVerify) {
	if info.Role == 0 {
		d.Type = -1
	} else {
		if info.Role <= 2 || info.Role == 7 {
			d.Type = 0
			d.Desc = info.Title
		} else {
			d.Type = 1
			d.Desc = info.Title
		}
	}
	return
}

// DefaultProfile .
var DefaultProfile = &accmdl.ProfileStatReply{
	Profile: &accmdl.Profile{
		Sex:  "保密",
		Rank: 10000,
		Face: "https://static.hdslb.com/images/member/noface.gif",
		Sign: "没签名",
	},
	LevelInfo: accmdl.LevelInfo{},
}
