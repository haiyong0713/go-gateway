package bnj

import (
	"go-gateway/app/app-svr/archive/service/api"
)

const (
	ActionTypeIncr = 1
	ActionTypeDecr = 2
	DecrLevelOne   = 1
	DecrLevelTwo   = 2
	DecrLevelThree = 3
	DecrLevelFour  = 4
	DecrLevelFive  = 5
)

var DecreaseMsgTypes = map[int]string{
	DecrLevelOne:   "喝了口汤",
	DecrLevelTwo:   "偷吃了%s",
	DecrLevelThree: "偷吃了一小口",
	DecrLevelFour:  "偷吃了一大口",
	DecrLevelFive:  "差点把锅端走",
}

type MemBnj20 struct {
	GameFinish  int64
	AppointCnt  int64
	HotpotValue int64
	HotpotLevel int64
	Arcs        map[int64]*api.Arc
}

type MainBnj20 struct {
	Sid              int64        `json:"sid"`
	ReservedCount    int64        `json:"reserved_count"`
	Value            int64        `json:"value"`
	HotPotLevel      int64        `json:"hot_pot_level"`
	HasReserved      int32        `json:"has_reserved"`
	Infos            []*InfoBnj20 `json:"infos"`
	TimelinePic      string       `json:"timeline_pic"`
	H5TimelinePic    string       `json:"h5_timeline_pic"`
	ShareTimelinePic string       `json:"share_timeline_pic"`
	BlockGame        int          `json:"block_game,omitempty"`
	BlockGameAction  int          `json:"block_game_action,omitempty"`
	Award            *AwardBnj20  `json:"award"`
}

type InfoBnj20 struct {
	Name         string      `json:"name"`
	Pic          []string    `json:"pic"`
	H5Pic        []string    `json:"h5_pic"`
	Detail       string      `json:"detail"`
	H5Detail     string      `json:"h5_detail"`
	SharePic     string      `json:"share_pic"`
	H5SharePic   string      `json:"h5_share_pic"`
	DynamicPic   string      `json:"dynamic_pic"`
	H5DynamicPic string      `json:"h5_dynamic_pic"`
	Arcs         []*ArcBnj20 `json:"arcs"`
}

type ArcBnj20 struct {
	Aid        int64        `json:"aid"`
	Title      string       `json:"title"`
	Pic        string       `json:"pic"`
	Owner      api.Author   `json:"owner"`
	Stat       ArcStatBnj20 `json:"stat"`
	RcmdReason string       `json:"rcmd_reason"`
}

type ArcStatBnj20 struct {
	View int32 `json:"view"`
}

type AwardBnj20 struct {
	HasMore       int               `json:"has_more"`
	FinalHasAward int               `json:"final_has_award"`
	List          []*AwardItemBnj20 `json:"list"`
}

type AwardItemBnj20 struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Pic         string `json:"pic"`
	CardPic     string `json:"card_pic"`
	Type        int    `json:"type"`
	Count       int64  `json:"count"`
	LinkText    string `json:"link_text"`
	LinkURL     string `json:"link_url"`
	HasUnlocked int    `json:"has_unlocked"`
	HasReward   int    `json:"has_reward"`
}

type MaterialRes struct {
	Blocked     int           `json:"blocked,omitempty"`
	NormalList  []string      `json:"normal_list"`
	SpecialList *MaterialSpec `json:"special_list"`
	RareHotDot  int           `json:"rare_hot_dot"`
	RareList    []*Material   `json:"rare_list"`
}

type MaterialSpec struct {
	Good     []string `json:"good"`
	GoodDesc []string `json:"good_desc"`
	Bad      []string `json:"bad"`
	BadDesc  []string `json:"bad_desc"`
}

type Material struct {
	ID          int64  `json:"id"`
	Pic         string `json:"pic"`
	H5Pic       string `json:"h5_pic"`
	Name        string `json:"name"`
	Desc        string `json:"desc"`
	SharePic    string `json:"share_pic"`
	CardPic     string `json:"card_pic"`
	HasUnlocked int    `json:"has_unlocked"`
	HasReceived int    `json:"has_received"`
	TaskID      int64  `json:"-"`
}

type IncreaseMaterial struct {
	ID       int64  `json:"id"`
	Pic      string `json:"pic"`
	H5Pic    string `json:"h5_pic"`
	SharePic string `json:"share_pic"`
	CardPic  string `json:"card_pic"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
}

type Action struct {
	Mid     int64  `json:"mid"`
	Type    int    `json:"type"`
	Num     int64  `json:"num"`
	Message string `json:"message"`
	Ts      int64  `json:"ts"`
}

type AwardAction struct {
	Mid          int64  `json:"mid"`
	ID           int64  `json:"id"`
	Type         int    `json:"type"`
	SourceID     string `json:"source_id"`
	SourceExpire int64  `json:"source_expire"`
	TaskID       int64  `json:"task_id"`
	Mirror       string `json:"mirror"`
}
