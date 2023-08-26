package model

import (
	"strconv"

	"go-common/library/time"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
)

const (
	// Meta.Type
	ArticleTypeNote = 2 //笔记
)

// Info struct.
type Info struct {
	Mid         string `json:"mid"`
	Name        string `json:"uname"`
	Sex         string `json:"sex"`
	Sign        string `json:"sign"`
	Avatar      string `json:"avatar"`
	Rank        string `json:"rank"`
	DisplayRank string `json:"DisplayRank"`
	LevelInfo   struct {
		Cur     int32       `json:"current_level"`
		Min     int         `json:"current_min"`
		NowExp  int         `json:"current_exp"`
		NextExp interface{} `json:"next_exp"`
	} `json:"level_info"`
	Pendant        accmdl.PendantInfo   `json:"pendant"`
	Nameplate      accmdl.NameplateInfo `json:"nameplate"`
	Official       accmdl.OfficialInfo  `json:"official"`
	OfficialVerify OfficialVerify       `json:"official_verify"`
	Vip            struct {
		Type          int32  `json:"vipType"`
		DueDate       int64  `json:"vipDueDate"`
		DueRemark     string `json:"dueRemark"`
		AccessStatus  int    `json:"accessStatus"`
		VipStatus     int32  `json:"vipStatus"`
		VipStatusWarn string `json:"vipStatusWarn"`
	} `json:"vip"`
	// article
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	PublishTime time.Time `json:"publish_time"`
	Following   bool      `json:"following"`
}

// FromCard from card.
func (i *Info) FromCard(c *accmdl.Card) {
	i.Mid = strconv.FormatInt(c.Mid, 10)
	i.Name = c.Name
	i.Sex = c.Sex
	i.Sign = c.Sign
	i.Avatar = c.Face
	i.Rank = strconv.FormatInt(int64(c.Rank), 10)
	i.DisplayRank = "0"
	i.LevelInfo.Cur = c.Level
	i.LevelInfo.NextExp = 0
	// i.LevelInfo.Min =
	i.Pendant = c.Pendant
	i.Nameplate = c.Nameplate
	i.Official = c.Official
	i.OfficialVerify = FromOfficial(c.Official)
	i.Vip.Type = c.Vip.Type
	i.Vip.VipStatus = c.Vip.Status
	i.Vip.DueDate = c.Vip.DueDate
}

// Meta struct.
type Meta struct {
	*artmdl.Meta
	Like int32 `json:"like"`
}

// ArticleUpInfo struct.
type ArticleUpInfo struct {
	ArtCount    int   `json:"art_count"`
	Follower    int64 `json:"follower"`
	IsFollowing bool  `json:"is_following"`
}
