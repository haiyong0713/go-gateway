package bwsonline

import "go-common/library/time"

const (
	CurrTypeEnergy    = 1
	CurrTypeCoin      = 2
	CurrAddTypeNormal = 0
	CurrAddTypeAuto   = 1
	DressHasEquip     = 1
	HadReward         = 1
	UsedTimeTypeAd    = 1
	UsedTimeTypeShare = 2
	UsedTimeTypeLed   = 3
	DressPosHead      = 1
	DressPosBody      = 2
	DressPosBottom    = 3
	UserPrintHad      = 1
	AwardPackageOwned = 1
	PackageTypeAward  = 0
	PackageTypePrint  = 1
	AwardTypeVip      = 1
	AwardTypeDress    = 2
	AwardTypeOffline  = 3
	AwardTypeCurrency = 4
	AwardTypeBBQ      = 5
	AwardTypeHeart    = 6
	AwardTypeSuit     = 7
	ReserveIsChecked  = 1
)

var PrintValue = map[int32]int64{
	1: 15,
	2: 35,
	3: 60,
}

type Main struct {
	Mid         int64        `json:"mid"`
	Name        string       `json:"name"`
	Face        string       `json:"face"`
	Energy      int64        `json:"energy"`
	Currency    int64        `json:"currency"`
	IsActivated int          `json:"is_activated"`
	Piece       []*UserPiece `json:"piece"`
	Dress       []*Dress     `json:"dress"`
}

type UserPiece struct {
	Pid   int64 `json:"pid"`
	Num   int64 `json:"num"`
	Level int64 `json:"level"`
}

type AwardPackageItem struct {
	*Award
	Owned int64 `json:"owned"`
}

type AwardPackageDetail struct {
	*AwardPackage
	Items   []*AwardPackageItem `json:"items"`
	Owned   int64               `json:"owned"`
	Total   int64               `json:"total"`
	Awarded int64               `json:"awarded"`
}

type UserAward struct {
	*Award
	State int64 `json:"state"`
}

type UserPrint struct {
	*Print
	PieceState int          `json:"piece_state"`
	Unlocked   int          `json:"unlocked"`
	UnlockCost []*UserPiece `json:"unlock_cost"`
}

type UserPrintDetail struct {
	*UserPrint
	Awards []*Award `json:"awards"`
}

type TicketBindRecord struct {
	Id             int64     `json:"id"`
	UserName       string    `json:"user_name"`
	Mid            int64     `json:"mid"`
	PersonalId     string    `json:"personal_id"`
	PersonalIdType int       `json:"personal_id_type"`
	PersonalIdSum  string    `json:"personal_id_sum"`
	State          int       `json:"state"`
	Ctime          time.Time `json:"ctime"`
	Mtime          time.Time `json:"mtime"`
}

type InterReserveOrder struct {
	Id             int64     `json:"order_id"`
	Mid            int64     `json:"mid"`
	TicketNo       string    `json:"ticket_no"`
	InterReserveId int64     `json:"inter_reserve_id"`
	OrderNo        string    `json:"order_no"`
	IsChecked      int       `json:"is_checked"`
	ReserveNo      int       `json:"reserve_no"`
	Ctime          time.Time `json:"ctime"`
	Mtime          time.Time `json:"mtime"`
}

type TicketInfo struct {
	Tel        string `json:"tel,omitempty"`
	Sid        int64  `json:"sid"`         //票种id
	SkuName    string `json:"sku_name"`    //票种名称
	ScreenName string `json:"screen_name"` //场次名称
	Type       int    `json:"type"`        //1: 正常购票，2:邀请函
	Ticket     string `json:"ticket"`      //票号
}

type TicketInfoFromHYG struct {
	Name       string        `json:"name"`
	List       []*TicketInfo `json:"list"`
	UpdateTime int64         `json:"update_time"`
}

//go:generate kratos t protoc --grpc model.proto
