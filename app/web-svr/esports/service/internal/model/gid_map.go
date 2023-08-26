package model

const (
	/** 枚举 **/
	/** ** oid类型 ** **/
	// 赛程
	OidContestType = int64(3)

	/** ** 软删状态 ** **/
	GidMapRecordDeleted    = 1
	GidMapRecordNotDeleted = 0
)

type GidMapModel struct {
	ID int64 `json:"id"`
	// 游戏id
	Gid int64 `json:"gid"`
	// 对象id
	Oid int64 `json:"oid"`
	// 对象类型
	Type int64 `json:"type"`

	IsDeleted int64 `json:"is_deleted"`
}
